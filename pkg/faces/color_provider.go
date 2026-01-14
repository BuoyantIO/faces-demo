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
	"os"

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

	err := cprv.SetColor(colorName)

	if err != nil {
		cprv.Warnf("Failed to set initial color: %v", err)
		err = cprv.SetColor("yellow")
	}

	// Wow this is unsatisfying.
	if err != nil {
		cprv.Warnf("Failed to set fallback color: %v", err)
		os.Exit(1)
	}

	return cprv
}

func (cprv *ColorProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	resp := ProviderResponseEmpty()
	resp.Add("color", cprv.GetColor())
	return resp
}

func (cprv *ColorProvider) GetColor() string {
	cprv.Lock()
	defer cprv.Unlock()

	return cprv.color
}

func (cprv *ColorProvider) SetColor(color string) error {
	if color == "" {
		return fmt.Errorf("color cannot be empty")
	}

	cprv.Lock()
	defer cprv.Unlock()

	newColor, found := utils.Colors.Lookup(color)

	if !found {
		cprv.Warnf("Unknown color '%s', not changing color", color)
		return fmt.Errorf("unknown color '%s'", color)
	}

	cprv.color = newColor
	cprv.Infof("Using color %s => %s", color, cprv.color)

	return nil
}
