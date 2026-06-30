#!/usr/bin/env bash
# json-envelope.sh — the shared JSON response envelope and exit-code
# convention for kit bin/* scripts (M2 Issue 8, ADR-047).
#
# Source this file from a bin/* script:
#     . "$(dirname "$0")/lib/json-envelope.sh"
#
# It defines the standard exit-code constants and a `print_envelope`
# helper. New scripts that support `--format json` SHOULD emit the
# standard envelope and use the subset of exit codes they need. The
# canonical description of both lives in docs/agent-contract.md.
#
# Standard envelope (one JSON object):
#   {
#     "skill":   "<script name>",
#     "version": "<envelope schema version>",
#     "status":  "<machine-readable status string>",
#     "outputs": { ... | [ ... ] },   # script-specific result payload
#     "next":    [ {"skill": "...", "args": "...", "when": "..."} ],
#     "errors":  [ {"code": "...", "message": "..."} ]
#   }
#
# Exit-code convention (a script uses the subset it needs):
#   0  success
#   1  domain failure        (the script ran, the answer is "no/failed")
#   2  invocation error      (bad flags, unreadable input, missing file)
#   3  auth/service failure  (gh/network/credential problem)
#   4  user cancellation     (operator declined at a prompt)
#
# bin/check-plan predates this convention and keeps its own
# {criteria-set, result, criteria[]} shape and 0/1/2 codes per
# ADR-043; it is the documented legacy surface, not a counter-example.

# Exit-code constants.
ENVELOPE_EXIT_OK=0
ENVELOPE_EXIT_DOMAIN=1
ENVELOPE_EXIT_INVOCATION=2
ENVELOPE_EXIT_AUTH=3
ENVELOPE_EXIT_CANCEL=4

# Envelope schema version. Bump only on a breaking shape change.
ENVELOPE_VERSION="1"

# print_envelope SKILL STATUS [OUTPUTS_JSON] [NEXT_JSON] [ERRORS_JSON]
#
# Assembles and prints the standard envelope. OUTPUTS_JSON / NEXT_JSON /
# ERRORS_JSON must be valid JSON (objects/arrays); they default to {} / []
# / []. Serialization and validation go through python3 so the emitted
# document is always well-formed; an invalid fragment fails closed with
# exit 2.
print_envelope() {
  local skill="$1" status="$2"
  local outputs="${3:-}" next="${4:-}" errors="${5:-}"
  [ -n "$outputs" ] || outputs='{}'
  [ -n "$next" ] || next='[]'
  [ -n "$errors" ] || errors='[]'
  ENV_SKILL="$skill" ENV_STATUS="$status" ENV_VERSION="$ENVELOPE_VERSION" \
  ENV_OUTPUTS="$outputs" ENV_NEXT="$next" ENV_ERRORS="$errors" \
  python3 -c '
import os, json, sys
def parse(name, default):
    raw = os.environ.get(name) or default
    try:
        return json.loads(raw)
    except Exception as e:
        sys.stderr.write("json-envelope: invalid %s JSON: %s\n" % (name, e))
        sys.exit(2)
doc = {
    "skill":   os.environ["ENV_SKILL"],
    "version": os.environ["ENV_VERSION"],
    "status":  os.environ["ENV_STATUS"],
    "outputs": parse("ENV_OUTPUTS", "{}"),
    "next":    parse("ENV_NEXT", "[]"),
    "errors":  parse("ENV_ERRORS", "[]"),
}
print(json.dumps(doc, indent=2, ensure_ascii=False))
'
}
