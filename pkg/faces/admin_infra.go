// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package faces

// admin_infra.go — Infrastructure overview endpoint.
//
// GET /api/infrastructure returns all service pods grouped by topology zone.
// Pod discovery uses the Kubernetes API directly (not headless DNS) so that
// ALL pods across ALL zones are returned regardless of DNS behaviour.
// Serving state (current emoji, color, publish rate) is fetched in parallel
// after the pod list is resolved.
//
// Zone classification:
//   - topology.kubernetes.io/zone label present → named zone card
//   - No zone, but Linkerd ExternalWorkload CRD matches the IP → "External Workload"
//   - No zone, regular in-cluster pod → "In-Cluster"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	colorpkg "github.com/BuoyantIO/faces-demo/v2/pkg/color"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InfraPod is the unified per-pod data shape for the infrastructure view.
type InfraPod struct {
	IP           string `json:"ip"`
	Name         string `json:"name,omitempty"`
	Zone         string `json:"zone,omitempty"`
	Region       string `json:"region,omitempty"`
	Node         string `json:"node,omitempty"`
	WorkloadType string `json:"workloadType,omitempty"` // "" | "externalworkload"

	// Serving state (service-specific). Smiley/Color are the CENTER value; the *Edge
	// fields are set ONLY when the pod's edge value differs from center, so the UI can
	// show both. (The faces services serve an independent center vs. edge face/color.)
	Smiley     string `json:"smiley,omitempty"`     // current center emoji HTML entity
	SmileyEdge string `json:"smileyEdge,omitempty"` // edge emoji, only when ≠ center
	Color      string `json:"color,omitempty"`      // current center color hex
	ColorEdge  string `json:"colorEdge,omitempty"`  // edge color, only when ≠ center
	Paused     bool   `json:"paused,omitempty"`
	RateMs     int64  `json:"publishIntervalMs,omitempty"`

	// Per-pod chaos state (nil when the pod doesn't support chaos or fetch failed)
	Chaos *chaosState `json:"chaos,omitempty"`

	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// InfraZone groups all pods within one availability zone (or workload class).
type InfraZone struct {
	Zone   string                `json:"zone"`
	Region string                `json:"region,omitempty"`
	Label  string                `json:"label"`
	Icon   string                `json:"icon"`
	Pods   map[string][]InfraPod `json:"pods"`
}

// InfraResponse is returned by GET /api/infrastructure.
type InfraResponse struct {
	Mode        string      `json:"mode"`
	HasTopology bool        `json:"hasTopology"`
	Zones       []InfraZone `json:"zones"`
}

// handleInfrastructure collects pod topology and serving state from all services
// in parallel, groups results by availability zone, and returns the full picture.
func (a *AdminProvider) handleInfrastructure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	// Respect any runtime faceMode override set via PUT /api/config
	mode := a.effectiveFaceMode()

	type svcResult struct {
		service string
		pods    []InfraPod
	}

	// ExternalWorkload IP index — best-effort, empty map when CRD absent
	var ewIPs map[string]string
	if a.k8s != nil {
		ewIPs = a.k8s.ExternalWorkloadIPs()
	}

	// ── Collect all results concurrently ──────────────────────────────────

	// Backend services (smiley, smiley2, color, color2, etc.) — discovered via
	// ListAllBackendComponents so secondary instances are included automatically.
	backendCh := make(chan map[string][]InfraPod, 1)
	go func() {
		backendCh <- a.infraBackendPods(ctx)
	}()

	type ctrl struct {
		svc  string
		hl   string
		url  string
		comp string
	}
	controlSvcs := []ctrl{
		{"publisher", a.publisherHeadless, a.svcURL("publisher", a.publisherURL), "face-publisher"},
		{"subscriber", a.subscriberHeadless, a.svcURL("subscriber", a.subscriberURL), "face-subscriber"},
	}
	ctrlCh := make(chan svcResult, len(controlSvcs)+2)

	if mode == "pubsub" {
		for _, c := range controlSvcs {
			c := c
			go func() {
				cs := a.queryAllPodsForComponent(c.hl, c.url, c.comp)
				pods := make([]InfraPod, len(cs.Pods))
				for i, p := range cs.Pods {
					pods[i] = controlPodToInfra(p)
				}
				ctrlCh <- svcResult{c.svc, pods}
			}()
		}
	} else {
		for range controlSvcs {
			ctrlCh <- svcResult{}
		}
	}

	// Classic face + GUI from K8s API
	go func() { ctrlCh <- svcResult{"face", a.infraK8sPods("face")} }()
	go func() { ctrlCh <- svcResult{"gui", a.infraK8sPods("faces-gui")} }()

	// Collect control/face/gui results
	allPods := make(map[string][]InfraPod)
	for i := 0; i < len(controlSvcs)+2; i++ {
		res := <-ctrlCh
		if res.service != "" && len(res.pods) > 0 {
			allPods[res.service] = res.pods
		}
	}

	// Merge backend results
	for svc, pods := range <-backendCh {
		allPods[svc] = pods
	}

	// ── Enrich with ExternalWorkload type ─────────────────────────────────
	if len(ewIPs) > 0 {
		for svc, pods := range allPods {
			for i := range pods {
				if _, ok := ewIPs[pods[i].IP]; ok {
					pods[i].WorkloadType = "externalworkload"
				}
			}
			allPods[svc] = pods
		}
	}

	// ── Enrich with per-pod chaos state for HTTP chaos services ────────────
	// smiley/color already fetched their chaos in fetchBackendState (reusing the
	// serving-state connection). Here we cover face/publisher/subscriber — pods
	// that support /chaos but weren't enriched above. gui has no chaos endpoint.
	{
		var wg sync.WaitGroup
		ewPorts := a.ewPortIndex()
		for svc := range allPods {
			if !strings.HasPrefix(svc, "face") && !strings.HasPrefix(svc, "publisher") && !strings.HasPrefix(svc, "subscriber") {
				continue
			}
			pods := allPods[svc]
			for i := range pods {
				if pods[i].Chaos != nil || pods[i].IP == "" || isLinkLocalIPv6(pods[i].IP) {
					continue
				}
				wg.Add(1)
				go func(p *InfraPod) {
					defer wg.Done()
					port := a.podPort
					if ep, ok := ewPorts[p.IP]; ok {
						port = ep
					}
					cs := a.fetchChaosState(fmt.Sprintf("http://%s:%s", p.IP, port))
					if cs.Available {
						p.Chaos = &cs
					}
				}(&pods[i])
			}
		}
		wg.Wait()
	}

	// ── Group by zone ──────────────────────────────────────────────────────
	// Three buckets:
	//   Named zone (topology label present) → AZ card        📍
	//   No zone, WorkloadType=="externalworkload" → EW card  🔗
	//   No zone, regular in-cluster pod → On-Premise card    🏢
	zoneMap := make(map[string]*InfraZone)
	hasTopology := false

	for svc, pods := range allPods {
		for _, pod := range pods {
			var key, label, icon string
			switch {
			case pod.Zone != "":
				key, label, icon = pod.Zone, pod.Zone, "📍"
				hasTopology = true
			case pod.WorkloadType == "externalworkload":
				key, label, icon = "__ew__", "External Workload", "🔗"
			default:
				key, label, icon = "__onprem__", "On-Premise", "🏢"
			}

			if _, ok := zoneMap[key]; !ok {
				zoneMap[key] = &InfraZone{
					Zone:   pod.Zone,
					Region: pod.Region,
					Label:  label,
					Icon:   icon,
					Pods:   make(map[string][]InfraPod),
				}
			}
			zoneMap[key].Pods[svc] = append(zoneMap[key].Pods[svc], pod)
		}
	}

	// Named zones alphabetically; External Workload before On-Premise; both last
	var zones []InfraZone
	for key, z := range zoneMap {
		if key != "__ew__" && key != "__onprem__" {
			zones = append(zones, *z)
		}
	}
	sortInfraZones(zones)
	if ew, ok := zoneMap["__ew__"]; ok {
		zones = append(zones, *ew)
	}
	if op, ok := zoneMap["__onprem__"]; ok {
		zones = append(zones, *op)
	}

	a.writeJSON(w, InfraResponse{
		Mode:        mode,
		HasTopology: hasTopology,
		Zones:       zones,
	})
}

// ── Backend pod collector ──────────────────────────────────────────────────

// infraBackendPods discovers every backend pod/workload and fetches serving state.
// Queries three sources and merges them — a service may appear in multiple sources
// (e.g. some replicas are K8s pods, others are ExternalWorkloads):
//
//  1. K8s pods via ListAllBackendComponents (smiley, smiley2, color, color2, …)
//  2. Linkerd ExternalWorkloads via ListExternalWorkloadComponents (off-cluster instances)
//  3. Headless DNS fallback for any service still undiscovered after 1+2
func (a *AdminProvider) infraBackendPods(ctx context.Context) map[string][]InfraPod {
	// Collect all topology entries per component from all K8s sources
	byComp := make(map[string][]PodTopology)

	if a.k8s != nil {
		// Source 1: K8s pods
		if pods, err := a.k8s.ListAllBackendComponents(); err == nil {
			for comp, kpods := range pods {
				byComp[comp] = append(byComp[comp], kpods...)
			}
		}
		// Source 2: ExternalWorkloads (not pods — off-cluster processes registered with Linkerd)
		for comp, ewPods := range a.k8s.ListExternalWorkloadComponents() {
			byComp[comp] = append(byComp[comp], ewPods...)
		}
	}

	result := make(map[string][]InfraPod)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for comp, kpods := range byComp {
		comp, kpods := comp, kpods
		wg.Add(1)
		go func() {
			defer wg.Done()
			pods := a.fetchBackendState(ctx, comp, kpods)
			if len(pods) > 0 {
				mu.Lock()
				result[comp] = pods
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Source 3: DNS fallback for services not found via K8s at all
	if _, ok := result["smiley"]; !ok {
		if smPods := a.infraSmileyFallback(ctx); len(smPods) > 0 {
			result["smiley"] = smPods
		}
	}
	if _, ok := result["color"]; !ok {
		if colPods := a.infraColorFallback(ctx); len(colPods) > 0 {
			result["color"] = colPods
		}
	}
	return result
}

// fetchBackendState pings each pod for its current serving state.
// Component type is inferred from the name prefix: "smiley*" → HTTP, "color*" → gRPC.
func (a *AdminProvider) fetchBackendState(ctx context.Context, comp string, kpods []PodTopology) []InfraPod {
	pods := make([]InfraPod, len(kpods))
	var wg sync.WaitGroup

	isSmiley := strings.HasPrefix(comp, "smiley")
	isColor := strings.HasPrefix(comp, "color")

	client := &http.Client{Timeout: 2 * time.Second}

	for i, kp := range kpods {
		wg.Add(1)
		go func(i int, kp PodTopology) {
			defer wg.Done()
			p := InfraPod{
				IP: kp.IP, Name: kp.Name,
				Zone: kp.Zone, Region: kp.Region, Node: kp.Node,
				WorkloadType: kp.WorkloadType,
			}
			// ExternalWorkloads declare their own port (typically 80);
			// K8s pods use a.podPort (typically 8000).
			port := a.podPort
			if kp.Port != "" {
				port = kp.Port
			}

			// Link-local IPv6 addresses (fe80::/10) require an interface scope ID
			// that is not available inside a Kubernetes pod. Skip the state query
			// and mark as available-but-state-unknown for ExternalWorkloads.
			if isLinkLocalIPv6(kp.IP) {
				p.Available = true // workload exists; state just can't be fetched
				p.Error = "state unavailable — link-local IPv6 not reachable from cluster"
				pods[i] = p
				return
			}

			podURL := fmt.Sprintf("http://%s:%s", kp.IP, port)

			switch {
			case isSmiley:
				resp, err := client.Get(podURL + "/center?row=0&col=0")
				if err != nil {
					p.Error = friendlyNetErr(err, kp.WorkloadType)
				} else {
					defer resp.Body.Close()
					var result struct {
						Smiley string `json:"smiley"`
					}
					if json.NewDecoder(resp.Body).Decode(&result) == nil {
						p.Smiley = result.Smiley
					}
					p.Available = true
				}
				// Edge emoji — record only when it differs from center (best-effort)
				if eResp, eErr := client.Get(podURL + "/edge?row=1&col=0"); eErr == nil {
					defer eResp.Body.Close()
					var er struct {
						Smiley string `json:"smiley"`
					}
					if json.NewDecoder(eResp.Body).Decode(&er) == nil && er.Smiley != "" && er.Smiley != p.Smiley {
						p.SmileyEdge = er.Smiley
					}
				}
				// Per-pod chaos state — reuse the same HTTP client/connection
				if cResp, cErr := client.Get(podURL + "/chaos"); cErr == nil {
					defer cResp.Body.Close()
					cs := chaosState{Available: true}
					if json.NewDecoder(cResp.Body).Decode(&cs) == nil {
						p.Chaos = &cs
					}
				}
			case isColor:
				addr := kp.IP + ":" + port // port already respects kp.Port override
				conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					p.Error = friendlyNetErr(err, kp.WorkloadType)
				} else {
					defer conn.Close()
					gCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
					defer cancel()
					colorClient := colorpkg.NewColorServiceClient(conn)
					r, err := colorClient.Center(gCtx, &colorpkg.ColorRequest{Row: 0, Column: 0})
					if err != nil {
						p.Error = err.Error()
					} else {
						p.Color = r.Color
						p.Available = true
					}
					// Edge color — record only when it differs from center (best-effort)
					eCtx, eCancel := context.WithTimeout(ctx, 2*time.Second)
					defer eCancel()
					if er, eErr := colorClient.Edge(eCtx, &colorpkg.ColorRequest{Row: 1, Column: 0}); eErr == nil {
						if er.Color != "" && er.Color != p.Color {
							p.ColorEdge = er.Color
						}
					}
					// Per-pod chaos state — reuse the same gRPC connection, fresh timeout
					chCtx, chCancel := context.WithTimeout(ctx, 2*time.Second)
					defer chCancel()
					if cs, cErr := colorClient.GetChaos(chCtx, &colorpkg.ChaosRequest{}); cErr == nil {
						buckets := make([]int, len(cs.DelayBuckets))
						for j, v := range cs.DelayBuckets {
							buckets[j] = int(v)
						}
						p.Chaos = &chaosState{
							ErrorFraction: int(cs.ErrorFraction),
							LatchFraction: int(cs.LatchFraction),
							MaxRate:       float64(cs.MaxRate),
							DelayBuckets:  buckets,
							Latched:       cs.Latched,
							Available:     true,
						}
					}
				}
			default:
				p.Available = kp.Phase == "Running"
			}
			pods[i] = p
		}(i, kp)
	}
	wg.Wait()
	return pods
}

// ── DNS fallbacks (used only when K8s API is unavailable) ─────────────────

func (a *AdminProvider) infraSmileyFallback(ctx context.Context) []InfraPod {
	urls := a.discoverPodURLs(a.smileyHeadless, a.svcURL("smiley", a.smileyURL))
	pods := make([]InfraPod, len(urls))
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 2 * time.Second}
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			ip := podURLToIP(u)
			p := InfraPod{IP: ip, Name: ip}
			resp, err := client.Get(u + "/center?row=0&col=0")
			if err != nil {
				p.Error = err.Error()
			} else {
				defer resp.Body.Close()
				var result struct {
					Smiley string `json:"smiley"`
				}
				if json.NewDecoder(resp.Body).Decode(&result) == nil {
					p.Smiley = result.Smiley
				}
				p.Available = true
			}
			pods[i] = p
		}(i, u)
	}
	wg.Wait()
	return pods
}

func (a *AdminProvider) infraColorFallback(ctx context.Context) []InfraPod {
	urls := a.discoverPodURLs(a.colorHeadless, a.colorGRPCAddr)
	pods := make([]InfraPod, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			addr := strings.TrimPrefix(u, "http://")
			ip := podURLToIP(u)
			p := InfraPod{IP: ip, Name: ip}
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				p.Error = err.Error()
				pods[i] = p
				return
			}
			defer conn.Close()
			gCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			r, err := colorpkg.NewColorServiceClient(conn).Center(gCtx, &colorpkg.ColorRequest{Row: 0, Column: 0})
			if err != nil {
				p.Error = err.Error()
			} else {
				p.Color = r.Color
				p.Available = true
			}
			pods[i] = p
		}(i, u)
	}
	wg.Wait()
	return pods
}

// infraK8sPods lists pods for a component using only the Kubernetes API.
func (a *AdminProvider) infraK8sPods(component string) []InfraPod {
	if a.k8s == nil {
		return nil
	}
	k8sPods, err := a.k8s.ListPodsForComponent(component)
	if err != nil || len(k8sPods) == 0 {
		return nil
	}
	pods := make([]InfraPod, len(k8sPods))
	for i, kp := range k8sPods {
		pods[i] = InfraPod{
			IP: kp.IP, Name: kp.Name,
			Zone: kp.Zone, Region: kp.Region, Node: kp.Node,
			WorkloadType: kp.WorkloadType,
			Available:    kp.Phase == "Running",
		}
	}
	return pods
}

// ── Helpers ────────────────────────────────────────────────────────────────

func controlPodToInfra(p podControlState) InfraPod {
	return InfraPod{
		IP: p.PodIP, Name: p.PodName,
		Zone: p.Zone, Region: p.Region, Node: p.Node,
		Paused: p.Paused, RateMs: p.PublishIntervalMs,
		Available: p.Available, Error: p.Error,
	}
}

// isLinkLocalIPv6 returns true for fe80::/10 addresses.
// These require a scope ID (e.g. %eth0) to route and are unreachable from
// inside a Kubernetes pod without one.
func isLinkLocalIPv6(ip string) bool {
	return len(ip) >= 4 &&
		(ip[0] == 'f' || ip[0] == 'F') &&
		(ip[1] == 'e' || ip[1] == 'E') &&
		(ip[2] == '8' || ip[2] == '9' ||
			ip[2] == 'a' || ip[2] == 'A' ||
			ip[2] == 'b' || ip[2] == 'B')
}

// friendlyNetErr returns a short human-readable error for network failures.
// For ExternalWorkloads it omits the raw URL so tooltips stay readable.
func friendlyNetErr(err error, workloadType string) string {
	msg := err.Error()
	if workloadType == "externalworkload" {
		// Drop the noisy URL prefix that Go includes in HTTP errors
		if idx := strings.Index(msg, "\": "); idx != -1 {
			msg = strings.TrimSpace(msg[idx+3:])
		}
		if len(msg) > 80 {
			msg = msg[:80] + "…"
		}
	}
	return msg
}

func sortInfraZones(zones []InfraZone) {
	for i := 1; i < len(zones); i++ {
		for j := i; j > 0 && zones[j].Zone < zones[j-1].Zone; j-- {
			zones[j], zones[j-1] = zones[j-1], zones[j]
		}
	}
}
