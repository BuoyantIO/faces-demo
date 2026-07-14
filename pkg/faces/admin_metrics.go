// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package faces

// admin_metrics.go — Prometheus text-format /metrics endpoint for faces-admin.
//
// Exposes pipeline health regardless of mode so Prometheus/Grafana can track
// the Faces demo over time.  Add faces-admin as a scrape target and you get
// historical charts for free.
//
// Scrape config example:
//   - job_name: faces-admin
//     static_configs:
//       - targets: ['faces-admin:80']
//     metrics_path: /metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ── Prometheus text builder ────────────────────────────────────────────────

type promWriter struct{ sb strings.Builder }

func (p *promWriter) gauge(name, help string) {
	p.sb.WriteString("# HELP " + name + " " + help + "\n")
	p.sb.WriteString("# TYPE " + name + " gauge\n")
}

func (p *promWriter) sample(name string, labels [][2]string, value float64) {
	p.sb.WriteString(name)
	if len(labels) > 0 {
		p.sb.WriteByte('{')
		for i, lp := range labels {
			if i > 0 {
				p.sb.WriteByte(',')
			}
			p.sb.WriteString(lp[0])
			p.sb.WriteString(`="`)
			p.sb.WriteString(strings.ReplaceAll(lp[1], `"`, `\"`))
			p.sb.WriteByte('"')
		}
		p.sb.WriteByte('}')
	}
	p.sb.WriteString(fmt.Sprintf(" %g\n", value))
}

func (p *promWriter) nl() { p.sb.WriteByte('\n') }

func boolGauge(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// ── Handler ────────────────────────────────────────────────────────────────

func (a *AdminProvider) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	// ── Collect data in parallel ───────────────────────────────────────────

	type healthResult struct {
		name   string
		status *serviceStatus
	}

	// Services to health-check depend on mode
	serviceList := []struct{ name, url string }{
		{"smiley", a.smileyURL + "/healthz"},
		{"gui", a.guiURL + "/ready"},
	}
	if a.faceMode == "pubsub" {
		serviceList = append(serviceList,
			struct{ name, url string }{"face-publisher", a.publisherURL + "/healthz"},
			struct{ name, url string }{"face-subscriber", a.subscriberURL + "/healthz"},
		)
	} else {
		serviceList = append(serviceList, struct{ name, url string }{"face", a.faceURL + "/healthz"})
	}

	healthCh := make(chan healthResult, len(serviceList)+3)
	for _, svc := range serviceList {
		go func(name, url string) {
			healthCh <- healthResult{name, a.checkHealth(url)}
		}(svc.name, svc.url)
	}
	go func() { healthCh <- healthResult{"color", a.checkColorHealth()} }()
	if a.faceMode == "pubsub" {
		go func() { healthCh <- healthResult{"mysql", a.checkDBHealth()} }()
		if a.queueType == "rabbitmq" {
			go func() { healthCh <- healthResult{"rabbitmq", a.checkRabbitMQHealth()} }()
		} else {
			go func() { healthCh <- healthResult{"redis", a.checkRedisHealth()} }()
		}
	}

	// DB depth and queue stats (pubsub only) — parallel with health checks
	type dbResult struct {
		pending, queued, acknowledged int64
		err                           error
	}
	dbCh := make(chan dbResult, 1)
	type queueResult struct {
		qs  *queueStats
		err error
	}
	qCh := make(chan queueResult, 1)

	if a.faceMode == "pubsub" {
		go func() {
			p, q, ack, err := a.dbQueueDepth(ctx)
			dbCh <- dbResult{p, q, ack, err}
		}()
		go func() {
			qs, err := a.queueStatsForMetrics(ctx)
			qCh <- queueResult{qs, err}
		}()
	} else {
		dbCh <- dbResult{}
		qCh <- queueResult{}
	}

	// Publisher/subscriber per-pod control state (pubsub only)
	var pubCtrl, subCtrl controlState
	var wg sync.WaitGroup
	if a.faceMode == "pubsub" {
		wg.Add(2)
		go func() { defer wg.Done(); pubCtrl = a.queryAllPods(a.publisherHeadless, a.publisherURL) }()
		go func() { defer wg.Done(); subCtrl = a.queryAllPods(a.subscriberHeadless, a.subscriberURL) }()
	}

	// Gather health results
	totalServices := len(serviceList) + 1 // +1 for color
	if a.faceMode == "pubsub" {
		totalServices += 2 // mysql + queue backend
	}
	healthResults := make(map[string]*serviceStatus, totalServices)
	for i := 0; i < totalServices; i++ {
		hr := <-healthCh
		healthResults[hr.name] = hr.status
	}

	dbRes := <-dbCh
	qRes := <-qCh
	wg.Wait()

	// ── Build Prometheus output ────────────────────────────────────────────

	p := &promWriter{}

	// Info metric — mode and queue backend
	p.gauge("faces_info", "Faces demo configuration (always 1, use labels for info)")
	p.sample("faces_info", [][2]string{
		{"mode", a.faceMode},
		{"queue_backend", a.queueType},
	}, 1)
	p.nl()

	// Service health
	p.gauge("faces_service_up", "1 if the service is healthy, 0 otherwise")
	for name, s := range healthResults {
		p.sample("faces_service_up", [][2]string{{"service", name}}, boolGauge(s.Healthy))
	}
	p.nl()

	p.gauge("faces_service_latency_ms", "Health check round-trip latency in milliseconds")
	for name, s := range healthResults {
		if s.LatencyMs > 0 {
			p.sample("faces_service_latency_ms", [][2]string{{"service", name}}, float64(s.LatencyMs))
		}
	}
	p.nl()

	// Pub/sub pipeline metrics
	if a.faceMode == "pubsub" {

		// MySQL 3-state row counts
		p.gauge("faces_db_rows", "Rows in MySQL face_queue table by state")
		if dbRes.err == nil {
			p.sample("faces_db_rows", [][2]string{{"state", "pending"}}, float64(dbRes.pending))
			p.sample("faces_db_rows", [][2]string{{"state", "queued"}}, float64(dbRes.queued))
			p.sample("faces_db_rows", [][2]string{{"state", "acknowledged"}}, float64(dbRes.acknowledged))
		}
		p.nl()

		// Queue backend depth
		if qRes.qs != nil {
			p.gauge("faces_queue_depth", "Current messages in the queue backend ready to be consumed")
			p.sample("faces_queue_depth", [][2]string{{"backend", a.queueType}}, float64(qRes.qs.Depth))
			p.nl()

			p.gauge("faces_queue_max_depth", "Configured maximum depth of the queue backend")
			p.sample("faces_queue_max_depth", [][2]string{{"backend", a.queueType}}, float64(qRes.qs.MaxDepth))
			p.nl()

			if qRes.qs.MaxDepth > 0 {
				p.gauge("faces_queue_fill_ratio", "Queue depth as a fraction of max depth (0.0–1.0)")
				p.sample("faces_queue_fill_ratio", [][2]string{{"backend", a.queueType}},
					float64(qRes.qs.Depth)/float64(qRes.qs.MaxDepth))
				p.nl()
			}

			if dbRes.err == nil && dbRes.queued >= 0 {
				p.gauge("faces_queue_stranded_rows",
					"Rows in MySQL state=queued but not present in the queue backend (evicted by depth cap)")
				stranded := dbRes.queued - qRes.qs.Depth
				if stranded < 0 {
					stranded = 0
				}
				p.sample("faces_queue_stranded_rows", [][2]string{{"backend", a.queueType}}, float64(stranded))
				p.nl()
			}

			if qRes.qs.ReadyRate > 0 || qRes.qs.DeliverRate > 0 {
				p.gauge("faces_queue_publish_rate", "Messages published to the queue per second (from management API)")
				p.sample("faces_queue_publish_rate", [][2]string{{"backend", a.queueType}}, qRes.qs.ReadyRate)
				p.nl()

				p.gauge("faces_queue_deliver_rate", "Messages delivered from the queue per second (from management API)")
				p.sample("faces_queue_deliver_rate", [][2]string{{"backend", a.queueType}}, qRes.qs.DeliverRate)
				p.nl()

				p.gauge("faces_queue_net_rate",
					"Publish rate minus deliver rate. Positive = queue filling; negative = queue draining; ~0 = stable")
				p.sample("faces_queue_net_rate", [][2]string{{"backend", a.queueType}},
					qRes.qs.ReadyRate-qRes.qs.DeliverRate)
				p.nl()
			}
		}

		// Publisher per-pod metrics
		if pubCtrl.Available && len(pubCtrl.Pods) > 0 {
			p.gauge("faces_publisher_paused", "1 if the publisher pod is paused, 0 if running")
			for _, pod := range pubCtrl.Pods {
				if pod.Available {
					p.sample("faces_publisher_paused", [][2]string{{"pod", pod.PodIP}}, boolGauge(pod.Paused))
				}
			}
			p.nl()

			p.gauge("faces_publisher_interval_ms",
				"Configured publish interval per pod in milliseconds (0 = flood mode)")
			for _, pod := range pubCtrl.Pods {
				if pod.Available {
					p.sample("faces_publisher_interval_ms", [][2]string{{"pod", pod.PodIP}},
						float64(pod.PublishIntervalMs))
				}
			}
			p.nl()

			p.gauge("faces_publisher_concurrency",
				"Concurrent publish goroutines per pod")
			for _, pod := range pubCtrl.Pods {
				if pod.Available && pod.PublishConcurrency > 0 {
					p.sample("faces_publisher_concurrency", [][2]string{{"pod", pod.PodIP}},
						float64(pod.PublishConcurrency))
				}
			}
			p.nl()
		}

		// Subscriber per-pod metrics
		if subCtrl.Available && len(subCtrl.Pods) > 0 {
			p.gauge("faces_subscriber_paused", "1 if the subscriber pod is paused, 0 if running")
			for _, pod := range subCtrl.Pods {
				if pod.Available {
					p.sample("faces_subscriber_paused", [][2]string{{"pod", pod.PodIP}}, boolGauge(pod.Paused))
				}
			}
			p.nl()
		}

		// Aggregate published/delivered counts from the DB
		if dbRes.err == nil {
			p.gauge("faces_messages_published_total",
				"Total messages written to MySQL (pending + queued + acknowledged)")
			p.sample("faces_messages_published_total", nil,
				float64(dbRes.pending+dbRes.queued+dbRes.acknowledged))
			p.nl()

			p.gauge("faces_messages_delivered_total",
				"Total messages delivered to the faces-gui (MySQL state=acknowledged)")
			p.sample("faces_messages_delivered_total", nil, float64(dbRes.acknowledged))
			p.nl()
		}
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, p.sb.String())
}

// queueStatsForMetrics fetches queue depth and rates from the configured backend.
// Returns a *queueStats struct (both depth and rate fields).
func (a *AdminProvider) queueStatsForMetrics(ctx context.Context) (*queueStats, error) {
	switch a.queueType {
	case "rabbitmq":
		return a.rabbitmqQueueStats()
	case "redis":
		depth, err := a.redisQueueDepth(ctx)
		if err != nil {
			return nil, err
		}
		return &queueStats{
			Backend:   "redis",
			Available: true,
			Depth:     depth,
			MaxDepth:  a.maxDepth,
		}, nil
	}
	return nil, fmt.Errorf("unknown queue type %q", a.queueType)
}
