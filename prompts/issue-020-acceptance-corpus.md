You are working in my `llm-wiki-kit` repository.

Context:
- A versioned Claude Code plugin paired with a bundled, self-contained Go
  `llm-wiki` engine that lets a technical researcher or team create and
  maintain portable, repository-native knowledge bundles in the Open Knowledge
  Format (OKF v0.1) — structured, validated, and agent-readable without a
  hosted wiki.
- Follow the rules in `CLAUDE.md`.
- The workflow model is described in `docs/workflow-kit/workflow-guide.md`.
- This is the Phase 2 (Install/init) closeout-support issue. The phase scope is
  `design/build-out-plan.md` §"Phase 2". The install/init behaviors themselves
  already exist with scattered per-command unit tests in `cmd/llm-wiki/`; what
  is missing is a named, criterion-traceable acceptance corpus that runs as a
  legible per-platform gate in CI, plus the gate-evidence artefact.

ADR:
- File: `design/adr/adr-002-platform-binary-selection.md`
- Decision: Five supported target platforms (darwin/amd64, darwin/arm64,
  linux/amd64, linux/arm64, windows/amd64); the CLI must run on each.
- File: `design/adr/adr-003-json-contract-and-exit-codes.md`
- Decision: One versioned JSON envelope `{contractVersion, operation, status,
  findings, affectedPaths, approval}` shared by every surface; six semantic
  exit-code buckets.
- File: `design/adr/adr-005-safe-filesystem-layer.md`
- Decision: Engine-managed state lives under `.llm-wiki/`; boundary-relative,
  root-escape-refusing filesystem access.
- File: `design/adr/adr-006-staged-mutation-transaction-model.md`
- Decision: Mutations run through a staged transaction (`txn.Begin` …
  commit/rollback); a pre-flight conflict refusal precedes `txn.Begin`, so a
  refusal mutates nothing.
- File: `design/adr/adr-009-install-upgrade-uninstall-ownership.md`
- Decision: Install writes a version-record manifest at `.llm-wiki/manifest.json`
  recording schema/plugin/cli/okf/profile versions plus per-asset ownership
  class and content hashes; the manifest never lists itself.

GitHub Issue:
- Title: New-repo + non-empty-repo install/init acceptance corpus
- Number: #20
- Milestone: Phase 2 — Install/init
- Labels: infra

Goal
Deliver a durable, criterion-traceable acceptance corpus for the Phase 2
install/init lifecycle and capture the Phase 2 gate evidence, so the phase exit
gate can be closed out (by #21) against legible per-platform proof.

Why it matters
Phase 2's exit gate requires acceptance criteria 2 (CLI runs on every supported
platform), 3 (install into new and non-empty repos with no file loss), and
4-core (init with the core profile) to pass on all five platforms, plus
criterion 1's install half (`--dry-run` full plan, zero mutation; ADR-009
version-record manifest). The behaviors already exist and have scattered unit
coverage, but there is no named acceptance corpus that maps one-to-one to those
criteria and runs as a named, cache-defeating gate step per matrix leg. This
issue supplies both the corpus and the evidence artefact.

Requirements
- Add `cmd/llm-wiki/acceptance_test.go` (package `main`) holding a named
  acceptance corpus with a stable `TestAcceptance` prefix, reusing the existing
  in-process harness (`exec`, `decodeEnvelope`, `snapshotTree`, `sha256Hex`,
  `installTargets`, `initTargets`). Each test is doc-commented with the
  criterion/ADR it proves; the file header declares it the Phase 2 gate corpus.
- Add a named CI step to the `test` job in `.github/workflows/test.yml` that runs
  the acceptance corpus per platform (`go test ./cmd/llm-wiki -run
  '^TestAcceptance' -count=1 -v`), leaving `cross-compile-smoke` untouched.
- Capture Phase 2 gate evidence in `notes/eval-issue-020.md` (criteria→test
  mapping, verification, honored non-goals, pointer to #21 for closeout).

Acceptance criteria
- The corpus is runnable via `go test ./cmd/llm-wiki -run '^TestAcceptance'`,
  producing one named PASS line per criterion.
- The CI matrix is green on all five ADR-002 target platforms with the named
  acceptance step present, plus `cross-compile-smoke` green.
- Phase 2 gate evidence is captured for closeout in `notes/eval-issue-020.md`.

Scope and constraints
- Primary folders to touch: `cmd/llm-wiki/` (test only), `.github/workflows/`,
  `prompts/`, `notes/`.
- Folders to avoid unless absolutely necessary: `internal/`, `design/`.
- Do NOT edit `design/state.md` — issue #21 owns the state refresh.
- Do NOT implement release builds (#17), closeout (#21), upgrade/uninstall, or
  any Phase 3+ behavior. No authoring/enrichment fixtures; no inspect/plan/apply.
- Core profile only (academic-research is Phase 4; custom is Phase 7).

Evaluation & testing requirements
- `gofmt -l .` empty; `go vet ./...`, `go build ./...`, `go test ./...` all clean
  under `GOTOOLCHAIN=local` with the pinned Go 1.24.13 toolchain.
- `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` shows every
  criterion test PASS.
- All existing tests must continue to pass.
- The gate evidence cites the green CI run URL.

Instructions for you
1. Read the relevant docs and existing files:
   - `CLAUDE.md`
   - the ADRs listed above
   - `cmd/llm-wiki/cli_test.go`, `cmd/llm-wiki/install_test.go`
     (the in-process harness the corpus reuses)
   - `internal/manifest/manifest.go` (manifest decoding)
   - `.github/workflows/test.yml`
2. Propose a short, step-by-step implementation PLAN for this issue.
3. Wait for my approval of the plan before making any edits.
4. After I approve, implement the plan:
   - keep changes focused on this issue's scope,
   - commit incrementally with messages referencing the ADR and issue.
5. At the end, provide an evaluation summary:
   - what changed,
   - verification steps performed,
   - any follow-up work needed for later issues,
   - exact commands I should run to inspect the result myself.

Do not start editing files until I explicitly approve your plan.
