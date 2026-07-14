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
	"log/slog"
	"os"

	"flag"

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

	// Define a command-line flag for the port number
	port := flag.Int("port", 8000, "the port number to listen on")
	flag.Parse()

	whisperAddr := utils.StringFromEnv("WHISPER_ADDRESS", "")
	enablePrometheus := utils.BoolFromEnv("ENABLE_PROMETHEUS", true)

	cprv := faces.NewColorProviderFromEnvironment()

	if whisperAddr != "" {
		nodeNumber := utils.IntFromEnv("WHISPER_NODE_NUMBER", 0)
		processNumber := utils.IntFromEnv("WHISPER_PROCESS_NUMBER", 3)

		cprv.EnableWhisper(whisperAddr, "color", nodeNumber, processNumber)
	}

	server := faces.NewColorServer(cprv)

	if enablePrometheus {
		faces.StartPrometheusServer()
	}

	// Chaos is now a gRPC method (GetChaos/UpdateChaos) on the same port as the
	// color service — no separate HTTP sidecar needed. The HTTP sidecar on port 8001
	// has been removed; all color traffic stays on the single gRPC port.

	err := server.Start(*port)

	if err != nil {
		slog.Error("Unable to serve gRPC", "error", err)
		os.Exit(1)
	}
}
