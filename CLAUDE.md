# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build and Run
- `go build -o crossposting-service ./cmd/crossposting-service` - Build the service
- `go run ./cmd/crossposting-service` - Run the service directly

### Testing
- `make test` - Run all tests with race detection
- `go test -race ./...` - Run all tests manually
- `go test -race ./service/app -run TestHandlerName` - Run a specific test
- `go test -race ./service/... -v` - Run tests with verbose output

### Linting and Formatting
- `make fmt` - Format code using gosimports
- `make lint` - Run go vet and golangci-lint
- `make tools` - Install required linting tools

### CI Pipeline Check
- `make ci` - Run the complete CI pipeline locally (test, lint, generate, fmt, tidy, check for uncommitted changes)

### Frontend
- `make frontend` - Build frontend files (Vue.js) and embed them in the Go binary

## High-level Architecture

This is a Nostr-to-Twitter crossposting service with the following architecture:

### Core Components

1. **Application Layer** (`service/app/`)
   - Handlers for HTTP endpoints and business logic
   - `Downloader` - Downloads Nostr events from relays
   - Event processors that convert Nostr notes to tweets
   - Authentication handlers for Twitter OAuth

2. **Domain Layer** (`service/domain/`)
   - Core business entities: `Event`, `PublicKey`, `Account`, `Tweet`, `LinkedPublicKey`
   - Tweet generation logic that converts Nostr events to Twitter-compatible format
   - Session management for authenticated users

3. **Adapters Layer** (`service/adapters/`)
   - SQLite repositories for persistence
   - Twitter API client for posting tweets
   - Nostr relay connections via websockets
   - Prometheus metrics adapter
   - Internal pubsub implementation

4. **Ports Layer** (`service/ports/`)
   - HTTP server and routing
   - SQLite-based pubsub for reliable message delivery
   - Memory-based pubsub for real-time events

### Data Flow

1. **Nostr Event Processing**:
   - Downloader connects to Nostr relays (fetched from Purple Pages)
   - Events are received and published to memory pubsub
   - ProcessReceivedEventHandler filters and processes valid notes
   - TweetCreatedEvent is published to SQLite pubsub queue

2. **Tweet Posting Flow**:
   - TweetCreatedEventSubscriber picks up events from queue
   - SendTweetHandler attempts to post to Twitter
   - Failed tweets are retried with exponential backoff
   - Old tweets (several days) are dropped

3. **User Management**:
   - Users authenticate via Twitter OAuth
   - Link Nostr public keys (npubs) to their Twitter account
   - Events from linked npubs are crossposted automatically

### Key Design Decisions

- **SQLite pubsub**: Provides persistent queueing for reliability
- **Relay discovery**: Uses Purple Pages API instead of user-provided relay lists
- **Error handling**: Graceful handling of Twitter API rate limits and auth failures
- **Frontend**: Vue.js SPA embedded in Go binary for easy deployment
- **Dependency injection**: Uses Google Wire for clean dependency management

## Important Patterns

- Repository interfaces defined in `app.go` for clean architecture
- Domain entities use value objects with validation in constructors
- Handlers follow a consistent pattern with metrics and error tracking
- All database operations use transactions via `TransactionProvider`