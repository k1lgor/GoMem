#!/bin/bash
# GoMem save-all — index entire project into persistent memory
# Usage: ./save-all.sh [project-directory]
# If no directory given, indexes the current directory.

set -e

PROJECT_DIR="${1:-.}"
GOMEM="$(dirname "$0")/../../gomem"

if [ ! -f "$GOMEM" ] && [ ! -f "$GOMEM.exe" ]; then
    # Try PATH
    GOMEM="$(which gomem 2>/dev/null || echo "")"
    if [ -z "$GOMEM" ]; then
        echo "Error: gomem binary not found."
        echo "Build it first: cd /path/to/gomem && go build -o gomem ./cmd/gomem"
        echo "Then place it in PATH or run from the gomem project root."
        exit 1
    fi
fi

# Resolve to absolute path
cd "$PROJECT_DIR"
PROJECT_DIR="$(pwd)"

echo "=== GoMem: Indexing $PROJECT_DIR ==="
echo ""

# Use the gomem binary's save-all command
"$GOMEM" save-all "$PROJECT_DIR"
