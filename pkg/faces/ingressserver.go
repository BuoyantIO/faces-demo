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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type IngressServer struct {
	BaseServer
	faceService string
}

type IngressResponse struct {
	Smiley  string   `json:"smiley"`
	Color   string   `json:"color"`
	Rate    string   `json:"rate"`
	Errors  []string `json:"errors"`
	Latency int64    `json:"latency"`
}

func NewIngressServer(serverName string) *IngressServer {
	srv := &IngressServer{
		BaseServer: BaseServer{
			Name: serverName,
		},
	}

	srv.SetupFromEnvironment()

	srv.RegisterNormal("/center/", srv.ingressGetHandler)
	srv.RegisterNormal("/edge/", srv.ingressGetHandler)

	return srv
}

func (srv *IngressServer) SetupFromEnvironment() {
	srv.BaseServer.SetupFromEnvironment()

	srv.faceService = utils.StringFromEnv("CELL_SERVICE", "cell")

	fmt.Printf("%s %s: faceService %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.faceService)
}

func (srv *IngressServer) ingressGetHandler(r *http.Request, rstat *BaseRequestStatus) *BaseServerResponse {
	start := time.Now()

	response := BaseServerResponse{
		StatusCode: http.StatusOK,
	}

	errors := []string{}

	smiley, _ := Smileys.Lookup(Defaults["smiley"])
	color := Colors.Lookup(Defaults["color"])
	rateStr := fmt.Sprintf("%.1f RPS", srv.CurrentRate())

	if rstat.IsRateLimited() {
		errors = append(errors, rstat.Message())
		smiley, _ = Smileys.Lookup(Defaults["smiley-ratelimit"])
		color = Colors.Lookup(Defaults["color-ratelimit"])
	} else {
		url := fmt.Sprintf("http://%s%s", srv.faceService, r.URL.Path)

		if srv.debugEnabled {
			fmt.Printf("%s %s: %s starting\n", time.Now().Format(time.RFC3339), srv.Name, url)
		}

		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to create request: %v", err))
			// No need to change smiley and color here.
		} else {
			// Copy headers from the original request
			req.Header.Add(srv.userHeaderName, r.Header.Get(srv.userHeaderName))
			req.Header.Add("User-Agent", r.Header.Get("User-Agent"))

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				errors = append(errors, fmt.Sprintf("request failed: %v", err))
				// No need to change smiley and color here.
			} else {
				defer resp.Body.Close()

				rcode := resp.StatusCode

				if srv.debugEnabled {
					fmt.Printf("%s %s: %s returned %d\n", time.Now().Format(time.RFC3339), srv.Name, url, rcode)
				}

				response.StatusCode = rcode

				body, err := io.ReadAll(resp.Body)

				if err != nil {
					errors = append(errors, fmt.Sprintf("failed to read response body: %v", err))
					// No need to change smiley and color here.
				} else {
					var ingressResp IngressResponse
					err := json.Unmarshal(body, &ingressResp)

					if err != nil {
						errors = append(errors, fmt.Sprintf("failed to unmarshal response: %v", err))
						// No need to change smiley and color here.
					} else {
						smiley = ingressResp.Smiley
						color = ingressResp.Color
						rateStr = ingressResp.Rate
						errors = ingressResp.Errors
					}
				}
			}
		}
	}

	end := time.Now()
	latency := end.Sub(start)

	response.Data = map[string]interface{}{
		"smiley":  smiley,
		"color":   color,
		"rate":    rateStr,
		"errors":  errors,
		"latency": latency.Milliseconds(),
	}

	if srv.debugEnabled {
		fmt.Printf("%s %s: %s, %s (%dms): %v\n", time.Now().Format(time.RFC3339), srv.Name, smiley, color, latency.Milliseconds(), errors)
	}

	return &response
}
