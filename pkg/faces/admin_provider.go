// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package faces

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"

	colorpkg "github.com/BuoyantIO/faces-demo/v2/pkg/color"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

// AdminProvider serves the faces-admin web UI and REST API.
// It is intentionally separate from BaseProvider — it never participates in
// the face request pipeline and has no chaos injection.
type AdminProvider struct {
	logger   *slog.Logger
	dataPath string
	mux      *http.ServeMux

	// connection state (nil when unavailable / classic mode)
	db          *sql.DB
	redisClient *redis.Client
	redisKey    string

	// service URLs for health checks and proxying
	smileyURL string
	colorURL  string
	// colorChaosURL removed — color chaos is now a gRPC method on port 8000
	colorGRPCAddr   string
	colorClient     colorpkg.ColorServiceClient
	publisherURL    string
	subscriberURL   string
	guiURL          string
	faceURL         string
	rabbitmqMgmtURL string
	rabbitmqQueue   string

	// headless service names for per-pod discovery
	publisherHeadless  string
	subscriberHeadless string
	smileyHeadless     string
	colorHeadless      string
	podPort            string

	faceMode  string // "pubsub" | "classic"
	queueType string // "redis" | "rabbitmq"
	maxDepth  int64

	// rolling history for sparklines (protected by mu)
	mu      sync.RWMutex
	history []pipelinePoint

	// runtime-editable service URL overrides (protected by cfgMu).
	// Takes precedence over the original env-var values when set.
	cfgMu        sync.RWMutex
	cfgOverrides map[string]string // key → overridden URL

	// keep a reference to the gRPC conn so we can swap it on color URL change
	colorGRPCConn *grpc.ClientConn

	// Kubernetes API client — nil when running outside a cluster or without RBAC
	k8s *k8sClient

	// auth (empty strings = auth disabled)
	adminUsername string
	adminPassword string
	sessions      *sessionStore

	// emojivoto integration (protected by emojivotoMu)
	emojivotoMu            sync.RWMutex
	emojivotoEndpoint      string
	emojivotoEnabled       bool
	emojivotoUpdateSmileys bool
	emojivotoLeader        string
	emojivotoLeaderboard   []emojivotoEntry
	emojivotoStatus        string // "unconfigured" | "ok" | "error"
	emojivotoError         string
	emojivotoSelectedPods  []string
}

type pipelinePoint struct {
	Timestamp    int64 `json:"ts"`
	Pending      int64 `json:"pending"`      // write-ahead buffer; should be ~0
	Queued       int64 `json:"queued"`       // pushed to queue, awaiting subscriber
	Acknowledged int64 `json:"acknowledged"` // consumed and delivered to GUI
	QueueDepth   int64 `json:"queueDepth"`   // current depth from the queue backend
}

// --- API response types ---

type serviceStatus struct {
	Healthy   bool   `json:"healthy"`
	LatencyMs int64  `json:"latencyMs,omitempty"`
	Error     string `json:"error,omitempty"`
}

type statusResponse struct {
	Mode         string                    `json:"mode"`
	QueueBackend string                    `json:"queueBackend"`
	Services     map[string]*serviceStatus `json:"services"`
	Timestamp    int64                     `json:"timestamp"`
}

type mysqlStats struct {
	Available    bool   `json:"available"`
	Pending      int64  `json:"pending"`      // write-ahead buffer; should be ~0
	Queued       int64  `json:"queued"`       // in pipeline, awaiting subscriber
	Acknowledged int64  `json:"acknowledged"` // delivered to GUI
	Total        int64  `json:"total"`
	Error        string `json:"error,omitempty"`
}

type queueStats struct {
	Backend     string  `json:"backend"`
	Available   bool    `json:"available"`
	Depth       int64   `json:"depth"`
	MaxDepth    int64   `json:"maxDepth"`
	ReadyRate   float64 `json:"readyRate,omitempty"`
	DeliverRate float64 `json:"deliverRate,omitempty"`
	Error       string  `json:"error,omitempty"`
}

type pipelineResponse struct {
	MySQL   mysqlStats      `json:"mysql"`
	Queue   queueStats      `json:"queue"`
	History []pipelinePoint `json:"history"`
	// derived: messages dropped = pending rows never acknowledged past retention
	Timestamp int64 `json:"timestamp"`
}

// podControlState is the per-replica view of a publisher or subscriber pod.
type podControlState struct {
	PodIP              string `json:"podIP"`
	PodName            string `json:"podName,omitempty"` // from Kubernetes API
	Node               string `json:"node,omitempty"`
	Zone               string `json:"zone,omitempty"`
	Region             string `json:"region,omitempty"`
	Available          bool   `json:"available"`
	Paused             bool   `json:"paused"`
	PublishIntervalMs  int64  `json:"publishIntervalMs"`
	PublishConcurrency int    `json:"publishConcurrency"`
	Pending            int64  `json:"pending"`      // should be ~0
	Queued             int64  `json:"queued"`       // awaiting subscriber
	Acknowledged       int64  `json:"acknowledged"` // delivered to GUI
	Error              string `json:"error,omitempty"`
}

// controlState is the aggregate view across all replicas of a service.
// Top-level fields reflect the primary (first-responding) pod so the UI
// slider always has something to show; Pods carries the per-replica detail.
type controlState struct {
	Available          bool              `json:"available"`
	Paused             bool              `json:"paused"`
	PublishIntervalMs  int64             `json:"publishIntervalMs"`
	PublishConcurrency int               `json:"publishConcurrency"`
	PodCount           int               `json:"podCount"`
	Pods               []podControlState `json:"pods,omitempty"`
	Pending            int64             `json:"pending"`      // should be ~0; non-zero = queue backend problem
	Queued             int64             `json:"queued"`       // in pipeline, awaiting subscriber
	Acknowledged       int64             `json:"acknowledged"` // delivered to GUI
	Error              string            `json:"error,omitempty"`
}

type controlsResponse struct {
	Publisher  controlState `json:"publisher"`
	Subscriber controlState `json:"subscriber"`
}

type configResponse struct {
	FaceMode      string `json:"faceMode"`
	QueueBackend  string `json:"queueBackend"`
	MaxDepth      int64  `json:"maxDepth"`
	SmileyURL     string `json:"smileyURL"`
	PublisherURL  string `json:"publisherURL,omitempty"`
	SubscriberURL string `json:"subscriberURL,omitempty"`
	GUIURL        string `json:"guiURL"`
	FaceURL       string `json:"faceURL"`
	ColorURL      string `json:"colorURL"`
	// Runtime context — useful for the operator info tooltip
	Namespace     string `json:"namespace,omitempty"` // K8s namespace the admin runs in
	K8sAvailable  bool   `json:"k8sAvailable"`        // true when in-cluster K8s API is reachable
	LinkerdMeshed bool   `json:"linkerdMeshed"`       // true when the Linkerd sidecar proxy is present
	AuthEnabled   bool   `json:"authEnabled"`         // true when ADMIN_USERNAME/PASSWORD are set
}

type emojivotoEntry struct {
	Votes     string `json:"votes"`
	Unicode   string `json:"unicode"`
	Shortcode string `json:"shortcode"`
}

type emojivotoConfigResponse struct {
	Endpoint      string           `json:"endpoint"`
	Enabled       bool             `json:"enabled"`
	UpdateSmileys bool             `json:"updateSmileys"`
	Status        string           `json:"status"` // "unconfigured" | "ok" | "error"
	Error         string           `json:"error,omitempty"`
	Leader        string           `json:"leader,omitempty"`
	Leaderboard   []emojivotoEntry `json:"leaderboard,omitempty"`
	SelectedPods  []string         `json:"selectedPods"`
}

// --- Constructor ---

func NewAdminProviderFromEnvironment() *AdminProvider {
	logger := slog.Default().With("component", "AdminProvider")

	faceMode := utils.StringFromEnv("FACE_MODE", "classic")
	queueType := utils.StringFromEnv("QUEUE_BACKEND", "rabbitmq")

	colorSvc := utils.StringFromEnv("COLOR_SERVICE", "color")
	// Normalise to host:port for gRPC — matches publisher pattern
	colorGRPCAddr := colorSvc
	if _, _, err := net.SplitHostPort(colorSvc); err != nil {
		colorGRPCAddr = colorSvc + ":80"
	}

	a := &AdminProvider{
		logger:    logger,
		dataPath:  utils.StringFromEnv("DATA_PATH", "/app/data"),
		faceMode:  faceMode,
		queueType: queueType,
		smileyURL: fmt.Sprintf("http://%s", utils.StringFromEnv("SMILEY_SERVICE", "smiley")),
		colorURL:  fmt.Sprintf("http://%s", colorSvc),
		// colorChaosURL removed — chaos now via gRPC GetChaos/UpdateChaos on port 8000
		colorGRPCAddr:      colorGRPCAddr,
		publisherURL:       fmt.Sprintf("http://%s", utils.StringFromEnv("PUBLISHER_SERVICE", "face-publisher")),
		subscriberURL:      fmt.Sprintf("http://%s", utils.StringFromEnv("SUBSCRIBER_SERVICE", "face-subscriber")),
		guiURL:             fmt.Sprintf("http://%s", utils.StringFromEnv("GUI_SERVICE", "faces-gui")),
		faceURL:            fmt.Sprintf("http://%s", utils.StringFromEnv("FACE_SERVICE", "face")),
		publisherHeadless:  utils.StringFromEnv("PUBLISHER_HEADLESS_SERVICE", "face-publisher-headless"),
		subscriberHeadless: utils.StringFromEnv("SUBSCRIBER_HEADLESS_SERVICE", "face-subscriber-headless"),
		smileyHeadless:     utils.StringFromEnv("SMILEY_HEADLESS_SERVICE", "smiley-headless"),
		colorHeadless:      utils.StringFromEnv("COLOR_HEADLESS_SERVICE", "color-headless"),
		podPort:            utils.StringFromEnv("POD_PORT", "8000"),
		adminUsername:      utils.StringFromEnv("ADMIN_USERNAME", ""),
		adminPassword:      utils.StringFromEnv("ADMIN_PASSWORD", ""),
		sessions:           newSessionStore(),
	}

	// gRPC color client — non-blocking, lazy connection
	if conn, err := grpc.NewClient(colorGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
		a.colorGRPCConn = conn
		a.colorClient = colorpkg.NewColorServiceClient(conn)
	} else {
		logger.Warn("AdminProvider: could not create gRPC color client", "error", err)
	}

	a.cfgOverrides = make(map[string]string)
	a.emojivotoStatus = "unconfigured"
	a.emojivotoSelectedPods = []string{}

	// Try in-cluster K8s API — only works inside a pod with a service account
	if k, err := newK8sClient(logger); err == nil {
		a.k8s = k
		logger.Info("AdminProvider: Kubernetes API available — pod topology enabled")
	} else {
		logger.Info("AdminProvider: Kubernetes API not available (running locally?)", "reason", err)
	}

	// MySQL — optional in classic mode but connect if available
	if faceMode == "pubsub" || utils.BoolFromEnv("ADMIN_CONNECT_DB", false) {
		host := utils.StringFromEnv("DB_HOST", "mysql")
		port := utils.StringFromEnv("DB_PORT", "3306")
		dbName := utils.StringFromEnv("DB_NAME", "faces")
		user := utils.StringFromEnv("DB_USER", "faces")
		password := utils.StringFromEnv("DB_PASSWORD", "")
		dsn := utils.StringFromEnv("DB_DSN", fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
			user, password, host, port, dbName,
		))
		db, err := sql.Open("mysql", dsn)
		if err == nil {
			db.SetMaxOpenConns(3)
			db.SetMaxIdleConns(2)
			db.SetConnMaxLifetime(5 * time.Minute)
			a.db = db
		} else {
			logger.Warn("AdminProvider: could not open MySQL", "error", err)
		}
	}

	// Redis
	if queueType == "redis" {
		addr := utils.StringFromEnv("REDIS_ADDRESS", "redis:6379")
		password := utils.StringFromEnv("REDIS_PASSWORD", "")
		db := utils.IntFromEnv("REDIS_DB", 0)
		a.redisKey = utils.StringFromEnv("REDIS_QUEUE_KEY", "faces:queue")
		a.maxDepth = int64(utils.IntFromEnv("REDIS_MAX_QUEUE_DEPTH", 5000))
		a.redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})
	}

	// RabbitMQ management API
	if queueType == "rabbitmq" {
		rmqHost := utils.StringFromEnv("RABBITMQ_HOST", "rabbitmq")
		rmqMgmtPort := utils.StringFromEnv("RABBITMQ_MANAGEMENT_PORT", "15672")
		rmqUser := utils.StringFromEnv("RABBITMQ_USER", "faces")
		rmqPass := utils.StringFromEnv("RABBITMQ_PASSWORD", "faces-password")
		rmqVhost := utils.StringFromEnv("RABBITMQ_VHOST", "/")
		a.rabbitmqQueue = utils.StringFromEnv("RABBITMQ_QUEUE", "faces.queue")
		a.maxDepth = int64(utils.IntFromEnv("RABBITMQ_MAX_QUEUE_DEPTH", 5000))

		// Encode vhost for URL — "/" becomes "%2F"
		encodedVhost := strings.ReplaceAll(rmqVhost, "/", "%2F")
		a.rabbitmqMgmtURL = fmt.Sprintf("http://%s:%s@%s:%s/api/queues/%s/%s",
			rmqUser, rmqPass, rmqHost, rmqMgmtPort, encodedVhost, a.rabbitmqQueue)
	}

	a.setupRoutes()
	logger.Info("AdminProvider: initialized", "mode", faceMode, "queueBackend", queueType)
	return a
}

func (a *AdminProvider) setupRoutes() {
	a.mux = http.NewServeMux()

	// Auth (always open — must be registered before the catch-all "/" handler)
	a.mux.HandleFunc("/login", a.handleLogin)
	a.mux.HandleFunc("/api/login", a.handleLogin)
	a.mux.HandleFunc("/api/logout", a.handleLogout)

	// Static files
	a.mux.HandleFunc("/", a.serveStatic)

	// REST API
	a.mux.HandleFunc("/api/status", a.handleStatus)
	a.mux.HandleFunc("/api/pipeline", a.handlePipeline)
	a.mux.HandleFunc("/api/smiley", a.handleSmiley)
	a.mux.HandleFunc("/api/smileypods", a.handleServicePods("smiley", &a.smileyHeadless, func() string { return a.smileyURL }))
	a.mux.HandleFunc("/api/color", a.handleColor)
	a.mux.HandleFunc("/api/colorpods", a.handleServicePods("color", &a.colorHeadless, func() string { return a.colorURL }))
	a.mux.HandleFunc("/api/facepods", a.handleFacePods)

	// Pub/sub flow controls — GET discovers all pods; PUT broadcasts to all pods
	a.mux.HandleFunc("/api/controls", a.handleControls)
	a.mux.HandleFunc("/api/controls/publisher",
		a.handleControlProxy(&a.publisherHeadless, func() string { return a.publisherURL }))
	a.mux.HandleFunc("/api/controls/subscriber",
		a.handleControlProxy(&a.subscriberHeadless, func() string { return a.subscriberURL }))

	// Chaos injection control — GET/PUT per service; GET /api/chaos aggregates all
	a.mux.HandleFunc("/api/chaos", a.handleAllChaos)
	a.mux.HandleFunc("/api/chaos/smiley",
		a.handleChaosProxy(&a.smileyHeadless, func() string { return a.svcURL("smiley", a.smileyURL) }))
	a.mux.HandleFunc("/api/chaos/color", a.handleColorChaosProxy()) // gRPC — no HTTP sidecar
	// Face = face-subscriber in pub/sub mode. Use subscriber headless so PUT
	// broadcasts to all pods. Falls back to VIP in classic mode (no headless).
	a.mux.HandleFunc("/api/chaos/face",
		a.handleChaosProxy(&a.subscriberHeadless, func() string { return a.svcURL("face", a.faceURL) }))
	a.mux.HandleFunc("/api/chaos/publisher",
		a.handleChaosProxy(&a.publisherHeadless, func() string { return a.svcURL("publisher", a.publisherURL) }))
	a.mux.HandleFunc("/api/chaos/subscriber",
		a.handleChaosProxy(&a.subscriberHeadless, func() string { return a.svcURL("subscriber", a.subscriberURL) }))

	// Runtime config editing
	a.mux.HandleFunc("/api/config", a.handleConfig)

	// Emojivoto integration
	a.mux.HandleFunc("/api/emojivoto", a.handleEmojivoto)

	// Per-pod current serving state (what emoji/color is each pod returning right now)
	a.mux.HandleFunc("/api/smileystate", a.handleSmileyState)
	a.mux.HandleFunc("/api/colorstate", a.handleColorState)

	// Infrastructure overview — all pods grouped by topology zone
	a.mux.HandleFunc("/api/infrastructure", a.handleInfrastructure)

	// Linky custom images — lists available images in DATA_PATH/linkys/
	a.mux.HandleFunc("/api/linkys", a.handleLinkys)

	// Maintenance endpoints
	a.mux.HandleFunc("/api/maintenance/db/status", a.handleDBStatus)
	a.mux.HandleFunc("/api/maintenance/db/migrate", a.handleDBMigrate)
	a.mux.HandleFunc("/api/maintenance/db/purge", a.handleDBPurge)
	a.mux.HandleFunc("/api/maintenance/queue/purge", a.handleQueuePurge)

	// Face proxy — the mini widget polls /face/... from each page
	a.mux.HandleFunc("/face/", a.handleFaceProxy)

	// Prometheus metrics — exposes pipeline health for scraping
	a.mux.HandleFunc("/metrics", a.handleMetrics)

	// Health
	a.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
}

func (a *AdminProvider) Start(addr string) error {
	a.logStartup()
	a.logger.Info("faces-admin: listening", "addr", addr)
	return http.ListenAndServe(addr, a.loggingMiddleware(
		authMiddleware(a.mux, a.sessions, a.adminUsername, a.adminPassword),
	))
}

// StartBackgroundPoller polls MySQL and queue every 2s to build history.
func (a *AdminProvider) StartBackgroundPoller() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			a.collectPoint()
		}
	}()
}

func (a *AdminProvider) collectPoint() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	point := pipelinePoint{Timestamp: time.Now().UnixMilli()}

	if a.db != nil {
		pending, queued, acked, err := a.dbQueueDepth(ctx)
		if err == nil {
			point.Pending = pending
			point.Queued = queued
			point.Acknowledged = acked
		}
	}

	depth, err := a.queueDepth(ctx)
	if err == nil {
		point.QueueDepth = depth
	}

	a.mu.Lock()
	a.history = append(a.history, point)
	if len(a.history) > 120 { // keep 4 minutes at 2s
		a.history = a.history[len(a.history)-120:]
	}
	a.mu.Unlock()
}

// --- DB helpers ---

func (a *AdminProvider) dbQueueDepth(ctx context.Context) (pending, queued, acknowledged int64, err error) {
	if a.db == nil {
		return 0, 0, 0, fmt.Errorf("db not connected")
	}
	row := a.db.QueryRowContext(ctx,
		`SELECT
			SUM(CASE WHEN state='pending'      THEN 1 ELSE 0 END),
			SUM(CASE WHEN state='queued'       THEN 1 ELSE 0 END),
			SUM(CASE WHEN state='acknowledged' THEN 1 ELSE 0 END)
		 FROM face_queue`)
	err = row.Scan(&pending, &queued, &acknowledged)
	return
}

// --- Queue depth helpers ---

func (a *AdminProvider) queueDepth(ctx context.Context) (int64, error) {
	switch a.queueType {
	case "redis":
		return a.redisQueueDepth(ctx)
	case "rabbitmq":
		stats, err := a.rabbitmqQueueStats()
		if err != nil {
			return 0, err
		}
		return stats.Depth, nil
	}
	return 0, fmt.Errorf("unknown queue type %q", a.queueType)
}

func (a *AdminProvider) redisQueueDepth(ctx context.Context) (int64, error) {
	if a.redisClient == nil {
		return 0, fmt.Errorf("redis not connected")
	}
	return a.redisClient.LLen(ctx, a.redisKey).Result()
}

type rabbitmqAPIResponse struct {
	Messages               int64 `json:"messages"`
	MessagesReady          int64 `json:"messages_ready"`
	MessagesUnacknowledged int64 `json:"messages_unacknowledged"`
	MessageStats           *struct {
		PublishDetails struct {
			Rate float64 `json:"rate"`
		} `json:"publish_details"`
		DeliverGetDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_get_details"`
	} `json:"message_stats"`
}

func (a *AdminProvider) rabbitmqQueueStats() (*queueStats, error) {
	if a.rabbitmqMgmtURL == "" {
		return nil, fmt.Errorf("rabbitmq management not configured")
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(a.rabbitmqMgmtURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rabbitmq management API: %d", resp.StatusCode)
	}

	var rmq rabbitmqAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&rmq); err != nil {
		return nil, err
	}

	qs := &queueStats{
		Backend:   "rabbitmq",
		Available: true,
		Depth:     rmq.MessagesReady,
		MaxDepth:  a.maxDepth,
	}
	if rmq.MessageStats != nil {
		qs.ReadyRate = rmq.MessageStats.PublishDetails.Rate
		qs.DeliverRate = rmq.MessageStats.DeliverGetDetails.Rate
	}
	return qs, nil
}

// --- HTTP handlers ---

func (a *AdminProvider) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := map[string]*serviceStatus{}

	// Always check smiley, color, gui
	services["smiley"] = a.checkHealth(a.svcURL("smiley", a.smileyURL) + "/healthz")
	services["color"] = a.checkColorHealth()
	services["gui"] = a.checkHealth(a.svcURL("gui", a.guiURL) + "/ready")

	mode := a.effectiveFaceMode()
	if mode == "pubsub" {
		services["face-publisher"] = a.checkHealth(a.svcURL("publisher", a.publisherURL) + "/healthz")
		services["face-subscriber"] = a.checkHealth(a.svcURL("subscriber", a.subscriberURL) + "/healthz")
		services["mysql"] = a.checkDBHealth()
		if a.queueType == "rabbitmq" {
			services["rabbitmq"] = a.checkRabbitMQHealth()
		} else {
			services["redis"] = a.checkRedisHealth()
		}
	} else {
		services["face"] = a.checkHealth(a.svcURL("face", a.faceURL) + "/healthz")
	}

	resp := statusResponse{
		Mode:         mode,
		QueueBackend: a.queueType,
		Services:     services,
		Timestamp:    time.Now().UnixMilli(),
	}
	a.writeJSON(w, resp)
}

func (a *AdminProvider) handlePipeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := pipelineResponse{Timestamp: time.Now().UnixMilli()}

	// MySQL stats
	if a.db != nil {
		pending, queued, acked, err := a.dbQueueDepth(ctx)
		if err != nil {
			resp.MySQL = mysqlStats{Available: false, Error: err.Error()}
		} else {
			resp.MySQL = mysqlStats{
				Available:    true,
				Pending:      pending,
				Queued:       queued,
				Acknowledged: acked,
				Total:        pending + queued + acked,
			}
		}
	} else {
		resp.MySQL = mysqlStats{Available: false, Error: "not connected"}
	}

	// Queue stats
	switch a.queueType {
	case "redis":
		depth, err := a.redisQueueDepth(ctx)
		if err != nil {
			resp.Queue = queueStats{Backend: "redis", Available: false, MaxDepth: a.maxDepth, Error: err.Error()}
		} else {
			resp.Queue = queueStats{
				Backend:   "redis",
				Available: true,
				Depth:     depth,
				MaxDepth:  a.maxDepth,
			}
		}
	case "rabbitmq":
		qs, err := a.rabbitmqQueueStats()
		if err != nil {
			resp.Queue = queueStats{Backend: "rabbitmq", Available: false, MaxDepth: a.maxDepth, Error: err.Error()}
		} else {
			resp.Queue = *qs
		}
	}

	a.mu.RLock()
	resp.History = append([]pipelinePoint{}, a.history...)
	a.mu.RUnlock()

	a.writeJSON(w, resp)
}

// svcURL returns the current (possibly overridden) value for a named service URL.
func (a *AdminProvider) svcURL(key, defaultURL string) string {
	a.cfgMu.RLock()
	defer a.cfgMu.RUnlock()
	if v, ok := a.cfgOverrides[key]; ok && v != "" {
		return v
	}
	return defaultURL
}

// editableConfig holds the settings that operators can change at runtime via PUT /api/config.
type editableConfig struct {
	FaceMode      string `json:"faceMode"` // "classic" | "pubsub" | "" (no change)
	SmileyURL     string `json:"smileyURL"`
	FaceURL       string `json:"faceURL"`
	GUIURL        string `json:"guiURL"`
	PublisherURL  string `json:"publisherURL"`
	SubscriberURL string `json:"subscriberURL"`
	ColorURL      string `json:"colorURL"` // display only; gRPC addr derived from this
}

// effectiveFaceMode returns the runtime faceMode, respecting any override set via PUT /api/config.
// Protected by cfgMu so concurrent mode changes and status reads are safe.
func (a *AdminProvider) effectiveFaceMode() string {
	a.cfgMu.RLock()
	m := a.cfgOverrides["faceMode"]
	a.cfgMu.RUnlock()
	if m != "" {
		return m
	}
	return a.faceMode
}

func (a *AdminProvider) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ns := ""
		if a.k8s != nil {
			ns = a.k8s.namespace
		}
		resp := configResponse{
			FaceMode:      a.effectiveFaceMode(),
			QueueBackend:  a.queueType,
			MaxDepth:      a.maxDepth,
			SmileyURL:     a.svcURL("smiley", a.smileyURL),
			PublisherURL:  a.svcURL("publisher", a.publisherURL),
			SubscriberURL: a.svcURL("subscriber", a.subscriberURL),
			GUIURL:        a.svcURL("gui", a.guiURL),
			FaceURL:       a.svcURL("face", a.faceURL),
			ColorURL:      a.svcURL("color", a.colorURL),
			Namespace:     ns,
			K8sAvailable:  a.k8s != nil,
			LinkerdMeshed: isLinkerdMeshed(),
			AuthEnabled:   a.adminUsername != "",
		}
		a.writeJSON(w, resp)

	case http.MethodPut:
		var req editableConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		a.cfgMu.Lock()
		if req.FaceMode == "classic" || req.FaceMode == "pubsub" {
			a.cfgOverrides["faceMode"] = req.FaceMode
		}
		if req.SmileyURL != "" {
			a.cfgOverrides["smiley"] = req.SmileyURL
		}
		if req.FaceURL != "" {
			a.cfgOverrides["face"] = req.FaceURL
		}
		if req.GUIURL != "" {
			a.cfgOverrides["gui"] = req.GUIURL
		}
		if req.PublisherURL != "" {
			a.cfgOverrides["publisher"] = req.PublisherURL
		}
		if req.SubscriberURL != "" {
			a.cfgOverrides["subscriber"] = req.SubscriberURL
		}
		if req.ColorURL != "" {
			a.cfgOverrides["color"] = req.ColorURL
		}
		a.cfgMu.Unlock()

		// If smiley or face URLs changed update the derived headless URLs used by
		// health checks and proxying. Other handlers call svcURL() so pick up
		// overrides automatically at call time — no extra work needed.
		a.logger.Info("config overrides applied via admin UI",
			"face_mode", a.cfgOverrides["faceMode"],
			"smiley_url", a.cfgOverrides["smiley"],
			"color_url", a.cfgOverrides["color"],
			"face_url", a.cfgOverrides["face"],
			"gui_url", a.cfgOverrides["gui"],
		)
		a.writeJSON(w, map[string]string{"message": "config updated"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSmiley proxies GET and PUT to the smiley service.
func (a *AdminProvider) handleSmiley(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return the known smiley names and their unicode values
		smileys := map[string]string{
			"Grinning":    "&#x1F603;",
			"Sleeping":    "&#x1F634;",
			"Cursing":     "&#x1F92C;",
			"Kaboom":      "&#x1F92F;",
			"HeartEyes":   "&#x1F60D;",
			"Neutral":     "&#x1F610;",
			"RollingEyes": "&#x1F644;",
			"Screaming":   "&#x1F631;",
			"Vomiting":    "&#x1F92E;",
		}
		a.writeJSON(w, smileys)

	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read body", http.StatusBadRequest)
			return
		}

		// Optional "pods" field selects specific pod IPs; omit for all pods.
		var req struct {
			Which  string   `json:"which"`
			Smiley string   `json:"smiley"`
			Pods   []string `json:"pods,omitempty"`
		}
		json.Unmarshal(body, &req)

		// Build target URLs — specific pods or all via headless DNS.
		// ExternalWorkloads are reached directly on their declared port (not podPort).
		ewPorts := a.ewPortIndex()
		var targets []string
		if len(req.Pods) > 0 {
			for _, ip := range req.Pods {
				port := a.podPort
				if p, ok := ewPorts[ip]; ok {
					port = p
				}
				targets = append(targets, fmt.Sprintf("http://%s:%s/", ip, port))
			}
		} else {
			for _, u := range a.discoverPodURLs(a.smileyHeadless, a.smileyURL) {
				if !strings.HasSuffix(u, "/") {
					u += "/"
				}
				targets = append(targets, u)
			}
		}

		// Strip "pods" from forwarded body
		fwdBody, _ := json.Marshal(map[string]string{"which": req.Which, "smiley": req.Smiley})

		type podRes struct{ ok bool }
		results := make([]podRes, len(targets))
		var wg sync.WaitGroup
		for i, target := range targets {
			wg.Add(1)
			go func(i int, target string) {
				defer wg.Done()
				fwdReq, err2 := http.NewRequestWithContext(r.Context(), http.MethodPut, target, bytes.NewReader(fwdBody))
				if err2 != nil {
					return
				}
				fwdReq.Header.Set("Content-Type", "application/json")
				c := &http.Client{Timeout: 5 * time.Second}
				resp, err2 := c.Do(fwdReq)
				if err2 == nil {
					results[i] = podRes{ok: resp.StatusCode < 300}
					resp.Body.Close()
				}
			}(i, target)
		}
		wg.Wait()

		succeeded := 0
		for _, res := range results {
			if res.ok {
				succeeded++
			}
		}
		if succeeded == 0 && len(targets) > 0 {
			a.logger.Warn("smiley apply failed — service unreachable",
				"emoji", req.Smiley, "pods_attempted", len(targets))
			http.Error(w, "smiley service unreachable", http.StatusBadGateway)
			return
		}
		a.logger.Info("smiley applied",
			"emoji", req.Smiley, "which", req.Which,
			"pods", len(targets), "succeeded", succeeded)
		a.writeJSON(w, map[string]interface{}{"pods": len(targets), "succeeded": succeeded})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Emojivoto integration ---

func (a *AdminProvider) handleEmojivoto(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.emojivotoMu.RLock()
		resp := emojivotoConfigResponse{
			Endpoint:      a.emojivotoEndpoint,
			Enabled:       a.emojivotoEnabled,
			UpdateSmileys: a.emojivotoUpdateSmileys,
			Status:        a.emojivotoStatus,
			Error:         a.emojivotoError,
			Leader:        a.emojivotoLeader,
			Leaderboard:   a.emojivotoLeaderboard,
			SelectedPods:  a.emojivotoSelectedPods,
		}
		if resp.Status == "" {
			resp.Status = "unconfigured"
		}
		if resp.SelectedPods == nil {
			resp.SelectedPods = []string{}
		}
		a.emojivotoMu.RUnlock()
		a.writeJSON(w, resp)

	case http.MethodPut:
		var req struct {
			Endpoint      string   `json:"endpoint"`
			Enabled       bool     `json:"enabled"`
			UpdateSmileys bool     `json:"updateSmileys"`
			SelectedPods  []string `json:"selectedPods"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		a.emojivotoMu.Lock()
		a.emojivotoEndpoint = req.Endpoint
		a.emojivotoEnabled = req.Enabled
		a.emojivotoUpdateSmileys = req.UpdateSmileys
		if req.SelectedPods != nil {
			a.emojivotoSelectedPods = req.SelectedPods
		} else {
			a.emojivotoSelectedPods = []string{}
		}
		if !req.Enabled || req.Endpoint == "" {
			a.emojivotoStatus = "unconfigured"
			a.emojivotoLeader = ""
			a.emojivotoLeaderboard = nil
			a.emojivotoError = ""
		}
		// Snapshot state needed for the immediate broadcast below.
		currentLeader := a.emojivotoLeader
		currentSelected := append([]string{}, a.emojivotoSelectedPods...)
		a.emojivotoMu.Unlock()

		// When updateSmileys is turned on (or pods are changed while it's on),
		// push the current leader immediately — don't wait for the next leader change.
		if req.UpdateSmileys && currentLeader != "" {
			go a.broadcastEmojivotoLeader(currentLeader, currentSelected)
		}

		a.logger.Info("emojivoto config updated",
			"endpoint", req.Endpoint,
			"enabled", req.Enabled,
			"update_smileys", req.UpdateSmileys,
			"selected_pods", len(req.SelectedPods))
		a.writeJSON(w, map[string]string{"message": "emojivoto config updated"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// StartEmojivotoPoller polls the emojivoto leaderboard every 5 seconds when enabled.
// When the leader changes, it broadcasts the new emoji to the configured smiley pods.
func (a *AdminProvider) StartEmojivotoPoller() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			a.emojivotoMu.RLock()
			endpoint := a.emojivotoEndpoint
			enabled := a.emojivotoEnabled
			a.emojivotoMu.RUnlock()
			if !enabled || endpoint == "" {
				continue
			}
			a.pollEmojivoto(endpoint)
		}
	}()
}

func (a *AdminProvider) pollEmojivoto(endpoint string) {
	url := strings.TrimRight(endpoint, "/") + "/api/leaderboard"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		a.emojivotoMu.Lock()
		a.emojivotoStatus = "error"
		a.emojivotoError = err.Error()
		a.emojivotoMu.Unlock()
		a.logger.Warn("emojivoto poll failed", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		a.emojivotoMu.Lock()
		a.emojivotoStatus = "error"
		a.emojivotoError = errMsg
		a.emojivotoMu.Unlock()
		return
	}

	// Votes may be a JSON string ("42") or a JSON number (42) depending on the version.
	// Decode with interface{} so both cases are handled without type errors.
	var raw []struct {
		Votes     interface{} `json:"votes"`
		Unicode   string      `json:"unicode"`
		Shortcode string      `json:"shortcode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		a.emojivotoMu.Lock()
		a.emojivotoStatus = "error"
		a.emojivotoError = "invalid response: " + err.Error()
		a.emojivotoMu.Unlock()
		return
	}
	entries := make([]emojivotoEntry, len(raw))
	for i, r := range raw {
		var votes string
		switch v := r.Votes.(type) {
		case string:
			votes = v
		case float64:
			votes = strconv.Itoa(int(v))
		default:
			votes = fmt.Sprintf("%v", r.Votes)
		}
		entries[i] = emojivotoEntry{Votes: votes, Unicode: r.Unicode, Shortcode: r.Shortcode}
	}
	a.logger.Info("emojivoto leaderboard fetched", "count", len(entries))

	top10 := entries
	if len(top10) > 10 {
		top10 = top10[:10]
	}

	var leader string
	if len(entries) > 0 {
		leader = entries[0].Unicode
	}

	a.emojivotoMu.Lock()
	prevLeader := a.emojivotoLeader
	a.emojivotoLeaderboard = top10
	a.emojivotoLeader = leader
	a.emojivotoStatus = "ok"
	a.emojivotoError = ""
	updateSmileys := a.emojivotoUpdateSmileys
	selectedPods := append([]string{}, a.emojivotoSelectedPods...)
	a.emojivotoMu.Unlock()

	if leader != "" && leader != prevLeader {
		a.logger.Info("emojivoto leader changed", "leader", leader, "prev", prevLeader)
	}
	if updateSmileys && leader != "" {
		a.broadcastEmojivotoLeader(leader, selectedPods)
	}
}

func (a *AdminProvider) broadcastEmojivotoLeader(emoji string, selectedPods []string) {
	ewPorts := a.ewPortIndex()
	var targets []string
	if len(selectedPods) > 0 {
		for _, ip := range selectedPods {
			port := a.podPort
			if p, ok := ewPorts[ip]; ok {
				port = p
			}
			targets = append(targets, fmt.Sprintf("http://%s:%s/", ip, port))
		}
	} else {
		for _, u := range a.discoverPodURLs(a.smileyHeadless, a.smileyURL) {
			if !strings.HasSuffix(u, "/") {
				u += "/"
			}
			targets = append(targets, u)
		}
	}

	if len(targets) == 0 {
		a.logger.Warn("emojivoto broadcast: no smiley pod targets found")
		return
	}

	fwdBody, _ := json.Marshal(map[string]string{"which": "all", "smiley": emoji})

	var wg sync.WaitGroup
	var mu sync.Mutex
	succeeded := 0
	for _, target := range targets {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodPut, target, bytes.NewReader(fwdBody))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			c := &http.Client{Timeout: 5 * time.Second}
			resp, err := c.Do(req)
			if err == nil {
				if resp.StatusCode < 300 {
					mu.Lock()
					succeeded++
					mu.Unlock()
				}
				resp.Body.Close()
			}
		}(target)
	}
	wg.Wait()

	a.logger.Info("emojivoto leader broadcast complete",
		"emoji", emoji,
		"pods", len(targets),
		"succeeded", succeeded)
}

// --- Health check helpers ---

func (a *AdminProvider) checkHealth(url string) *serviceStatus {
	start := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return &serviceStatus{Healthy: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	healthy := resp.StatusCode == http.StatusOK
	s := &serviceStatus{Healthy: healthy, LatencyMs: latency}
	if !healthy {
		s.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return s
}

// checkColorHealth probes the color gRPC service by making a real Center call.
// Any response code other than Unavailable means the server is reachable.
func (a *AdminProvider) checkColorHealth() *serviceStatus {
	if a.colorClient == nil {
		return &serviceStatus{Healthy: false, Error: "gRPC client not configured"}
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := a.colorClient.Center(ctx, &colorpkg.ColorRequest{Row: 0, Column: 0})
	latency := time.Since(start).Milliseconds()
	if err == nil {
		return &serviceStatus{Healthy: true, LatencyMs: latency}
	}
	st, _ := grpcstatus.FromError(err)
	// Unavailable = transport-level failure (no connection); anything else means
	// the server answered and is alive (e.g. Internal from chaos injection is fine).
	if st.Code() != codes.Unavailable {
		return &serviceStatus{Healthy: true, LatencyMs: latency}
	}
	return &serviceStatus{Healthy: false, Error: st.Message()}
}

func (a *AdminProvider) checkDBHealth() *serviceStatus {
	if a.db == nil {
		return &serviceStatus{Healthy: false, Error: "not connected"}
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := a.db.PingContext(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return &serviceStatus{Healthy: false, Error: err.Error()}
	}
	return &serviceStatus{Healthy: true, LatencyMs: latency}
}

func (a *AdminProvider) checkRedisHealth() *serviceStatus {
	if a.redisClient == nil {
		return &serviceStatus{Healthy: false, Error: "not connected"}
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := a.redisClient.Ping(ctx).Err()
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return &serviceStatus{Healthy: false, Error: err.Error()}
	}
	return &serviceStatus{Healthy: true, LatencyMs: latency}
}

func (a *AdminProvider) checkRabbitMQHealth() *serviceStatus {
	if a.rabbitmqMgmtURL == "" {
		return &serviceStatus{Healthy: false, Error: "management API not configured"}
	}
	start := time.Now()
	_, err := a.rabbitmqQueueStats()
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return &serviceStatus{Healthy: false, Error: err.Error()}
	}
	return &serviceStatus{Healthy: true, LatencyMs: latency}
}

// discoverPodURLs resolves the headless service DNS to get all pod IPs.
// Falls back to the plain service URL if DNS fails or returns nothing.
func (a *AdminProvider) discoverPodURLs(headless, fallbackURL string) []string {
	if headless != "" {
		addrs, err := net.LookupHost(headless)
		if err == nil && len(addrs) > 0 {
			urls := make([]string, len(addrs))
			for i, addr := range addrs {
				host := addr
				if strings.Contains(addr, ":") {
					host = "[" + addr + "]" // IPv6
				}
				urls[i] = fmt.Sprintf("http://%s:%s", host, a.podPort)
			}
			return urls
		}
		a.logger.Debug("pod discovery: DNS lookup failed, using service URL",
			"headless", headless, "error", err)
	}
	return []string{fallbackURL}
}

// fetchOnePodControl calls GET /control on a single URL and returns the result.
func (a *AdminProvider) fetchOnePodControl(podURL string) podControlState {
	ps := podControlState{PodIP: podURLToIP(podURL)}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(podURL + "/control")
	if err != nil {
		ps.Error = err.Error()
		return ps
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&ps); err != nil {
		ps.Error = err.Error()
		return ps
	}
	ps.Available = true
	return ps
}

// queryAllPods discovers pods via headless DNS and queries each one in parallel.
// The returned controlState aggregates the per-pod results.
func (a *AdminProvider) queryAllPods(headless, fallbackURL string) controlState {
	return a.queryAllPodsForComponent(headless, fallbackURL, "")
}

// queryAllPodsForComponent is like queryAllPods but also enriches results with
// Kubernetes topology (pod name, node, zone, region) when the K8s client is available.
func (a *AdminProvider) queryAllPodsForComponent(headless, fallbackURL, component string) controlState {
	urls := a.discoverPodURLs(headless, fallbackURL)

	podStates := make([]podControlState, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			podStates[i] = a.fetchOnePodControl(url)
		}(i, url)
	}
	wg.Wait()

	// Enrich with Kubernetes topology when available
	if a.k8s != nil && component != "" {
		if k8sPods, err := a.k8s.ListPodsForComponent(component); err == nil {
			idx := BuildIPIndex(k8sPods)
			for i := range podStates {
				if info, ok := idx[podStates[i].PodIP]; ok {
					podStates[i].PodName = info.Name
					podStates[i].Node = info.Node
					podStates[i].Zone = info.Zone
					podStates[i].Region = info.Region
				}
			}
		}
	}

	cs := controlState{PodCount: len(podStates), Pods: podStates}
	for _, p := range podStates {
		if p.Available {
			cs.Available = true
			// Primary (first available) pod drives the top-level values used by the UI slider
			if cs.PublishIntervalMs == 0 && p.PublishIntervalMs > 0 {
				cs.PublishIntervalMs = p.PublishIntervalMs
				cs.PublishConcurrency = p.PublishConcurrency
			}
			// Use the first available pod's DB counters as aggregate
			// (all pods share the same MySQL view, so values should be identical)
			if cs.Queued == 0 && p.Queued > 0 {
				cs.Pending = p.Pending
				cs.Queued = p.Queued
				cs.Acknowledged = p.Acknowledged
			}
		}
	}
	// Paused = ALL available pods are paused
	availCount := 0
	pausedCount := 0
	for _, p := range podStates {
		if p.Available {
			availCount++
			if p.Paused {
				pausedCount++
			}
		}
	}
	if availCount > 0 {
		cs.Paused = pausedCount == availCount
	}
	if !cs.Available {
		cs.Error = fmt.Sprintf("%d pod(s) unreachable", len(podStates))
	}
	return cs
}

// handleControls queries all publisher and subscriber pods in parallel.
func (a *AdminProvider) handleControls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type result struct {
		key   string
		state controlState
	}
	ch := make(chan result, 2)

	go func() {
		cs := a.queryAllPodsForComponent(a.publisherHeadless, a.svcURL("publisher", a.publisherURL), "face-publisher")
		ch <- result{"publisher", cs}
	}()
	go func() {
		cs := a.queryAllPodsForComponent(a.subscriberHeadless, a.svcURL("subscriber", a.subscriberURL), "face-subscriber")
		ch <- result{"subscriber", cs}
	}()

	out := controlsResponse{}
	for i := 0; i < 2; i++ {
		res := <-ch
		switch res.key {
		case "publisher":
			out.Publisher = res.state
		case "subscriber":
			out.Subscriber = res.state
		}
	}
	a.writeJSON(w, out)
}

// handleControlProxy broadcasts GET/PUT to every pod of the target service.
// For GET it returns the first pod's response; for PUT it sends to all pods.
func (a *AdminProvider) handleControlProxy(headless *string, serviceURL func() string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// GET: just proxy to the service (single pod is fine for status checks)
			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Get(serviceURL() + "/control")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			w.Write(body)

		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "could not read body", http.StatusBadRequest)
				return
			}

			hl := ""
			if headless != nil {
				hl = *headless
			}
			podURLs := a.discoverPodURLs(hl, serviceURL())

			// Decode the request so we can log what changed
			var ctrlReq struct {
				Paused            *bool  `json:"paused"`
				PublishIntervalMs *int64 `json:"publishIntervalMs"`
			}
			json.Unmarshal(body, &ctrlReq)

			// Broadcast PUT to every pod in parallel
			type podResult struct {
				statusCode int
				err        error
			}
			results := make([]podResult, len(podURLs))
			var wg sync.WaitGroup
			for i, pu := range podURLs {
				wg.Add(1)
				go func(i int, pu string) {
					defer wg.Done()
					req, err := http.NewRequestWithContext(r.Context(), http.MethodPut,
						pu+"/control", bytes.NewReader(body))
					if err != nil {
						results[i] = podResult{err: err}
						return
					}
					req.Header.Set("Content-Type", "application/json")
					client := &http.Client{Timeout: 5 * time.Second}
					resp, err := client.Do(req)
					if err != nil {
						results[i] = podResult{err: err}
						return
					}
					resp.Body.Close()
					results[i] = podResult{statusCode: resp.StatusCode}
				}(i, pu)
			}
			wg.Wait()

			// Summarise: success if all pods responded OK
			succeeded, failed := 0, 0
			for _, res := range results {
				if res.err == nil && res.statusCode >= 200 && res.statusCode < 300 {
					succeeded++
				} else {
					failed++
				}
			}

			// Determine which service this targets from the URL path
			svcName := "service"
			if strings.Contains(r.URL.Path, "publisher") {
				svcName = "publisher"
			} else if strings.Contains(r.URL.Path, "subscriber") {
				svcName = "subscriber"
			}
			logArgs := []any{"service", svcName, "pods", len(podURLs), "succeeded", succeeded}
			if ctrlReq.Paused != nil {
				if *ctrlReq.Paused {
					logArgs = append(logArgs, "action", "pause")
				} else {
					logArgs = append(logArgs, "action", "resume")
				}
			}
			if ctrlReq.PublishIntervalMs != nil {
				logArgs = append(logArgs, "interval_ms", *ctrlReq.PublishIntervalMs)
			}
			a.logger.Info("control updated", logArgs...)

			a.writeJSON(w, map[string]interface{}{
				"pods":      len(podURLs),
				"succeeded": succeeded,
				"failed":    failed,
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// chaosState mirrors the JSON returned by the /chaos endpoint on each service.
type chaosState struct {
	ErrorFraction int     `json:"errorFraction"`
	LatchFraction int     `json:"latchFraction"`
	MaxRate       float64 `json:"maxRate"`
	DelayBuckets  []int   `json:"delayBuckets"`
	Latched       bool    `json:"latched"`
	Available     bool    `json:"available"`
	Error         string  `json:"error,omitempty"`
}

// handleAllChaos returns the current chaos state for all services in a single call.
// GET /api/chaos → { smiley: {...}, color: {...}, face: {...}, publisher: {...}, subscriber: {...} }
func (a *AdminProvider) handleAllChaos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type namedResult struct {
		name  string
		state chaosState
	}
	type serviceSpec struct {
		name    string
		url     func() string // nil = use gRPC
		useGRPC bool
	}
	services := []serviceSpec{
		{"smiley", func() string { return a.svcURL("smiley", a.smileyURL) }, false},
		{"color", nil, true}, // gRPC via existing colorClient
		{"face", func() string { return a.svcURL("face", a.faceURL) }, false},
		{"publisher", func() string { return a.svcURL("publisher", a.publisherURL) }, false},
		{"subscriber", func() string { return a.svcURL("subscriber", a.subscriberURL) }, false},
	}

	ch := make(chan namedResult, len(services))
	for _, svc := range services {
		svc := svc
		go func() {
			var st chaosState
			if svc.useGRPC {
				st = a.fetchColorChaosStateGRPC()
			} else {
				st = a.fetchChaosState(svc.url())
			}
			ch <- namedResult{svc.name, st}
		}()
	}

	out := make(map[string]chaosState, len(services))
	for range services {
		res := <-ch
		out[res.name] = res.state
	}
	a.writeJSON(w, out)
}

func (a *AdminProvider) fetchChaosState(serviceURL string) chaosState {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(serviceURL + "/chaos")
	if err != nil {
		return chaosState{Available: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	var state chaosState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return chaosState{Available: false, Error: err.Error()}
	}
	state.Available = true
	return state
}

// fetchColorChaosStateGRPC queries the color service's chaos state via gRPC.
// The color service no longer has an HTTP chaos sidecar — chaos is a first-class
// gRPC method on the same port 8000 connection used for UpdateColor.
func (a *AdminProvider) fetchColorChaosStateGRPC() chaosState {
	if a.colorClient == nil {
		return chaosState{Available: false, Error: "color gRPC client not connected"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := a.colorClient.GetChaos(ctx, &colorpkg.ChaosRequest{})
	if err != nil {
		return chaosState{Available: false, Error: err.Error()}
	}
	buckets := make([]int, len(resp.DelayBuckets))
	for i, v := range resp.DelayBuckets {
		buckets[i] = int(v)
	}
	return chaosState{
		Available:     true,
		ErrorFraction: int(resp.ErrorFraction),
		LatchFraction: int(resp.LatchFraction),
		MaxRate:       float64(resp.MaxRate),
		DelayBuckets:  buckets,
		Latched:       resp.Latched,
	}
}

// handleColorChaosProxy handles GET/PUT /api/chaos/color via gRPC instead of HTTP.
// Uses the existing color gRPC connection (port 8000) — no HTTP chaos sidecar needed.
func (a *AdminProvider) handleColorChaosProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			st := a.fetchColorChaosStateGRPC()
			a.writeJSON(w, st)

		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "could not read body", http.StatusBadRequest)
				return
			}
			// Parse the standard chaos body (same JSON schema used by other services)
			var req struct {
				ErrorFraction *int     `json:"errorFraction"`
				LatchFraction *int     `json:"latchFraction"`
				MaxRate       *float64 `json:"maxRate"`
				DelayBuckets  []int    `json:"delayBuckets"`
				ForceUnlatch  bool     `json:"forceUnlatch"`
				Pods          []string `json:"pods,omitempty"`
			}
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			// Build the gRPC request (zero values = "don't change"; caller sends explicit zeros to clear)
			grpcReq := &colorpkg.ChaosState{ForceUnlatch: req.ForceUnlatch}
			if req.ErrorFraction != nil {
				grpcReq.ErrorFraction = int32(*req.ErrorFraction)
			}
			if req.LatchFraction != nil {
				grpcReq.LatchFraction = int32(*req.LatchFraction)
			}
			if req.MaxRate != nil {
				grpcReq.MaxRate = float32(*req.MaxRate)
			}
			if req.DelayBuckets != nil {
				for _, v := range req.DelayBuckets {
					grpcReq.DelayBuckets = append(grpcReq.DelayBuckets, int32(v))
				}
			}

			// Discover all color pod IPs — prefer K8s API (all pods across all zones/instances)
			// over headless DNS, which may only return one IP due to gRPC connection pinning.
			podIPs := req.Pods
			if len(podIPs) == 0 && a.k8s != nil {
				if byComp, err2 := a.k8s.ListAllBackendComponents(); err2 == nil {
					for comp, pods := range byComp {
						if comp == "color" || strings.HasPrefix(comp, "color") {
							for _, p := range pods {
								podIPs = append(podIPs, p.IP)
							}
						}
					}
				}
			}
			// Fallback: headless DNS
			if len(podIPs) == 0 {
				if addrs, dnsErr := net.LookupHost(a.colorHeadless); dnsErr == nil {
					podIPs = addrs
				}
			}

			succeeded := 0
			var wg sync.WaitGroup
			var mu sync.Mutex
			for _, ip := range podIPs {
				wg.Add(1)
				go func(ip string) {
					defer wg.Done()
					addr := ip + ":" + a.podPort
					conn, connErr := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
					if connErr != nil {
						a.logger.Debug("color chaos gRPC: dial failed", "ip", ip, "error", connErr)
						return
					}
					defer conn.Close()
					ctx2, cancel := context.WithTimeout(r.Context(), 5*time.Second)
					defer cancel()
					if _, rpcErr := colorpkg.NewColorServiceClient(conn).UpdateChaos(ctx2, grpcReq); rpcErr == nil {
						mu.Lock()
						succeeded++
						mu.Unlock()
					} else {
						a.logger.Debug("color chaos gRPC: UpdateChaos failed", "ip", ip, "error", rpcErr)
					}
				}(ip)
			}
			wg.Wait()

			if len(podIPs) == 0 {
				// Last resort: use the existing persistent VIP connection
				ctx2, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()
				if _, vipErr := a.colorClient.UpdateChaos(ctx2, grpcReq); vipErr == nil {
					succeeded = 1
				}
				podIPs = []string{"vip"}
				a.logger.Warn("color chaos: no pod IPs discovered, fell back to VIP (only one pod updated)")
			}

			a.logger.Info("chaos updated via gRPC", "service", "color",
				"pods", len(podIPs), "succeeded", succeeded)
			a.writeJSON(w, map[string]interface{}{
				"pods": len(podIPs), "succeeded": succeeded, "failed": len(podIPs) - succeeded,
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleChaosProxy proxies GET/PUT /api/chaos/{service} to the target service's /chaos endpoint.
// For GET it proxies to the service VIP and returns the raw response.
// For PUT it broadcasts to every pod in parallel (via headless DNS) then returns a summary.
// chaosPodPort overrides a.podPort for per-pod URL construction — pass "8001" for color,
// which runs its chaos HTTP server on a separate port alongside gRPC.
func (a *AdminProvider) handleChaosProxy(headless *string, serviceURL func() string, chaosPodPort ...string) http.HandlerFunc {
	podPort := a.podPort
	if len(chaosPodPort) > 0 && chaosPodPort[0] != "" {
		podPort = chaosPodPort[0]
	}
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Get(serviceURL() + "/chaos")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			w.Write(body)

		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "could not read body", http.StatusBadRequest)
				return
			}

			// Optional pods field: specific IPs to target instead of broadcasting to all.
			var podFilter struct {
				Pods []string `json:"pods,omitempty"`
			}
			json.Unmarshal(body, &podFilter)

			var podURLs []string
			if len(podFilter.Pods) > 0 {
				// Per-pod targeting: build URLs from the specified IPs
				for _, ip := range podFilter.Pods {
					podURLs = append(podURLs, fmt.Sprintf("http://%s:%s", ip, podPort))
				}
				// Strip pods field from forwarded body so services don't see it
				if stripped, err2 := stripField(body, "pods"); err2 == nil {
					body = stripped
				}
			} else {
				// K8s API — discovers ALL instances (e.g. smiley, smiley2, smiley3) like
				// handleColorChaosProxy does for color. Component prefix is derived from the
				// headless service name by stripping the "-headless" suffix.
				if a.k8s != nil && headless != nil && *headless != "" {
					componentPrefix := strings.TrimSuffix(*headless, "-headless")
					if byComp, err2 := a.k8s.ListAllBackendComponents(); err2 == nil {
						ewPorts := a.ewPortIndex()
						for comp, pods := range byComp {
							if comp == componentPrefix || strings.HasPrefix(comp, componentPrefix) {
								for _, p := range pods {
									port := podPort
									if ep, ok := ewPorts[p.IP]; ok {
										port = ep
									}
									podURLs = append(podURLs, fmt.Sprintf("http://%s:%s", p.IP, port))
								}
							}
						}
					}
				}

				// Fallback: headless DNS (used when K8s client unavailable or returns nothing)
				if len(podURLs) == 0 {
					hl := ""
					if headless != nil {
						hl = *headless
					}
					if hl != "" {
						addrs, dnsErr := net.LookupHost(hl)
						if dnsErr == nil && len(addrs) > 0 {
							for _, addr := range addrs {
								host := addr
								if strings.Contains(addr, ":") {
									host = "[" + addr + "]"
								}
								podURLs = append(podURLs, fmt.Sprintf("http://%s:%s", host, podPort))
							}
						}
					}
				}

				if len(podURLs) == 0 {
					podURLs = []string{serviceURL()}
				}
			}

			type podResult struct {
				statusCode int
				err        error
			}
			results := make([]podResult, len(podURLs))
			var wg sync.WaitGroup
			for i, pu := range podURLs {
				wg.Add(1)
				go func(i int, pu string) {
					defer wg.Done()
					req, err := http.NewRequestWithContext(r.Context(), http.MethodPut,
						pu+"/chaos", bytes.NewReader(body))
					if err != nil {
						results[i] = podResult{err: err}
						return
					}
					req.Header.Set("Content-Type", "application/json")
					client := &http.Client{Timeout: 5 * time.Second}
					resp, err := client.Do(req)
					if err != nil {
						results[i] = podResult{err: err}
						return
					}
					resp.Body.Close()
					results[i] = podResult{statusCode: resp.StatusCode}
				}(i, pu)
			}
			wg.Wait()

			succeeded, failed := 0, 0
			for _, res := range results {
				if res.err == nil && res.statusCode >= 200 && res.statusCode < 300 {
					succeeded++
				} else {
					failed++
				}
			}

			// Log which service and what changed
			svcName := "service"
			if strings.Contains(r.URL.Path, "smiley") {
				svcName = "smiley"
			} else if strings.Contains(r.URL.Path, "color") {
				svcName = "color"
			} else if strings.Contains(r.URL.Path, "face") && !strings.Contains(r.URL.Path, "publisher") && !strings.Contains(r.URL.Path, "subscriber") {
				svcName = "face"
			} else if strings.Contains(r.URL.Path, "publisher") {
				svcName = "publisher"
			} else if strings.Contains(r.URL.Path, "subscriber") {
				svcName = "subscriber"
			}
			var req struct {
				ErrorFraction *int     `json:"errorFraction"`
				LatchFraction *int     `json:"latchFraction"`
				MaxRate       *float64 `json:"maxRate"`
				ForceUnlatch  bool     `json:"forceUnlatch"`
			}
			json.Unmarshal(body, &req)
			logArgs := []any{"service", svcName, "pods", len(podURLs), "succeeded", succeeded}
			if req.ErrorFraction != nil {
				logArgs = append(logArgs, "errorFraction", *req.ErrorFraction)
			}
			if req.LatchFraction != nil {
				logArgs = append(logArgs, "latchFraction", *req.LatchFraction)
			}
			if req.MaxRate != nil {
				logArgs = append(logArgs, "maxRate", *req.MaxRate)
			}
			if req.ForceUnlatch {
				logArgs = append(logArgs, "forceUnlatch", true)
			}
			a.logger.Info("chaos updated", logArgs...)

			a.writeJSON(w, map[string]interface{}{
				"pods":      len(podURLs),
				"succeeded": succeeded,
				"failed":    failed,
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// ewPortIndex returns a map from ExternalWorkload IP → declared port string.
// Used when applying smiley/color to specific pod IPs — EWs must be reached
// on their own port (typically 80), not the K8s pod port (8000).
// Returns nil when K8s is unavailable.
func (a *AdminProvider) ewPortIndex() map[string]string {
	if a.k8s == nil {
		return nil
	}
	idx := make(map[string]string)
	for _, pods := range a.k8s.ListExternalWorkloadComponents() {
		for _, p := range pods {
			if p.Port != "" {
				idx[p.IP] = p.Port
			}
		}
	}
	return idx
}

// stripField removes a single top-level key from a JSON object body.
// Used to strip the admin-only "pods" field before forwarding to services.
func stripField(body []byte, field string) ([]byte, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return body, err
	}
	delete(m, field)
	return json.Marshal(m)
}

// podURLToIP strips the scheme and port from a pod URL for display.
func podURLToIP(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	if host, _, err := net.SplitHostPort(url); err == nil {
		return strings.Trim(host, "[]")
	}
	return url
}

// handleFacePods returns face pods via K8s API (component=face label, edge-type).
// Face uses component-type:edge rather than backend, so ListAllBackendComponents
// does not find it. Falls back to the face VIP URL (no headless service for face).
// sortPodsStable orders pods deterministically (by name, then IP) so the UI's
// pod pills/selectors don't reshuffle on every poll. K8s list + map iteration
// order is not stable across calls.
func sortPodsStable(pods []PodTopology) {
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Name != pods[j].Name {
			return pods[i].Name < pods[j].Name
		}
		return pods[i].IP < pods[j].IP
	})
}

func (a *AdminProvider) handleFacePods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.k8s != nil {
		if pods, err := a.k8s.ListPodsForComponent("face"); err == nil && len(pods) > 0 {
			sortPodsStable(pods)
			a.writeJSON(w, pods)
			return
		}
	}
	// Fallback: no K8s / no pods found — return empty so UI shows "no pods"
	a.writeJSON(w, []PodTopology{})
}

// handleServicePods returns a handler that lists pods for a given service.
// Returns K8s pods for the named component AND all components that share the
// same prefix (smiley → smiley + smiley2 + smiley3, color → color + color2 …).
// Also includes Linkerd ExternalWorkloads. Falls back to headless DNS when K8s
// is unavailable.
func (a *AdminProvider) handleServicePods(svcName string, headless *string, fallback func() string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.k8s != nil {
			var pods []PodTopology

			// Query all backend components and collect every instance whose name
			// starts with svcName (e.g. "smiley" matches smiley, smiley2, smiley3).
			if byComp, err := a.k8s.ListAllBackendComponents(); err == nil {
				for comp, kpods := range byComp {
					if comp == svcName || strings.HasPrefix(comp, svcName) {
						pods = append(pods, kpods...)
					}
				}
			}

			// Also include ExternalWorkloads registered for any matching component
			ewByComp := a.k8s.ListExternalWorkloadComponents()
			for comp, ewPods := range ewByComp {
				if comp == svcName || strings.HasPrefix(comp, svcName) {
					pods = append(pods, ewPods...)
				}
			}

			if len(pods) > 0 {
				// Stable order across polls — byComp is a map (random iteration order),
				// which otherwise makes pod pills jump around in the UI between refreshes.
				sortPodsStable(pods)
				a.writeJSON(w, pods)
				return
			}
		}

		// Fall back to headless DNS (IP only, no topology or type info)
		hl := ""
		if headless != nil {
			hl = *headless
		}
		urls := a.discoverPodURLs(hl, fallback())
		if len(urls) == 0 {
			a.writeJSON(w, []PodTopology{})
			return
		}
		pods := make([]PodTopology, len(urls))
		for i, u := range urls {
			ip := podURLToIP(u)
			pods[i] = PodTopology{IP: ip, Name: ip}
		}
		a.writeJSON(w, pods)
	}
}

// handleColor returns the known color map (GET) or updates color via gRPC (PUT).
func (a *AdminProvider) handleColor(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		colors := map[string]string{
			"grey":     "#BBBBBB",
			"darkblue": "#4477AA",
			"blue":     "#66CCEE",
			"green":    "#228833",
			"yellow":   "#CCBB44",
			"red":      "#EE6677",
			"purple":   "#AA3377",
		}
		a.writeJSON(w, colors)

	case http.MethodPut:
		var req struct {
			Which string   `json:"which"`
			Color string   `json:"color"`
			Pods  []string `json:"pods,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Build list of gRPC targets.
		// ExternalWorkloads are dialled on their declared port (not podPort).
		ewPorts := a.ewPortIndex()
		var grpcTargets []string
		if len(req.Pods) > 0 {
			for _, ip := range req.Pods {
				port := a.podPort
				if p, ok := ewPorts[ip]; ok {
					port = p
				}
				grpcTargets = append(grpcTargets, ip+":"+port)
			}
		} else {
			for _, u := range a.discoverPodURLs(a.colorHeadless, a.colorGRPCAddr) {
				// discoverPodURLs returns http://ip:port style; strip scheme for gRPC
				addr := strings.TrimPrefix(u, "http://")
				grpcTargets = append(grpcTargets, addr)
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		succeeded := 0
		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstRPCErr error // first non-nil gRPC error for diagnostic reporting
		for _, addr := range grpcTargets {
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				conn, err2 := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err2 != nil {
					return
				}
				defer conn.Close()
				client := colorpkg.NewColorServiceClient(conn)
				_, err2 = client.UpdateColor(ctx, &colorpkg.ColorUpdate{Which: req.Which, Color: req.Color})
				if err2 == nil {
					mu.Lock()
					succeeded++
					mu.Unlock()
				} else {
					mu.Lock()
					if firstRPCErr == nil {
						firstRPCErr = err2
					}
					mu.Unlock()
				}
			}(addr)
		}
		wg.Wait()

		if succeeded == 0 && len(grpcTargets) > 0 {
			// Distinguish between connectivity failures and rejected calls (e.g. unknown
			// color name). InvalidArgument means the service is reachable but rejected
			// the request — return 400 with the actual reason rather than a misleading 502.
			if firstRPCErr != nil && grpcstatus.Code(firstRPCErr) == codes.InvalidArgument {
				errMsg := grpcstatus.Convert(firstRPCErr).Message()
				a.logger.Warn("color apply rejected — invalid argument",
					"color", req.Color, "reason", errMsg)
				http.Error(w, "color rejected: "+errMsg, http.StatusBadRequest)
				return
			}
			a.logger.Warn("color apply failed — service unreachable",
				"color", req.Color, "pods_attempted", len(grpcTargets))
			http.Error(w, "color service unreachable", http.StatusBadGateway)
			return
		}
		a.logger.Info("color applied",
			"color", req.Color, "which", req.Which,
			"pods", len(grpcTargets), "succeeded", succeeded)
		a.writeJSON(w, map[string]interface{}{"pods": len(grpcTargets), "succeeded": succeeded})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFaceProxy proxies GET /face/... to the configured face service so the
// in-page widget can consume faces without needing direct service access.
func (a *AdminProvider) handleFaceProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /face/center?row=0&col=0 → http://face/center?row=0&col=0
	tail := strings.TrimPrefix(r.URL.Path, "/face")
	if tail == "" {
		tail = "/center"
	}
	target := a.svcURL("face", a.faceURL) + tail
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Propagate user header so the face service isn't confused
	req.Header.Set("X-Faces-User", "admin-widget")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Return a well-formed error JSON so the widget can display a state
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"smiley": "&#x1F92C;",
			"color":  "#BBBBBB",
			"errors": []string{err.Error()},
			"status": 503,
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// handleDBStatus returns the MySQL connection state plus per-state row counts.
func (a *AdminProvider) handleDBStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.db == nil {
		a.writeJSON(w, map[string]interface{}{"connected": false, "error": "not configured"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	start := time.Now()
	if err := a.db.PingContext(ctx); err != nil {
		a.writeJSON(w, map[string]interface{}{"connected": false, "error": err.Error()})
		return
	}
	latency := time.Since(start).Milliseconds()

	pending, queued, acked, err := a.dbQueueDepth(ctx)
	if err != nil {
		a.writeJSON(w, map[string]interface{}{"connected": true, "latencyMs": latency, "schemaError": err.Error()})
		return
	}
	a.writeJSON(w, map[string]interface{}{
		"connected":    true,
		"latencyMs":    latency,
		"pending":      pending,
		"queued":       queued,
		"acknowledged": acked,
	})
}

// handleDBMigrate runs the idempotent CREATE TABLE + ALTER TABLE — safe to call any time.
// Useful when the MySQL container was restarted and the table needs to be recreated.
func (a *AdminProvider) handleDBMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.db == nil {
		a.writeJSON(w, map[string]interface{}{"ok": false, "error": "database not connected"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := a.ensureFaceQueueSchema(ctx); err != nil {
		a.writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	a.logger.Info("database migration executed — face_queue table is current")
	a.writeJSON(w, map[string]interface{}{"ok": true, "message": "Migration complete — face_queue table and 3-state ENUM are current"})
}

// ensureFaceQueueSchema runs the idempotent CREATE TABLE + 3-state ENUM ALTER.
// Shared by the migrate and purge handlers so a purge also leaves the table
// ready for the demo.
func (a *AdminProvider) ensureFaceQueueSchema(ctx context.Context) error {
	if _, err := a.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS face_queue (
		id               VARCHAR(36)  NOT NULL,
		smiley           TEXT         NOT NULL,
		color            TEXT         NOT NULL,
		errors           TEXT         NULL,
		state            ENUM('pending','queued','acknowledged') NOT NULL DEFAULT 'pending',
		created_at       BIGINT       NOT NULL,
		acknowledged_at  BIGINT       NULL,
		PRIMARY KEY (id),
		INDEX idx_state_created (state, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	// Ensure 3-state ENUM on existing installs — idempotent
	_, _ = a.db.ExecContext(ctx, `ALTER TABLE face_queue
		MODIFY COLUMN state ENUM('pending','queued','acknowledged') NOT NULL DEFAULT 'pending'`)
	return nil
}

// handleDBPurge deletes all rows from face_queue — a hard reset for demo use.
// Use when you want a clean counter slate without redeploying the pipeline.
func (a *AdminProvider) handleDBPurge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.db == nil {
		a.writeJSON(w, map[string]interface{}{"ok": false, "error": "database not connected"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Ensure the table exists (and the ENUM is current) BEFORE deleting, so a purge
	// also prepares a fresh/empty schema for the demo — no separate "Run Migration"
	// step needed. DELETE would otherwise fail if the table was missing.
	if err := a.ensureFaceQueueSchema(ctx); err != nil {
		a.writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	res, err := a.db.ExecContext(ctx, "DELETE FROM face_queue")
	if err != nil {
		a.writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	a.logger.Info("database purged and schema ensured", "rows_deleted", n)
	a.writeJSON(w, map[string]interface{}{
		"ok":           true,
		"rows_deleted": n,
		"message":      fmt.Sprintf("Purged %d rows; face_queue table ready for the demo", n),
	})
}

// handleQueuePurge removes all messages from the queue backend (RabbitMQ or Redis).
func (a *AdminProvider) handleQueuePurge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	switch a.queueType {
	case "redis":
		if a.redisClient == nil {
			a.writeJSON(w, map[string]interface{}{"ok": false, "error": "Redis not connected"})
			return
		}
		n, err := a.redisClient.Del(ctx, a.redisKey).Result()
		if err != nil {
			a.writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		a.logger.Info("queue purged", "backend", "redis", "key", a.redisKey, "entries_removed", n)
		a.writeJSON(w, map[string]interface{}{"ok": true, "message": fmt.Sprintf("Deleted key %q (%d entries removed)", a.redisKey, n)})

	case "rabbitmq":
		if a.rabbitmqMgmtURL == "" {
			a.writeJSON(w, map[string]interface{}{"ok": false, "error": "RabbitMQ management not configured"})
			return
		}
		// Management API: DELETE /api/queues/{vhost}/{queue}/contents
		purgeURL := a.rabbitmqMgmtURL + "/contents"
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, purgeURL, nil)
		if err != nil {
			a.writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			a.writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			a.logger.Info("queue purged", "backend", "rabbitmq", "queue", a.rabbitmqQueue)
			a.writeJSON(w, map[string]interface{}{"ok": true, "message": fmt.Sprintf("Queue %q purged", a.rabbitmqQueue)})
		} else {
			body, _ := io.ReadAll(resp.Body)
			a.writeJSON(w, map[string]interface{}{"ok": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))})
		}

	default:
		a.writeJSON(w, map[string]interface{}{"ok": false, "error": "unknown queue type: " + a.queueType})
	}
}

// podServing holds what a single pod is currently returning to callers.
type podServing struct {
	IP         string `json:"ip"`
	Name       string `json:"name,omitempty"`
	Zone       string `json:"zone,omitempty"`
	Smiley     string `json:"smiley,omitempty"`     // center HTML entity, e.g. &#x1F603;
	SmileyEdge string `json:"smileyEdge,omitempty"` // edge emoji, only when ≠ center
	Color      string `json:"color,omitempty"`      // center hex, e.g. #66CCEE
	ColorEdge  string `json:"colorEdge,omitempty"`  // edge color, only when ≠ center
	Error      string `json:"error,omitempty"`
}

// handleSmileyState GETs /center on each smiley pod and returns the smiley value
// it is currently configured to serve. Useful for showing per-pod emoji status.
// Discovery uses the K8s API (all pods across all zones + ExternalWorkloads) and
// falls back to headless DNS when K8s is unavailable.
func (a *AdminProvider) handleSmileyState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Prefer K8s API — discovers all smiley* instances (smiley, smiley2, smiley3 …)
	ipIndex := make(map[string]PodTopology)
	var urls []string
	if a.k8s != nil {
		if byComp, err := a.k8s.ListAllBackendComponents(); err == nil {
			for comp, pods := range byComp {
				if comp == "smiley" || strings.HasPrefix(comp, "smiley") {
					for _, p := range pods {
						if _, seen := ipIndex[p.IP]; !seen {
							ipIndex[p.IP] = p
							urls = append(urls, fmt.Sprintf("http://%s:%s", p.IP, a.podPort))
						}
					}
				}
			}
		}
		// Also include ExternalWorkloads for any smiley* component
		for comp, ewPods := range a.k8s.ListExternalWorkloadComponents() {
			if comp == "smiley" || strings.HasPrefix(comp, "smiley") {
				for _, p := range ewPods {
					if _, seen := ipIndex[p.IP]; !seen {
						ipIndex[p.IP] = p
						port := a.podPort
						if p.Port != "" {
							port = p.Port
						}
						urls = append(urls, fmt.Sprintf("http://%s:%s", p.IP, port))
					}
				}
			}
		}
	}
	// Fallback to headless DNS when K8s returned nothing
	if len(urls) == 0 {
		urls = a.discoverPodURLs(a.smileyHeadless, a.svcURL("smiley", a.smileyURL))
	}

	states := make([]podServing, len(urls))
	client := &http.Client{Timeout: 2 * time.Second}
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			ip := podURLToIP(u)
			ps := podServing{IP: ip, Name: ip}
			if info, ok := ipIndex[ip]; ok {
				ps.Name = info.Name
				ps.Zone = info.Zone
			}
			if isLinkLocalIPv6(ip) {
				ps.Error = "state unavailable — link-local IPv6"
				states[i] = ps
				return
			}
			resp, err := client.Get(u + "/center?row=0&col=0")
			if err != nil {
				ps.Error = friendlyNetErr(err, ipIndex[ip].WorkloadType)
				states[i] = ps
				return
			}
			defer resp.Body.Close()
			var result struct {
				Smiley string `json:"smiley"`
			}
			if json.NewDecoder(resp.Body).Decode(&result) == nil {
				ps.Smiley = result.Smiley
			}
			// Edge emoji — record only when it differs from center (best-effort)
			if eResp, eErr := client.Get(u + "/edge?row=1&col=0"); eErr == nil {
				defer eResp.Body.Close()
				var er struct {
					Smiley string `json:"smiley"`
				}
				if json.NewDecoder(eResp.Body).Decode(&er) == nil && er.Smiley != "" && er.Smiley != ps.Smiley {
					ps.SmileyEdge = er.Smiley
				}
			}
			states[i] = ps
		}(i, u)
	}
	wg.Wait()
	a.writeJSON(w, states)
}

// handleColorState calls gRPC Center() on each color pod and returns its current color.
// Discovery uses the K8s API (all pods across all zones + ExternalWorkloads) and
// falls back to headless DNS when K8s is unavailable.
func (a *AdminProvider) handleColorState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Prefer K8s API — discovers all color* instances (color, color2, color3 …)
	ipIndex := make(map[string]PodTopology)
	var urls []string
	if a.k8s != nil {
		if byComp, err := a.k8s.ListAllBackendComponents(); err == nil {
			for comp, pods := range byComp {
				if comp == "color" || strings.HasPrefix(comp, "color") {
					for _, p := range pods {
						if _, seen := ipIndex[p.IP]; !seen {
							ipIndex[p.IP] = p
							urls = append(urls, fmt.Sprintf("http://%s:%s", p.IP, a.podPort))
						}
					}
				}
			}
		}
		for comp, ewPods := range a.k8s.ListExternalWorkloadComponents() {
			if comp == "color" || strings.HasPrefix(comp, "color") {
				for _, p := range ewPods {
					if _, seen := ipIndex[p.IP]; !seen {
						ipIndex[p.IP] = p
						port := a.podPort
						if p.Port != "" {
							port = p.Port
						}
						urls = append(urls, fmt.Sprintf("http://%s:%s", p.IP, port))
					}
				}
			}
		}
	}
	if len(urls) == 0 {
		urls = a.discoverPodURLs(a.colorHeadless, a.colorGRPCAddr)
	}

	states := make([]podServing, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			// u may have http:// prefix from discoverPodURLs — strip it for gRPC
			addr := strings.TrimPrefix(u, "http://")
			ip := podURLToIP(u)
			ps := podServing{IP: ip, Name: ip}
			if info, ok := ipIndex[ip]; ok {
				ps.Name = info.Name
				ps.Zone = info.Zone
			}
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				ps.Error = err.Error()
				states[i] = ps
				return
			}
			defer conn.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			colorClient := colorpkg.NewColorServiceClient(conn)
			resp, err := colorClient.Center(ctx, &colorpkg.ColorRequest{Row: 0, Column: 0})
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Color = resp.Color
			}
			// Edge color — record only when it differs from center (best-effort)
			eCtx, eCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer eCancel()
			if er, eErr := colorClient.Edge(eCtx, &colorpkg.ColorRequest{Row: 1, Column: 0}); eErr == nil {
				if er.Color != "" && er.Color != ps.Color {
					ps.ColorEdge = er.Color
				}
			}
			states[i] = ps
		}(i, u)
	}
	wg.Wait()
	a.writeJSON(w, states)
}

// handleLinkys returns the list of PNG/GIF filenames available in DATA_PATH/linkys/.
// The admin emoji picker uses this to build the Linkerd category dynamically —
// drop any image file into that directory and it appears in the picker automatically.
func (a *AdminProvider) handleLinkys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	linkysDir := filepath.Join(a.dataPath, "linkys")
	entries, err := os.ReadDir(linkysDir)
	if err != nil {
		a.writeJSON(w, []string{})
		return
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".gif") ||
			strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
			names = append(names, e.Name())
		}
	}
	if names == nil {
		names = []string{}
	}
	a.writeJSON(w, names)
}

// --- Static file serving ---

func (a *AdminProvider) serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	absDataPath, err := filepath.Abs(a.dataPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	relPath := strings.TrimPrefix(urlPath, "/")
	candidate := filepath.Join(absDataPath, relPath)
	absCandidate, err := filepath.Abs(filepath.Clean(candidate))
	if err != nil || !strings.HasPrefix(absCandidate, absDataPath+string(os.PathSeparator)) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	info, err := os.Stat(absCandidate)
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	raw, err := os.ReadFile(absCandidate)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ctype := contentType(absCandidate)
	w.Header().Set("Content-Type", ctype)
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		w.Write(raw)
	}
}

func contentType(path string) string {
	switch filepath.Ext(path) {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".ico":
		return "image/x-icon"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func (a *AdminProvider) writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		a.logger.Warn("writeJSON failed", "error", err)
	}
}
