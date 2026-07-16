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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// MCPServer is a mostly-generic MCP server implementation -- it would be
// slightly cleaner to factor out the Faces service URLs, but broadly
// speaking this is a pretty reusable server that happens to have some
// app-specific configuration in it.
type MCPServer struct {
	name  string
	tools []Tool

	// For SSE mode
	mu            sync.RWMutex
	sseClients    map[chan []byte]bool
	clientCounter int

	// Faces-specific service URLs
	smileyURL string
	colorURL  string
	faceURL   string
}

// NewMCPServerFromEnvironment creates a new MCPServer, reading the
// app-specific configuration (Faces service URLs) from the environment.
func NewMCPServerFromEnvironment(tools []Tool) *MCPServer {
	// Make a defensive copy to avoid external mutation
	toolsCopy := make([]Tool, len(tools))
	copy(toolsCopy, tools)

	return &MCPServer{
		name:       "faces-mcp",
		tools:      toolsCopy,
		smileyURL:  getEnv("SMILEY_URL", "http://smiley"),
		colorURL:   getEnv("COLOR_URL", "color:80"),
		faceURL:    getEnv("FACE_URL", "http://face"),
		sseClients: make(map[chan []byte]bool),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// listTools returns the tool definitions that this MCP server supports.
// It makes a copy mostly out of sheer paranoia.
func (mcpsrv *MCPServer) listTools() ToolsListResponse {
	defs := make([]ToolDefinition, len(mcpsrv.tools))

	for i, tool := range mcpsrv.tools {
		defs[i] = ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		}
	}

	return ToolsListResponse{Tools: defs}
}

// callTool routes tool calls to their respective handlers
func (mcpsrv *MCPServer) callTool(ctx context.Context, name string, args map[string]interface{}) (CallToolResponse, error) {
	for _, tool := range mcpsrv.tools {
		if tool.Name() == name {
			return tool.Handle(ctx, mcpsrv, args)
		}
	}
	return CallToolResponse{}, fmt.Errorf("unknown tool: %s", name)
}

// processJSONRPCRequest handles a JSON-RPC 2.0 request and returns the response.
// Returns nil for notifications (which don't require a response).
//
// This is used by both stdio and SSE handlers.
func (mcpsrv *MCPServer) processJSONRPCRequest(ctx context.Context, request map[string]interface{}) interface{} {
	method, ok := request["method"].(string)
	if !ok {
		log.Printf("Missing method in request")
		return nil
	}

	id := request["id"]

	switch method {
	case "initialize":
		params, _ := request["params"].(map[string]interface{})
		log.Printf("Initialize request with params: %v", params)

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "faces-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "notifications/initialized":
		log.Printf("Client initialized")
		// No response for notifications
		return nil

	case "tools/list":
		log.Printf("Tools list request")
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"tools": mcpsrv.listTools().Tools,
			},
		}

	case "tools/call":
		params, ok := request["params"].(map[string]interface{})
		if !ok {
			log.Printf("Missing params in request")
			return map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			}
		}

		name, _ := params["name"].(string)
		args, _ := params["arguments"].(map[string]interface{})

		log.Printf("Tools call request for %s", name)

		result, err := mcpsrv.callTool(ctx, name, args)

		if err != nil {
			return map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32603,
					"message": err.Error(),
				},
			}
		}

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  result,
		}

	default:
		log.Printf("Unknown method: %s", method)

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]interface{}{
				"code":    -32601,
				"message": fmt.Sprintf("Method not found: %s", method),
			},
		}
	}
}

// handleStdio handles stdio-based MCP communication
func (mcpsrv *MCPServer) handleStdio() {
	log.Printf("%s MCP Server starting in stdio mode", mcpsrv.name)
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Received request: %s", line)

		var request map[string]interface{}
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		response := mcpsrv.processJSONRPCRequest(context.Background(), request)

		// Write response (skip for notifications)
		if response != nil {
			responseBytes, _ := json.Marshal(response)
			fmt.Println(string(responseBytes))
		}
	}
}

// handleSSE handles Server-Sent Events connection for MCP
func (mcpsrv *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s MCP Server starting in SSE mode", mcpsrv.name)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan []byte, 10)

	mcpsrv.mu.Lock()
	mcpsrv.sseClients[clientChan] = true
	clientID := mcpsrv.clientCounter
	mcpsrv.clientCounter++
	mcpsrv.mu.Unlock()

	defer func() {
		mcpsrv.mu.Lock()
		delete(mcpsrv.sseClients, clientChan)
		mcpsrv.mu.Unlock()
		close(clientChan)
	}()

	log.Printf("SSE client %d connected", clientID)

	// Send endpoint discovery message
	endpoint := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "endpoint",
		"params": map[string]interface{}{
			"endpoint": "/messages",
		},
	}
	endpointBytes, _ := json.Marshal(endpoint)
	fmt.Fprintf(w, "data: %s\n\n", string(endpointBytes))
	w.(http.Flusher).Flush()

	// Keep connection alive and send messages
	for {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", string(msg))
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			log.Printf("SSE client %d disconnected", clientID)
			return
		}
	}
}

// handleMessages handles incoming messages from SSE clients
func (mcpsrv *MCPServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := mcpsrv.processJSONRPCRequest(r.Context(), request)

	// No response for notifications
	if response == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
