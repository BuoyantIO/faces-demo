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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type GUIServer struct {
	BaseServer
	dataPath string
	bgColor  string
	hideKey  bool
	showPods bool
}

func NewGUIServer(serverName string) *GUIServer {
	srv := &GUIServer{
		BaseServer: BaseServer{
			Name: serverName,
		},
	}

	srv.SetupFromEnvironment()
	// srv.SetUpdater(srv.updater)

	srv.RegisterCustom("/", srv.guiGetHandler)

	return srv
}

func (srv *GUIServer) SetupFromEnvironment() {
	srv.BaseServer.SetupFromEnvironment()

	srv.dataPath = utils.StringFromEnv("DATA_PATH", "/app/data")
	srv.bgColor = utils.StringFromEnv("COLOR", "white")
	srv.hideKey = utils.BoolFromEnv("HIDE_KEY", false)
	srv.showPods = utils.BoolFromEnv("SHOW_PODS", false)

	fmt.Printf("%s %s: dataPath %s\n", time.Now().Format(time.RFC3339), srv.Name, srv.dataPath)
	fmt.Printf("%s %s: bgColor %s\n", time.Now().Format(time.RFC3339), srv.Name, srv.bgColor)
	fmt.Printf("%s %s: hideKey %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.hideKey)
	fmt.Printf("%s %s: showPods %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.showPods)
}

func (srv *GUIServer) guiGetHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	fmt.Printf("%s %s: GET %s\n", time.Now().Format(time.RFC3339), srv.Name, r.URL.Path)
	fmt.Printf("  userHeaderName: %s\n", srv.userHeaderName)
	fmt.Printf("  headers: %s\n", r.Header)

	// Default user comes from the header value used to load the GUI, which
	// might be "".
	user := r.Header.Get(srv.userHeaderName)

	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "unknown"
	}

	fmt.Printf("  user: %s\n", user)
	fmt.Printf("  userAgent: %s\n", userAgent)

	podID := srv.hostIP
	rcode := http.StatusNotFound
	rtext := fmt.Sprintf("%s not found", r.URL.Path)
	rtype := "text/html"

	// Handle readiness checks first (they're simple).
	if r.URL.Path == "/ready" {
		end := time.Now()
		latencyMs := end.Sub(start).Milliseconds()

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set(srv.userHeaderName, user)
		w.Header().Set("X-Faces-User-Agent", userAgent)
		w.Header().Set("X-Faces-Latency", strconv.FormatInt(latencyMs, 10))
		w.Header().Set("X-Faces-Pod", podID)
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "Ready and waiting!")
		return
	} else if r.URL.Path == "/" {
		// "/" means "index.html", which we expect to find in the dataPath.
		rcode = http.StatusOK

		indexPath := filepath.Join(srv.dataPath, "index.html")

		raw, err := os.ReadFile(indexPath)

		if err != nil {
			rcode = http.StatusNotFound
			rtype = "text/plain"
			rtext = fmt.Sprintf("error loading %s: %s", indexPath, err)
		} else {
			rtext = string(raw)
			rtext = strings.ReplaceAll(rtext, "%%{color}", srv.bgColor)
			rtext = strings.ReplaceAll(rtext, "%%{hide_key}", fmt.Sprintf("%v", srv.hideKey))
			rtext = strings.ReplaceAll(rtext, "%%{show_pods}", fmt.Sprintf("%v", srv.showPods))
			rtext = strings.ReplaceAll(rtext, "%%{user}", user)
			rtext = strings.ReplaceAll(rtext, "%%{user_header}", fmt.Sprintf("%v", srv.userHeaderName))
			rtext = strings.ReplaceAll(rtext, "%%{user_agent}", userAgent)
		}
	} else if strings.HasPrefix(r.URL.Path, "/face/") {
		// /face/ is a special case: we forward it to the face workload. This is
		// here because it allows running the demo without an ingress controller.
		// (Obviously, this is _NOT_ a good idea outside of demos!)
		reqStart := time.Now()

		url := fmt.Sprintf("http://face/%s", r.URL.Path[6:])

		rq := r.URL.RawQuery

		if rq != "" {
			url = fmt.Sprintf("%s?%s", url, rq)
		}

		user := r.Header.Get(srv.userHeaderName)
		if user == "" {
			user = "unknown"
		}
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "unknown"
		}

		if srv.debugEnabled {
			fmt.Printf("...%s: starting\n", url)
		}

		response, err := http.Get(url)
		if err != nil {
			rcode = http.StatusInternalServerError
			rtext = err.Error()
			rtype = "text/plain"
		} else {
			rcode = response.StatusCode
			rtextBytes, err := io.ReadAll(response.Body)
			if err != nil {
				rcode = http.StatusInternalServerError
				rtext = err.Error()
				rtype = "text/plain"
			} else {
				rtext = string(rtextBytes)
				rtype = response.Header.Get("Content-Type")

				podID = response.Header.Get("X-Faces-Pod")
				if podID == "" {
					podID = srv.hostIP
				}
			}
			response.Body.Close()
		}

		reqEnd := time.Now()
		reqLatencyMs := reqEnd.Sub(reqStart).Milliseconds()

		if srv.debugEnabled {
			fmt.Printf("...%s (%dms): %d\n", url, reqLatencyMs, rcode)
		}
	} else if strings.HasSuffix(r.URL.Path, ".svg") {
		// We'll serve SVG files from the dataPath, though really we don't
		// currently use this.
		filePath := filepath.Join(srv.dataPath, r.URL.Path[1:])
		file, err := os.Open(filePath)
		if err != nil {
			rtext = fmt.Sprintf("%s not found??", r.URL.Path)
			rtype = "text/plain"
			rcode = http.StatusNotFound
		} else {
			defer file.Close()

			fileInfo, _ := file.Stat()
			rdata := make([]byte, fileInfo.Size())
			_, err = file.Read(rdata)
			if err != nil {
				rtext = fmt.Sprintf("Exception: %s", err)
				rtype = "text/plain"
				rcode = http.StatusInternalServerError
			} else {
				rcode = http.StatusOK
				rtype = "image/svg+xml"
				w.Write(rdata)
				return
			}
		}
	}

	end := time.Now()
	latencyMs := end.Sub(start).Milliseconds()

	w.Header().Set("Content-Type", rtype)
	w.Header().Set(srv.userHeaderName, user)
	w.Header().Set("X-Faces-User-Agent", userAgent)
	w.Header().Set("X-Faces-Latency", strconv.FormatInt(latencyMs, 10))
	w.Header().Set("X-Faces-Pod", podID)
	w.WriteHeader(rcode)

	if rtext != "" {
		fmt.Fprint(w, rtext)
	}
}
