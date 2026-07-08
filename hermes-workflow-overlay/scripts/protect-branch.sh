#!/usr/bin/env bash
# Enable hard branch protection on main.
#
# Usage:
#   protect-branch.sh owner/repo [required-check-name ...]
#
# Defaults to requiring the "guard" check. Pass your test-matrix job names
# too, e.g.:
#   protect-branch.sh olivermorgan2/llm-wiki-kit guard "test (linux-amd64)" "test (windows-amd64)"
#
# Requires: gh authenticated with admin rights on the repo.
# Run this AFTER any history surgery (it blocks force pushes) and AFTER the
# guard workflow has run once (so the check name exists).

set -euo pipefail

REPO="${1:?usage: protect-branch.sh owner/repo [check ...]}"
shift || true
CHECKS=("$@")
[ ${#CHECKS[@]} -eq 0 ] && CHECKS=("guard")

ctx=$(printf '"%s",' "${CHECKS[@]}")
ctx="[${ctx%,}]"

gh api -X PUT "repos/$REPO/branches/main/protection" --input - <<EOF
{
  "required_status_checks": { "strict": true, "contexts": $ctx },
  "enforce_admins": true,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_linear_history": false,
  "required_conversation_resolution": false
}
EOF

echo "Protected main on $REPO. Required checks: ${CHECKS[*]}"
echo "Note: PR review-approval is intentionally not required (autonomous agent"
echo "cannot self-approve); the guard + test checks are the merge gate."
