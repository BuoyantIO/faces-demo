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

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"flag"
	"fmt"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var server *faces.BaseServer

	service := os.Getenv("FACES_SERVICE")

	switch service {
	case "gui":
		server = &faces.NewGUIServer("GUIServer").BaseServer

	case "face":
		server = &faces.NewFaceServer("FaceServer").BaseServer

	case "smiley":
		server = &faces.NewSmileyServer("SmileyServer").BaseServer

	case "color":
		server = &faces.NewColorServer("ColorServer").BaseServer

	case "load":
		fmt.Printf("Running load generator")

	default:
		log.Fatalf("Unknown service: %s", service)
	}

	if server != nil {
		fmt.Printf("%s %s (generic), commit %s, built at %s\n", server.Name, version, commit, date)

		// Define a command-line flag for the port number
		port := flag.Int("port", 8000, "the port number to listen on")
		flag.Parse()

		// Use the port number from the command line flag
		addr := fmt.Sprintf(":%d", *port)
		server.ListenAndServe(addr)
	} else {
		target := os.Getenv("LOAD_TARGET")
		rps := os.Getenv("LOAD_RPS")
		debug, _ := strconv.ParseBool(os.Getenv("LOAD_DEBUG"))

		// Convert rps to an integer
		rpsInt, err := strconv.Atoi(rps)
		if err != nil {
			log.Fatalf("Failed to convert rps to an integer: %v", err)
		}

		// Create a ticker to control the rate of requests
		ticker := time.NewTicker(time.Second / time.Duration(rpsInt))
		defer ticker.Stop()
		count := 0

		// Start a goroutine to send requests
		for range ticker.C {
			go func() {
				// Make a GET request to http://target/
				resp, err := http.Get(fmt.Sprintf("http://%s/", target))
				if err != nil {
					log.Fatalf("Failed to make request: %v", err)
				}
				defer resp.Body.Close()

				// Read the response body
				body, _ := io.ReadAll(resp.Body)

				if debug {
					fmt.Printf("http://%s/ %d %s\n", target, resp.StatusCode, string(body))
				}

				count++

				if count >= (rpsInt * 10) {
					fmt.Printf("Sent %d requests\n", count)
					count = 0
				}
			}()
		}
	}
}
