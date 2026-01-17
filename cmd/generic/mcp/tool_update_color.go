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

// UpdateColorTool implements the Tool interface for updating the color
// returned by the Faces color service.
type UpdateColorTool struct{}

func (t *UpdateColorTool) Name() string {
	return "update_color"
}

func (t *UpdateColorTool) Description() string {
	return "Update the color for the color service via gRPC"
}

func (t *UpdateColorTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"color": map[string]interface{}{
				"type":        "string",
				"description": "Color name or hex code (e.g., blue or #66CCEE)",
			},
			"which": map[string]interface{}{
				"type":        "string",
				"description": "Which color to update: center, edge, or all",
				"enum":        []string{"center", "edge", "all"},
				"default":     "all",
			},
		},
		"required": []string{"color"},
	}
}

func (t *UpdateColorTool) Handle(ctx context.Context, s *MCPServer, args map[string]interface{}) (CallToolResponse, error) {
	newColor := args["color"].(string)
	which := "all" // default
	if w, ok := args["which"].(string); ok {
		which = w
	}

	conn, err := grpc.NewClient(s.colorURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return CallToolResponse{}, err
	}
	defer conn.Close()

	client := color.NewColorServiceClient(conn)

	// Update color (note that UpdateColor does the right thing with
	// "all" vs "center"/"edge").
	req := &color.ColorUpdate{
		Which: which,
		Color: newColor,
	}

	resp, err := client.UpdateColor(ctx, req)
	if err != nil {
		return CallToolResponse{}, err
	}

	return CallToolResponse{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Color updated successfully (%s): %s", resp.Which, resp.Color),
			},
		},
	}, nil
}
