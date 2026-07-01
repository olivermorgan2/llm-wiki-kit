# Evaluation — Issue #3: Core-profile `validate` (OKF vs profile, three severities)

**Branch:** `issue-003-core-profile-validate` (cut from synced `main`).
**ADR:** [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) (one
engine; ruleset-tagged, three-severity, profile-configurable findings; fixed
`core defaults → profile overrides → baseline suppression` precedence; baseline
hard boundary). YAML toolchain [ADR-001](../design/adr/adr-001-go-toolchain-and-yaml.md);
envelope + exit codes [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md).
**Method:** strict TDD (red → green per unit), one commit per plan step.

## What changed

**YAML adapter (`internal/yamladapter/`)**
- `goccy.go` — concrete `Adapter` backed by `github.com/goccy/go-yaml`
  (pinned `v1.8.10`, ADR-001 exact-patch pin). Real `Unmarshal` (ignores
  unknown fields); `Marshal` is a documented not-implemented stub (round-trip
  preservation is criterion 6 / Slice 2). goccy is imported **only** here.

**Validation engine (`internal/validate/`)**
- `frontmatter.go` — `splitFrontmatter`: splits the leading `---`-fenced YAML
  block from the body; missing/unterminated blocks are structural failures.
- `rules.go` — OKF ruleset (`okf-yaml-parse` error/never-suppressible;
  `okf-type-present` error) and core-profile ruleset (`core-required-title`,
  `core-required-description`, `core-field-type` error; `core-recommended-missing`
  suggestion; `core-kebab-filename` warning), each tagged with its ruleset. Parse
  failure short-circuits a page to a single `okf-yaml-parse` finding. Per-field
  rules aggregate to one finding each for a unique `{ruleset,code,path}`
  fingerprint; required-string and field-type stay non-overlapping.
- `severity.go` — `Resolve` (profile-override layer; nil = identity for
  core-only) and `StatusFor` (findings → status).
- `baseline.go` — `Baseline`, `Fingerprint` (`{ruleset,code,path}`), and
  `ApplyBaseline` (differential filter + hard boundary).
- `validate.go` — real `Engine`: `New(yamladapter.Adapter)` and
  `Run(fs.FS)` walking `*.md` deterministically (lexical order, recurses
  subdirectories) and returning core-default findings.

**CLI (`cmd/llm-wiki/`)**
- `cli.go` — `runValidate` wired to the engine: optional target dir (default
  `.`), `Resolve(core-only)`, no adoption baseline (release-gate semantics),
  envelope findings/status, mapped exit code; usage text refreshed.

**Docs**
- `docs/contract/validation.md` — rule tables (OKF vs profile), severity
  defaults, precedence, baseline boundary, status/exit mapping.

## Commits (in order)

- `bfa89fa` feat(yamladapter): goccy-backed Unmarshal adapter (ADR-001, #3)
- `ef120eb` feat(validate): frontmatter split + OKF base rules (ADR-004, #3)
- `252132a` feat(validate): core-profile field rules (ADR-004, #3)
- `c3f5928` feat(validate): severity precedence resolution (ADR-004, #3)
- `3d0840d` feat(validate): baseline differential filter + hard boundary (ADR-004, #3)
- `e49d25b` feat(validate): status + exit-code mapping (ADR-003/004, #3)
- `ba77326` feat(cli): wire validate to engine — findings, status, exit code (#3)
- `11ecdf2` docs(contract): document validate rulesets, severities, baseline boundary (#3)

## Verification performed

- `gofmt -l internal cmd` clean; `go vet ./...` OK; `go build ./...` OK
  (Go 1.24.13).
- `go test ./...` — all packages pass. Issue #3 adds ~40 test runs in
  `internal/validate`, 6 in `internal/yamladapter`, and updates the CLI validate
  tests to controlled temp fixtures.
- Compiled-binary smoke on synthetic wikis:
  - clean / suggestion-only page → `success`, **exit 0**;
  - warning-only page (non-kebab filename, otherwise complete) →
    `success-with-warnings`, **exit 1**;
  - malformed YAML (`title: {broken`) → `validation-failure`, **exit 2**;
  - a page missing both `type` and `title` emits an `okf`-tagged
    `okf-type-present` **and** a `profile`-tagged `core-required-title` in one
    run (criterion 5).

## Acceptance criteria

**Criterion 5 — OKF vs profile reported separately.**
- [x] Every finding carries a `ruleset` tag (`okf`/`profile`); OKF rules
      (`okf-*`) and profile rules (`core-*`) are non-overlapping and render
      distinctly. Proven by `TestOKFAndProfileFindingsTaggedDistinctly` and the
      mixed-page smoke.

**Criterion 7 — malformed YAML / missing required field fails validation.**
- [x] Malformed / missing / unterminated frontmatter → `okf-yaml-parse` error →
      `validation-failure`/exit 2, and is **never** baseline-suppressible.
- [x] Missing required fields (`type`, `title`, `description`) → error; in the
      default (release-gate) CLI path they always fail the gate.

## Deviation from the plan (validation-revealed)

The plan's smoke used `title: [broken` to trigger `okf-yaml-parse`. goccy
`v1.8.10` parses an unterminated flow **sequence** (`[broken`) leniently as the
plain scalar string `"[broken"`, so that input does **not** parse-fail. An
unterminated flow **mapping** (`title: {broken`) is unambiguously malformed and
goccy rejects it, so tests and the smoke use `{broken`. This preserves the
plan's intent (malformed YAML → `okf-yaml-parse` → exit 2) faithfully; no ADR or
contract change. Documented here rather than silently substituted.

## Non-goals honored

- No broken-link detection (#4); no link scanning.
- No safe-filesystem layer / `fsafe` (#5); validation only **reads** via
  `fs.FS`/`os.DirFS`. `internal/fsafe/fsafe.go` untouched.
- No shipped `profiles/` fixtures (#6); tests use `fstest.MapFS` / temp dirs.
- No academic/custom profiles, per-type page rules, required-section or
  `source`/`claim` provenance checks; unknown `type` values accepted.
- No profile-file loading (ADR-007); `internal/profile` stays a skeleton;
  overrides are injected data in tests, not loaded from disk.
- No `--baseline` flag or on-disk baseline format; the baseline **primitive +
  boundary** are built and unit-tested only. The default CLI path runs
  release-gate semantics (no baseline) — this bounded scope is stated, not
  silently narrowed.
- `Marshal` / unknown-field round-trip preservation (criterion 6) deferred to
  Slice 2; `Marshal` is a documented not-implemented stub.
- Reserved `index.md`/`log.md` semantics not special-cased; all `.md` treated as
  ordinary OKF pages.
- No new contract surface: reused `contract.Finding`/`Ruleset`/`Severity`/
  `Status`, `contract.New`, `contract.ExitCodeForStatus`, and the existing
  `--json`/`LLM_WIKI_JSON`/`emit` handling.

## Follow-ups (out of scope for #3)

- **Broken-link detection** at configured severity → issue #4 (ADR-004,
  criterion 8).
- **Baseline CLI surface** — `--baseline` flag, on-disk baseline format,
  adoption-mode UX (the primitive lands here; the surface is later work).
- **Profile-file loading** and one-level inheritance resolution → ADR-007 /
  Slice 3.
- `Marshal` / round-trip preservation → Slice 2 (criterion 6).
- `design/architecture.md` still does not exist; for the next `/workflow-docs`
  run note the now-real `internal/validate` engine and the concrete
  `internal/yamladapter` goccy implementation (new goccy dependency).

## Commands to reproduce

```bash
export PATH="$HOME/sdk/go1.24.13/bin:$PATH"   # or your own Go 1.24.13
cd /Users/hermes/llm-wiki-kit
gofmt -l internal cmd && go vet ./... && go build ./... && go test ./...

# compiled-binary smoke (honest exit codes; go run wraps non-zero as 1)
BIN="$(mktemp -d)/llm-wiki"; go build -o "$BIN" ./cmd/llm-wiki
W="$(mktemp -d)"
printf -- '---\ntype: concept\ntitle: Alpha\ndescription: A page.\ntimestamp: t\ntags: [x]\naliases: [y]\nresource: r\n---\n# Alpha\n' > "$W/alpha.md"
"$BIN" validate "$W"; echo "exit=$?"                       # success, exit 0
printf -- '---\ntype: concept\ntitle: {broken\n---\n' > "$W/bad.md"
"$BIN" validate "$W" --json; echo "exit=$?"                # validation-failure, exit 2
```

## Next step

`/pr-review-packager` (or `gh pr create`) to open a PR from this branch
referencing ADR-004 (plus ADR-001/003) and `Closes #3`, then a Codex review
loop before merge.
