#!/usr/bin/env bash
# Apply the hermes-workflow-overlay to a kit-bootstrapped project.
#
# Usage:
#   install-overlay.sh /path/to/project
#
# Idempotent-ish: skips knowledge/ if present, refuses to double-append
# the CLAUDE.md section. Does NOT commit and does NOT enable branch
# protection — do those explicitly:
#   fresh repo:    commit directly as part of initial setup
#   existing repo: land via PR
# then, after the guard workflow has run once:
#   scripts/protect-branch.sh owner/repo guard "<test job names...>"

set -euo pipefail

TARGET="${1:?usage: install-overlay.sh /path/to/project}"
HERE="$(cd "$(dirname "$0")/.." && pwd)"

[ -d "$TARGET/.git" ] || { echo "error: $TARGET is not a git repo"; exit 1; }
[ -f "$TARGET/CLAUDE.md" ] || { echo "error: $TARGET has no CLAUDE.md — bootstrap the kit first"; exit 1; }

# 1. Guard workflow
mkdir -p "$TARGET/.github/workflows"
cp "$HERE/.github/workflows/workflow-guard.yml" "$TARGET/.github/workflows/"
echo "installed .github/workflows/workflow-guard.yml"

# 2. Knowledge layer skeleton
if [ -d "$TARGET/knowledge" ]; then
  echo "skipped knowledge/ (already exists)"
else
  cp -r "$HERE/knowledge" "$TARGET/knowledge"
  echo "installed knowledge/ skeleton"
fi

# 3. CLAUDE.md hardening section
if grep -q "Hermes hardened workflow" "$TARGET/CLAUDE.md"; then
  echo "skipped CLAUDE.md append (overlay section already present)"
else
  cat "$HERE/CLAUDE-overlay.md" >> "$TARGET/CLAUDE.md"
  echo "appended overlay section to CLAUDE.md"
fi

echo
echo "Done. Next: commit these changes (direct on a fresh repo, PR on an"
echo "existing one), let 'guard' run once, then run protect-branch.sh."
