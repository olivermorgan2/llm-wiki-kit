#!/usr/bin/env bash
# .github/hooks/pre-commit.sh
# Runs llm-wiki validate and outputs JSON findings and metrics.

set -euo pipefail

# Resolve repo root
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

BIN=".bin.llm-wiki"  # assume built binary
if [[ ! -x "$BIN" ]]; then
    echo '{"findings": ["llm-wiki binary not found"], "metrics": {}}'
    exit 1
fi

OUTPUT=$("$BIN" validate 2>&1) || true
# Wrap raw output into JSON
printf '{"findings": [%q], "metrics": {}}' "$OUTPUT"
