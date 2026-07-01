# Evaluation — Issue #4: Broken-link detection at configured severity

**Branch:** `issue-004-broken-link-detection` (cut from synced `main`).
**ADR:** [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) (one
engine; ruleset-tagged, three-severity, profile-configurable findings; fixed
`core defaults → profile overrides → baseline suppression` precedence). Envelope +
exit codes [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md).
**Method:** strict TDD (red → green per unit), incremental commits.

Backs **Phase 1 (Foundation)** exit gate / **acceptance criterion 8**:
`llm-wiki validate` detects broken **intra-wiki** links and reports each at the
severity the active profile configures, through the ADR-003 envelope and exit
codes. Default severity is **warning** (ADR-004 FR8); a profile may promote it to
**error**, flipping the run to the validation-failure exit code.

## Confirmed decisions (carried from the plan)

1. **Bundle-root-relative resolution.** A link target is interpreted relative to
   the bundle root regardless of the linking page's location, matching the PRD's
   canonical-output form and the wiki mental model.
2. **Tag `Ruleset=profile`, `Code=core-broken-link`, default `Severity=warning`.**
   Intra-wiki bundle-relative links are a kit/core-profile concept (absent from
   the external OKF SPEC); broken links do not invalidate base OKF conformance.

## What changed

**Validation engine (`internal/validate/`)**
- `links.go` *(new)* — `linkRules(pagePath, body, exists)`: extracts inline
  `[text](target)` links from the markdown **body**, classifies each target, and
  emits **one aggregated** `core-broken-link` warning per page naming every broken
  target. Helpers `isIntraWiki` (skip URI-scheme / protocol-relative / pure
  `#fragment` / empty targets) and `resolveTarget` (strip `#fragment`/`?query`,
  trim a leading `/`, `path.Clean`, treat `../`-escaping as unresolved). Image
  links `![alt](t)` are skipped via an explicit captured `!` group (RE2 has no
  lookbehind).
- `rules.go` — added `codeCoreBrokenLink = "core-broken-link"` to the rule-code
  block. No other change.
- `validate.go` — `Run`'s single walk now records **every** bundle file path (not
  just `.md`) so link targets resolve against the full bundle, and evaluates pages
  **after** the walk (links may point forward) preserving deterministic lexical
  order. Each page runs `evaluatePage` (unchanged) plus `linkRules` over its split
  body; a page whose frontmatter fails to split already yields `okf-yaml-parse` and
  contributes no link findings.

**Tests**
- `links_test.go` *(new)* — extraction, broken vs valid, external/mailto/fragment
  skip, image skip, `#fragment`/`?query` stripping, root-escape, bundle-root
  resolution, tag/severity/code, per-page aggregation.
- `validate_test.go` — engine-integration over `fstest.MapFS` (broken + all-valid).
- `severity_test.go` — promotion precedence: default warning →
  success-with-warnings; `Resolve({core-broken-link: error})` → validation-failure.

## How it reuses existing machinery (no new engine/contract code)

- **Default severity:** the rule stamps `warning` directly (ADR-004 FR8), like
  `core-kebab-filename`.
- **Profile override / promotion:** handled by existing `Resolve(findings,
  overrides)` — `{core-broken-link: error}` re-stamps to error. No new code.
- **Status → exit:** existing `StatusFor` yields success-with-warnings (exit 1) at
  warning, validation-failure (exit 2) once promoted.
- **Baseline:** `ApplyBaseline` already suppresses by `{ruleset, code, path}`; the
  per-page aggregation keeps that fingerprint unique.
- **Envelope:** emitted through the ADR-003 envelope unchanged; the `profile`
  ruleset tag routes it into the profile bucket, distinct from `okf`.

No changes to `contract/`, `baseline.go`, `severity.go` signatures, profile
loading, or `cmd/llm-wiki/cli.go` behavior.

## Verification performed

- `go build ./...` OK; `go vet ./...` OK; `gofmt -l .` empty (Go 1.24.13).
- `go test ./...` — all packages pass (adds 12 test funcs in `internal/validate`).
- Compiled-binary smoke on a synthetic wiki:
  - page linking `claims/missing.md` (absent) + `claims/real.md` (present) →
    single `profile`/`warning`/`core-broken-link` finding on the linking page →
    `success-with-warnings`, **exit 1**;
  - remove the broken link → `success`, **exit 0**.

## Deviation from the plan (declared)

The plan's prose said to thread an `exists` predicate **into `evaluatePage`** and
call `linkRules` there. Doing so would change `evaluatePage`'s signature and break
the ~15 existing call sites in `rules_okf_test.go` / `rules_profile_test.go` — two
files the plan's own **in-scope file table does not list**. To honor the declared
file surface and keep the PR tight, `linkRules(path, body, exists)` (exact plan
signature) is instead wired from **`Run`**, which is the natural owner of the
bundle-wide path-set. Behavior, findings, severity/override/baseline reuse, and
every acceptance criterion are identical; `evaluatePage` and the existing per-page
unit tests are untouched. Documented here rather than silently substituted.

## Acceptance criteria (issue #4)

- [x] **AC #1 — tag/severity/code.** `Ruleset=profile`, `Severity=warning`,
      `Code=core-broken-link` (`TestLinkRulesTagSeverityCode`, CLI smoke).
- [x] **AC #2 — configured severity / promotion.** Default warning → exit 1;
      profile override to error → validation-failure / exit 2
      (`TestBrokenLinkPromotionFlipsStatus`).
- [x] **AC #3 — valid links pass.** Resolvable targets raise nothing
      (`TestLinkRulesValidTargetNoFinding`, `TestRunAllLinksValidNoBrokenLinkFinding`).
- [x] Criterion 8 — broken intra-wiki links detected and reported through the
      ADR-003 envelope/exit codes (engine-integration test + CLI smoke).

## Non-goals honored

- **No network / external liveness** — external schemes are skipped, never fetched.
- **No index / link-graph** (ADR-010) — no cross-page graph, backlinks, or
  ordering beyond the existing walk.
- **No filesystem-safety enforcement** (ADR-005 / #5) — root-escaping `../` targets
  merely count as unresolved; the filesystem is never touched for them.
- **No example/fixture tree** (#6) — in-memory `fstest.MapFS` / inline bytes only.
- **No new contract surface / broad profile-loader** — reused `Finding`,
  `Ruleset`, `Severity`, `Resolve`, `StatusFor`, `ApplyBaseline`, exit mapping.
  The runtime override is a minimal configuration seam, not a profile-file loader
  (ADR-007 remains deferred): `cmd/llm-wiki/cli.go` reads `LLM_WIKI_SEVERITY`
  (comma-separated `code=severity` pairs) into the `Resolve` override map, mirroring
  the existing `LLM_WIKI_JSON` env toggle.

## Codex review follow-up — end-to-end configured-severity path

The initial cut demonstrated broken-link promotion only at the engine seam
(`Resolve(findings, {core-broken-link: error})` in a unit test); the CLI still
called `Resolve(findings, nil)`, so no runtime path proved a configured override
flips the emitted envelope/exit code. Fixed by wiring `LLM_WIKI_SEVERITY` into
`runValidate` and adding CLI-level tests:
`TestValidateBrokenLinkDefaultIsWarning` (default → warning / exit 1),
`TestValidateBrokenLinkPromotedToErrorViaConfig` (`core-broken-link=error` →
validation-failure / exit 2, severity `error` on the envelope finding), plus
`severityOverrides` parser tests. Scope unchanged: no profile-file loader, no
network/liveness, no link graph, no fs-safety, no fixture tree.

## Follow-ups (out of scope for #4)

- **Reference-style links** `[text][ref]`, **autolinks** `<…>`, **raw HTML**
  `<a href>`, and inline-link **titles** `[t](u "title")` are not extracted yet.
- **Anchor/fragment validation** — `#fragment` is stripped, not checked.
- **Profile-file loading** — feeding overrides from a real on-disk profile file
  belongs with the profile loader (ADR-007), not this issue. The `LLM_WIKI_SEVERITY`
  env seam proves the configured-severity contract without that loader.
- `design/architecture.md` still does not exist; for the next `/workflow-docs` run
  note the new `internal/validate` intra-wiki broken-link rule.

## Commands to reproduce

```bash
export PATH="$HOME/sdk/go1.24.13/bin:$PATH"   # or your own Go 1.24.13
cd /Users/hermes/llm-wiki-kit
gofmt -l . && go vet ./... && go build ./... && go test ./...

# compiled-binary smoke (honest exit codes; go run wraps non-zero as 1)
BIN="$(mktemp -d)/llm-wiki"; go build -o "$BIN" ./cmd/llm-wiki
W="$(mktemp -d)"; mkdir -p "$W/concepts" "$W/claims"
printf -- '---\ntype: concept\ntitle: Alpha\ndescription: A page.\ntimestamp: t\ntags: [x]\naliases: [y]\nresource: r\n---\n# Alpha\nSee [real](claims/real.md) and [gone](claims/missing.md).\n' > "$W/concepts/alpha.md"
printf -- '---\ntype: concept\ntitle: Real\ndescription: A page.\ntimestamp: t\ntags: [x]\naliases: [y]\nresource: r\n---\n# Real\n' > "$W/claims/real.md"
"$BIN" validate "$W" --json; echo "exit=$?"   # core-broken-link warning, exit 1
```

## Next step

`/pr-review-packager` (or `gh pr create`) to open a PR from this branch
referencing ADR-004 (plus ADR-003) and `Closes #4`, then a Codex review loop
before merge.
