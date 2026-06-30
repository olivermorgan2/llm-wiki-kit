# ADR-001: Adopt Go 1.24.x and goccy/go-yaml for the engine toolchain

**Status:** proposed
**Date:** 2026-06-30

## Context

The `llm-wiki` engine is a bundled, self-contained Go CLI — the deterministic
authority for parsing, validation, indexing, and safe filesystem writes
(PRD §1; `knowledge/project-brief.md`). Two toolchain choices must be ratified
before the first engine implementation issue: the Go language line and the YAML
library. Binding constraints:

- Acceptance criterion 6 (addendum 001) requires unknown frontmatter fields
  **and comments** to survive authoring and enrichment — i.e. YAML
  **round-trip preservation**.
- `gopkg.in/yaml.v3` is archived / unmaintained as of 2025.
- Users must not install Go, Python, Node, or third-party libraries (PRD §9).

MVP planning assumption-locked **Go 1.24.x + `github.com/goccy/go-yaml`** as the
ADR-001 *candidate* (open-questions Q2; addendum 002 A5; build-out-plan
§Assumptions and §"Decisions needing ADRs" #1) — explicitly reversible and to be
**ratified or revised here**.

## Options considered

### Option A: Go 1.24.x + github.com/goccy/go-yaml

- Pros: node/AST-aware parsing enables comment- and unknown-field-preserving
  round-trips, satisfying criterion 6 directly; actively maintained; Go 1.24.x
  is a conservative current-stable line with broad CI-runner coverage on all
  five target platforms; static single-binary builds need no user runtime
  (PRD §9); the YAML choice can be confined behind one internal adapter.
- Cons: goccy's node API is heavier than yaml.v3's plain Marshal/Unmarshal; a
  current Go line must be tracked for patch updates (Go has no LTS); a
  third-party YAML library is one more supply-chain dependency to checksum.

### Option B: Go 1.24.x + gopkg.in/yaml.v3

- Pros: historically the default, widely understood Marshal/Unmarshal API;
  zero migration for engineers who already know it.
- Cons: archived/unmaintained since 2025 (no security fixes); its decode model
  drops comments and is not reliably node-aware, so round-trip preservation
  (criterion 6) would need a brittle bolt-on; stakes a hard acceptance
  criterion on an abandoned library.

## Decision

Adopt **Option A** — Go 1.24.x with `github.com/goccy/go-yaml`. Criterion 6's
round-trip requirement makes a node-aware library non-negotiable, and yaml.v3's
archived status removes it as a responsible default; Go 1.24.x gives broad,
runtime-free cross-platform builds. The exact patch is pinned in `go.mod` at the
first-engine-issue time. YAML behavior is confined to the engine's YAML adapter
behind an internal interface, so a future revision changes only that adapter and
`go.mod`. This **ratifies** the MVP assumption-lock (Q2).

## Consequences

- Easier: criterion 6 round-trip preservation has a library built for it;
  runtime-free single-binary distribution on all five platforms; one place
  (the YAML adapter) owns all serialization behavior.
- Harder: engineers work against goccy's node API rather than plain struct
  marshalling; the team tracks the Go release cadence to keep 1.24.x patched.
- Maintain: the `go.mod` pin and the YAML-adapter interface; a round-trip
  fixture (Phase 3 gate, per `knowledge/risks.md`) guards regressions; the
  goccy dependency is checksum-tracked as supply-chain hygiene (see ADR-002).
- Deferred / validation implications: this ADR ratifies open-question Q2 —
  record in `knowledge/log.md`. Round-trip is verified by the criterion-6
  fixture owned in the build-out plan; a later toolchain change re-opens only
  this ADR. Markdown-parser choice and non-YAML serialization are out of scope.
