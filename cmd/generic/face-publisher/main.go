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
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	utils.InitLogging()

	port := flag.Int("port", 8000, "the port number to listen on")
	flag.Parse()

	whisperAddr := utils.StringFromEnv("WHISPER_ADDRESS", "")
	enablePrometheus := utils.BoolFromEnv("ENABLE_PROMETHEUS", true)

	// Connect to MySQL + Redis, run DDL.  Panics on failure.
	fprv := faces.NewFacePublisherFromEnvironment()

	if whisperAddr != "" {
		nodeNumber := utils.IntFromEnv("WHISPER_NODE_NUMBER", 0)
		processNumber := utils.IntFromEnv("WHISPER_PROCESS_NUMBER", 1)
		fprv.EnableWhisper(whisperAddr, "face-publisher", nodeNumber, processNumber)
	}

	if enablePrometheus {
		faces.StartPrometheusServer()
	}

	// Start the HTTP server first so liveness/readiness probes (/healthz) respond
	// immediately.  WarmQueue, the janitor, the queue monitor, and the publish loops
	// all start in a background goroutine — they run after the server is already
	// accepting connections.
	//
	// WarmQueue is still guaranteed to complete before StartPublishLoops begins, so
	// the queue is pre-filled before any new messages are produced.
	server := faces.NewBaseHTTPServer(&fprv.BaseProvider)
	fprv.RegisterAdminRoutes(server)

	go func() {
		fprv.WarmQueue()
		go fprv.RunJanitor()
		go fprv.RunQueueMonitor()
		fprv.StartPublishLoops()
	}()

	if err := server.Start(fmt.Sprintf(":%d", *port)); err != nil {
		slog.Error(fmt.Sprintf("Unable to serve HTTP: %v", err))
		os.Exit(1)
	}
}
