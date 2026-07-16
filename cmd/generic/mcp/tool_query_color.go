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

	"github.com/BuoyantIO/faces-demo/v2/pkg/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QueryColorTool implements the Tool interface for querying the Faces
// color service.
type QueryColorTool struct{}

func (t *QueryColorTool) Name() string {
	return "query_color"
}

func (t *QueryColorTool) Description() string {
	return "Query the color service to get the current color via gRPC"
}

func (t *QueryColorTool) InputSchema() map[string]interface{} {
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

func (t *QueryColorTool) Handle(ctx context.Context, s *MCPServer, args map[string]interface{}) (CallToolResponse, error) {
	subrequest := args["subrequest"].(string)
	row := int32(0)
	col := int32(0)

	if r, ok := args["row"].(float64); ok {
		row = int32(r)
	}
	if c, ok := args["column"].(float64); ok {
		col = int32(c)
	}

	conn, err := grpc.NewClient(s.colorURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return CallToolResponse{}, err
	}
	defer conn.Close()

	client := color.NewColorServiceClient(conn)

	req := &color.ColorRequest{
		Row:    row,
		Column: col,
	}

	var resp *color.ColorResponse
	if subrequest == "center" {
		resp, err = client.Center(ctx, req)
	} else {
		resp, err = client.Edge(ctx, req)
	}

	if err != nil {
		return CallToolResponse{}, err
	}

	return CallToolResponse{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Color service response: color=%s, rate=%s", resp.Color, resp.Rate),
			},
		},
	}, nil
}
