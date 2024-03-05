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
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"fmt"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rotaryAPin := utils.IntFromEnv("ROTARY_A_PIN", 5)
	rotaryBPin := utils.IntFromEnv("ROTARY_B_PIN", 6)
	buttonPin := utils.IntFromEnv("BUTTON_PIN", 4)
	ledGreenPin := utils.IntFromEnv("LED_GREEN_PIN", 19)
	ledRedPin := utils.IntFromEnv("LED_RED_PIN", 13)

	hw, err := NewHardwareStuff(rotaryAPin, rotaryBPin, buttonPin, ledGreenPin, ledRedPin)

	if err != nil {
		log.Fatal(fmt.Sprintf("Could not initialize hardware: %s", err))
	}

	defer hw.Close()

	fmt.Printf("%s hardware rotaryAPin %d\n", time.Now().Format(time.RFC3339), hw.rotaryAPin)
	fmt.Printf("%s hardware rotaryBPin %d\n", time.Now().Format(time.RFC3339), hw.rotaryBPin)
	fmt.Printf("%s hardware buttonPin %d\n", time.Now().Format(time.RFC3339), hw.buttonPin)
	fmt.Printf("%s hardware ledGreenPin %d\n", time.Now().Format(time.RFC3339), hw.ledGreenPin)
	fmt.Printf("%s hardware ledRedPin %d\n", time.Now().Format(time.RFC3339), hw.ledRedPin)

	var server *faces.BaseServer

	service := os.Getenv("FACES_SERVICE")

	switch service {
	case "face":
		server = &faces.NewFaceServer("FaceServer").BaseServer

	case "smiley":
		server = &faces.NewSmileyServer("SmileyServer").BaseServer

	case "color":
		server = &faces.NewColorServer("ColorServer").BaseServer

	default:
		log.Fatalf("Unknown service: %s", service)
	}

	hw.Watch(server.ErrorFraction(), server.Latched())

	server.SetPreHook(func(srv *faces.BaseServer, r *http.Request, rstat *faces.BaseRequestStatus) bool {
		srv.Lock()
		defer srv.Unlock()

		if srv.ErrorFraction() != hw.serverErrorFraction {
			srv.SetErrorFraction(hw.serverErrorFraction)
			fmt.Printf("%s %s: errorFraction %d\n", time.Now().Format(time.RFC3339), srv.Name, srv.ErrorFraction())
		}

		if srv.Latched() != hw.serverLatched {
			srv.SetLatched(hw.serverLatched)
			fmt.Printf("%s %s: latched %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.Latched())
		}

		if rstat.IsErrored() || rstat.IsRateLimited() {
			hw.ledOn("red")
		} else {
			hw.ledOn("green")
		}

		return true
	})

	server.SetPostHook(func(srv *faces.BaseServer, r *http.Request, rstat *faces.BaseRequestStatus) bool {
		hw.ledOff("red")
		hw.ledOff("green")

		return true
	})

	fmt.Printf("%s %s (pi), commit %s, built at %s\n", server.Name, version, commit, date)

	// Define a command-line flag for the port number
	port := flag.Int("port", 8000, "the port number to listen on")
	flag.Parse()

	// Use the port number from the command line flag
	addr := fmt.Sprintf(":%d", *port)
	server.ListenAndServe(addr)
}
