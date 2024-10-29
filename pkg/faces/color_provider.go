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
	"log/slog"
	"net/http"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type ColorProvider struct {
	BaseProvider
	color string
}

func NewColorProviderFromEnvironment() *ColorProvider {
	cprv := &ColorProvider{}
	cprv.SetLogger(slog.Default().With(
		"provider", "ColorProvider",
	))

	cprv.BaseProvider.SetupFromEnvironment()

	colorName := utils.StringFromEnv("COLOR", "blue")
	cprv.color = Colors.Lookup(colorName)

	cprv.Infof("Using %s => %s\n", colorName, cprv.color)
	return cprv
}

func (cprv *ColorProvider) Get(row, col int) ProviderResponse {
	start := time.Now()

	resp := ProviderResponse{
		StatusCode: http.StatusOK,
	}

	cprv.CheckUnlatch(start)
	rstat := cprv.CheckRequestStatus()

	cprv.DelayIfNeeded()

	if rstat.IsRateLimited() {
		cprv.Warnf("RATELIMIT(%d, %d) => %s\n", row, col, rstat.Message())

		resp.StatusCode = http.StatusTooManyRequests
		resp.Body = rstat.Message()
	} else if rstat.IsErrored() {
		msg := rstat.Message()

		if msg == "" {
			msg = fmt.Sprintf("Color error! (error fraction %d%%)", cprv.errorFraction)
		}

		cprv.Warnf("ERROR(%d, %d) => %d, %s\n", row, col, rstat.StatusCode(), msg)

		resp.StatusCode = rstat.StatusCode()
		resp.Body = msg
	} else {
		cprv.Infof("OK(%d, %d) => %d, %s\n", row, col, rstat.StatusCode(), cprv.color)

		resp.StatusCode = rstat.StatusCode()
		resp.Body = cprv.color
	}

	return resp
}
