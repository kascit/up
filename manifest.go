package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type AppEntry struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Icon     string `json:"icon"`
	Category string `json:"category,omitempty"`
	MinRole  string `json:"minRole"`
}

type ManifestResponse struct {
	Version string     `json:"version"`
	Apps    []AppEntry `json:"apps"`
}

var defaultManifest = ManifestResponse{
	Version: "2",
	Apps: []AppEntry{
		{Name: "Home", URL: "https://dhanur.me", Icon: "fa-solid fa-house", MinRole: "guest"},
		{Name: "Linkr", URL: "https://linkr.dhanur.me", Icon: "fa-solid fa-link", MinRole: "guest"},
		{Name: "Tasks", URL: "https://tasks.dhanur.me", Icon: "fa-solid fa-clipboard-list", MinRole: "guest"},
		{Name: "Auth", URL: "https://auth.dhanur.me", Icon: "fa-solid fa-fingerprint", MinRole: "guest", Category: "Core"},
		{Name: "Status", URL: "https://up.dhanur.me", Icon: "fa-solid fa-heart-pulse", MinRole: "guest", Category: "Core"},
	},
}

func loadManifest() ManifestResponse {
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		return defaultManifest
	}

	var manifest ManifestResponse
	if err := json.Unmarshal(data, &manifest); err != nil {
		slog.Error("Failed to parse manifest.json, using defaults", "err", err)
		return defaultManifest
	}
	if manifest.Version == "" {
		manifest.Version = defaultManifest.Version
	}
	if len(manifest.Apps) == 0 {
		slog.Warn("manifest.json contains no apps, using defaults")
		return defaultManifest
	}

	slog.Info("Loaded apps manifest from manifest.json", "count", len(manifest.Apps), "version", manifest.Version)
	return manifest
}

func roleLevel(role string) int {
	switch strings.ToLower(role) {
	case "admin":
		return 2
	case "user":
		return 1
	default:
		return 0
	}
}

func filterByRole(manifest ManifestResponse, role string) ManifestResponse {
	level := roleLevel(role)
	filtered := ManifestResponse{Version: manifest.Version, Apps: make([]AppEntry, 0, len(manifest.Apps))}
	for _, app := range manifest.Apps {
		if roleLevel(app.MinRole) <= level {
			filtered.Apps = append(filtered.Apps, app)
		}
	}
	return filtered
}

func handleManifest(manifest ManifestResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("role")))
		if role == "" {
			role = "guest"
		}

		filtered := filterByRole(manifest, role)
		w.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=86400")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filtered); err != nil {
			http.Error(w, "failed to encode manifest", http.StatusInternalServerError)
		}
	}
}
