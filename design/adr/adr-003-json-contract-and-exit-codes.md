# ADR-003: Versioned JSON skill–engine contract (v1) and a fixed exit-code set

**Status:** accepted
**Date:** 2026-06-30

## Context

The plugin must define a stable, documented contract between skills and the Go
engine; skills are thin adapters and core behavior must be callable directly
from the CLI without Claude Code (PRD §12). Every structured response must
identify **contract version, operation, status, findings, affected paths, and
any required approval** (PRD §12; acceptance criterion 14). The CLI must document
stable exit codes for at least: success, success-with-warnings, validation
failure, approval required, invalid invocation/configuration, and
system/filesystem failure (PRD §12).

MVP assumption-locked: the contract **starts at v1** with **no backward-compat
guarantee until first public release**; skills and engine ship in lockstep in
one versioned plugin (open-questions Q8; addendum 002 A4; build-out-plan #3).
This addresses the "divergent findings across skills/hooks/CI/CLI" risk
(one deterministic implementation → identical findings, criterion 15;
`knowledge/risks.md`). The Codex PRD review flagged the contract as blocking.

## Options considered

### Option A: One versioned JSON envelope shared by all surfaces, plus a fixed six-value exit-code set; break freely pre-release, freeze at first release

- Pros: one envelope `{contractVersion, operation, status, findings,
  affectedPaths, approval}` satisfies criterion 14 directly; an identical
  envelope across skills/hooks/CI/CLI is what makes "same findings for same
  state" (criterion 15) achievable; six documented exit codes map 1:1 to PRD
  §12's required outcomes; lockstep shipping makes pre-release breakage cheap and
  honest (addendum 002 A4); human-readable output stays the default for direct
  terminal use.
- Cons: every machine surface emits/parses the full envelope even for trivial
  commands; freezing at v1 means a future breaking change needs a v2 and a
  migration story; exit-code semantics become a stable public surface forever.

### Option B: Per-command ad-hoc JSON shapes with conventional 0/non-zero exit codes

- Pros: less boilerplate per command; familiar Unix exit convention.
- Cons: no single contract to validate against (criterion 14 fails); "approval
  required" and "validation failure" collapse into one non-zero code, losing the
  deterministic outcome distinctions PRD §12 demands; divergent shapes invite the
  cross-surface drift the project must avoid.

## Decision

Adopt **Option A** — one versioned JSON envelope carrying `{contractVersion,
operation, status, findings, affectedPaths, approval}` shared by every surface,
with a fixed, documented exit-code set covering success / success-with-warnings /
validation-failure / approval-required / invalid-invocation /
system-or-filesystem-failure. v1 may break during pre-release (skills and engine
move together) and freezes its compatibility promise at first public release, per
addendum 002 A4. This is the contract every adapter and CI invoker depends on and
the precondition for identical findings across surfaces (criterion 15). Advances
open-question Q8.

**Output-mode selection.** Human-readable text is the **default** for a direct
interactive terminal; the JSON envelope is opt-in via an explicit **`--json`**
flag (or equivalent, e.g. a `LLM_WIKI_JSON` env toggle) available on **every**
command. Machine surfaces (skills, hooks, CI) always pass `--json`. So "one
envelope shared by every surface" means one *schema* whenever JSON is emitted —
not that JSON is the unconditional default output.

**Exit-code values.** This ADR fixes the **six semantic buckets** (success /
success-with-warnings / validation-failure / approval-required /
invalid-invocation / system-or-filesystem-failure) and their 1:1 mapping to PRD
§12 outcomes. The **exact numeric values** are **deferred to the implementation
issue**, which must publish the stable code→meaning table before Phase 1 closes
(the "documented stable exit codes" acceptance hook); once published they are a
frozen public surface.

## Consequences

- Easier: criterion 14 conformance is a single schema to test; one envelope makes
  cross-surface parity (criterion 15) tractable; exit codes give callers
  deterministic branching.
- Harder: all machine surfaces carry full-envelope overhead; post-release
  evolution requires a versioned v2 plus migration.
- Maintain: the envelope schema, the exit-code table, and a contract-conformance
  fixture; exit-code semantics are a stable public surface from first release.
- Deferred / validation implications: advances Q8 (record in `knowledge/log.md`).
  Hash-binding of mutation plans and stale-plan rejection (criterion 13) are
  referenced here but **specified in ADR-006**. Backward-compat policy beyond v1
  is deferred to a future ADR. Criteria 14 and 15 are the validation hooks.
