// SPDX-FileCopyrightText: 2026 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Buoyant Inc.
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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// UpdateSmileyTool implements the Tool interface for updating the smiley
// returned by the Faces smiley service.
type UpdateSmileyTool struct{}

func (t *UpdateSmileyTool) Name() string {
	return "update_smiley"
}

func (t *UpdateSmileyTool) Description() string {
	return "Update the smiley for the smiley service via HTTP PUT"
}

func (t *UpdateSmileyTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"smiley": map[string]interface{}{
				"type":        "string",
				"description": "HTML entity or Unicode for smiley (e.g., &#x1F603; or U+1F603)",
			},
			"which": map[string]interface{}{
				"type":        "string",
				"description": "Which smiley to update: center, edge, or all",
				"enum":        []string{"center", "edge", "all"},
				"default":     "all",
			},
		},
		"required": []string{"smiley"},
	}
}

func (t *UpdateSmileyTool) Handle(ctx context.Context, s *MCPServer, args map[string]interface{}) (CallToolResponse, error) {
	newSmiley := args["smiley"].(string)
	which := "all" // default
	if w, ok := args["which"].(string); ok {
		which = w
	}

	updateData := map[string]string{
		"smiley": newSmiley,
		"which":  which,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return CallToolResponse{}, err
	}

	url := fmt.Sprintf("%s/", s.smileyURL)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return CallToolResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CallToolResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CallToolResponse{}, err
	}

	return CallToolResponse{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Smiley updated successfully (%s): %s", which, string(body)),
			},
		},
	}, nil
}
