# SearXNG Configuration Placeholder

## Where to Configure

Add this section to your `/Users/wavegoodvybe/.picoclaw/config.json`:

```json
{
  "searxng": {
    "url": "http://localhost:8080"
  }
}
```

## Setting up SearXNG

SearXNG is a privacy-respecting metasearch engine that runs locally.

### Option 1: Docker (Recommended)

```bash
# Pull the image
docker pull searxng/searxng

# Run SearXNG
docker run -d \
  --name searxng \
  -p 8080:8080 \
  searxng/searxng
```

### Option 2: Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3'
services:
  searxng:
    image: searxng/searxng
    ports:
      - "8080:8080"
    volumes:
      - ./searxng:/etc/searxng
    restart: unless-stopped
```

Then run:
```bash
docker-compose up -d
```

### Option 3: Native Installation

See: https://docs.searxng.org/admin/installation.html

## Verify SearXNG is Running

```bash
curl http://localhost:8080
```

## Test Search with PicoClaw

```bash
cd ~/.picoclaw/workspace
./bin/search "hello world"
```

This should create `web/last_search.json` with search results.
