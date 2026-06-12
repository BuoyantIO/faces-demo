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

package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"fmt"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	Name = "load"
)

func main() {
	utils.InitLogging()

	target := os.Getenv("LOAD_TARGET")
	rps := os.Getenv("LOAD_RPS")

	if target == "" || rps == "" {
		slog.Error(fmt.Sprintf("%s: LOAD_TARGET and LOAD_RPS must be set", Name))
		os.Exit(1)
	}

	debug, _ := strconv.ParseBool(os.Getenv("LOAD_DEBUG"))

	// Convert rps to an integer
	rpsInt, err := strconv.Atoi(rps)
	if err != nil {
		slog.Error("%s: failed to convert rps to an integer: %v", Name, err)
		os.Exit(1)
	}

	hostName, err := os.Hostname()

	if err != nil {
		hostName = "unknown"
	}

	hostName = utils.StringFromEnv("HOSTNAME", hostName)

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of requests sent",
		},
		[]string{"provider", "hostname", "target", "status"},
	)

	requestErrorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_errors_total",
			Help: "Total number of requests that we couldn't send",
		},
		[]string{"provider", "hostname", "target"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Histogram of request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "hostname", "key"},
	)

	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestErrorsTotal)
	prometheus.MustRegister(requestDuration)

	faces.StartPrometheusServer()

	// Use a ticker goroutine to send requests
	ticker := time.NewTicker(time.Second / time.Duration(rpsInt))
	defer ticker.Stop()
	count := 0

	go func() {
		for range ticker.C {
			go func() {
				start := time.Now()

				// Make a GET request to http://target/
				resp, err := http.Get(fmt.Sprintf("http://%s/", target))
				if err != nil {
					slog.Warn(fmt.Sprintf("%s: failed to make request: %v", Name, err))
					requestErrorsTotal.WithLabelValues(Name, hostName, target).Inc()
					return
				}
				defer resp.Body.Close()

				// Read the response body
				body, _ := io.ReadAll(resp.Body)

				end := time.Now()
				delta := end.Sub(start)

				requestsTotal.WithLabelValues(Name, hostName, target, fmt.Sprintf("%03d", resp.StatusCode)).Inc()
				requestDuration.WithLabelValues(Name, hostName, target).Observe(delta.Seconds())

				if debug {
					fmt.Printf("http://%s/ %d %s\n", target, resp.StatusCode, string(body))
				}

				count++
			}()
		}
	}()

	// Start a second ticker to print out the metrics every 10 seconds
	log_period := 10
	ticker2 := time.NewTicker(time.Duration(log_period) * time.Second)
	defer ticker.Stop()
	last_count := 0

	for range ticker2.C {
		go func() {
			since_last := count - last_count
			last_count = count

			rps := float64(since_last) / float64(log_period)
			slog.Info(fmt.Sprintf("%s: %.02f RPS (%d requests / %d seconds)", Name, rps, since_last, log_period))
		}()
	}
}
