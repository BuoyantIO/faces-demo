// SPDX-FileCopyrightText: 2024 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2024 Buoyant Inc.
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
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type BaseRequestStatus struct {
	errored     bool
	ratelimited bool
	latched     bool
	message     string
	statusCode  int
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

type BaseServerResponse struct {
	StatusCode int
	Data       map[string]interface{}
}

// BaseServerHandler is a function that takes an http.Request and a
// BaseRequestStatus and returns a BaseServerResponse. It's used as a GET
// handler for BaseServer.
type BaseServerHandler func(*http.Request, *BaseRequestStatus) *BaseServerResponse

// A Hook is a function that takes an http.Request and a BaseRequestStatus and
// returns a boolean indicating whether the request should proceed. If desired,
// the Hook may modify the BaseRequestStatus in place to set the status code and
// a message to hand back to the client.
type Hook func(*BaseServer, *http.Request, *BaseRequestStatus) bool

type BaseServer struct {
	Name            string
	lock            sync.Mutex
	latchFraction   int
	latched         bool
	delayBuckets    []int
	errorFraction   int
	errorText       string
	debugEnabled    bool
	lastRequestTime time.Time
	hostIP          string
	rateCounter     *utils.RateCounter
	maxRate         float64
	server          *http.ServeMux
	preHook         Hook
	postHook        Hook
	userHeaderName  string
}

func NewBaseServer(serverName string) *BaseServer {
	srv := &BaseServer{
		Name: serverName,
	}

	srv.SetupFromEnvironment()

	return srv
}

func (srv *BaseServer) Lock() {
	srv.lock.Lock()
}

func (srv *BaseServer) Unlock() {
	srv.lock.Unlock()
}

func (srv *BaseServer) SetPreHook(hook Hook) {
	srv.preHook = hook
}

func (srv *BaseServer) SetPostHook(hook Hook) {
	srv.postHook = hook
}

func (srv *BaseServer) ErrorFraction() int {
	return srv.errorFraction
}

func (srv *BaseServer) SetErrorFraction(fraction int) {
	srv.errorFraction = fraction
}

func (srv *BaseServer) Latched() bool {
	return srv.latched
}

func (srv *BaseServer) SetLatched(latched bool) {
	srv.latched = latched
}

// SetDebug sets the debugEnabled flag on the server.
func (srv *BaseServer) SetDebug(debugEnabled bool) {
	srv.debugEnabled = debugEnabled
	fmt.Printf("%s %s: debug set to %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.debugEnabled)
}

func (srv *BaseServer) SetupFromEnvironment() {
	fmt.Printf("%s %s: setupFromEnvironment starting\n", time.Now().Format(time.RFC3339), srv.Name)

	if srv.server == nil {
		srv.server = http.NewServeMux()
	}

	if srv.errorText == "" {
		srv.errorText = "Error fraction triggered"
	}

	srv.RegisterCustom("/rl", srv.rlGetHandler)

	srv.userHeaderName = utils.StringFromEnv("USER_HEADER_NAME", "X-Faces-User")
	srv.hostIP = utils.StringFromEnv("HOST_IP", utils.StringFromEnv("HOSTNAME", "unknown"))

	delayBucketsStr := utils.StringFromEnv("DELAY_BUCKETS", "")

	if delayBucketsStr != "" {
		delayBuckets := strings.Split(delayBucketsStr, ",")
		for _, bucketStr := range delayBuckets {
			bucket, err := strconv.Atoi(bucketStr)
			if err == nil {
				if bucket < 0 {
					bucket = 0
				}

				srv.delayBuckets = append(srv.delayBuckets, bucket)
			}
		}
	}

	srv.errorFraction = utils.PercentageFromEnv("ERROR_FRACTION", 0)
	srv.latchFraction = utils.PercentageFromEnv("LATCH_FRACTION", 0)

	srv.debugEnabled = utils.BoolFromEnv("DEBUG_ENABLED", false)

	srv.maxRate = utils.FloatFromEnv("MAX_RATE", 0.0)

	if srv.maxRate >= 0.1 {
		srv.rateCounter = utils.NewRateCounter(10)
	}

	fmt.Printf("%s %s: booted on %s\n", time.Now().Format(time.RFC3339), srv.Name, srv.hostIP)
	fmt.Printf("%s %s: delay_buckets %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.delayBuckets)
	fmt.Printf("%s %s: error_fraction %d\n", time.Now().Format(time.RFC3339), srv.Name, srv.errorFraction)
	fmt.Printf("%s %s: latch_fraction %d\n", time.Now().Format(time.RFC3339), srv.Name, srv.latchFraction)
	fmt.Printf("%s %s: debug_enabled %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.debugEnabled)
	fmt.Printf("%s %s: max_rate %f\n", time.Now().Format(time.RFC3339), srv.Name, srv.maxRate)
	fmt.Printf("%s %s: userHeaderName %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.userHeaderName)
}

func (srv *BaseServer) ListenAndServe(port string) {
	fmt.Printf("%s %s: listening on %s\n", time.Now().Format(time.RFC3339), srv.Name, port)

	http.ListenAndServe(port, srv.server)
}

func (srv *BaseServer) RegisterNormal(path string, handler BaseServerHandler) {
	customHandler := func(w http.ResponseWriter, r *http.Request) {
		srv.standardGetHandler(w, r, handler)
	}

	srv.server.HandleFunc(path, srv.customHandler(customHandler))
}

func (srv *BaseServer) RegisterCustom(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	srv.server.HandleFunc(path, srv.customHandler(handler))
}

func (srv *BaseServer) customHandler(userFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			srv.standardHeaders(w, r, http.StatusOK, "")
		} else if r.Method == "GET" {
			userFunc(w, r)
		} else {
			http.Error(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
		}
	}
}

func (srv *BaseServer) Latch() {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	srv.latched = true
}

func (srv *BaseServer) IsLatched() bool {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	return srv.latched
}

func (srv *BaseServer) Unlatch() {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	srv.latched = false
}

// CurrentRate gets the current rate from the rate counter, if one is active,
// otherwise it returns 0.0.
func (srv *BaseServer) CurrentRate() float64 {
	if srv.rateCounter != nil {
		return srv.rateCounter.CurrentRate()
	}
	return 0.0
}

func (srv *BaseServer) CheckUnlatch(now time.Time) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	if srv.latched {
		// How long has it been since our last request?
		delta := now.Sub(srv.lastRequestTime)

		if delta.Seconds() > 30 {
			// It's been thirty full seconds since our last request. If we
			// were latched into the error state, it's time to come out.
			srv.latched = false
		}
	}
}

func (srv *BaseServer) DelayIfNeeded() {
	if len(srv.delayBuckets) > 0 {
		delayMs := srv.delayBuckets[rand.Intn(len(srv.delayBuckets))]
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}

func (srv *BaseServer) CheckRequestStatus(w http.ResponseWriter, r *http.Request) *BaseRequestStatus {
	// We need to figure out if we're going to send an error.
	rstat := &BaseRequestStatus{
		statusCode: http.StatusOK,
	}

	start := time.Now()

	// Is a rate limiter active? We do this first because if the rate limiter
	// trips, we want the service to be unable to do _any_ processing, including
	// checking for other errors.

	if srv.rateCounter != nil {
		srv.rateCounter.Mark(start)
		rate := srv.rateCounter.CurrentRate()

		if rate >= srv.maxRate {
			// Bzzzt! Rate limited.
			rstat.ratelimited = true
		}
	}

	// OK, if rate limiting didn't get us, we might still have an error.
	if !rstat.ratelimited {
		// If we've gotten latched into an error state, we're definitely sending
		// an error.

		if srv.IsLatched() {
			rstat.latched = true
			rstat.errored = true
			rstat.message = "Latched into error state"
			rstat.statusCode = 599
		} else if srv.errorFraction > 0 {
			// Not latched, but there's a chance of an error here too.
			if rand.Intn(100) <= srv.errorFraction {
				if srv.debugEnabled {
					fmt.Printf("%s %s: error fraction triggered\n", time.Now().Format(time.RFC3339), srv.Name)
				}

				// Yup. Error.
				rstat.errored = true
				rstat.message = srv.errorText
				rstat.statusCode = 500

				// We might get latched here, too.
				if srv.latchFraction > 0 && rand.Intn(100) <= srv.latchFraction {
					srv.Latch()
					rstat.message = "Error fraction triggered and latched!"
					rstat.latched = true
					rstat.statusCode = 599
				}
			}
		}
	}

	return rstat
}

func (srv *BaseServer) rlGetHandler(w http.ResponseWriter, r *http.Request) {
	rateStr := "N/A"

	if srv.rateCounter != nil {
		rate := srv.rateCounter.CurrentRate()
		rateStr = fmt.Sprintf("%.1f", rate)
	}

	srv.StandardResponse(w, r, &BaseServerResponse{
		StatusCode: http.StatusOK,
		Data: map[string]interface{}{
			"rl": rateStr,
		},
	})
}

func (srv *BaseServer) standardGetHandler(w http.ResponseWriter, r *http.Request, userFunc BaseServerHandler) {
	start := time.Now()

	srv.CheckUnlatch(start)
	rstat := srv.CheckRequestStatus(w, r)

	var response *BaseServerResponse

	if srv.preHook != nil {
		proceed := srv.preHook(srv, r, rstat)

		if !proceed {
			// fmt.Printf("%s: preHook short-circuit with status %d\n\n", time.Now().Format(time.RFC3339), srv.Name, r.statusCode)

			rstat.errored = true
		}
	}

	srv.DelayIfNeeded()

	if !rstat.errored {
		// OK! We'll generate a nice JSON body for this after we call the user
		// function to find out what data to return. (Note that we DO call the
		// user function if our internal rate limiter fired.)

		response = userFunc(r, rstat)
	}

	if srv.postHook != nil {
		// The postHook can't actually trigger a short-circuit.
		if !srv.postHook(srv, r, rstat) {
			fmt.Printf("%s %s: postHook error: status %d, message %s\n", time.Now().Format(time.RFC3339), srv.Name, rstat.statusCode, rstat.message)
		}
	}

	srv.lastRequestTime = start

	if rstat.errored {
		srv.StandardError(w, r, rstat.statusCode, rstat.message)
	} else {
		srv.StandardResponse(w, r, response)
	}
}

func (srv *BaseServer) StandardError(w http.ResponseWriter, r *http.Request, statusCode int, responseBody string) {
	srv.standardHeaders(w, r, statusCode, "text/plain")
	w.Write([]byte(responseBody))
}

func (srv *BaseServer) StandardResponse(w http.ResponseWriter, r *http.Request, response *BaseServerResponse) {
	rdict := map[string]interface{}{
		"path":           r.URL.Path,
		"client_address": r.RemoteAddr,
		"method":         r.Method,
		"headers":        r.Header,
		"status":         response.StatusCode,
	}
	if response.Data != nil {
		for key, value := range response.Data {
			rdict[key] = value
		}
	}

	responseBodyBytes, err := json.Marshal(rdict)
	responseType := "application/json"

	if err != nil {
		// This should be "impossible".
		responseBody := fmt.Sprintf("Error marshalling response: %s", err)
		responseBodyBytes = []byte(responseBody)
		responseType = "text/plain"
	}

	srv.standardHeaders(w, r, response.StatusCode, responseType)
	w.Write([]byte(responseBodyBytes))
}

func (srv *BaseServer) standardHeaders(w http.ResponseWriter, r *http.Request, statusCode int, contentType string) {
	if srv.debugEnabled {
		fmt.Printf("%s %s: %s %s %d\n", time.Now().Format(time.RFC3339), srv.Name, r.Method, r.URL.Path, statusCode)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set(srv.userHeaderName, r.Header.Get(srv.userHeaderName))
	w.Header().Set("User-Agent", r.Header.Get("User-Agent"))
	w.Header().Set("X-Faces-Pod", srv.hostIP)
	w.WriteHeader(statusCode)
}
