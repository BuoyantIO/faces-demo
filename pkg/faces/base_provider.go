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
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type ProviderResponse struct {
	StatusCode int
	Body       string
}

type BaseProvider struct {
	lock           sync.Mutex
	logger         *slog.Logger
	delayBuckets   []int
	errorFraction  int
	latchFraction  int
	maxRate        float64
	userHeaderName string
	hostIP         string
	debugEnabled   bool

	latched         bool
	rateCounter     *utils.RateCounter
	lastRequestTime time.Time
}

func (bprv *BaseProvider) SetupFromEnvironment() {
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

	bprv.debugEnabled = utils.BoolFromEnv("DEBUG_ENABLED", false)

	bprv.maxRate = utils.FloatFromEnv("MAX_RATE", 0.0)

	if bprv.maxRate >= 0.1 {
		bprv.rateCounter = utils.NewRateCounter(10)
	}

	bprv.userHeaderName = utils.StringFromEnv("USER_HEADER_NAME", "X-Faces-User")
	bprv.hostIP = utils.StringFromEnv("HOST_IP", utils.StringFromEnv("HOSTNAME", "unknown"))

	bprv.Infof("booted on %s", bprv.hostIP)
	bprv.Infof("delay_buckets %v", bprv.delayBuckets)
	bprv.Infof("error_fraction %d", bprv.errorFraction)
	bprv.Infof("latch_fraction %d", bprv.latchFraction)
	bprv.Infof("debug_enabled %v", bprv.debugEnabled)
	bprv.Infof("max_rate %f", bprv.maxRate)
	bprv.Infof("userHeaderName %v", bprv.userHeaderName)
}

func (bprv *BaseProvider) Infof(format string, args ...interface{}) {
	bprv.logger.Info(fmt.Sprintf(format, args...))
}

func (bprv *BaseProvider) Debugf(format string, args ...interface{}) {
	bprv.logger.Debug(fmt.Sprintf(format, args...))
}

func (bprv *BaseProvider) Warnf(format string, args ...interface{}) {
	bprv.logger.Warn(fmt.Sprintf(format, args...))
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
func (bprv *BaseProvider) DelayIfNeeded() {
	if len(bprv.delayBuckets) > 0 {
		delayMs := bprv.delayBuckets[rand.Intn(len(bprv.delayBuckets))]
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
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

	return rstat
}
