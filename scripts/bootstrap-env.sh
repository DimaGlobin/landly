#!/bin/bash
# Bootstrap .env files from .env.example if they don't exist
# This script ensures developers have working .env files out of the box

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🔧 Bootstrapping environment files..."

# Bootstrap backend .env
BACKEND_ENV="$ROOT_DIR/apps/backend/.env"
BACKEND_EXAMPLE="$ROOT_DIR/apps/backend/.env.example"

if [ ! -f "$BACKEND_ENV" ]; then
    if [ -f "$BACKEND_EXAMPLE" ]; then
        echo "  📝 Creating apps/backend/.env from .env.example"
        cp "$BACKEND_EXAMPLE" "$BACKEND_ENV"
    else
        echo "  ⚠️  Warning: apps/backend/.env.example not found"
    fi
else
    echo "  ✅ apps/backend/.env already exists (skipping)"
fi

# Bootstrap frontend .env
FRONTEND_ENV="$ROOT_DIR/apps/frontend/.env"
FRONTEND_EXAMPLE="$ROOT_DIR/apps/frontend/.env.example"

if [ ! -f "$FRONTEND_ENV" ]; then
    if [ -f "$FRONTEND_EXAMPLE" ]; then
        echo "  📝 Creating apps/frontend/.env from .env.example"
        cp "$FRONTEND_EXAMPLE" "$FRONTEND_ENV"
    else
        echo "  ⚠️  Warning: apps/frontend/.env.example not found"
    fi
else
    echo "  ✅ apps/frontend/.env already exists (skipping)"
fi

echo "✅ Environment bootstrap complete!"

