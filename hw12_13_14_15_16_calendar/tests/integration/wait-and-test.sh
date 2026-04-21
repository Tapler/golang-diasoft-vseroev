#!/bin/sh
set -e

echo "Waiting for services to be ready..."

# Ждем API
until curl -s http://calendar:8080/api/v1/events/day/2025-01-01 > /dev/null 2>&1; do
  echo "Waiting for Calendar API..."
  sleep 2
done

echo "Calendar API is ready!"

# Ждем еще немного для полной инициализации
sleep 5

echo "Running integration tests..."
go test -v -count=1 -timeout=3m ./...
