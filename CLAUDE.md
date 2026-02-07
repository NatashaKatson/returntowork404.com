# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

returntowork404.com is a humorous landing page for parents returning to work after parental leave. It has two parts:

1. **Landing page** (`index.html` at root) — a static, self-contained single-page scrolling site hosted on GitHub Pages (CNAME: returntowork404.com). All CSS/JS is inline.
2. **"What Did I Miss?" app** (`static/` + Go backend) — a separate web app where users pick an industry and time period, and get an AI-generated summary of what they missed. This runs as a Docker-deployed service.

## Build & Run Commands

```bash
make dev          # Run Go backend locally (needs GEMINI_API_KEY env var)
make build        # Build Docker images
make up           # Start all services (reproxy + api)
make down         # Stop all services
make restart      # Rebuild and restart
make test         # Run Go tests: go test -v ./...
make logs         # Tail Docker logs
```

Local dev requires: `export GEMINI_API_KEY=<key>`.

## Architecture

The Go backend (module name: `whatdidimiss`) uses chi router and has three packages:

- **`main.go`** — HTTP server setup, chi router with middleware (logger, recoverer, CORS, gzip). Serves `static/` files and mounts `/api` routes.
- **`handlers/`** — API handler with `POST /api/catchup` and `GET /api/health`. Validates industry/time_period against hardcoded allowlists, checks in-memory cache, falls back to Gemini API. Cache key format: `{industry}:{time_period}`.
- **`gemini/`** — HTTP client for Google Gemini API. Builds a prompt from industry + time period and returns the text response.
- **`cache/`** — In-memory cache with Get/Set/Close. TTL is 7 days. Cache is lost on server restart.

Production uses Docker Compose with two services: **reproxy** (reverse proxy with auto Let's Encrypt SSL, serves static files, proxies `/api/*`) and **api** (Go binary).

## Adding Industries or Time Periods

Update three places in sync:
1. `handlers/api.go` — add to `validIndustries`/`validTimePeriods` slices and their label maps
2. `static/index.html` — add `<option>` elements to the corresponding `<select>`

## Docker

Multi-stage `Dockerfile`: builds with `golang:1.22-alpine`, runs on `alpine:3.19`. Binary is named `server`. Static files are copied into the image at `./static`.

## Known Stale Docs

`README.md` still references Redis (commands, architecture diagram, env vars, local dev section). It needs to be updated to reflect the in-memory cache change.

## Environment Variables

| Variable | Required | Default |
|----------|----------|---------|
| `GEMINI_API_KEY` | Yes | — |
| `PORT` | No | 8080 |
| `ACME_EMAIL` | No | admin@returntowork404.com |
