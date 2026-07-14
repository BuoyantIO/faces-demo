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
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/color"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// FacePublisherProvider continuously calls smiley and color, writes each
// assembled FaceMessage to MySQL (write-ahead), then pushes it into the
// configured queue backend (RabbitMQ by default, Redis optionally).
// The HTTP server it exposes is for liveness/readiness probes only.
type FacePublisherProvider struct {
	BaseProvider
	smileyService       string
	colorService        string
	colorClient         color.ColorServiceClient
	store               *FaceStore
	queue               QueueBackend
	publishIntervalMs   atomic.Int64
	publishConcurrency  int
	janitorInterval     time.Duration
	paused              atomic.Bool
	fillThreshold       float64       // WarmQueue when queue depth < maxDepth * fillThreshold
	monitorInterval     time.Duration // how often to poll queue depth
}

// NewFacePublisherFromEnvironment creates and fully initialises the publisher.
// Panics if MySQL or the queue backend cannot be reached.
func NewFacePublisherFromEnvironment() *FacePublisherProvider {
	fprv := &FacePublisherProvider{
		BaseProvider: BaseProvider{
			Name: "FacePublisher",
			Key:  "FacePublisher",
		},
	}

	fprv.SetLogger(slog.Default().With("provider", "FacePublisherProvider"))
	fprv.SetGetHandler(fprv.Get)
	fprv.BaseProvider.SetupFromEnvironment()

	fprv.smileyService = utils.StringFromEnv("SMILEY_SERVICE", "smiley")
	fprv.colorService = utils.StringFromEnv("COLOR_SERVICE", "color")

	// Normalise colorService to host:port.
	if _, _, err := net.SplitHostPort(fprv.colorService); err != nil {
		addr := net.ParseIP(fprv.colorService)
		if addr != nil {
			if strings.Contains(fprv.colorService, ":") {
				fprv.colorService = fmt.Sprintf("[%s]:80", fprv.colorService)
			} else {
				fprv.colorService = fmt.Sprintf("%s:80", fprv.colorService)
			}
		} else {
			fprv.colorService = fmt.Sprintf("%s:80", fprv.colorService)
		}
	}

	// Persistent gRPC connection — reused by all publish goroutines to avoid
	// TCP+HTTP/2 setup overhead on every loop iteration.
	colorConn, err := grpc.NewClient(fprv.colorService,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(fmt.Sprintf("FacePublisher: cannot connect to color service %s: %v", fprv.colorService, err))
	}
	fprv.colorClient = color.NewColorServiceClient(colorConn)

	store, err := NewFaceStoreFromEnvironment()
	if err != nil {
		panic(fmt.Sprintf("FacePublisher: cannot connect to MySQL: %v", err))
	}
	fprv.store = store

	queue, err := NewQueueBackendFromEnvironment(fprv.logger)
	if err != nil {
		panic(fmt.Sprintf("FacePublisher: cannot initialise queue backend: %v", err))
	}
	fprv.queue = queue

	fprv.publishIntervalMs.Store(int64(utils.IntFromEnv("PUBLISH_INTERVAL_MS", 50)))
	fprv.publishConcurrency = utils.IntFromEnv("PUBLISH_CONCURRENCY", 2)
	fprv.janitorInterval = time.Duration(utils.IntFromEnv("JANITOR_INTERVAL_M", 60)) * time.Minute
	fprv.fillThreshold = utils.FloatFromEnv("QUEUE_FILL_THRESHOLD", 0.75)
	fprv.monitorInterval = time.Duration(utils.IntFromEnv("QUEUE_MONITOR_INTERVAL_S", 30)) * time.Second

	fprv.Infof("FacePublisher: smiley http://%s  color grpc://%s", fprv.smileyService, fprv.colorService)
	fprv.Infof("FacePublisher: intervalMs=%d  concurrency=%d", fprv.publishIntervalMs.Load(), fprv.publishConcurrency)

	return fprv
}

// Get handles a single on-demand HTTP GET (probe / smoke-test traffic).
func (fprv *FacePublisherProvider) Get(prvReq *ProviderRequest) ProviderResponse {
	ctx := context.Background()
	msg := fprv.buildFaceMessage(prvReq)

	if err := fprv.writeAndPublish(ctx, msg); err != nil {
		fprv.Warnf("FacePublisher.Get: writeAndPublish: %v", err)
	}

	resp := ProviderResponseEmpty()
	resp.Add("smiley", msg.Smiley)
	resp.Add("color", msg.Color)
	for _, e := range msg.Errors {
		resp.AddError(e)
	}
	return resp
}

// WarmQueue re-populates the queue backend from all unacknowledged MySQL rows
// (state='pending' OR state='queued').  Must be called synchronously in main()
// before StartPublishLoops, and can also be triggered on demand via PUT /control
// with {"warm": true} — useful after a manual queue purge without restarting the pod.
//
// Re-pushing 'queued' rows is intentional: 'queued' means the message was
// successfully pushed at some point, but the queue may have been purged or
// restarted since then.  Re-pushing is idempotent from the GUI's perspective —
// if the message is already in the queue, the subscriber simply delivers it again.
func (fprv *FacePublisherProvider) WarmQueue() {
	ctx := context.Background()

	msgs, err := fprv.store.UnacknowledgedMessages(ctx)
	if err != nil {
		fprv.Warnf("FacePublisher: warm-up: query failed: %v", err)
		return
	}

	fprv.Infof("FacePublisher: warm-up: re-hydrating queue with %d unacknowledged DB rows", len(msgs))

	if err := fprv.queue.Warm(ctx, msgs); err != nil {
		fprv.Warnf("FacePublisher: warm-up: %v", err)
	}

	fprv.Infof("FacePublisher: warm-up: done")
}

// RunPublishLoop continuously produces messages at publishInterval.
// Designed to run as a goroutine after WarmQueue completes.
func (fprv *FacePublisherProvider) RunPublishLoop() {
	prvReq := &ProviderRequest{
		subrequest: "center",
		user:       "publisher",
		userAgent:  "face-publisher",
		row:        0,
		col:        0,
	}
	ctx := context.Background()

	for {
		if fprv.paused.Load() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		start := time.Now()
		msg := fprv.buildFaceMessage(prvReq)

		if err := fprv.writeAndPublish(ctx, msg); err != nil {
			fprv.Warnf("FacePublisher: publish loop: %v", err)
		}

		// Pace against the interval END-TO-END: the admin slider's msg/s target
		// assumes one message per interval, so the time already spent calling
		// smiley/color and writing to MySQL + the queue counts toward it.
		// If the work exceeds the interval the loop is saturated — no sleep,
		// it simply runs at its natural maximum.
		if ms := fprv.publishIntervalMs.Load(); ms > 0 {
			if remaining := time.Duration(ms)*time.Millisecond - time.Since(start); remaining > 0 {
				time.Sleep(remaining)
			}
		}
	}
}

// StartPublishLoops spawns publishConcurrency goroutines of RunPublishLoop.
func (fprv *FacePublisherProvider) StartPublishLoops() {
	fprv.Infof("FacePublisher: starting %d publish goroutine(s) at %dms interval",
		fprv.publishConcurrency, fprv.publishIntervalMs.Load())
	for i := 0; i < fprv.publishConcurrency; i++ {
		go fprv.RunPublishLoop()
	}
}

// RegisterAdminRoutes mounts the /control endpoint on the given server.
// GET /control  — returns paused state + queue depth + publish config.
// PUT /control  — body {"paused": true} to pause, {"paused": false} to resume.
func (fprv *FacePublisherProvider) RegisterAdminRoutes(server *BaseHTTPServer) {
	server.AddRoute("/control", fprv.handleControl)
	fprv.BaseProvider.RegisterChaosRoutes(server)
}

func (fprv *FacePublisherProvider) handleControl(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		pending, queued, acknowledged, _ := fprv.store.QueueDepth(ctx)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"paused":             fprv.paused.Load(),
			"publishIntervalMs":  fprv.publishIntervalMs.Load(),
			"publishConcurrency": fprv.publishConcurrency,
			"pending":            pending,
			"queued":             queued,
			"acknowledged":       acknowledged,
		})

	case http.MethodPut:
		var body struct {
			Paused            *bool  `json:"paused"`
			PublishIntervalMs *int64 `json:"publishIntervalMs"`
			Warm              bool   `json:"warm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.Paused != nil {
			fprv.paused.Store(*body.Paused)
			fprv.Infof("FacePublisher: paused=%v via admin control", *body.Paused)
		}
		if body.PublishIntervalMs != nil && *body.PublishIntervalMs >= 0 {
			fprv.publishIntervalMs.Store(*body.PublishIntervalMs)
			fprv.Infof("FacePublisher: publishIntervalMs=%d via admin control", *body.PublishIntervalMs)
		}
		if body.Warm {
			// Re-hydrate the queue from MySQL without restarting the pod.
			// Useful after a manual queue purge or queue backend restart.
			fprv.Infof("FacePublisher: manual WarmQueue triggered via admin control")
			go fprv.WarmQueue()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"paused":            fprv.paused.Load(),
			"publishIntervalMs": fprv.publishIntervalMs.Load(),
			"warming":           body.Warm,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// RunJanitor periodically deletes old acknowledged rows from MySQL.
func (fprv *FacePublisherProvider) RunJanitor() {
	ctx := context.Background()

	for {
		time.Sleep(fprv.janitorInterval)

		n, err := fprv.store.DeleteOldRows(ctx)
		if err != nil {
			fprv.Warnf("FacePublisher: janitor: %v", err)
			continue
		}

		if n > 0 {
			fprv.Infof("FacePublisher: janitor: deleted %d old rows", n)
		}
	}
}

// RunQueueMonitor polls the queue depth every monitorInterval and calls
// WarmQueue when the depth falls below fillThreshold × maxDepth.
//
// This recovers automatically from:
//   - Manual queue purges (RabbitMQ management UI "Purge" button)
//   - Queue pod restarts that wipe ephemeral storage
//   - Any other event that drains the queue without consuming messages
//
// Designed to run as a goroutine alongside StartPublishLoops.
func (fprv *FacePublisherProvider) RunQueueMonitor() {
	ctx := context.Background()
	maxDepth := fprv.queue.MaxDepth()
	threshold := int64(float64(maxDepth) * fprv.fillThreshold)

	fprv.Infof("FacePublisher: queue monitor started — warming when depth < %d (%.0f%% of %d), poll every %s",
		threshold, fprv.fillThreshold*100, maxDepth, fprv.monitorInterval)

	for {
		time.Sleep(fprv.monitorInterval)

		depth, err := fprv.queue.Depth(ctx)
		if err != nil {
			fprv.Warnf("FacePublisher: queue monitor: depth check failed: %v", err)
			continue
		}

		if depth < threshold {
			fprv.Infof("FacePublisher: queue monitor: depth %d < threshold %d — triggering WarmQueue", depth, threshold)
			fprv.WarmQueue()
		}
	}
}

// buildFaceMessage calls smiley and color in parallel and assembles a
// FaceMessage with error fallbacks.
func (fprv *FacePublisherProvider) buildFaceMessage(prvReq *ProviderRequest) *FaceMessage {
	smileyCh := make(chan *FaceResponse)
	colorCh := make(chan *FaceResponse)

	go func() { smileyCh <- fprv.makeSmileyRequest(prvReq) }()
	go func() { colorCh <- fprv.makeColorRequest(prvReq) }()

	smileyResp := <-smileyCh
	colorResp := <-colorCh

	msg := &FaceMessage{
		ID:        NewFaceMessageID(),
		Timestamp: time.Now().UnixMilli(),
	}

	if smileyResp.statusCode != http.StatusOK {
		msg.Errors = append(msg.Errors, fmt.Sprintf("smiley: %s", smileyResp.data))
		smileyName := mapStatus("smiley", smileyResp.statusCode)
		msg.Smiley, _ = utils.Smileys.Lookup(smileyName)
	} else {
		msg.Smiley = smileyResp.data
	}

	if colorResp.statusCode != http.StatusOK {
		msg.Errors = append(msg.Errors, fmt.Sprintf("color: %s", colorResp.data))
		colorName := mapStatus("color", colorResp.statusCode)
		msg.Color, _ = utils.Colors.Lookup(colorName)
	} else {
		msg.Color = colorResp.data
	}

	return msg
}

// writeAndPublish is the two-phase write-ahead commit.
// Phase 1 (MySQL INSERT) is fatal — the message is not pushed if this fails.
// Phase 2 (queue Push) is non-fatal — the message is safe in MySQL.
func (fprv *FacePublisherProvider) writeAndPublish(ctx context.Context, msg *FaceMessage) error {
	// Phase 1: write-ahead — message is durable before it enters the queue.
	if err := fprv.store.Insert(ctx, msg); err != nil {
		return fmt.Errorf("writeAndPublish: MySQL insert: %w", err)
	}

	// Phase 2: push to queue.
	if err := fprv.queue.Push(ctx, msg); err != nil {
		// Push failed — leave row as 'pending' so WarmQueue picks it up on restart.
		fprv.Warnf("FacePublisher: queue push failed (message safe in MySQL for recovery): %v", err)
		return nil
	}

	// Phase 3: mark as queued — transitions 'pending' → 'queued' to show the
	// message made it into the queue.  'pending' now only means "push hasn't
	// happened yet" (crash-recovery window, typically milliseconds).
	// 'acknowledged' is reserved for when the subscriber actually delivers
	// the message to the GUI.
	if err := fprv.store.MarkQueued(ctx, msg.ID); err != nil {
		fprv.Warnf("FacePublisher: mark-queued failed (non-fatal): %v", err)
	}

	return nil
}

// --- back-end call helpers ---------------------------------------------------

func (fprv *FacePublisherProvider) makeSmileyRequest(prvReq *ProviderRequest) *FaceResponse {
	start := time.Now()
	url := fmt.Sprintf("http://%s/%s/?row=%d&col=%d", fprv.smileyService, prvReq.subrequest, prvReq.row, prvReq.col)

	fprv.Debugf("HTTP starting (%s) %s", prvReq.InfoStr(), url)

	failed := false
	rcode := http.StatusOK
	rtext := ""
	var response *http.Response
	var ok bool

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		failed = true
		rcode = http.StatusInternalServerError
		rtext = fmt.Sprintf("couldn't create request to %s: %s", fprv.smileyService, err)
	}

	if !failed {
		req.Header.Set(fprv.userHeaderName, prvReq.user)
		req.Header.Set("User-Agent", prvReq.userAgent)

		response, err = http.DefaultClient.Do(req)
		if err != nil {
			failed = true
			rcode = http.StatusInternalServerError
			rtext = fmt.Sprintf("couldn't make request to %s: %s", fprv.smileyService, err)
		}
	}

	if !failed {
		defer response.Body.Close()

		rcode = response.StatusCode
		body, _ := io.ReadAll(response.Body)

		fprv.Debugf("HTTP %s status %d", url, rcode)

		if rcode != http.StatusOK {
			failed = true
			bstr := ""
			if len(body) > 0 {
				bstr = fmt.Sprintf(" (%s)", string(body))
			}
			rtext = fmt.Sprintf("error from %s: %03d%s", fprv.smileyService, rcode, bstr)
		}

		if !failed {
			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				failed = true
				rtext = fmt.Sprintf("couldn't decode response from %s: %s", fprv.smileyService, err)
			}

			if !failed {
				rtext, ok = data["smiley"].(string)
				if !ok {
					rtext = fmt.Sprintf("no smiley in response from %s", fprv.smileyService)
				}
			}
		}
	}

	latency := time.Since(start)
	fprv.Debugf("HTTP %s done (%d, %dms -- %s)", url, rcode, latency.Milliseconds(), rtext)

	return &FaceResponse{statusCode: rcode, latency: latency, data: rtext}
}

func (fprv *FacePublisherProvider) makeColorRequest(prvReq *ProviderRequest) *FaceResponse {
	md := metadata.New(map[string]string{"x-faces-user": prvReq.user})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	colorReq := &color.ColorRequest{
		Row:    int32(prvReq.row),
		Column: int32(prvReq.col),
	}

	fprv.Debugf("gRPC starting (%s) %s", prvReq.InfoStr(), fprv.colorService)

	var (
		colorResp *color.ColorResponse
		err       error
	)
	if prvReq.subrequest == "center" {
		colorResp, err = fprv.colorClient.Center(ctx, colorReq)
	} else {
		colorResp, err = fprv.colorClient.Edge(ctx, colorReq)
	}

	if err != nil {
		fprv.Debugf("gRPC (%s) failed: %s", prvReq.InfoStr(), err)
		return &FaceResponse{
			statusCode: http.StatusInternalServerError,
			data:       fmt.Sprintf("couldn't get color from %s: %s", fprv.colorService, err),
		}
	}

	fprv.Debugf("gRPC (%s) succeeded: %s", prvReq.InfoStr(), colorResp.Color)
	return &FaceResponse{statusCode: http.StatusOK, data: colorResp.Color}
}
