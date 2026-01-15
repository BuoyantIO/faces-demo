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
	"log/slog"
	"net/http"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

type SmileyProvider struct {
	BaseProvider
	smilies map[string]string
}

func NewSmileyProviderFromEnvironment() *SmileyProvider {
	sprv := &SmileyProvider{
		BaseProvider: BaseProvider{
			Name: "Smiley",
		},
		smilies: make(map[string]string),
	}

	sprv.SetLogger(slog.Default().With(
		"provider", "SmileyProvider",
	))

	sprv.SetGetHandler(sprv.Get)

	sprv.BaseProvider.SetupFromEnvironment()

	// Set the initial smilies by hand: we explicitly want to use the
	// fallback smiley if anything goes wrong here.
	smileyName := utils.StringFromEnv("SMILEY", "Grinning")
	smiley, _ := utils.Smileys.Lookup(smileyName)

	sprv.Infof("Starting with smiley %s => %s", smileyName, smiley)

	sprv.smilies["center"] = smiley
	sprv.smilies["edge"] = smiley

	// This isn't really ideal.
	sprv.Key = smileyName

	// Set up PUT handler for emoji updates
	sprv.BaseProvider.SetHTTPPutHandler(sprv.HandlePutRequest)

	return sprv
}

func (sprv *SmileyProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	// Error fraction, latching, and rate limiting are all handled by the base
	// provider

	resp := ProviderResponseEmpty()
	resp.Add("smiley", sprv.GetSmiley(prvReq.subrequest))

	return resp
}

func (sprv *SmileyProvider) GetSmiley(which string) string {
	sprv.Lock()
	defer sprv.Unlock()

	smiley, found := sprv.smilies[which]

	if !found {
		sprv.Warnf("Unknown smiley key '%s', returning center smiley", which)
		smiley = sprv.smilies["center"]
	}

	return smiley
}

func (sprv *SmileyProvider) SetSmiley(which string, smiley string) error {
	if which != "all" && which != "center" && which != "edge" {
		return fmt.Errorf("unknown smiley key '%s'", which)
	}

	if smiley == "" {
		return fmt.Errorf("smiley cannot be empty")
	}

	sprv.Lock()
	defer sprv.Unlock()

	newSmiley, found := utils.Smileys.Lookup(smiley)

	if !found {
		sprv.Warnf("Unknown %s smiley %s, not changing smiley", which, smiley)
		return fmt.Errorf("unknown smiley '%s'", smiley)
	}

	if which == "all" {
		sprv.smilies["center"] = newSmiley
		sprv.smilies["edge"] = newSmiley
	} else {
		sprv.smilies[which] = newSmiley
	}

	sprv.Infof("Set smiley '%s' to %s => %s", which, smiley, newSmiley)

	return nil
}

// HandlePutRequest processes HTTP PUT requests to update the smiley emoji.
func (sprv *SmileyProvider) HandlePutRequest(w http.ResponseWriter, r *http.Request) {
	// Grab the new smiley from the request body...
	var updateData struct {
		Which  string `json:"which"`
		Smiley string `json:"smiley"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// ...and update the smiley accordingly.
	err = sprv.SetSmiley(updateData.Which, updateData.Smiley)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to set smiley: %v", err), http.StatusBadRequest)
		return
	}

	// Finally, return a success response.
	resp := ProviderResponseEmpty()
	resp.Add("which", updateData.Which)
	resp.Add("smiley", sprv.GetSmiley(updateData.Which))
	resp.Add("message", "Smiley updated successfully")

	// I don't think this can really fail, but handle it just in case.
	respJSON, err := json.Marshal(resp.Data)

	if err != nil {
		sprv.Warnf("Failed to marshal update response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to marshal update response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respJSON)
}
