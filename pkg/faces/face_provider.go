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
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/color"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type FaceProvider struct {
	BaseProvider
	smileyService string
	colorService  string
}

type FaceResponse struct {
	statusCode int
	latency    time.Duration
	data       string
}

func mapStatus(name string, statusCode int) string {
	keys := []string{
		fmt.Sprintf("%s-%03d", name, statusCode),
		fmt.Sprintf("%s-%dxx", name, statusCode/100),
		fmt.Sprintf("%s-error", name),
	}

	for _, key := range keys {
		if val, ok := utils.Defaults[key]; ok {
			return val
		}
	}

	return utils.Defaults[name]
}

func NewFaceProviderFromEnvironment() *FaceProvider {
	fprv := &FaceProvider{
		BaseProvider: BaseProvider{
			Name: "Face",
			Key:  "Face",
		},
	}

	fprv.SetLogger(slog.Default().With(
		"provider", "FaceProvider",
	))

	fprv.SetGetHandler(fprv.Get)

	fprv.BaseProvider.SetupFromEnvironment()

	fprv.smileyService = utils.StringFromEnv("SMILEY_SERVICE", "smiley")
	fprv.colorService = utils.StringFromEnv("COLOR_SERVICE", "color")

	_, _, err := net.SplitHostPort(fprv.colorService)

	if err != nil {
		// Most likely we're missing the port, so try to default it.
		addr := net.ParseIP(fprv.colorService)

		if addr != nil {
			// Is this an IPv6 address?
			if strings.Contains(fprv.colorService, ":") {
				fprv.colorService = fmt.Sprintf("[%s]:80", fprv.colorService)
			} else {
				fprv.colorService = fmt.Sprintf("%s:80", fprv.colorService)
			}
		} else {
			// Probably a hostname.
			fprv.colorService = fmt.Sprintf("%s:80", fprv.colorService)
		}
	}

	fprv.Infof("Face: smileyService http://%s", fprv.smileyService)
	fprv.Infof("Face: colorService grpc://%s", fprv.colorService)

	return fprv
}

func (fprv *FaceProvider) makeSmileyRequest(prvReq *ProviderRequest) *FaceResponse {
	start := time.Now()

	url := fmt.Sprintf("http://%s/%s/?row=%d&col=%d", fprv.smileyService, prvReq.subrequest, prvReq.row, prvReq.col)

	fprv.Debugf("HTTP starting (%s) %s", prvReq.InfoStr(), url)

	failed := false
	rcode := http.StatusOK
	rtext := ""
	var response *http.Response
	var ok bool

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		failed = true
		rcode = http.StatusInternalServerError
		rtext = fmt.Sprintf("couldn't create request to %s: %s", fprv.smileyService, err)
	}

	if !failed {
		req.Header.Set(fprv.userHeaderName, prvReq.user)
		req.Header.Set("User-Agent", prvReq.userAgent)

		response, err = http.DefaultClient.Do(req)

		if err != nil {
			failed = true
			rcode = http.StatusInternalServerError
			rtext = fmt.Sprintf("couldn't make request to %s: %s", fprv.smileyService, err)
		}
	}

	if !failed {
		defer response.Body.Close()

		rcode = response.StatusCode
		body, _ := io.ReadAll(response.Body)

		fprv.Debugf("HTTP %s status %d", url, rcode)

		if rcode != http.StatusOK {
			failed = true

			bstr := ""

			if len(body) > 0 {
				bstr = fmt.Sprintf(" (%s)", string(body))
			}

			rtext = fmt.Sprintf("error from %s: %03d%s", fprv.smileyService, rcode, bstr)
		}

		if !failed {
			// Decode the response body as JSON into a map[string]interface{} called data.
			var data map[string]interface{}
			err := json.Unmarshal(body, &data)

			if err != nil {
				failed = true
				rtext = fmt.Sprintf("couldn't decode response from %s: %s", fprv.smileyService, err)
			}

			if !failed {
				rtext, ok = data["smiley"].(string)

				if !ok {
					failed = true
					rtext = fmt.Sprintf("no smiley in response from %s", fprv.smileyService)
				}
			}
		}
	}

	end := time.Now()
	latency := end.Sub(start)

	fprv.Debugf("HTTP %s done (%d, %dms -- %s)", url, rcode, latency.Milliseconds(), rtext)

	return &FaceResponse{
		statusCode: rcode,
		latency:    latency,
		data:       rtext,
	}
}

func (fprv *FaceProvider) makeColorRequest(prvReq *ProviderRequest) *FaceResponse {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(fprv.colorService, opts...)

	if err != nil {
		return &FaceResponse{
			statusCode: http.StatusInternalServerError,
			data:       fmt.Sprintf("couldn't connect to %s: %s", fprv.colorService, err),
		}
	}

	defer conn.Close()

	client := color.NewColorServiceClient(conn)

	// Anything linked to this variable will transmit request headers.
	md := metadata.New(map[string]string{"x-faces-user": prvReq.user})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	colorReq := &color.ColorRequest{
		Row:    int32(prvReq.row),
		Column: int32(prvReq.col),
	}

	var colorResp *color.ColorResponse

	fprv.Debugf("gRPC starting (%s) %s", prvReq.InfoStr(), fprv.colorService)

	if prvReq.subrequest == "center" {
		colorResp, err = client.Center(ctx, colorReq)
	} else {
		colorResp, err = client.Edge(ctx, colorReq)
	}

	if err != nil {
		fprv.Debugf("gRPC (%s) failed: %s", prvReq.InfoStr(), err)

		return &FaceResponse{
			statusCode: http.StatusInternalServerError,
			data:       fmt.Sprintf("couldn't get color from %s: %s", fprv.colorService, err),
		}
	} else {
		fprv.Debugf("gRPC (%s) succeeded: %s", prvReq.InfoStr(), colorResp.Color)

		return &FaceResponse{
			statusCode: http.StatusOK,
			data:       colorResp.Color,
		}
	}
}

func (sprv *FaceProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	resp := ProviderResponseEmpty()
	var smiley string
	var color string

	// Make HTTP Get requests to the smiley service and the color service in parallel using goroutines
	smileyCh := make(chan *FaceResponse)
	colorCh := make(chan *FaceResponse)

	go func() {
		smileyCh <- sprv.makeSmileyRequest(prvReq)
	}()

	go func() {
		colorCh <- sprv.makeColorRequest(prvReq)
	}()

	// Wait for the responses from both services
	smileyResp := <-smileyCh

	if smileyResp.statusCode != http.StatusOK {
		resp.AddError(fmt.Sprintf("smiley: %s", smileyResp.data))

		smileyName := mapStatus("smiley", smileyResp.statusCode)
		smiley, _ = utils.Smileys.Lookup(smileyName)

		sprv.Debugf("(%s) smiley status %d => %s (%s)", prvReq.InfoStr(), smileyResp.statusCode, smileyName, smiley)
	} else {
		smiley = smileyResp.data
	}

	colorResp := <-colorCh

	if colorResp.statusCode != http.StatusOK {
		resp.AddError(fmt.Sprintf("color: %s", colorResp.data))

		colorName := mapStatus("color", colorResp.statusCode)
		color, _ = utils.Colors.Lookup(colorName)

		sprv.Debugf("(%s) color status %d => %s (%s)", prvReq.InfoStr(), colorResp.statusCode, colorName, color)
	} else {
		color = colorResp.data
	}

	resp.Add("smiley", smiley)
	resp.Add("color", color)

	sprv.Debugf("(%s) %v", prvReq.InfoStr(), resp.Data)

	return resp
}
