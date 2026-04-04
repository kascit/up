# UP — dhanur.me status monitor

Real-time service health dashboard for all `dhanur.me` services. Written in Go, zero external dependencies, deployed on Render.

## Architecture

- **Backend**: Pure Go standard library (`net/http`, `log/slog`, `sync`). No frameworks.
- **Health checks**: Concurrent goroutines probe every service every **30 seconds**.
- **History**: 90 data points per service (~45 min rolling window), stored in-memory.
- **API**: `GET /api/status` → JSON snapshot; `GET /api/health` → self-health.
- **Frontend**: Single embedded HTML file served at `/`, auto-refreshes every 30 s.

## Adding / Editing Services

Create a `services.json` file next to the binary:

```json
[
  {
    "name":        "my-service",
    "displayName": "My Service",
    "url":         "https://my-service.dhanur.me/api/health",
    "homepageUrl": "https://my-service.dhanur.me",
    "icon":        "fa-solid fa-server",
    "category":    "Core"
  }
]
```

If `services.json` is absent the compiled-in defaults are used (authy, tasks, linkr, dhanur.me).

## Running locally

```bash
go run .
# open http://localhost:8080
```

## Render deployment

The `render.yaml` in this repo configures a Go web service. Just connect the GitHub repo in the Render dashboard and it deploys automatically on every push to `main`.

Set `up.dhanur.me` as the custom domain for the service in Render.
