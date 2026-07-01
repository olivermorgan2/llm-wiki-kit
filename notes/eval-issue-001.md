# Evaluation — Issue #1: llm-wiki CLI skeleton + versioned JSON-contract spine

**Branch:** `issue-001-cli-contract-spine` (cut from `docs/adr-001-005-engine-foundations`,
where the accepted ADR-001/003/004/005 live — they are not yet on `main`).
**ADRs:** ADR-001 (Go toolchain + YAML seam), ADR-003 (JSON contract + exit
codes), with the finding shape from ADR-004 and the fsafe seam from ADR-005.
**Method:** strict TDD (red → green per unit), incremental commits.

## What changed

**Toolchain / module**
- `go.mod` — new module `github.com/olivermorgan2/llm-wiki-kit`, Go patch pinned
  `go 1.24.13` (ADR-001). No third-party dependency added: `goccy/go-yaml` is
  deferred (see Follow-ups) because the skeleton parses no YAML yet.

**Contract spine (`internal/contract/`)**
- `envelope.go` — the v1 `Envelope` carrying exactly the six ADR-003 fields
  (`contractVersion, operation, status, findings, affectedPaths, approval`);
  `Finding`/`Severity`/`Ruleset` shape (ADR-004); `Approval`; `New()`
  (empty slices, not null); `WriteJSON` (indented, deterministic, trailing NL).
- `exitcode.go` — six semantic buckets as frozen named constants
  (`0 success … 5 system-or-filesystem-failure`), `ExitCodeForStatus`, and the
  canonical `ExitCodes` table.
- `docs/contract/exit-codes.md` — the published, frozen code→meaning table and
  envelope reference; a test asserts it matches the constants.

**Engine seams (`internal/`)**
- `yamladapter/yamladapter.go` — the ADR-001 `Adapter` interface confining YAML
  behind one seam (concrete goccy impl deferred).
- `profile/profile.go` — `Loader` that routes all YAML through the injected
  adapter (white-box test enforces the wiring).
- `validate/validate.go` — the single-pass `Engine`; `Run` returns an empty
  (non-nil) `[]contract.Finding` — no rules yet (ADR-004).
- `fsafe/fsafe.go` — the ADR-005 `Gate` chokepoint interface (guards deferred).

**CLI (`cmd/llm-wiki/`)**
- `main.go` — thin `main` → `run` → `os.Exit`.
- `cli.go` — `run(args, stdout, stderr) int`; global `--json` flag (any
  position) and `LLM_WIKI_JSON` toggle; `version` and no-op `validate`
  (wired to `internal/validate`); unknown command → `invalid-invocation`.

**Housekeeping**
- `.gitignore` — ignore Go build artifacts (root binary, `*.test`, `*.out`).

## Commits (in order)

- `85f5a85` feat(contract): add versioned JSON envelope and exit-code buckets (ADR-003, ADR-004, #1)
- `86e8cf1` feat(engine): add internal validate/profile/fsafe seams and YAML adapter interface (ADR-001, ADR-004, ADR-005, #1)
- `950673a` feat(cli): add llm-wiki CLI skeleton with --json envelope output (ADR-003, #1)
- `78620fb` chore: ignore Go build artifacts (#1)

## Verification performed

- `go build ./...` — OK (Go 1.24.13).
- `go vet ./...` — OK. `gofmt` — clean.
- `go test ./...` — all packages pass (27 tests: 9 contract, 1 profile,
  1 validate, 9 CLI, + subtests; `fsafe`/`yamladapter` are interface-only,
  verified by compilation).
- Manual smoke of the built binary: `version` (human + `--json`),
  `validate --json`, `LLM_WIKI_JSON=1 validate`, unknown command (human exit 4
  + `--json` invalid-invocation envelope). Output matches the documented
  envelope example; exit codes 0 and 4 confirmed.

## Acceptance criteria (issue #1)

- [x] `go build ./...` succeeds on Go 1.24.x with the patch pinned in `go.mod`.
- [x] Trivial command with `--json` emits an envelope with all six fields;
      without `--json` prints human-readable text.
- [x] Each of the six exit-code buckets is a named constant with a test
      asserting its PRD §12 outcome mapping; the documented table matches.
- [x] No validation rules, filesystem guards, or binary selection implemented.

## Follow-ups (out of scope for #1)

- **`goccy/go-yaml` concrete adapter + dependency** — deferred to the validation
  issue (#3/#4). The `yamladapter.Adapter` seam is in place; the node-aware
  implementation that preserves unknown fields on round-trip (criterion 6) lands
  when real YAML parsing does, along with pinning goccy in `go.mod`.
- **fsafe guards + traversal/symlink fixtures** — later Phase 1 issue (#5, ADR-005).
- **OS/arch detection + checksum verification** — separate Phase 1 issue (#2, ADR-002).
- **ADR-006 (staged cross-file mutation)** must be drafted before any
  cross-file mutation work; this skeleton implements none.
- **Toolchain provisioning:** Go was not installed on this host; it was
  installed self-contained at `~/sdk/go1.24.13`. CI/other machines need a Go
  1.24.13 toolchain (ADR-002 cross-platform matrix will formalize this).

## Commands to reproduce

```bash
export PATH="$HOME/sdk/go1.24.13/bin:$PATH"   # or your own Go 1.24.13
cd /Users/hermes/llm-wiki-kit
go build ./... && go vet ./... && go test ./...
go run ./cmd/llm-wiki version
go run ./cmd/llm-wiki version --json
go run ./cmd/llm-wiki validate --json
LLM_WIKI_JSON=1 go run ./cmd/llm-wiki validate
go run ./cmd/llm-wiki frobnicate; echo "exit=$?"
```

## Next step

`/pr-review-packager` to draft a PR from this branch, referencing ADR-001 and
ADR-003. (Note the branch base above: it sits on the ADR docs branch, not
`main`.)
