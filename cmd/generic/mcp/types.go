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

import "context"

// Tool interface defines the contract for all MCP tools
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Handle(ctx context.Context, s *MCPServer, args map[string]interface{}) (CallToolResponse, error)
}

// ToolDefinition represents a tool definition for MCP protocol responses
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolsListResponse is the response for the tools/list method
type ToolsListResponse struct {
	Tools []ToolDefinition `json:"tools"`
}

// CallToolRequest represents an MCP tools/call request
type CallToolRequest struct {
	Params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	} `json:"params"`
}

// CallToolResponse represents the response from a tool call
type CallToolResponse struct {
	Content []map[string]interface{} `json:"content"`
}
