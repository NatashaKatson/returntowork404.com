# What Did I Miss?

A web app that catches you up on what happened in your industry while you were away, powered by Claude AI.

## Quick Start (Production)

### 1. Server Setup (DigitalOcean $5 Droplet)

```bash
# SSH into your server
ssh root@your-server-ip

# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Docker Compose
apt install docker-compose-plugin

# Clone the repo
git clone https://github.com/yourusername/whatdidimiss.git
cd whatdidimiss

# Setup environment
make setup

# Edit .env and add your Claude API key
nano .env
```

### 2. Configure DNS

Point your domain `returntowork404.com` to your server's IP address:
- A record: `@` → `your-server-ip`
- A record: `www` → `your-server-ip`

### 3. Deploy

```bash
make up
```

The app will automatically obtain SSL certificates from Let's Encrypt.

## Commands

```bash
make help        # Show all commands
make build       # Build Docker images
make up          # Start services
make down        # Stop services
make restart     # Rebuild and restart
make logs        # View logs
make update      # Pull latest code and restart
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                     Internet                         │
└─────────────────────┬───────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│  Reproxy (HTTPS + Static Files)                     │
│  - Auto SSL via Let's Encrypt                       │
│  - Serves static files (/, /style.css, etc.)        │
│  - Proxies /api/* to backend                        │
└─────────────────────┬───────────────────────────────┘
                      │
           ┌──────────┴──────────┐
           │                     │
           ▼                     ▼
┌──────────────────┐   ┌──────────────────┐
│  Go API Server   │   │  Static Files    │
│  (port 8080)     │   │  (HTML/CSS/JS)   │
└────────┬─────────┘   └──────────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌────────┐ ┌────────────┐
│In-Mem  │ │ Claude API │
│ Cache  │ │            │
└────────┘ └────────────┘
```

## API

### POST /api/catchup

Request:
```json
{
  "industry": "software-development",
  "time_period": "1-year"
}
```

Response:
```json
{
  "summary": "...",
  "industry": "Software Development",
  "period": "1 year",
  "cached": false
}
```

### Valid Industries
- `software-development`
- `marketing`
- `healthcare`
- `legal`

### Valid Time Periods
- `6-months`
- `1-year`
- `2-3-years`
- `5-years`
- `10-years`

## Adding More Industries

Edit `handlers/api.go`:

```go
var validIndustries = []string{
    "software-development",
    "marketing",
    "healthcare",
    "legal",
    "finance",        // Add new ones here
    "education",
}

var industryLabels = map[string]string{
    // ... existing entries
    "finance":   "Finance",
    "education": "Education",
}
```

Then update `static/index.html` to add the new `<option>` elements.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CLAUDE_API_KEY` | Yes | Your Anthropic API key |
| `ACME_EMAIL` | No | Email for Let's Encrypt notifications |
| `PORT` | No | API server port (default: 8080) |

## Local Development

```bash
# Set API key
export CLAUDE_API_KEY=sk-ant-api03-xxxxx

# Run the server
make dev

# Visit http://localhost:8080
```

Note: The app uses an in-memory cache (7-day TTL) — no external dependencies like Redis are needed. Cache is lost on server restart.

## License

MIT
