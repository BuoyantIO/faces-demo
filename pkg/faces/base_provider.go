// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2025 Buoyant Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.  You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package faces

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	"github.com/BuoyantIO/faces-demo/v2/pkg/whisper"
	"github.com/prometheus/client_golang/prometheus"
)

// Glowing stuff
type GlowMsg struct {
	Node    int
	Process int
	OK      bool
	Value   int
}

const (
	CmdActivity = 1
)

type BaseRequestStatus struct {
	errored     bool
	ratelimited bool
	latched     bool
	message     string
	statusCode  int
	delayMs     int
}

func (rstat *BaseRequestStatus) IsErrored() bool {
	return rstat.errored
}

func (rstat *BaseRequestStatus) IsRateLimited() bool {
	return rstat.ratelimited
}

func (rstat *BaseRequestStatus) IsLatched() bool {
	return rstat.latched
}

func (rstat *BaseRequestStatus) Message() string {
	return rstat.message
}

func (rstat *BaseRequestStatus) StatusCode() int {
	return rstat.statusCode
}

func (rstat *BaseRequestStatus) DelayMs() int {
	return rstat.delayMs
}

type ProviderRequest struct {
	subrequest string
	user       string
	userAgent  string
	row        int
	col        int
}

func (prvReq *ProviderRequest) InfoStr() string {
	return fmt.Sprintf("%d, %d - %s, %s", prvReq.row, prvReq.col, prvReq.subrequest, prvReq.user)
}

type ProviderResponse struct {
	StatusCode int
	Data       map[string]interface{}
}

func ProviderResponseNotImplemented() ProviderResponse {
	return ProviderResponse{
		StatusCode: http.StatusNotImplemented,
		Data: map[string]interface{}{
			"errors": []string{"not implemented"},
		},
	}
}

func ProviderResponseMethodNotAllowed(method string) ProviderResponse {
	return ProviderResponse{
		StatusCode: http.StatusMethodNotAllowed,
		Data: map[string]interface{}{
			"errors": []string{fmt.Sprintf("method %s not allowed", method)},
		},
	}
}

func ProviderResponseEmpty() ProviderResponse {
	return ProviderResponse{
		StatusCode: http.StatusOK,
		Data:       map[string]interface{}{},
	}
}

func (pr *ProviderResponse) Add(key string, value interface{}) {
	pr.Data[key] = value
}

func (pr *ProviderResponse) AddError(error string) {
	errors, exists := pr.Data["errors"]

	if !exists {
		errors = []string{}
	}

	pr.Data["errors"] = append(errors.([]string), error)
}

func (pr *ProviderResponse) GetString(key string) string {
	value, exists := pr.Data[key]

	if !exists {
		return ""
	}

	return value.(string)
}

func (pr *ProviderResponse) GetErrors() string {
	value, exists := pr.Data["errors"]

	if !exists {
		return ""
	}

	return strings.Join(value.([]string), ", ")
}

type ProviderInterface interface {
	SetupFromEnvironment()
	Get(prvReq *ProviderRequest) ProviderResponse
	Center(prvReq *ProviderRequest) ProviderResponse
	Edge(prvReq *ProviderRequest) ProviderResponse
}

type HTTPGetHandler func(w http.ResponseWriter, r *http.Request)
type ProviderGetHandler func(prvReq *ProviderRequest) ProviderResponse

// A ProviderUpdater is a function that updates the provider's state based on external
// things. It does _not_ get to short-circuit the request; it only gets to update
// state.
type ProviderUpdater func(*BaseProvider)

// A ProviderHook is a function that takes a ProviderRequest and a
// BaseRequestStatus and returns a boolean indicating whether the request
// should proceed. If desired, the ProviderHook may modify the
// BaseRequestStatus in place to set the status code and a message to hand
// back to the client.
type ProviderHook func(*BaseProvider, *ProviderRequest, *BaseRequestStatus) bool

type BaseProvider struct {
	Name               string // Name of this kind of provider
	Key                string // Descriptive string for this provider instance
	lock               sync.Mutex
	logger             *slog.Logger
	delayBuckets       []int
	errorFraction      int
	latchFraction      int
	maxRate            float64
	userHeaderName     string
	hostIP             string
	hostName           string
	debugEnabled       bool
	providerGetHandler ProviderGetHandler
	httpGetHandler     HTTPGetHandler

	updaters  []ProviderUpdater
	preHooks  []ProviderHook
	postHooks []ProviderHook

	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec

	latched         bool
	rateCounter     *utils.RateCounter
	lastRequestTime time.Time

	whisper              *whisper.Whisper
	whisperNodeNumber    int
	whisperProcessNumber int
	whisperSelfID        uint32
	whisperServerID      uint32
}

func (bprv *BaseProvider) SetupBasicsFromEnvironment() {
	bprv.debugEnabled = utils.BoolFromEnv("DEBUG_ENABLED", false)

	bprv.userHeaderName = utils.StringFromEnv("USER_HEADER_NAME", "X-Faces-User")
	bprv.hostIP = utils.StringFromEnv("HOST_IP", utils.StringFromEnv("HOSTNAME", "unknown"))

	hostname, err := os.Hostname()

	if err != nil {
		hostname = "unknown"
	}

	bprv.hostName = utils.StringFromEnv("HOSTNAME", hostname)

	bprv.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of requests received",
		},
		[]string{"provider", "hostname", "key", "status"},
	)

	bprv.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Histogram of request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "hostname", "key"},
	)

	prometheus.MustRegister(bprv.requestsTotal)
	prometheus.MustRegister(bprv.requestDuration)

	bprv.Infof("booted on %s (%s)", bprv.hostName, bprv.hostIP)
	bprv.Infof("userHeaderName %v", bprv.userHeaderName)
	bprv.Infof("debug_enabled %v", bprv.debugEnabled)
}

func (bprv *BaseProvider) SetupFromEnvironment() {
	bprv.SetupBasicsFromEnvironment()

	delayBucketsStr := utils.StringFromEnv("DELAY_BUCKETS", "")

	if delayBucketsStr != "" {
		delayBuckets := strings.Split(delayBucketsStr, ",")
		for _, bucketStr := range delayBuckets {
			bucket, err := strconv.Atoi(bucketStr)
			if err == nil {
				if bucket < 0 {
					bucket = 0
				}

				bprv.delayBuckets = append(bprv.delayBuckets, bucket)
			}
		}
	}

	bprv.errorFraction = utils.PercentageFromEnv("ERROR_FRACTION", 0)
	bprv.latchFraction = utils.PercentageFromEnv("LATCH_FRACTION", 0)

	bprv.maxRate = utils.FloatFromEnv("MAX_RATE", 0.0)

	if bprv.maxRate >= 0.1 {
		bprv.rateCounter = utils.NewRateCounter(10)
	}

	bprv.Infof("delay_buckets %v", bprv.delayBuckets)
	bprv.Infof("error_fraction %d", bprv.errorFraction)
	bprv.Infof("latch_fraction %d", bprv.latchFraction)
	bprv.Infof("max_rate %f", bprv.maxRate)
}

func (bprv *BaseProvider) EnableWhisper(whisperAddr string, name string, nodeNumber int, processNumber int) {
	w, err := whisper.NewWhisperWithOptions(whisperAddr, whisper.DefaultPort)

	if err != nil {
		bprv.Warnf("Could not enable whisper: %s", err)
		return
	}

	bprv.whisper = w
	bprv.whisperNodeNumber = nodeNumber
	bprv.whisperProcessNumber = processNumber

	// Server we'll send to, which is a constant
	bprv.whisperServerID = bprv.whisper.GetHashID([]byte("glowsrv"))
	bprv.Infof("Whisper: server ID 0x%08X", bprv.whisperServerID)

	// Local ID based on hostname
	localName := fmt.Sprintf("%s-%d-%d", name, nodeNumber, processNumber)

	bprv.whisperSelfID = crc32.ChecksumIEEE([]byte(localName))
	bprv.whisper.SetID(bprv.whisperSelfID)
	bprv.Infof("Whisper enabled: %s", bprv.whisper.String())
}

func (bprv *BaseProvider) Announce(ok bool, value int) error {
	// No-op if whisper is not enabled
	if bprv.whisper == nil {
		return nil
	}

	msg := GlowMsg{
		Node:    bprv.whisperNodeNumber,
		Process: bprv.whisperProcessNumber,
		OK:      ok,
		Value:   value,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		bprv.Warnf("failed to marshal message: %v", err)
		return err
	}
	// fmt.Printf("ANNOUNCE: JSON payload: %s\n", string(jsonData))

	return bprv.whisper.Send(bprv.whisperServerID, CmdActivity, jsonData)
}

// SetGetHandler sets the function that will be called to get the data for a
// given provider request.
func (bprv *BaseProvider) SetGetHandler(handler ProviderGetHandler) {
	bprv.providerGetHandler = handler
}

// SetHTTPGetHandler sets the function that will be called to handle HTTP GET
// requests. This is lower-level than SetGetHandler; the default HTTP GET handler
// calls the provider-level Get handler (set with SetGetHandler) for requests
// that don't trip the error fraction or get rate limited.
func (bprv *BaseProvider) SetHTTPGetHandler(handler HTTPGetHandler) {
	bprv.httpGetHandler = handler
}

func (bprv *BaseProvider) AddUpdater(updater ProviderUpdater) {
	bprv.updaters = append(bprv.updaters, updater)
}

func (bprv *BaseProvider) AddPreHook(hook ProviderHook) {
	bprv.preHooks = append(bprv.preHooks, hook)
}

func (bprv *BaseProvider) AddPostHook(hook ProviderHook) {
	bprv.postHooks = append(bprv.postHooks, hook)
}

func (bprv *BaseProvider) Infof(format string, args ...interface{}) {
	bprv.logger.Info(bprv.Name + ": " + fmt.Sprintf(format, args...))
}

func (bprv *BaseProvider) Debugf(format string, args ...interface{}) {
	bprv.logger.Debug(bprv.Name + ": " + fmt.Sprintf(format, args...))
}

func (bprv *BaseProvider) Warnf(format string, args ...interface{}) {
	bprv.logger.Warn(bprv.Name + ": " + fmt.Sprintf(format, args...))
}

func (bprv *BaseProvider) SetLogger(logger *slog.Logger) {
	bprv.logger = logger
}

func (bprv *BaseProvider) SetDebug(debug bool) {
	bprv.debugEnabled = debug
}

func (bprv *BaseProvider) Lock() {
	bprv.lock.Lock()
}

func (bprv *BaseProvider) Unlock() {
	bprv.lock.Unlock()
}

func (bprv *BaseProvider) IsLatched() bool {
	return bprv.latched
}

func (bprv *BaseProvider) SetLatched(latched bool) {
	bprv.latched = latched
}

func (bprv *BaseProvider) GetUserHeaderName() string {
	return bprv.userHeaderName
}

func (bprv *BaseProvider) ErrorFraction() int {
	return bprv.errorFraction
}

func (bprv *BaseProvider) SetErrorFraction(fraction int) {
	bprv.errorFraction = fraction
}

// CheckUnlatch checks to see if we should unlatch the provider.
func (bprv *BaseProvider) CheckUnlatch(now time.Time) {
	bprv.lock.Lock()
	defer bprv.lock.Unlock()

	if bprv.latched {
		// How long has it been since our last request?
		delta := now.Sub(bprv.lastRequestTime)

		if delta.Seconds() > 30 {
			// It's been thirty full seconds since our last request. If we
			// were latched into the error state, it's time to come out.
			bprv.latched = false
		}
	}
}

// DelayIfNeeded delays if there are delay buckets set.
func (bprv *BaseProvider) DelayIfNeeded(rstat *BaseRequestStatus) {
	if rstat.delayMs > 0 {
		time.Sleep(time.Duration(rstat.delayMs) * time.Millisecond)
	}
}

// CheckRequestStatus checks the state of the provider and decides whether
// it's OK to have the request continue, or whether it should be failed for
// various reasons:
//
// - If maxRate is set, then we first check the rate counter to see if we
//   need to fail due to rate limiting.
// - Otherwise, if we're latched into an error state, then we immediately fail.
// - Otherwise, if errorFraction is set, then errorFraction% of requests will fail
//   and, if latchFraction is set, then every error has a latchFraction % chance
//   to latch the error state.

func (bprv *BaseProvider) CheckRequestStatus() *BaseRequestStatus {
	// We need to figure out if we're going to send an error.
	rstat := &BaseRequestStatus{
		// It's true that not every provider uses HTTP, but we're going
		// to use the HTTP status codes as a common way to signal errors.
		statusCode: http.StatusOK,
	}

	start := time.Now()

	// Is a rate limiter active? We do this first because if the rate limiter
	// trips, we want the service to be unable to do _any_ processing, including
	// checking for other errors.

	if bprv.rateCounter != nil {
		bprv.rateCounter.Mark(start)
		rate := bprv.rateCounter.CurrentRate()

		if rate >= bprv.maxRate {
			// Bzzzt! Rate limited.
			rstat.ratelimited = true
			rstat.message = fmt.Sprintf("Rate limited (%.1f RPS > max %.1f RPS)", rate, bprv.maxRate)
		}
	}

	// OK, if rate limiting didn't get us, we might still have an error.
	if !rstat.ratelimited {
		// If we've gotten latched into an error state, we're definitely sending
		// an error.

		if bprv.IsLatched() {
			rstat.latched = true
			rstat.errored = true
			rstat.message = "Latched into error state"
			rstat.statusCode = 599
		} else if bprv.errorFraction > 0 {
			// Not latched, but there's a chance of an error here too.
			if rand.Intn(100) <= bprv.errorFraction {
				bprv.Debugf("error fraction triggered")

				// Yup. Error.
				rstat.errored = true
				rstat.message = "" // No message, the provider will fill this in.
				rstat.statusCode = 500

				// We might get latched here, too.
				if bprv.latchFraction > 0 && rand.Intn(100) <= bprv.latchFraction {
					bprv.SetLatched(true)

					rstat.latched = true
					rstat.message = "Latched into error state"
					rstat.statusCode = 599
				}
			}
		}
	}

	if len(bprv.delayBuckets) > 0 {
		delayMs := bprv.delayBuckets[rand.Intn(len(bprv.delayBuckets))]
		rstat.delayMs = delayMs
	}

	return rstat
}

func (bprv *BaseProvider) HandleRequest(start time.Time, prvReq *ProviderRequest) ProviderResponse {
	resp := ProviderResponseEmpty()

	bprv.CheckUnlatch(start)

	if bprv.updaters != nil {
		for _, updater := range bprv.updaters {
			updater(bprv)
		}
	}

	rstat := bprv.CheckRequestStatus()

	if bprv.whisper != nil {
		succeeded := !(rstat.IsErrored() || rstat.IsRateLimited())
		err := bprv.Announce(succeeded, int(rstat.DelayMs()))

		if err != nil {
			bprv.Warnf("Could not send whisper announce: %s", err)
		}
	}

	if bprv.preHooks != nil {
		for _, hook := range bprv.preHooks {
			proceed := hook(bprv, prvReq, rstat)

			if !proceed {
				bprv.Debugf("pre-hook short-circuited with status %03d, message %s", rstat.StatusCode(), rstat.Message())

				rstat.errored = true
			}
		}
	}

	bprv.DelayIfNeeded(rstat)

	if rstat.IsRateLimited() {
		bprv.Debugf("RATELIMIT(%s) => %s", prvReq.InfoStr(), rstat.Message())

		resp.StatusCode = http.StatusTooManyRequests
		resp.AddError(rstat.Message())
	} else if rstat.IsErrored() {
		msg := rstat.Message()

		if msg == "" {
			msg = fmt.Sprintf("%s error! (error fraction %d%%)", bprv.Name, bprv.errorFraction)
		}

		bprv.Debugf("ERROR(%s) => %d, %s", prvReq.InfoStr(), rstat.StatusCode(), msg)

		resp.StatusCode = rstat.StatusCode()
		resp.AddError(msg)
	} else {
		resp = bprv.providerGetHandler(prvReq)

		dataJSON, err := json.Marshal(resp.Data)

		if err != nil {
			bprv.Warnf("couldn't marshal data: %s", err)
			dataJSON = []byte("{????}")
		}

		if resp.StatusCode == http.StatusOK {
			bprv.Debugf("OK(%s) => %d, %s", prvReq.InfoStr(), resp.StatusCode, string(dataJSON))
		} else {
			bprv.Debugf("FAIL(%s) => %d, %s", prvReq.InfoStr(), resp.StatusCode, string(dataJSON))
		}
	}

	if bprv.postHooks != nil {
		for _, hook := range bprv.postHooks {
			if !hook(bprv, prvReq, rstat) {
				bprv.Debugf("post-hook errored with status %03d, message %s", rstat.StatusCode(), rstat.Message())
			}
		}
	}

	end := time.Now()
	delta := end.Sub(start)

	bprv.lastRequestTime = end

	bprv.requestsTotal.WithLabelValues(bprv.Name, bprv.hostName, bprv.Key, fmt.Sprintf("%03d", resp.StatusCode)).Inc()
	bprv.requestDuration.WithLabelValues(bprv.Name, bprv.hostName, bprv.Key).Observe(delta.Seconds())

	return resp
}
