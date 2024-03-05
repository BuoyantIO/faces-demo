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
	"log"
	"os"

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

	default:
		log.Fatalf("Unknown service: %s", service)
	}

	fmt.Printf("%s %s (generic), commit %s, built at %s\n", server.Name, version, commit, date)

	// Define a command-line flag for the port number
	port := flag.Int("port", 8000, "the port number to listen on")
	flag.Parse()

	// Use the port number from the command line flag
	addr := fmt.Sprintf(":%d", *port)
	server.ListenAndServe(addr)
}
