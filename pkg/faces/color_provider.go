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
	"log/slog"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type ColorProvider struct {
	BaseProvider
	colors map[string]string
}

func NewColorProviderFromEnvironment() *ColorProvider {
	cprv := &ColorProvider{
		BaseProvider: BaseProvider{
			Name: "Color",
		},
		colors: make(map[string]string),
	}

	cprv.SetLogger(slog.Default().With(
		"provider", "ColorProvider",
	))

	cprv.SetGetHandler(cprv.Get)

	cprv.BaseProvider.SetupFromEnvironment()

	// Set the initial colors by hand: we explicitly want to use the
	// fallback color if anything goes wrong here.
	colorName := utils.StringFromEnv("COLOR", "blue")
	color, _ := utils.Colors.Lookup(colorName)

	cprv.Infof("Starting with color %s => %s", colorName, color)

	cprv.colors["center"] = color
	cprv.colors["edge"] = color

	// This isn't really ideal.
	cprv.Key = colorName

	return cprv
}

func (cprv *ColorProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	resp := ProviderResponseEmpty()
	resp.Add("color", cprv.GetColor(prvReq.subrequest))
	return resp
}

func (cprv *ColorProvider) GetColor(which string) string {
	cprv.Lock()
	defer cprv.Unlock()

	color, found := cprv.colors[which]

	if !found {
		cprv.Warnf("Unknown color key '%s', returning center color", which)
		color = cprv.colors["center"]
	}

	return color
}

func (cprv *ColorProvider) SetColor(which string, color string) (string, error) {
	if which != "all" && which != "center" && which != "edge" {
		return "", fmt.Errorf("unknown color key '%s'", which)
	}

	if color == "" {
		return "", fmt.Errorf("color cannot be empty")
	}

	// It's safe to do the lookup without holding the lock, since
	// utils.Colors is immutable.
	newColor, found := utils.Colors.Lookup(color)

	if !found {
		cprv.Warnf("Unknown %s color %s, not changing color", which, color)
		return "", fmt.Errorf("unknown color '%s'", color)
	}

	cprv.Lock()
	defer cprv.Unlock()

	if which == "all" {
		cprv.colors["center"] = newColor
		cprv.colors["edge"] = newColor
	} else {
		cprv.colors[which] = newColor
	}

	cprv.Infof("Set color '%s' to %s => %s", which, color, newColor)

	return newColor, nil
}
