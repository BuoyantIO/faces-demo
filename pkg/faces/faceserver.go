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
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FaceServer struct {
	BaseServer
	smileyService string
	colorService  string
}

type FaceResponse struct {
	statusCode int
	latency    time.Duration
	data       string
}

func NewFaceServer(serverName string) *FaceServer {
	srv := &FaceServer{
		BaseServer: BaseServer{
			Name: serverName,
		},
	}

	srv.SetupFromEnvironment()

	srv.RegisterNormal("/", srv.faceGetHandler)

	return srv
}

func (srv *FaceServer) SetupFromEnvironment() {
	srv.BaseServer.SetupFromEnvironment()

	srv.smileyService = utils.StringFromEnv("SMILEY_SERVICE", "smiley")
	srv.colorService = utils.StringFromEnv("COLOR_SERVICE", "color")

	_, _, err := net.SplitHostPort(srv.colorService)

	if err != nil {
		// Most likely we're missing the port, so try to default it.
		addr := net.ParseIP(srv.colorService)

		if addr != nil {
			// Is this an IPv6 address?
			if strings.Contains(srv.colorService, ":") {
				srv.colorService = fmt.Sprintf("[%s]:80", srv.colorService)
			} else {
				srv.colorService = fmt.Sprintf("%s:80", srv.colorService)
			}
		} else {
			// Probably a hostname.
			srv.colorService = fmt.Sprintf("%s:80", srv.colorService)
		}
	}

	fmt.Printf("%s %s: smileyService %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.smileyService)
	fmt.Printf("%s %s: colorService %v\n", time.Now().Format(time.RFC3339), srv.Name, srv.colorService)
}

func (srv *FaceServer) makeRequest(user string, userAgent string, service string, keyword string, subrequest string, row int, col int) *FaceResponse {
	start := time.Now()

	url := fmt.Sprintf("http://%s/%s/?row=%d&col=%d", service, subrequest, row, col)

	if srv.debugEnabled {
		fmt.Printf("%s %s: %s starting\n", time.Now().Format(time.RFC3339), srv.Name, url)
	}

	failed := false
	rcode := http.StatusOK
	rtext := ""
	var response *http.Response
	var ok bool

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		failed = true
		rcode = http.StatusInternalServerError
		rtext = fmt.Sprintf("couldn't create request to %s: %s", service, err)
	}

	if !failed {
		req.Header.Set(srv.userHeaderName, user)
		req.Header.Set("User-Agent", userAgent)

		response, err = http.DefaultClient.Do(req)

		if err != nil {
			failed = true
			rcode = http.StatusInternalServerError
			rtext = fmt.Sprintf("couldn't make request to %s: %s", service, err)
		}
	}

	if !failed {
		defer response.Body.Close()

		rcode = response.StatusCode
		body, _ := io.ReadAll(response.Body)

		if srv.debugEnabled {
			fmt.Printf("%s %s: %s status %d\n", time.Now().Format(time.RFC3339), srv.Name, url, rcode)
		}

		if rcode != http.StatusOK {
			failed = true

			bstr := ""

			if len(body) > 0 {
				bstr = fmt.Sprintf(" (%s)", string(body))
			}

			rtext = fmt.Sprintf("error from %s: %03d%s", service, rcode, bstr)
		}

		if !failed {
			// Decode the response body as JSON into a map[string]interface{} called data.
			var data map[string]interface{}
			err := json.Unmarshal(body, &data)

			if err != nil {
				failed = true
				rtext = fmt.Sprintf("couldn't decode response from %s: %s", service, err)
			}

			if !failed {
				rtext, ok = data[keyword].(string)

				if !ok {
					failed = true
					rtext = fmt.Sprintf("no %s in response from %s", keyword, service)
				}
			}
		}
	}

	end := time.Now()
	latency := end.Sub(start)

	if srv.debugEnabled {
		fmt.Printf("%s %s: %s done (%d, %dms -- %s)\n", time.Now().Format(time.RFC3339), srv.Name, url, rcode, latency.Milliseconds(), rtext)
	}

	return &FaceResponse{
		statusCode: rcode,
		latency:    latency,
		data:       rtext,
	}
}

func mapStatus(name string, statusCode int) string {
	keys := []string{
		fmt.Sprintf("%s-%03d", name, statusCode),
		fmt.Sprintf("%s-%dxx", name, statusCode/100),
		fmt.Sprintf("%s-error", name),
	}

	for _, key := range keys {
		if val, ok := Defaults[key]; ok {
			return val
		}
	}

	return Defaults[name]
}

func (srv *FaceServer) faceGetHandler(r *http.Request, rstat *BaseRequestStatus) *BaseServerResponse {
	start := time.Now()

	response := BaseServerResponse{
		StatusCode: http.StatusOK,
	}

	fmt.Printf("%s %s: request path: %s, query string: %s\n", time.Now().Format(time.RFC3339), srv.Name, r.URL.Path, r.URL.RawQuery)

	// Our request URL should start with /center/ or /edge/, and we want to
	// propagate that to our smiley and color services.
	subrequest := strings.Split(r.URL.Path, "/")[1]

	errors := []string{}
	var smiley string
	var color string
	var smileyOK bool

	rateStr := fmt.Sprintf("%.1f RPS", srv.CurrentRate())

	query := r.URL.Query()
	query_row := query.Get("row")
	query_column := query.Get("col")

	row := -1
	column := -1

	if query_row != "" {
		r, err := strconv.Atoi(query_row)

		if err == nil {
			row = r
		} else {
			fmt.Printf("%s %s: couldn't parse row %s, using -1: %s\n", time.Now().Format(time.RFC3339), srv.Name, query_row, err)
		}
	}

	if query_column != "" {
		c, err := strconv.Atoi(query_column)

		if err == nil {
			column = c
		} else {
			fmt.Printf("%s %s: couldn't parse column %s, using -1: %s\n", time.Now().Format(time.RFC3339), srv.Name, query_column, err)
		}
	}

	if rstat.IsRateLimited() {
		errors = append(errors, rstat.Message())
		smiley, smileyOK = Smileys.Lookup(Defaults["smiley-ratelimit"])
		color = Colors.Lookup(Defaults["color-ratelimit"])
	} else {
		user := r.Header.Get(srv.userHeaderName)

		if user == "" {
			user = "unknown"
		}

		userAgent := r.Header.Get("user-agent")

		if userAgent == "" {
			userAgent = "unknown"
		}

		// Make HTTP Get requests to the smiley service and the color service in parallel using goroutines
		smileyCh := make(chan *FaceResponse)
		colorCh := make(chan *FaceResponse)

		go func() {
			smileyCh <- srv.makeRequest(user, userAgent, srv.smileyService, "smiley", subrequest, row, column)
		}()

		go func() {
			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}

			conn, err := grpc.NewClient(srv.colorService, opts...)

			if err != nil {
				errors = append(errors, fmt.Sprintf("color: %s", err))
				colorCh <- &FaceResponse{
					statusCode: http.StatusInternalServerError,
					data:       fmt.Sprintf("couldn't connect to %s: %s", srv.colorService, err),
				}
				return
			}

			defer conn.Close()

			client := NewColorServiceClient(conn)
			colorReq := &ColorRequest{
				Row:    int32(row),
				Column: int32(column),
			}

			var colorResp *ColorResponse

			if srv.debugEnabled {
				fmt.Printf("%s %s: starting gRPC to %s for %s (row %d col %d)\n",
					time.Now().Format(time.RFC3339), srv.Name, srv.colorService, subrequest, colorReq.Row, colorReq.Column)
			}

			if subrequest == "center" {
				colorResp, err = client.Center(context.Background(), colorReq)
			} else {
				colorResp, err = client.Edge(context.Background(), colorReq)
			}

			if err != nil {
				if srv.debugEnabled {
					fmt.Printf("%s %s: gRPC failed: %s\n", time.Now().Format(time.RFC3339), srv.Name, err)
				}

				errors = append(errors, fmt.Sprintf("color: %s", err))
				colorCh <- &FaceResponse{
					statusCode: http.StatusInternalServerError,
					data:       fmt.Sprintf("couldn't get color from %s: %s", srv.colorService, err),
				}
			} else {
				if srv.debugEnabled {
					fmt.Printf("%s %s: gRPC got %#v\n", time.Now().Format(time.RFC3339), srv.Name, colorResp)
				}

				colorCh <- &FaceResponse{
					statusCode: http.StatusOK,
					data:       colorResp.Color,
				}
			}
		}()

		// Wait for the responses from both services
		smileyResp := <-smileyCh

		if smileyResp.statusCode != http.StatusOK {
			errors = append(errors, fmt.Sprintf("smiley: %s", smileyResp.data))
			mapped := mapStatus("smiley", smileyResp.statusCode)
			smiley, smileyOK = Smileys.Lookup(mapped)

			if srv.debugEnabled {
				fmt.Printf("%s %s: mapped smiley %d to %s (%s, %v)\n",
					time.Now().Format(time.RFC3339), srv.Name, smileyResp.statusCode, mapped, smiley, smileyOK)
			}
		} else {
			smiley = smileyResp.data
			smileyOK = true
		}

		colorResp := <-colorCh

		if colorResp.statusCode != http.StatusOK {
			errors = append(errors, fmt.Sprintf("color: %s", colorResp.data))
			color = Colors.Lookup(mapStatus("color", colorResp.statusCode))
		} else {
			color = colorResp.data
		}
	}

	if !smileyOK {
		// Something bizarre happened with the smiley lookup?
		smiley, _ = Smileys.Lookup("Vomiting")
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
