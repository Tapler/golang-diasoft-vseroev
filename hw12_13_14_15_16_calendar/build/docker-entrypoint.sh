#!/bin/sh
set -e

# Ждем, пока DNS не будет доступен
echo "Waiting for DNS to be ready..."
until getent hosts postgres > /dev/null 2>&1; do
  echo "Waiting for postgres DNS resolution..."
  sleep 2
done

echo "DNS is ready, starting application..."
exec "$@"
