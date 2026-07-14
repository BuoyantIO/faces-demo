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
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

// FaceSubscriberProvider serves faces-gui requests by atomically popping the
// oldest message from the queue backend and acknowledging the corresponding
// MySQL row.  The Kubernetes Service named "face" routes to these pods.
type FaceSubscriberProvider struct {
	BaseProvider
	store  *FaceStore
	queue  QueueBackend
	paused atomic.Bool
}

// NewFaceSubscriberFromEnvironment creates and fully initialises the
// subscriber.  Panics if MySQL or the queue backend cannot be reached.
func NewFaceSubscriberFromEnvironment() *FaceSubscriberProvider {
	fprv := &FaceSubscriberProvider{
		BaseProvider: BaseProvider{
			Name: "FaceSubscriber",
			Key:  "FaceSubscriber",
		},
	}

	fprv.SetLogger(slog.Default().With("provider", "FaceSubscriberProvider"))
	fprv.SetGetHandler(fprv.Get)
	fprv.BaseProvider.SetupFromEnvironment()

	store, err := NewFaceStoreFromEnvironment()
	if err != nil {
		panic(fmt.Sprintf("FaceSubscriber: cannot connect to MySQL: %v", err))
	}
	fprv.store = store

	queue, err := NewQueueBackendFromEnvironment(fprv.logger)
	if err != nil {
		panic(fmt.Sprintf("FaceSubscriber: cannot initialise queue backend: %v", err))
	}
	fprv.queue = queue

	backend := utils.StringFromEnv("QUEUE_BACKEND", "rabbitmq")
	fprv.Infof("FaceSubscriber: queue backend=%s", backend)

	return fprv
}

// Get is the three-step consume sequence called for each GUI poll:
//  1. Pop from queue backend
//  2. Acknowledge row in MySQL (non-fatal)
//  3. Return smiley/color to GUI
func (fprv *FaceSubscriberProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	ctx := context.Background()

	// Return empty-queue response immediately when paused so the GUI shows
	// the neutral "queue dry" state without consuming any messages.
	if fprv.paused.Load() {
		return fprv.emptyQueueResponse()
	}

	// Step 1: pop
	msg, err := fprv.queue.Pop(ctx)
	if err != nil {
		return ProviderResponse{
			StatusCode: http.StatusInternalServerError,
			Data: map[string]interface{}{
				"errors": []string{fmt.Sprintf("queue pop: %v", err)},
			},
		}
	}
	if msg == nil {
		return fprv.emptyQueueResponse()
	}

	// Step 2: acknowledge in MySQL (non-fatal — response still returned on failure)
	if msg.ID == "" {
		fprv.Warnf("FaceSubscriber: message has no ID, skipping ack")
	} else if err := fprv.store.Acknowledge(ctx, msg.ID); err != nil {
		fprv.Warnf("FaceSubscriber: ack failed for %s: %v", msg.ID, err)
	}

	// Step 3: respond
	resp := ProviderResponseEmpty()
	resp.Add("smiley", msg.Smiley)
	resp.Add("color", msg.Color)
	for _, e := range msg.Errors {
		resp.AddError(e)
	}
	return resp
}

// RegisterAdminRoutes mounts the /control endpoint on the given server.
// GET /control  — returns paused state + queue depth.
// PUT /control  — body {"paused": true} to pause, {"paused": false} to resume.
func (fprv *FaceSubscriberProvider) RegisterAdminRoutes(server *BaseHTTPServer) {
	server.AddRoute("/control", fprv.handleControl)
	fprv.BaseProvider.RegisterChaosRoutes(server)
}

func (fprv *FaceSubscriberProvider) handleControl(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		pending, queued, acknowledged, _ := fprv.store.QueueDepth(ctx)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"paused":       fprv.paused.Load(),
			"pending":      pending,
			"queued":       queued,
			"acknowledged": acknowledged,
		})

	case http.MethodPut:
		var body struct {
			Paused bool `json:"paused"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		fprv.paused.Store(body.Paused)
		fprv.Infof("FaceSubscriber: subscribing paused=%v via admin control", body.Paused)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"paused": body.Paused})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// emptyQueueResponse returns HTTP 503 with the neutral (Cursing) smiley and
// grey color — a clearly transient "queue empty" state distinct from a 200 OK.
func (fprv *FaceSubscriberProvider) emptyQueueResponse() ProviderResponse {
	smiley, _ := utils.Smileys.Lookup(utils.Defaults["smiley"])
	clr, _ := utils.Colors.Lookup(utils.Defaults["color"])
	return ProviderResponse{
		StatusCode: http.StatusServiceUnavailable,
		Data: map[string]interface{}{
			"smiley": smiley,
			"color":  clr,
		},
	}
}

