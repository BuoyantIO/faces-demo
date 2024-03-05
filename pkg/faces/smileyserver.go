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
	"net/http"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type SmileyServer struct {
	BaseServer
	smiley string
}

func NewSmileyServer(serverName string) *SmileyServer {
	srv := &SmileyServer{
		BaseServer: BaseServer{
			Name: serverName,
		},
	}

	srv.SetupFromEnvironment()
	// srv.SetUpdater(srv.updater)

	srv.RegisterNormal("/", srv.smileyGetHandler)

	return srv
}

func (srv *SmileyServer) SetupFromEnvironment() {
	srv.BaseServer.SetupFromEnvironment()

	smileyKey := utils.StringFromEnv("SMILEY", "Smiling")

	smiley, ok := Smileys[smileyKey]

	if !ok {
		smileyKey = "Neutral"
		smiley = Smileys[smileyKey]
	}

	srv.smiley = smiley

	fmt.Printf("%s %s: smiley %s\n", time.Now().Format(time.RFC3339), srv.Name, smileyKey)
}

func (srv *SmileyServer) smileyGetHandler(r *http.Request, rstat *BaseRequestStatus) *BaseServerResponse {
	// The only error we need to handle here is the internal rate limiter.
	if rstat.ratelimited {
		errstr := fmt.Sprintf("Rate limited (%.1f RPS > max %.1f RPS)", srv.CurrentRate(), srv.maxRate)
		return &BaseServerResponse{
			StatusCode: http.StatusTooManyRequests,
			Data: map[string]interface{}{
				"smiley": Defaults["smiley-ratelimit"],
				"rate":   fmt.Sprintf("%.1f RPS", srv.CurrentRate()),
				"errors": []string{errstr},
			},
		}
	}

	// Errors have already been handled, so this is always just a simple
	// success response.
	return &BaseServerResponse{
		StatusCode: http.StatusOK,
		Data: map[string]interface{}{
			"smiley": srv.smiley,
			"rate":   fmt.Sprintf("%.1f RPS", srv.CurrentRate()),
			"errors": []string{},
		},
	}
}
