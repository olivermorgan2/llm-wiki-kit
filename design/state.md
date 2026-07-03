# Project State

_Last updated: 2026-07-02_

## Current status

- **Phase 1 / Foundation: complete.**
- **Next phase: Phase 2 / Install-init.**

## Phase 1 / Foundation — complete

Foundation delivered the deterministic `llm-wiki` CLI spine: versioned JSON
contract, platform/artifact verification, core-profile validation with a
three-severity model, broken-link detection, a mandatory safe-filesystem
gate, and the core-profile + security acceptance fixtures.

### Issues (all closed)

| Issue | Title | ADR |
|-------|-------|-----|
| #1 | Scaffold deterministic `llm-wiki` CLI skeleton + versioned JSON-contract spine | ADR-001, ADR-003 |
| #2 | OS/arch detection + release-artifact checksum verification | ADR-002 |
| #3 | Core-profile validate: OKF-vs-profile reporting at three severities | ADR-004 |
| #4 | Broken-link detection at configured severity | ADR-004 |
| #5 | Safe filesystem layer: atomic writes, bounded scope, symlink/path-traversal rejection | ADR-005 |
| #6 | Core-profile fixtures + security testdata: traversal/symlink | ADR-004, ADR-005 |

### PRs (all merged)

| PR | Merged | Closes |
|----|--------|--------|
| #7 | 2026-07-01 | #1 |
| #8 | 2026-07-01 | #2 |
| #9 | 2026-07-01 | #3 |
| #10 | 2026-07-01 | #4 |
| #11 | 2026-07-01 | #5 |
| #12 | 2026-07-01 | #6 |

### ADRs adopted

- ADR-001 — Go toolchain and YAML
- ADR-002 — Platform binary selection
- ADR-003 — JSON contract and exit codes
- ADR-004 — Validation and severity model
- ADR-005 — Safe filesystem layer

## Validation evidence (2026-07-02, repo at `main` = `2b8f988`)

Toolchain: `go1.24.13 darwin/arm64` (`GOTOOLCHAIN=local`).

- `go test ./...` — **PASS** (all 8 packages ok):
  `cmd/gen-checksums`, `cmd/llm-wiki`, `internal/contract`,
  `internal/fsafe`, `internal/platform`, `internal/profile`,
  `internal/validate`, `internal/yamladapter`.
- `go build ./...` — **PASS** (exit 0).
- `go vet ./...` — **PASS** (exit 0).
- `go run ./cmd/llm-wiki validate --json testdata/fixtures/core-valid` —
  **PASS**: `status: success`, empty `findings`, contract `v1`.

## Next phase — Phase 2 / Install-init

Foundation is closed out; work moves to the install/init surface.

## Notes

- **CI is deferred to Phase 2.** There is no CI workflow yet; Phase 1
  validation is the local evidence recorded above.
