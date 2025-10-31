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

package faces

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type GUIProvider struct {
	BaseProvider
	dataPath    string
	faceService string
	bgColor     string
	hideKey     bool
	showPods    bool
	gridSize    int
	edgeSize    int
	startActive bool
}

func NewGUIProviderFromEnvironment() *GUIProvider {
	gprv := &GUIProvider{
		BaseProvider: BaseProvider{
			Name: "GUI",
			Key:  "GUI",
		},
	}

	gprv.SetLogger(slog.Default().With(
		"provider", "GUIProvider",
	))

	gprv.SetGetHandler(gprv.Get)

	gprv.BaseProvider.SetupBasicsFromEnvironment()

	gprv.dataPath = utils.StringFromEnv("DATA_PATH", "/app/data")
	gprv.faceService = utils.StringFromEnv("FACE_SERVICE", "face")
	gprv.bgColor = utils.StringFromEnv("COLOR", "white")
	gprv.hideKey = utils.BoolFromEnv("HIDE_KEY", false)
	gprv.showPods = utils.BoolFromEnv("SHOW_PODS", false)
	gprv.gridSize = utils.IntFromEnv("GRID_SIZE", 4)
	gprv.edgeSize = utils.IntFromEnv("EDGE_SIZE", 1)
	gprv.startActive = utils.BoolFromEnv("START_ACTIVE", true)

	gprv.Infof("dataPath %s", gprv.dataPath)
	gprv.Infof("bgColor %s", gprv.bgColor)
	gprv.Infof("hideKey %v", gprv.hideKey)
	gprv.Infof("showPods %v", gprv.showPods)
	gprv.Infof("gridSize %d", gprv.gridSize)
	gprv.Infof("edgeSize %d", gprv.edgeSize)
	gprv.Infof("startActive %v", gprv.startActive)

	return gprv
}

// This should never ever be called.
func (gprv *GUIProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	return ProviderResponseNotImplemented()
}

func (gprv *GUIProvider) HTTPGetHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	userAgent := r.Header.Get("user-agent")

	if userAgent == "" {
		userAgent = "unknown"
	}

	user := r.Header.Get(gprv.userHeaderName)

	if user == "" {
		user = "unknown"
	}

	podID := gprv.hostIP
	podName := gprv.hostName

	key := "unknown"
	rcode := http.StatusNotFound
	rtext := fmt.Sprintf("%s not found", r.URL.Path)
	rtype := "text/html"

	// Handle readiness checks first (they're simple).
	if r.URL.Path == "/ready" {
		key = "ready"
		rcode = http.StatusOK
		rtype = "text/plain"
		rtext = "Ready and waiting!"
	} else if (r.Method == "GET") && strings.HasPrefix(r.URL.Path, "/face/") {
		// /face/ is a special case: we forward it to the face workload. This is
		// here because it allows running the demo without an ingress controller.
		// (Obviously, this is _NOT_ a good idea outside of demos!)

		key = "face"
		reqStart := time.Now()

		url := fmt.Sprintf("http://%s/%s", gprv.faceService, r.URL.Path[6:])

		rq := r.URL.RawQuery

		if rq != "" {
			url = fmt.Sprintf("%s?%s", url, rq)
		}

		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			rcode = http.StatusInternalServerError
			rtext = fmt.Sprintf("could not create new Request? %s", err.Error())
			rtype = "text/plain"
		} else {
			for key, values := range r.Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}

			gprv.Debugf("...%s: starting", url)

			response, err := http.DefaultClient.Do(req)

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
						podID = gprv.hostIP
					}
				}
				response.Body.Close()
			}

			reqEnd := time.Now()
			reqLatencyMs := reqEnd.Sub(reqStart).Milliseconds()

			gprv.Debugf("...%s (%dms): %d", url, reqLatencyMs, rcode)
		}
	} else if r.Method == "GET" && r.URL.RawQuery == "" {
		// Try to read the file from our dataPath.
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}

		key = "static"
		interpolate := false

		if r.URL.Path == "/index.html" {
			key = "index"
			interpolate = true
		}

		filePath := filepath.Join(gprv.dataPath, strings.TrimPrefix(r.URL.Path, "/"))
		raw, err := os.ReadFile(filePath)

		if err != nil {
			gprv.Infof("%s: file not found", filePath)

			rcode = http.StatusNotFound
			rtype = "text/plain"
			rtext = fmt.Sprintf("error loading %s: %s", filePath, err)
		} else {
			gprv.Debugf("%s: loaded", filePath)

			rcode = http.StatusOK
			rtext = string(raw)

			switch filepath.Ext(filePath) {
			case ".html":
				rtype = "text/html"
			case ".css":
				rtype = "text/css"
			case ".js":
				rtype = "application/javascript"
			case ".json":
				rtype = "application/json"
			case ".png":
				rtype = "image/png"
			case ".jpg", ".jpeg":
				rtype = "image/jpeg"
			case ".gif":
				rtype = "image/gif"
			case ".svg":
				rtype = "image/svg+xml"
			case ".ico":
				rtype = "image/x-icon"
			default:
				rtype = "application/octet-stream"
			}

			if interpolate {
				gprv.Debugf("%s: interpolating", filePath)
				rtext = strings.ReplaceAll(rtext, "%%{color}", gprv.bgColor)
				rtext = strings.ReplaceAll(rtext, "%%{hide_key}", fmt.Sprintf("%v", gprv.hideKey))
				rtext = strings.ReplaceAll(rtext, "%%{show_pods}", fmt.Sprintf("%v", gprv.showPods))
				rtext = strings.ReplaceAll(rtext, "%%{grid_size}", fmt.Sprintf("%d", gprv.gridSize))
				rtext = strings.ReplaceAll(rtext, "%%{edge_size}", fmt.Sprintf("%d", gprv.edgeSize))
				rtext = strings.ReplaceAll(rtext, "%%{start_active}", fmt.Sprintf("%v", gprv.startActive))
				rtext = strings.ReplaceAll(rtext, "%%{user}", user)
				rtext = strings.ReplaceAll(rtext, "%%{user_header}", fmt.Sprintf("%v", gprv.userHeaderName))
				rtext = strings.ReplaceAll(rtext, "%%{user_agent}", userAgent)
			}
		}
	}

	end := time.Now()
	latency := end.Sub(start)

	gprv.requestsTotal.WithLabelValues(gprv.Name, gprv.hostName, key, fmt.Sprintf("%03d", rcode)).Inc()
	gprv.requestDuration.WithLabelValues(gprv.Name, gprv.hostName, key).Observe(latency.Seconds())

	gprv.Debugf("GET %s %03d (user %s, user-agent %s)", r.URL.Path, rcode, user, userAgent)

	w.Header().Set("Content-Type", rtype)
	w.Header().Set(gprv.userHeaderName, user)
	w.Header().Set("X-Faces-User-Agent", userAgent)
	w.Header().Set("X-Faces-Latency", strconv.FormatInt(latency.Milliseconds(), 10))
	w.Header().Set("X-Faces-Pod", podID)
	w.Header().Set("X-Faces-PodName", podName)
	w.WriteHeader(rcode)

	if rtext != "" {
		fmt.Fprint(w, rtext)
	}
}
