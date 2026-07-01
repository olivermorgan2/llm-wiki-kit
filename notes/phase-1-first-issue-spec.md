# Phase 1 — first implementation issue (pre-issue spec)

**Status:** local pre-issue spec — **no GitHub issue created yet.**
**Created:** 2026-07-01 (on acceptance of ADR-001–005).
**Owner:** Oliver.

> This is a working spec in `notes/`, not a finalized per-issue prompt. Per-issue
> prompts are GitHub-issue-derived and immutable, and live in `prompts/issue-NNN-*.md`
> (written by `/prepare-issue`). This file bridges accepted ADRs → the first
> Phase 1 issue so `/issue-planner` + `/prepare-issue` have a concrete starting
> point. It references, and does not restate,
> [`design/build-out-plan.md`](../design/build-out-plan.md) §"Phase 1" and the
> accepted ADRs.

## Working title

**Scaffold the deterministic `llm-wiki` CLI skeleton + versioned JSON-contract spine.**

## Why this is the first issue

Phase 1 (Slice 0, must-pass, blocks everything) is the deterministic engine every
other surface calls. The very first slice of that is the **CLI skeleton + contract
shell** — the callable-without-Claude-Code authority that later Phase 1 work
(`validate`, filesystem safety) plugs into. It is enabled entirely by ADRs now
**accepted**: ADR-001 (toolchain), ADR-003 (contract + exit codes), ADR-004
(validation/severity *model* the envelope carries), ADR-005 (filesystem-safety
*primitives interface* the skeleton wires but does not yet implement in depth).

Binary selection (ADR-002 — OS/arch detect + checksum verify) is a **separate
Phase 1 issue**; it is intentionally out of this first issue to keep the skeleton
reviewable. This issue must not drift into installer/asset-ownership work
(that is **ADR-009**, not authorized by ADR-002).

## Enabling ADRs (all accepted 2026-07-01)

| ADR | What it authorizes for this issue |
|---|---|
| [ADR-001](../design/adr/adr-001-go-toolchain-and-yaml.md) | Go 1.24.x module + `goccy/go-yaml` behind an internal YAML adapter interface. Pin the exact Go patch in `go.mod` here. |
| [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md) | The versioned JSON envelope `{contractVersion, operation, status, findings, affectedPaths, approval}`; `--json` opt-in on every command (human-readable default); six exit-code **semantic buckets**. |
| [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) | The finding shape the envelope carries: ruleset tag (OKF vs profile) + three severities (Error/Warning/Suggestion). Skeleton defines the finding type; rule content is later Phase 1 work. |
| [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) | The `fsafe` chokepoint **interface** (canonicalize + boundary-check + symlink-resolve + per-file atomic write). Skeleton wires the seam; full guard fixtures are a later Phase 1 issue. |

## Scope (this issue)

- Go 1.24.x module init; `go.mod` pin; `cmd/llm-wiki` entry point.
- `internal/` package seams per the build-out plan repo layout: `contract/`,
  `validate/`, `profile/`, `fsafe/` (stubs/interfaces are fine where full
  behavior is a later Phase 1 issue).
- `internal/contract`: the versioned JSON envelope type + serializer, the six
  exit-code **semantic buckets** as named constants, and a `--json` flag (or
  `LLM_WIKI_JSON` toggle) honored on every command.
- A trivial command (e.g. `llm-wiki version` and a no-op `validate` shell) that
  emits a schema-valid envelope under `--json` and a human-readable form by
  default — proving the contract is callable **without Claude Code**.
- Unit tests: envelope round-trips its schema; each exit-code bucket maps to the
  intended PRD §12 outcome; `--json` vs default output selection.

## Out of scope (name explicitly in the issue to prevent scope creep)

- **Real validation rules / broken-link detection** — later Phase 1 issue (ADR-004).
- **Filesystem-safety guard fixtures** (traversal/symlink) — later Phase 1 issue (ADR-005).
- **OS/arch detection + checksum verification** — separate Phase 1 issue (ADR-002).
- **Install / init / upgrade / uninstall / asset ownership** — Phase 2 / **ADR-009**.
- **Staged mutation (`inspect`/`plan`/`apply`), hash-binding** — Phase 3 / **ADR-006**.

## Acceptance-criteria linkage

Advances (does not by itself close) criterion **14** (structured output conforms
to the contract + documented exit codes). Lays the spine criteria **5, 7, 8, 17**
build on in later Phase 1 issues.

## Carry-forwards (must not be lost)

1. **ADR-003 numeric exit-code table.** This issue fixes the six *semantic
   buckets* as named constants; the **exact numeric values** must be published as
   a stable, documented code→meaning table **before Phase 1 closes** (ADR-003
   deferred them to implementation). Once published they are a frozen public
   surface.
2. **ADR-006 before any cross-file mutation.** This skeleton must not implement
   multi-file transaction / staged-mutation semantics. The cross-file
   transaction model (staging manifest, commit ordering, recovery/rollback,
   partial-commit detection, hash-bound stale-plan rejection) is **owned by
   ADR-006, which must be drafted first**. ADR-005 gives per-file atomicity only.
3. **ADR-002 ≠ ADR-009.** Accepted ADR-002 authorizes ship/select/verify of the
   platform binary **only**. Full installer / asset-ownership and release
   signing/provenance remain **ADR-009** — do not begin that work under an ADR-002
   banner.

## Next workflow steps

1. Draft **ADR-006** (staged mutation model) — next-needed ADR; required before
   Phase 3 authoring and any cross-file mutation.
2. File the Phase 1 backlog as GitHub issues via `/issue-planner` (build-out plan
   §"Foundation — Phase 1"), then `/prepare-issue` this first issue into
   `prompts/issue-NNN-cli-contract-spine.md`.
3. Implement via `/claude-issue-executor` (plan-first, one issue per session).
