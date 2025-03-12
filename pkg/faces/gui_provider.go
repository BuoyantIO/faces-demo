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
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

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

type GUIProvider struct {
	BaseProvider
	dataPath string
	bgColor  string
	hideKey  bool
	showPods bool
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
	gprv.bgColor = utils.StringFromEnv("COLOR", "white")
	gprv.hideKey = utils.BoolFromEnv("HIDE_KEY", false)
	gprv.showPods = utils.BoolFromEnv("SHOW_PODS", false)

	gprv.Infof("dataPath %s", gprv.dataPath)
	gprv.Infof("bgColor %s", gprv.bgColor)
	gprv.Infof("hideKey %v", gprv.hideKey)
	gprv.Infof("showPods %v", gprv.showPods)

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
	} else if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		// "/" means "index.html", which we expect to find in the dataPath.
		key = "index"
		rcode = http.StatusOK

		indexPath := filepath.Join(gprv.dataPath, "index.html")
		gprv.Debugf("...loading %s", indexPath)

		raw, err := os.ReadFile(indexPath)

		if err != nil {
			rcode = http.StatusNotFound
			rtype = "text/plain"
			rtext = fmt.Sprintf("error loading %s: %s", indexPath, err)
		} else {
			rtext = string(raw)
			rtext = strings.ReplaceAll(rtext, "%%{color}", gprv.bgColor)
			rtext = strings.ReplaceAll(rtext, "%%{hide_key}", fmt.Sprintf("%v", gprv.hideKey))
			rtext = strings.ReplaceAll(rtext, "%%{show_pods}", fmt.Sprintf("%v", gprv.showPods))
			rtext = strings.ReplaceAll(rtext, "%%{user}", user)
			rtext = strings.ReplaceAll(rtext, "%%{user_header}", fmt.Sprintf("%v", gprv.userHeaderName))
			rtext = strings.ReplaceAll(rtext, "%%{user_agent}", userAgent)
		}
	} else if strings.HasPrefix(r.URL.Path, "/face/") {
		// /face/ is a special case: we forward it to the face workload. This is
		// here because it allows running the demo without an ingress controller.
		// (Obviously, this is _NOT_ a good idea outside of demos!)
		key = "face"
		reqStart := time.Now()

		url := fmt.Sprintf("http://face/%s", r.URL.Path[6:])

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
