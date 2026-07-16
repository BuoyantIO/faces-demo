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
	"fmt"
	"io"
	"net/http"
)

// QueryFaceTool implements the Tool interface for querying the Faces
// face service.
type QueryFaceTool struct{}

func (t *QueryFaceTool) Name() string {
	return "query_face"
}

func (t *QueryFaceTool) Description() string {
	return "Query the face service to get a combined face response (smiley + color)"
}

func (t *QueryFaceTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"subrequest": map[string]interface{}{
				"type":        "string",
				"description": "The subrequest type (center or edge)",
				"enum":        []string{"center", "edge"},
			},
			"row": map[string]interface{}{
				"type":        "integer",
				"description": "Row number for the request",
			},
			"column": map[string]interface{}{
				"type":        "integer",
				"description": "Column number for the request",
			},
		},
		"required": []string{"subrequest"},
	}
}

func (t *QueryFaceTool) Handle(ctx context.Context, s *MCPServer, args map[string]interface{}) (CallToolResponse, error) {
	subrequest := args["subrequest"].(string)
	row := 0
	col := 0

	if r, ok := args["row"].(float64); ok {
		row = int(r)
	}
	if c, ok := args["column"].(float64); ok {
		col = int(c)
	}

	url := fmt.Sprintf("%s/%s/?row=%d&col=%d", s.faceURL, subrequest, row, col)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return CallToolResponse{}, err
	}

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
				"text": fmt.Sprintf("Face service response: %s", string(body)),
			},
		},
	}, nil
}
