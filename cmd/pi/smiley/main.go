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
	"log"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/raspberry_pi"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	utils.InitLogging()

	// Initialize hardware
	hw, err := raspberry_pi.NewAutomaticHardwareStuff()

	if err != nil {
		log.Fatal(fmt.Sprintf("Could not initialize hardware: %s", err))
	}

	defer hw.Close()

	// Define a command-line flag for the port number
	port := flag.Int("port", 8000, "the port number to listen on")
	flag.Parse()

	sprv := faces.NewSmileyProviderFromEnvironment()
	sprv.SetPreHook(hw.PreHook)
	sprv.SetPostHook(hw.PostHook)

	hw.Watch(sprv.ErrorFraction(), sprv.IsLatched())

	server := faces.NewBaseHTTPServer(&sprv.BaseProvider)
	server.Start(fmt.Sprintf(":%d", *port))

	faces.StartPrometheusServer()

	// Wait for servers to finish
	select {}
}
