// SPDX-FileCopyrightText: 2026 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Buoyant Inc.
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
	"fmt"
	"log"
	"net/http"
	"os"
)

// Faces MCP Server main entry point
func main() {
	// Create the MCP server with all the tools we know about.
	server := NewMCPServerFromEnvironment([]Tool{
		&QuerySmileyTool{},
		&QueryColorTool{},
		&QueryFaceTool{},
		&UpdateSmileyTool{},
		&UpdateColorTool{},
	})

	// Check transport mode: SSE (the default) or stdio
	mode := getEnv("TRANSPORT", "sse")

	if mode == "sse" {
		// SSE mode - HTTP server with Server-Sent Events
		port := getEnv("PORT", "3000")
		addr := fmt.Sprintf(":%s", port)

		http.HandleFunc("/sse", server.handleSSE)
		http.HandleFunc("/messages", server.handleMessages)

		log.Printf("Faces MCP Server starting on %s (SSE mode)", addr)
		log.Printf("Smiley URL: %s", server.smileyURL)
		log.Printf("Color URL: %s", server.colorURL)
		log.Printf("Face URL: %s", server.faceURL)

		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		// stdio mode (default) - standard MCP protocol
		log.SetOutput(os.Stderr) // Send logs to stderr to avoid interfering with stdio protocol

		log.Printf("Faces MCP Server starting in stdio mode")
		log.Printf("Smiley URL: %s", server.smileyURL)
		log.Printf("Color URL: %s", server.colorURL)
		log.Printf("Face URL: %s", server.faceURL)

		server.handleStdio()
	}
}
