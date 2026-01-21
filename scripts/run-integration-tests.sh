#!/bin/bash

set -e

echo "🚀 Starting Integration Test Environment..."

# Check if Docker Compose is running
if ! docker-compose ps | grep -q "Up"; then
  echo "📦 Starting Docker Compose..."
  make up
  
  echo "⏳ Waiting for services to be healthy..."
  sleep 30
  
  # Wait for gateway to be healthy
  echo "🔍 Checking Gateway health..."
  max_retries=30
  retry=0
  while [ $retry -lt $max_retries ]; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
      echo "✅ Gateway is healthy!"
      break
    fi
    echo "⏳ Waiting for gateway... ($retry/$max_retries)"
    sleep 2
    retry=$((retry + 1))
  done
  
  if [ $retry -eq $max_retries ]; then
    echo "❌ Gateway did not become healthy"
    docker-compose logs gateway
    exit 1
  fi
else
  echo "✅ Docker Compose is already running"
fi

echo ""
echo "🧪 Running Integration Tests..."
cd services/gateway
go test -v ./test/integration/... -timeout 10m

echo ""
echo "✅ Integration tests completed!"
