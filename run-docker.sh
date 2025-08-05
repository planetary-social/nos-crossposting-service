#!/bin/bash

echo "Starting Nos Crossposting Service with Docker Compose..."
echo "========================================================"
echo ""

# Check if Twitter keys are provided via environment
if [ -z "$CROSSPOSTING_TWITTER_KEY" ] || [ "$CROSSPOSTING_TWITTER_KEY" = "your-twitter-key-here" ]; then
    echo "⚠️  WARNING: No Twitter API keys set!"
    echo ""
    echo "The service will run in DEVELOPMENT mode with a fake Twitter adapter."
    echo "To use real Twitter API, set these environment variables:"
    echo ""
    echo "  export CROSSPOSTING_TWITTER_KEY=your-actual-key"
    echo "  export CROSSPOSTING_TWITTER_KEY_SECRET=your-actual-secret"
    echo ""
    echo "Then run this script again."
    echo ""
fi

echo "Services:"
echo "  • Web UI:      http://localhost:8008"
echo "  • Metrics:     http://localhost:8008 (port 8009 not exposed)"
echo "  • Redis:       localhost:6379"
echo "  • Nostr Relay: ws://localhost:7777"
echo ""
echo "Press Ctrl+C to stop"
echo "========================================================"
echo ""

# Run docker compose
docker compose up