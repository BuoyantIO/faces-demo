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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BaseHTTPServer struct {
	provider *BaseProvider
	mux      *http.ServeMux
}

func NewBaseHTTPServer(provider *BaseProvider) *BaseHTTPServer {
	bsrv := &BaseHTTPServer{provider: provider}

	bsrv.mux = http.NewServeMux()
	bsrv.mux.HandleFunc("/", bsrv.handleRequest)

	provider.SetHTTPGetHandler(bsrv.defaultGetHandler)

	return bsrv
}

func (bsrv *BaseHTTPServer) Start(addr string) error {
	bsrv.provider.Infof("Starting server on %s", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: bsrv.mux,
	}

	return httpServer.ListenAndServe()
}

func (bsrv *BaseHTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodHead {
		bsrv.StandardResponse(w, r, ProviderResponseEmpty())
	} else if r.Method == http.MethodGet {
		bsrv.provider.httpGetHandler(w, r)
	} else {
		bsrv.StandardResponse(w, r, ProviderResponseMethodNotAllowed(r.Method))
	}
}

func (bsrv *BaseHTTPServer) defaultGetHandler(w http.ResponseWriter, r *http.Request) {
	prv := bsrv.provider
	start := time.Now()

	// Our request URL should start with /center/ or /edge/, and we want to
	// propagate that to our smiley and color services.
	subrequest := strings.Split(r.URL.Path, "/")[1]

	userAgent := r.Header.Get("user-agent")

	if userAgent == "" {
		userAgent = "unknown"
	}

	// Parse the query
	query := r.URL.Query()
	query_row := query.Get("row")
	query_col := query.Get("col")

	row := -1
	col := -1

	if query_row != "" {
		r, err := strconv.Atoi(query_row)

		if err == nil {
			row = r
		} else {
			prv.Warnf("couldn't parse row '%s', using -1: %s\n", query_row, err)
		}
	}

	if query_col != "" {
		c, err := strconv.Atoi(query_col)

		if err == nil {
			col = c
		} else {
			prv.Warnf("couldn't parse column '%s', using -1: %s\n", query_col, err)
		}
	}

	user := r.Header.Get(prv.userHeaderName)

	if user == "" {
		user = "unknown"
	}

	prvReq := &ProviderRequest{
		subrequest: subrequest,
		user:       user,
		userAgent:  userAgent,
		row:        row,
		col:        col,
	}

	resp := prv.HandleRequest(start, prvReq)

	bsrv.StandardResponse(w, r, resp)
}

func (bsrv *BaseHTTPServer) StandardError(w http.ResponseWriter, r *http.Request, statusCode int, responseBody string) {
	bsrv.standardHeaders(w, r, statusCode, "text/plain")
	w.Write([]byte(responseBody))
}

func (bsrv *BaseHTTPServer) StandardResponse(w http.ResponseWriter, r *http.Request, response ProviderResponse) {
	rdict := map[string]interface{}{
		"path":           r.URL.Path,
		"client_address": r.RemoteAddr,
		"method":         r.Method,
		"headers":        r.Header,
		"status":         response.StatusCode,
	}

	if response.Data != nil {
		for key, value := range response.Data {
			rdict[key] = value
		}
	}

	responseBodyBytes, err := json.Marshal(rdict)
	responseType := "application/json"

	if err != nil {
		// This should be "impossible".
		responseBody := fmt.Sprintf("Error marshalling response: %s", err)
		responseBodyBytes = []byte(responseBody)
		responseType = "text/plain"
	}

	bsrv.standardHeaders(w, r, response.StatusCode, responseType)
	w.Write([]byte(responseBodyBytes))
}

func (bsrv *BaseHTTPServer) standardHeaders(w http.ResponseWriter, r *http.Request, statusCode int, contentType string) {
	// bsrv.provider.Debugf("%s: %s %s %d\n", bsrv.provider.Name, r.Method, r.URL.Path, statusCode)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set(bsrv.provider.userHeaderName, r.Header.Get(bsrv.provider.userHeaderName))
	w.Header().Set("User-Agent", r.Header.Get("User-Agent"))
	w.Header().Set("X-Faces-Pod", bsrv.provider.hostIP)
	w.WriteHeader(statusCode)
}
