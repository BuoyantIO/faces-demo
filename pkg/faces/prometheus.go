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
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Start the HTTP server for Prometheus metrics.
func StartPrometheusServer() {
	promServer := &http.Server{
		Handler: promhttp.Handler(),
		Addr:    ":9090", // Prometheus scrapes metrics from this port
	}

	go func() {
		if err := promServer.ListenAndServe(); err != nil {
			log.Fatalf("Unable to start a http server: %v", err)
		}
	}()
}
