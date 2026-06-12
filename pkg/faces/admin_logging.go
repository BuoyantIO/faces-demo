// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package faces

// admin_logging.go — HTTP access logging middleware + startup summary.
//
// Log levels:
//   WARN  — any 4xx/5xx response
//   INFO  — all mutations (PUT/POST), meaningful GETs (navigation/page-open signals)
//   DEBUG — high-frequency polling GETs that fire every few seconds

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// silentPaths are GET endpoints polled automatically by the browser — noisy at INFO.
var silentPaths = map[string]bool{
	"/api/status":         true, // every 3 s
	"/api/pipeline":       true, // every 3 s
	"/api/controls":       true, // every 3 s (pub/sub mode)
	"/api/infrastructure": true, // every 10 s
	"/api/smileystate":    true, // controls page refresh
	"/api/colorstate":     true, // controls page refresh
	"/healthz":            true, // K8s probes
	"/metrics":            true, // Prometheus scrapes
}

// responseRecorder wraps http.ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	code    int
	written bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.code = code
		rr.written = true
	}
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) statusCode() int {
	if rr.code == 0 {
		return http.StatusOK // implicit 200 when WriteHeader was never called
	}
	return rr.code
}

// loggingMiddleware wraps the admin mux with structured access logging.
func (a *AdminProvider) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		status := rw.statusCode()
		ms := time.Since(start).Milliseconds()

		level := slog.LevelInfo
		switch {
		case status >= 400:
			level = slog.LevelWarn // errors always visible
		case r.Method == http.MethodGet &&
			(silentPaths[r.URL.Path] ||
				isStaticAsset(r.URL.Path) ||
				strings.HasPrefix(r.URL.Path, "/face/")):
			level = slog.LevelDebug // noisy polling / static files
		}

		a.logger.Log(r.Context(), level, "http",
			"method", r.Method,
			"path",   r.URL.Path,
			"status", status,
			"ms",     ms,
		)
	})
}

// isLinkerdMeshed returns true when the Linkerd sidecar proxy is present in this
// pod. Detection: the Linkerd proxy always binds its admin server to localhost:4191.
// A successful TCP connect means the proxy is running; any error means no mesh.
func isLinkerdMeshed() bool {
	conn, err := net.DialTimeout("tcp", "localhost:4191", 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// isStaticAsset returns true for paths that serve the SPA bundle.
func isStaticAsset(path string) bool {
	if path == "/" {
		return true
	}
	for _, ext := range []string{".css", ".js", ".html", ".ico", ".png", ".gif", ".webp", ".svg"} {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// logStartup emits a structured summary of the admin configuration at startup.
// Called once from Start() so operators can confirm everything is wired correctly.
func (a *AdminProvider) logStartup() {
	mode := a.effectiveFaceMode()

	a.logger.Info("faces-admin starting",
		"mode",          mode,
		"k8s_available", a.k8s != nil,
	)

	if a.k8s != nil {
		a.logger.Info("faces-admin: kubernetes",
			"namespace", a.k8s.namespace,
		)
	}

	a.logger.Info("faces-admin: services",
		"smiley", a.smileyURL,
		"color",  a.colorURL,
		"gui",    a.guiURL,
		"face",   a.faceURL,
	)

	if mode == "pubsub" {
		a.logger.Info("faces-admin: pub/sub pipeline",
			"queue_backend", a.queueType,
			"publisher",     a.publisherURL,
			"subscriber",    a.subscriberURL,
		)
		if a.db != nil {
			a.logger.Info("faces-admin: MySQL connected")
		}
	}
}
