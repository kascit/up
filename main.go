package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// ─── Config ───────────────────────────────────────────────────────────────────

type ServiceConfig struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`        // health endpoint
	HomepageURL string `json:"homepageUrl"` // publicly visible link
	Icon        string `json:"icon"`        // FontAwesome class
	Category    string `json:"category"`    // "Core", "Tools", etc.
}

// Default services – the admin can override by providing services.json alongside the binary.
var defaultServices = []ServiceConfig{
	{
		Name:        "authy",
		DisplayName: "Authy",
		URL:         "https://auth.dhanur.me/api/health",
		HomepageURL: "https://auth.dhanur.me",
		Icon:        "fa-solid fa-fingerprint",
		Category:    "Core",
	},
	{
		Name:        "tasks",
		DisplayName: "Tasks",
		URL:         "https://tasks.dhanur.me/actuator/health",
		HomepageURL: "https://tasks.dhanur.me",
		Icon:        "fa-solid fa-clipboard-list",
		Category:    "Core",
	},
	{
		Name:        "linkr",
		DisplayName: "Linkr",
		URL:         "https://linkr.dhanur.me/health",
		HomepageURL: "https://linkr.dhanur.me",
		Icon:        "fa-solid fa-link",
		Category:    "Core",
	},
	{
		Name:        "main",
		DisplayName: "dhanur.me",
		URL:         "https://dhanur.me",
		HomepageURL: "https://dhanur.me",
		Icon:        "fa-solid fa-house",
		Category:    "Core",
	},
}

func loadServices() []ServiceConfig {
	data, err := os.ReadFile("services.json")
	if err != nil {
		return defaultServices
	}
	var svcs []ServiceConfig
	if err := json.Unmarshal(data, &svcs); err != nil {
		slog.Error("Failed to parse services.json, using defaults", "err", err)
		return defaultServices
	}
	slog.Info("Loaded services from services.json", "count", len(svcs))
	return svcs
}

// ─── Status Types ─────────────────────────────────────────────────────────────

type ServiceStatus string

const (
	StatusUp      ServiceStatus = "up"
	StatusDown    ServiceStatus = "down"
	StatusDegraded ServiceStatus = "degraded"
	StatusUnknown ServiceStatus = "unknown"
)

const (
	// Industry-standard: keep 90 data points per service (90 × 30s = 45 min window)
	historyLen  = 90
	checkPeriod = 30 * time.Second
	httpTimeout = 8 * time.Second
)

type CheckResult struct {
	Status    ServiceStatus `json:"status"`
	Latency   int64         `json:"latencyMs"` // milliseconds, -1 if unreachable
	Timestamp time.Time     `json:"timestamp"`
}

type ServiceState struct {
	Config  ServiceConfig `json:"config"`
	Current CheckResult   `json:"current"`
	History []CheckResult `json:"history"` // newest-first, capped at historyLen
	Uptime  float64       `json:"uptimePercent"`
}

// ─── Monitor ──────────────────────────────────────────────────────────────────

type Monitor struct {
	mu       sync.RWMutex
	services map[string]*ServiceState
	order    []string // keep deterministic iteration order
	client   *http.Client
}

func NewMonitor(cfgs []ServiceConfig) *Monitor {
	m := &Monitor{
		services: make(map[string]*ServiceState, len(cfgs)),
		order:    make([]string, 0, len(cfgs)),
		client: &http.Client{
			Timeout: httpTimeout,
			// Don't follow redirects – a redirect usually means something changed
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	for _, cfg := range cfgs {
		m.order = append(m.order, cfg.Name)
		m.services[cfg.Name] = &ServiceState{
			Config:  cfg,
			Current: CheckResult{Status: StatusUnknown, Latency: -1, Timestamp: time.Now()},
			History: make([]CheckResult, 0, historyLen),
		}
	}
	return m
}

// probe checks a single service and returns a CheckResult.
func (m *Monitor) probe(cfg ServiceConfig) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return CheckResult{Status: StatusDown, Latency: -1, Timestamp: time.Now()}
	}
	req.Header.Set("User-Agent", "up-monitor/1.0 (+https://up.dhanur.me)")

	start := time.Now()
	resp, err := m.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return CheckResult{Status: StatusDown, Latency: -1, Timestamp: time.Now()}
	}
	defer resp.Body.Close()

	var status ServiceStatus
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		status = StatusUp
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		// 3xx from our own apps is unexpected – treat as degraded
		status = StatusDegraded
	case resp.StatusCode == 503:
		// Spring Boot reports 503 when DB is down but app is alive
		status = StatusDegraded
	default:
		status = StatusDown
	}

	// High latency → degrade (industry threshold: >2s is sluggish)
	if status == StatusUp && latency > 2000 {
		status = StatusDegraded
	}

	return CheckResult{Status: status, Latency: latency, Timestamp: time.Now()}
}

// checkAll probes every service concurrently and commits results.
func (m *Monitor) checkAll() {
	type result struct {
		name string
		cr   CheckResult
	}

	results := make(chan result, len(m.order))
	for _, name := range m.order {
		go func(s *ServiceState) {
			cr := m.probe(s.Config)
			results <- result{name: s.Config.Name, cr: cr}
		}(m.services[name])
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for range m.order {
		r := <-results
		s := m.services[r.name]
		s.Current = r.cr

		// Prepend to history (newest-first)
		s.History = append([]CheckResult{r.cr}, s.History...)
		if len(s.History) > historyLen {
			s.History = s.History[:historyLen]
		}

		// Recalculate uptime over entire stored window
		up := 0
		for _, h := range s.History {
			if h.Status == StatusUp {
				up++
			}
		}
		if len(s.History) > 0 {
			s.Uptime = float64(up) / float64(len(s.History)) * 100
		}
	}
}

// Start launches the background poller.
func (m *Monitor) Start() {
	// Fire immediately on startup, then tick.
	m.checkAll()
	ticker := time.NewTicker(checkPeriod)
	go func() {
		for range ticker.C {
			m.checkAll()
		}
	}()
}

// Snapshot returns a safe copy for JSON serialisation.
func (m *Monitor) Snapshot() []*ServiceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*ServiceState, 0, len(m.order))
	for _, name := range m.order {
		s := m.services[name]
		snap := *s // shallow copy is fine: slices are only appended/replaced
		out = append(out, &snap)
	}
	return out
}

// ─── HTTP Handlers ────────────────────────────────────────────────────────────

func overallStatus(states []*ServiceState) ServiceStatus {
	down, degraded := 0, 0
	for _, s := range states {
		switch s.Current.Status {
		case StatusDown:
			down++
		case StatusDegraded:
			degraded++
		}
	}
	switch {
	case down > 0:
		return StatusDown
	case degraded > 0:
		return StatusDegraded
	default:
		return StatusUp
	}
}

type APIResponse struct {
	Overall   ServiceStatus  `json:"overall"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Services  []*ServiceState `json:"services"`
}

func handleAPI(mon *Monitor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		states := mon.Snapshot()
		resp := APIResponse{
			Overall:   overallStatus(states),
			UpdatedAt: time.Now().UTC(),
			Services:  states,
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(resp)
	}
}

// Self-health endpoint so this service can be monitored too.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

func handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

type authStatus struct {
	Authenticated bool   `json:"authenticated"`
	Role          string `json:"role"`
}

func fetchAuthStatus(r *http.Request, authServiceURL string) (authStatus, error) {
	statusURL := strings.TrimRight(authServiceURL, "/") + "/api/status"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, statusURL, nil)
	if err != nil {
		return authStatus{}, err
	}

	if cookie := r.Header.Get("Cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if authz := r.Header.Get("Authorization"); authz != "" {
		req.Header.Set("Authorization", authz)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return authStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return authStatus{}, fmt.Errorf("auth status endpoint returned %d", resp.StatusCode)
	}

	var status authStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return authStatus{}, err
	}

	return status, nil
}

func requireAdminDashboard(authServiceURL string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := fetchAuthStatus(r, authServiceURL)
		nextURL := url.QueryEscape(currentRequestURL(r))
		if err != nil || !status.Authenticated {
			http.Redirect(w, r, strings.TrimRight(authServiceURL, "/")+"/login?next="+nextURL, http.StatusFound)
			return
		}
		if status.Role != "admin" {
			http.Redirect(w, r, strings.TrimRight(authServiceURL, "/")+"/verify?next="+nextURL, http.StatusFound)
			return
		}
		next(w, r)
	}
}

func currentRequestURL(r *http.Request) string {
	scheme := "https"
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.ToLower(forwarded)
	} else if r.TLS == nil {
		scheme = "http"
	}
	return scheme + "://" + r.Host + r.URL.RequestURI()
}

func isAllowedOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "dhanur.me" {
		return true
	}
	return strings.HasSuffix(host, ".dhanur.me")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	svcs := loadServices()
	mon := NewMonitor(svcs)
	mon.Start()
	manifest := loadManifest()
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "https://auth.dhanur.me"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /api/status", handleAPI(mon))
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/manifest", handleManifest(manifest))
	mux.HandleFunc("GET /.well-known/web-app-origin-association", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"web_apps":[{"manifest":"https://dhanur.me/icons/site.webmanifest","details":{"paths":["/*"]}}]}`))
	})

	addr := ":" + port
	slog.Info("UP monitor started", "addr", addr, "services", len(svcs))
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		slog.Error("Server failed", "err", err)
		os.Exit(1)
	}
}
