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
	"log/slog"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type ColorProvider struct {
	BaseProvider
	color string
}

func NewColorProviderFromEnvironment() *ColorProvider {
	cprv := &ColorProvider{
		BaseProvider: BaseProvider{
			Name: "Color",
		},
	}

	cprv.SetLogger(slog.Default().With(
		"provider", "ColorProvider",
	))

	cprv.SetGetHandler(cprv.Get)

	cprv.BaseProvider.SetupFromEnvironment()

	colorName := utils.StringFromEnv("COLOR", "blue")
	cprv.Key = colorName
	cprv.color = utils.Colors.Lookup(colorName)

	cprv.Infof("Using color %s => %s", colorName, cprv.color)
	return cprv
}

func (cprv *ColorProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	resp := ProviderResponseEmpty()
	resp.Add("color", cprv.color)
	return resp
}
