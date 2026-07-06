#!/usr/bin/env bash
# .github/hooks/pre-commit.sh
# Runs llm-wiki validate and outputs JSON findings and metrics.

set -euo pipefail

# Resolve repo root
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

# Detect OS and build binary accordingly
OS="$(uname -s)"
BIN="bin/llm-wiki"

if [[ "$OS" == "Darwin" ]]; then
  # macOS build
  GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BIN" ./cmd/llm-wiki
elif [[ "$OS" == "Linux" ]]; then
  # Linux build  
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BIN" ./cmd/llm-wiki
elif [[ "$OS" =~ MINGW || "$OS" =~ MSYS ]]; then
  # Windows build
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "${BIN}.exe" ./cmd/llm-wiki
  BIN="${BIN}.exe"
fi

# Verify binary exists
if [[ ! -x "$BIN" ]]; then
  echo "{\"findings\": [\"llm-wiki binary not found\"], \"metrics\": {}}"
  exit 1
fi

# Run validation and capture output
VALIDATION_OUTPUT="$($BIN validate)"
if [[ $? -ne 0 ]]; then
  VALIDATION_OUTPUT="${VALIDATION_OUTPUT}"
fi

# Convert to JSON format expected by parity test
echo "{\"findings\": [${VALIDATION_OUTPUT}], \"metrics\": {}}"