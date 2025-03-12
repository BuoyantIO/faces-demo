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
	"log/slog"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type SmileyProvider struct {
	BaseProvider
	smiley string
}

func NewSmileyProviderFromEnvironment() *SmileyProvider {
	sprv := &SmileyProvider{
		BaseProvider: BaseProvider{
			Name: "Smiley",
		},
	}

	sprv.SetLogger(slog.Default().With(
		"provider", "SmileyProvider",
	))

	sprv.SetGetHandler(sprv.Get)

	sprv.BaseProvider.SetupFromEnvironment()

	smileyName := utils.StringFromEnv("SMILEY", "Grinning")
	sprv.Key = smileyName
	sprv.smiley = utils.Smileys.Lookup(smileyName)

	smileyNameUsed := utils.Smileys.LookupValue(sprv.smiley)

	sprv.Infof("Using smiley %s", smileyNameUsed)
	return sprv
}

func (sprv *SmileyProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	resp := ProviderResponseEmpty()
	resp.Add("smiley", sprv.smiley)

	return resp
}
