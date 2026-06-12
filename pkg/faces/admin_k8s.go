// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package faces

// admin_k8s.go — lightweight Kubernetes API client for the admin portal.
//
// Uses the in-cluster service-account token directly rather than pulling in
// k8s.io/client-go, keeping the binary small and avoiding dependency churn.
// Only needs two API calls: list pods (namespaced) and get node (cluster-scoped).

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// PodTopology holds Kubernetes-sourced identity and placement data for a pod.
type PodTopology struct {
	Name         string `json:"name"`
	IP           string `json:"ip"`
	Node         string `json:"node,omitempty"`
	Zone         string `json:"zone,omitempty"`
	Region       string `json:"region,omitempty"`
	Phase        string `json:"phase,omitempty"`
	WorkloadType string `json:"workloadType,omitempty"` // "pod" | "externalworkload"
	Port         string `json:"port,omitempty"`         // overrides default pod port (ExternalWorkloads use a different port than K8s pods)
}

// k8sClient makes authenticated REST calls to the in-cluster API server.
type k8sClient struct {
	host       string
	token      string
	namespace  string
	httpClient *http.Client
	logger     *slog.Logger

	// pod + node cache keyed by label selector
	mu        sync.RWMutex
	podCache  map[string][]PodTopology
	podCacheT map[string]time.Time
	nodeCache map[string]struct{ zone, region string }
	cacheTTL  time.Duration
}

func newK8sClient(logger *slog.Logger) (*k8sClient, error) {
	tokenB, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, fmt.Errorf("read service-account token: %w", err)
	}
	nsB, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	caB, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caB)

	return &k8sClient{
		host:       "https://kubernetes.default.svc",
		token:      strings.TrimSpace(string(tokenB)),
		namespace:  strings.TrimSpace(string(nsB)),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: pool},
			},
		},
		logger:    logger.With("component", "k8sClient"),
		podCache:  make(map[string][]PodTopology),
		podCacheT: make(map[string]time.Time),
		nodeCache: make(map[string]struct{ zone, region string }),
		cacheTTL:  30 * time.Second,
	}, nil
}

// ListPodsForComponent returns topology-enriched info for all running pods
// matching `faces.buoyant.io/component=<component>` in the admin's namespace.
// Results are cached for 30 s to avoid hammering the API on every poll cycle.
func (c *k8sClient) ListPodsForComponent(component string) ([]PodTopology, error) {
	selector := "faces.buoyant.io/component=" + component

	c.mu.RLock()
	if pods, ok := c.podCache[selector]; ok {
		if time.Since(c.podCacheT[selector]) < c.cacheTTL {
			c.mu.RUnlock()
			return pods, nil
		}
	}
	c.mu.RUnlock()

	// ── List pods ─────────────────────────────────────────────────────────
	podURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pods?labelSelector=%s&fieldSelector=status.phase=Running",
		c.host, c.namespace, url.QueryEscape(selector))

	var podList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
			Status struct {
				PodIP string `json:"podIP"`
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := c.get(podURL, &podList); err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	// ── Fetch node topology (cached per node) ──────────────────────────────
	var pods []PodTopology
	for _, item := range podList.Items {
		ip := item.Status.PodIP
		if ip == "" {
			continue
		}
		zone, region := c.nodeTopology(item.Spec.NodeName)
		pods = append(pods, PodTopology{
			Name:   item.Metadata.Name,
			IP:     ip,
			Node:   item.Spec.NodeName,
			Zone:   zone,
			Region: region,
			Phase:  item.Status.Phase,
		})
	}

	c.mu.Lock()
	c.podCache[selector] = pods
	c.podCacheT[selector] = time.Now()
	c.mu.Unlock()

	return pods, nil
}

// nodeTopology reads zone/region labels from the named node (cached indefinitely
// since node labels change very rarely in a running cluster).
func (c *k8sClient) nodeTopology(nodeName string) (zone, region string) {
	if nodeName == "" {
		return
	}

	c.mu.RLock()
	if info, ok := c.nodeCache[nodeName]; ok {
		c.mu.RUnlock()
		return info.zone, info.region
	}
	c.mu.RUnlock()

	var node struct {
		Metadata struct {
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
	}
	nodeURL := fmt.Sprintf("%s/api/v1/nodes/%s", c.host, nodeName)
	if err := c.get(nodeURL, &node); err != nil {
		c.logger.Debug("k8s: get node failed", "node", nodeName, "error", err)
		return
	}

	zone   = node.Metadata.Labels["topology.kubernetes.io/zone"]
	region = node.Metadata.Labels["topology.kubernetes.io/region"]

	c.mu.Lock()
	c.nodeCache[nodeName] = struct{ zone, region string }{zone, region}
	c.mu.Unlock()
	return
}

// get makes an authenticated GET to the K8s API and decodes JSON into v.
func (c *k8sClient) get(rawURL string, v any) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// ListAllBackendComponents queries all pods labelled
// `faces.buoyant.io/component-type=backend` and returns them grouped by
// component name. This discovers smiley, smiley2, smiley3, color, color2, color3,
// etc. automatically without hardcoding each name.
func (c *k8sClient) ListAllBackendComponents() (map[string][]PodTopology, error) {
	const cacheKey = "__backend_components__"
	c.mu.RLock()
	if pods, ok := c.podCache[cacheKey]; ok {
		if time.Since(c.podCacheT[cacheKey]) < c.cacheTTL {
			c.mu.RUnlock()
			// Re-group from cached flat list
			return groupByComponent(pods), nil
		}
	}
	c.mu.RUnlock()

	selector := "faces.buoyant.io/component-type=backend"
	podURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pods?labelSelector=%s&fieldSelector=status.phase=Running",
		c.host, c.namespace, url.QueryEscape(selector))

	var podList struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
			Status struct {
				PodIP string `json:"podIP"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := c.get(podURL, &podList); err != nil {
		return nil, fmt.Errorf("list backend pods: %w", err)
	}

	var flat []PodTopology
	for _, item := range podList.Items {
		ip := item.Status.PodIP
		if ip == "" {
			continue
		}
		zone, region := c.nodeTopology(item.Spec.NodeName)
		component := item.Metadata.Labels["faces.buoyant.io/component"]
		flat = append(flat, PodTopology{
			Name:   item.Metadata.Name,
			IP:     ip,
			Node:   item.Spec.NodeName,
			Zone:   zone,
			Region: region,
			// Store component in Phase field temporarily for grouping, then clear it.
			// Cleaner: use a helper struct, but we avoid adding fields to PodTopology.
			Phase: component,
		})
	}

	c.mu.Lock()
	c.podCache[cacheKey] = flat
	c.podCacheT[cacheKey] = time.Now()
	c.mu.Unlock()

	return groupByComponent(flat), nil
}

func groupByComponent(flat []PodTopology) map[string][]PodTopology {
	out := make(map[string][]PodTopology)
	for _, p := range flat {
		comp := p.Phase // component name stored temporarily in Phase
		entry := p
		entry.Phase = "Running" // restore correct Phase value
		out[comp] = append(out[comp], entry)
	}
	return out
}

// ExternalWorkloadIPs returns a set of IPs registered as Linkerd ExternalWorkloads
// in the admin namespace. Returns empty map (not error) if the CRD is not installed.
// listExternalWorkloads fetches and caches ExternalWorkload resources as a flat
// []PodTopology slice where Phase holds the component label value temporarily
// (same pattern as ListAllBackendComponents). Returns nil when CRD is absent.
func (c *k8sClient) listExternalWorkloads() []PodTopology {
	const cacheKey = "__externalworkloads__"

	c.mu.RLock()
	if cached, ok := c.podCache[cacheKey]; ok {
		if time.Since(c.podCacheT[cacheKey]) < c.cacheTTL {
			c.mu.RUnlock()
			return cached
		}
	}
	c.mu.RUnlock()

	// API group is workload.linkerd.io (Buoyant Enterprise Linkerd).
	// Previously ext.linkerd.io — that was wrong and caused silent 404/403 failures.
	ewURL := fmt.Sprintf("%s/apis/workload.linkerd.io/v1beta1/namespaces/%s/externalworkloads",
		c.host, c.namespace)

	var ewList struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				WorkloadIPs []struct {
					IP string `json:"ip"`
				} `json:"workloadIPs"`
				Ports []struct {
					Port int `json:"port"`
				} `json:"ports"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := c.get(ewURL, &ewList); err != nil {
		// 403 = RBAC missing the ext.linkerd.io rule; 404 = CRD not installed.
		// Both are logged at Warn so operators can diagnose without enabling debug logging.
		if strings.Contains(err.Error(), "HTTP 403") {
			c.logger.Warn("k8s: RBAC missing ext.linkerd.io/externalworkloads permission — "+
				"ExternalWorkloads will appear as On-Premise. "+
				"Run: helm upgrade with the updated faces-admin-rbac.yaml to grant access.")
		} else {
			c.logger.Debug("k8s: ExternalWorkload CRD not available (Linkerd not installed or different version)",
				"error", err)
		}
		return nil
	}

	var flat []PodTopology
	for _, ew := range ewList.Items {
		comp := ew.Metadata.Labels["faces.buoyant.io/component"]
		ewPort := ""
		if len(ew.Spec.Ports) > 0 && ew.Spec.Ports[0].Port > 0 {
			ewPort = fmt.Sprintf("%d", ew.Spec.Ports[0].Port)
		}

		// Pick the single best IP from workloadIPs — one entry per ExternalWorkload.
		// Preference: IPv4 > global IPv6 > link-local IPv6 (fe80::/10).
		// This avoids duplicate entries when a VM has both an IPv4 and a link-local
		// IPv6 address registered (the link-local is unreachable from inside the cluster).
		bestIP := bestWorkloadIP(ew.Spec.WorkloadIPs)
		if bestIP != "" {
			flat = append(flat, PodTopology{
				IP:           bestIP,
				Name:         ew.Metadata.Name,
				Phase:        comp, // component name carried through cache
				WorkloadType: "externalworkload",
				Port:         ewPort,
			})
		}
	}

	c.mu.Lock()
	c.podCache[cacheKey] = flat
	c.podCacheT[cacheKey] = time.Now()
	c.mu.Unlock()

	return flat
}

// ExternalWorkloadIPs returns a map of IP → ExternalWorkload name for all
// registered ExternalWorkloads. Returns empty map when the CRD is absent.
func (c *k8sClient) ExternalWorkloadIPs() map[string]string {
	m := make(map[string]string)
	for _, p := range c.listExternalWorkloads() {
		m[p.IP] = p.Name
	}
	return m
}

// ListExternalWorkloadComponents returns ExternalWorkloads grouped by
// their `faces.buoyant.io/component` label — same shape as ListAllBackendComponents.
// This lets the infra view discover smiley/color ExternalWorkloads alongside pods.
func (c *k8sClient) ListExternalWorkloadComponents() map[string][]PodTopology {
	out := make(map[string][]PodTopology)
	for _, p := range c.listExternalWorkloads() {
		comp := p.Phase // Phase holds component name in cache
		if comp == "" {
			continue
		}
		out[comp] = append(out[comp], PodTopology{
			IP:           p.IP,
			Name:         p.Name,
			WorkloadType: "externalworkload",
			Phase:        "Running",
			Port:         p.Port, // preserve port override (e.g. 80 for ExternalWorkloads)
		})
	}
	return out
}

// bestWorkloadIP selects the most routable IP from a workloadIPs list.
// Preference order: IPv4 → global IPv6 → link-local IPv6.
// Returns "" when the list is empty.
func bestWorkloadIP(wips []struct{ IP string `json:"ip"` }) string {
	best := ""
	bestScore := -1 // higher = better
	for _, w := range wips {
		ip := w.IP
		if ip == "" {
			continue
		}
		score := 0
		if !strings.Contains(ip, ":") {
			score = 2 // IPv4 — most reliably routable in K8s contexts
		} else if !isLinkLocalIPv6(ip) {
			score = 1 // global IPv6
		}
		// link-local IPv6 scores 0
		if score > bestScore {
			bestScore = score
			best = ip
		}
	}
	return best
}

// BuildIPIndex returns a map from pod IP → PodTopology for fast lookup.
func BuildIPIndex(pods []PodTopology) map[string]PodTopology {
	m := make(map[string]PodTopology, len(pods))
	for _, p := range pods {
		m[p.IP] = p
	}
	return m
}
