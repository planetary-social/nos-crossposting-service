# Local Development Setup

This guide helps you get the Nos Crossposting Service running locally.

## Quick Start with Docker Compose (Recommended)

This is the easiest way to run the service with all dependencies:

```bash
# Run with Docker Compose
./run-docker.sh

# Or manually:
docker compose up
```

This will start:
- **Crossposting Service** on http://localhost:8008
- **Redis** on localhost:6379
- **Nostr Relay** on ws://localhost:7777

### Using Real Twitter API Keys

To use real Twitter API keys (instead of the fake development adapter):

```bash
# Set your Twitter API credentials
export CROSSPOSTING_TWITTER_KEY=your-actual-key
export CROSSPOSTING_TWITTER_KEY_SECRET=your-actual-secret

# Run the service
./run-docker.sh
```

## Alternative: Run Locally without Docker

If you prefer to run without Docker, you'll need:
1. Redis running locally
2. Go 1.21 or later

```bash
# Start Redis (in a separate terminal)
redis-server

# Set required environment variables
export REDIS_URL=redis://localhost:6379
export CROSSPOSTING_TWITTER_KEY=test
export CROSSPOSTING_TWITTER_KEY_SECRET=test
export CROSSPOSTING_DATABASE_PATH=./local-data/crossposting.sqlite
export CROSSPOSTING_ENVIRONMENT=DEVELOPMENT
export CROSSPOSTING_PUBLIC_FACING_ADDRESS=http://localhost:8008/

# Run the service
go run ./cmd/crossposting-service
```

## What's Running?

Once started, you can access:

- **Web UI**: http://localhost:8008 - The main interface for managing crossposting
- **API Endpoints**:
  - `/api/current-user` - Current user info
  - `/api/current-user/public-keys` - Manage linked Nostr public keys
  - `/login` - Twitter OAuth login

## Development Mode

When `CROSSPOSTING_ENVIRONMENT=DEVELOPMENT` is set:
- Uses a fake Twitter adapter (doesn't actually post to Twitter)
- Returns hardcoded Twitter account details
- Perfect for testing without real API keys

## Troubleshooting

1. **Port already in use**: Stop any existing services on ports 8008, 6379, or 7777
2. **Docker not found**: Install Docker Desktop
3. **Build errors**: Make sure you have Go 1.21 or later
4. **Redis connection errors**: Ensure Redis is running if not using Docker