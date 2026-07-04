You are working in my `llm-wiki-kit` repository.

Context:
- A versioned Claude Code plugin paired with a bundled, self-contained Go
  `llm-wiki` engine that lets a technical researcher or team create and
  maintain portable, repository-native knowledge bundles in the Open Knowledge
  Format (OKF v0.1) — structured, validated, and agent-readable without a
  hosted wiki.
- Follow the rules in `CLAUDE.md`.
- This is a Phase 3 (Authoring + staged mutation) acceptance-corpus issue. The
  phase scope is `design/build-out-plan.md` §"Phase 3". The Phase 3 behaviors
  (`page inspect`/`plan`/`apply`, the citation model, the citation-loss gate, the
  authoring adapter) already shipped in #34/#35/#36/#37/#38/#42 with scattered
  per-command and skill-flow tests; what is missing is a named,
  criterion-traceable acceptance corpus that runs as a legible per-platform gate
  in CI, mirroring the Phase 2 corpus, plus the gate-evidence artefact.

ADR:
- File: `design/adr/adr-003-json-contract-and-exit-codes.md`
- Decision: One versioned JSON envelope `{contractVersion, operation, status,
  findings, affectedPaths, approval}` shared by every surface, plus one optional
  page/plan/apply payload; six semantic exit-code buckets.
- File: `design/adr/adr-004-validation-severity-model.md`
- Decision: A three-level configurable severity model; a citation resolution
  failure is a warning by default (relevant to criterion 9's negative control).
- File: `design/adr/adr-006-staged-mutation-transaction-model.md`
- Decision: Mutations run through a staged transaction; `page plan` previews a
  diff and stages under `.llm-wiki/staging/<txn>/`, `page apply` commits it; a
  plan binds the target's base state, so an out-of-band change makes apply reject
  as stale with zero mutation; an identical re-plan is a no-op; the OKF round-trip
  is loss-free (unknown fields preserved).
- File: `design/adr/adr-008-provenance-and-citation-model.md`
- Decision: Sourced claims in an evidence context carry citations resolved by an
  offline resolver; unresolved/malformed/duplicate targets raise
  `core-citation-*` findings.

GitHub Issue:
- Title: Phase 3 acceptance corpus: criteria 6, 9, 10, 11, 12, 13
- Number: #39
- Milestone: Phase 3 — Authoring + staged mutation
- Labels: feature

Goal
Deliver a durable, criterion-traceable acceptance corpus for Phase 3 exit
criteria 6, 9, 10, 11, 12, 13, and capture the gate evidence, so the phase exit
gate can be closed out (by #40) against legible per-platform proof.

Why it matters
Phase 3's exit gate requires these six acceptance criteria to pass with named,
criterion-traceable evidence. The behaviors already exist and have scattered
unit/skill-flow coverage, but there is no named acceptance corpus that maps
one-to-one to those criteria and runs as a named, cache-defeating gate step per
matrix leg. This issue supplies both the corpus and the evidence artefact.

Requirements
- Add `cmd/llm-wiki/acceptance_phase3_test.go` (package `main`) holding six named
  tests with the stable `TestAcceptance` prefix
  (`TestAcceptanceCriterion{6,9,10,11,12,13}...`), each doc-commented with the
  criterion/ADR it proves and written as an end-to-end journey (init → plan →
  apply → verify), reusing only the existing in-package harness/helpers/fixtures
  (`exec`, `decodeEnvelope`, `initBundle`, `writeContentFile`, `snapshotAll`,
  `changedPaths`, `allUnder`, `hashFile`, and the `authoringDraft` const kept
  byte-aligned with `skills/wiki-authoring/SKILL.md`). Key deltas beyond the
  existing unit tests: criterion 6 asserts unknown fields survive into the
  committed live file (both edit and new-page directions); criterion 9 adds a
  negative control (missing in-bundle citation target ⇒ `core-citation-unresolved`,
  asserted on findings not exit code); criterion 13 proves zero mutation with a
  whole-tree snapshot around the rejected apply.
- In `.github/workflows/test.yml`, only rename the existing named acceptance step
  `Acceptance corpus (Phase 2 gate evidence)` → `Acceptance corpus (Phase 2+3 gate
  evidence)` and refresh its comment; the `^TestAcceptance` run already selects the
  new tests — do not change the command.
- Capture Phase 3 gate evidence in `notes/eval-issue-039.md` (criteria→test
  mapping, verification, honored non-goals, honest #29/#30 CI caveats, pointer to
  #40 for closeout).

Scope and constraints
- Primary folders to touch: `cmd/llm-wiki/` (test only), `.github/workflows/`,
  `prompts/`, `notes/`.
- No new engine behavior. If a new acceptance test exposes an engine gap, STOP and
  report it for the owning issue / a follow-up — do not fix engine behavior here.
- Do NOT edit `design/state.md` or `knowledge/` — issue #40 owns the closeout.
- Do NOT touch `skills/`, the adapter, or enrichment; no enrichment-half of
  criterion 6 (Phase 5), no academic-research profile (Phase 4), no criterion
  outside {6, 9, 10, 11, 12, 13}. No fixes for #29 or #30.

Evaluation & testing requirements
- `gofmt -l .` empty; `go vet ./...`, `go build ./...`, `go test ./...` all clean
  under `GOTOOLCHAIN=local` with the pinned Go 1.24.13 toolchain.
- `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` shows every criterion
  test PASS.
- All existing tests must continue to pass.
- The gate evidence records the CI run honestly, including the #29 (windows
  perm-mode) and #30 (macos-amd64 undispatched) caveats if still present.
