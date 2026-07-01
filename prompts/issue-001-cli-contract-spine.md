You are working in my `llm-wiki-kit` repository.

Context:
- A versioned Claude Code plugin paired with a bundled, self-contained Go
  `llm-wiki` engine that lets a technical researcher or team create and
  maintain portable, repository-native knowledge bundles in the Open Knowledge
  Format (OKF v0.1) — structured, validated, and agent-readable without a
  hosted wiki.
- Follow the rules in `CLAUDE.md`.
- The workflow model is described in `docs/workflow-kit/workflow-guide.md`.
- This is the first Phase 1 implementation issue (Slice 0 — must-pass, blocks
  everything). The local pre-issue spec is `notes/phase-1-first-issue-spec.md`;
  the phase scope is `design/build-out-plan.md` §"Phase 1: Engine + contract
  spine". Reference them; do not restate them here.

ADR:
- File: `design/adr/adr-001-go-toolchain-and-yaml.md`
- Decision: Adopt Go 1.24.x with `github.com/goccy/go-yaml`; keep YAML behind an
  internal adapter interface; pin the exact Go patch in `go.mod` at this
  first-engine-issue time.
- File: `design/adr/adr-003-json-contract-and-exit-codes.md`
- Decision: One versioned JSON envelope `{contractVersion, operation, status,
  findings, affectedPaths, approval}` shared by every surface; `--json` is
  opt-in on every command (human-readable is the default); six semantic
  exit-code buckets (success / success-with-warnings / validation-failure /
  approval-required / invalid-invocation / system-or-filesystem-failure). The
  exact numeric exit-code values are deferred to this implementation issue.

GitHub Issue:
- Title: Scaffold deterministic llm-wiki CLI skeleton + versioned JSON-contract spine (ADR-001, ADR-003)
- Number: #1
- Milestone: Foundation
- Labels: feature

Goal
Stand up a `llm-wiki` binary on Go 1.24.x that exposes the `internal/` engine
package seams and emits a schema-valid JSON envelope under `--json`
(human-readable by default) — proving the versioned contract is callable
without Claude Code.

Why it matters
Phase 1 (Slice 0) is the deterministic engine every other surface calls. The
CLI skeleton + contract shell is the authority that `validate`, binary
selection, and the filesystem gate later plug into. This issue advances
acceptance criterion 14 (structured output conforms to the contract +
documented exit codes) and lays the spine that criteria 5, 7, 8, and 17 build
on in later Phase 1 issues. It rides on the ADR-001 Go/YAML assumption-lock
(Q2), which is reversible and confined to the internal YAML adapter.

Requirements
- `go mod init` on Go 1.24.x; pin the exact Go patch in `go.mod`; add the
  `cmd/llm-wiki` entry point.
- Create the `internal/` package seams per the build-out-plan repository
  layout: `contract/`, `validate/`, `profile/`, `fsafe/`. Stubs / interfaces
  are acceptable where the full behaviour is a later Phase 1 issue; the YAML
  access in `profile`/`validate` seams must go through an internal YAML adapter
  interface (ADR-001), not `goccy/go-yaml` directly at call sites.
- `internal/contract`: the versioned JSON envelope type + serializer carrying
  `{contractVersion, operation, status, findings, affectedPaths, approval}`,
  and the six exit-code semantic buckets as named constants.
- Honor a `--json` flag (or `LLM_WIKI_JSON` toggle) on every command; default
  to human-readable output.
- Provide a trivial command surface (e.g. `llm-wiki version` plus a no-op
  `validate` shell) that emits a schema-valid envelope under `--json`.
- Publish the ADR-003 numeric exit-code code→meaning table as a documented,
  stable surface (this issue owns the deferred numeric values; once published
  they are frozen). A short contract/exit-code doc is acceptable.

Acceptance criteria
- `go build ./...` succeeds on Go 1.24.x with the patch pinned in `go.mod`.
- The trivial command with `--json` emits an envelope carrying all six contract
  fields; the same command without `--json` prints human-readable text.
- Each of the six exit-code buckets is a named constant with a test asserting
  its PRD §12 outcome mapping, and the documented code→meaning table matches
  those constants.
- No product validation rules, filesystem guards, or binary selection are
  implemented (those are separate Phase 1 issues #2–#6).

Scope and constraints
- Primary folders to touch: `cmd/llm-wiki/`, `internal/contract/`,
  `internal/validate/`, `internal/profile/`, `internal/fsafe/`, and `go.mod`.
- Folders to avoid unless absolutely necessary: `profiles/`, `testdata/`,
  `bin/`, `.github/`, `design/`, `knowledge/`, `.claude/`.
- Do NOT implement: real validation/broken-link rules (issues #3/#4), the
  filesystem-safety guards (issue #5), OS/arch detection + checksum (issue #2),
  any install / init / upgrade / uninstall / asset-ownership work (ADR-009), or
  any staged cross-file mutation / `inspect`-`plan`-`apply` / hash-binding
  (ADR-006, which must be drafted before any cross-file mutation).

Evaluation & testing requirements
- Unit tests: envelope round-trips its JSON schema; each exit-code bucket maps
  to its intended PRD §12 outcome; `--json` vs default output selection.
- `go build ./...` and `go test ./...` pass on Go 1.24.x.
- All existing tests must continue to pass.
- If a change cannot be unit tested, document the manual verification.

Instructions for you
1. Read the relevant docs and existing files:
   - `CLAUDE.md`
   - `design/adr/adr-001-go-toolchain-and-yaml.md`
   - `design/adr/adr-003-json-contract-and-exit-codes.md`
   - `notes/phase-1-first-issue-spec.md` and `design/build-out-plan.md`
     §"Phase 1"
   - any existing modules under `cmd/` and `internal/` (none exist yet — this
     is a greenfield scaffold).
2. Propose a short, step-by-step implementation PLAN for this issue, including:
   - new files or modules to create,
   - existing files to modify,
   - key functions or structures (the envelope type, exit-code constants),
   - your verification or test plan.
3. Wait for my approval of the plan before making any edits.
4. After I approve, implement the plan:
   - keep changes focused on this issue's scope,
   - commit incrementally with messages referencing the ADR and issue
     (e.g. "feat(contract): add JSON envelope (ADR-003, #1)").
5. At the end, provide an evaluation summary:
   - what changed,
   - verification steps performed,
   - any follow-up work needed for later issues,
   - exact commands I should run to inspect the result myself.

Do not start editing files until I explicitly approve your plan.
