# Evaluation — Issue #39: Phase 3 acceptance corpus (criteria 6, 9, 10, 11, 12, 13)

**Branch:** `test/39-phase3-acceptance-corpus` (cut from synced `main`, `c3a8b9e`).
**PR:** [#49](https://github.com/olivermorgan2/llm-wiki-kit/pull/49)
**ADRs:** [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md)
(envelope + exit codes), [ADR-004](../design/adr/adr-004-validation-severity-model.md)
(severity model — the citation negative control is a warning),
[ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md)
(staged plan/apply, base-state binding, loss-free round-trip),
[ADR-008](../design/adr/adr-008-provenance-and-citation-model.md)
(offline citation resolver).
**Method:** durable, criterion-traceable acceptance corpus over the existing
in-process CLI harness, mirroring the Phase 2 corpus; the existing named
per-platform CI gate step now carries the Phase 3 evidence too.

Backs the **Phase 3 (Authoring + staged mutation)** exit gate
(`design/build-out-plan.md` §"Phase 3"), whose named exit criteria include
acceptance criteria **6, 9, 10, 11, 12, 13**.

## What this delivers

The Phase 3 behaviors already shipped in #34/#35/#36/#37/#38 with scattered
per-command and skill-flow tests. What was missing is a **named, criterion-mapped
acceptance corpus** carrying the stable `TestAcceptance` prefix so CI can
name-select it and print one legible PASS line per criterion per platform, plus
this gate-evidence artefact. #39 adds exactly that — **no product behavior
changes** (test-only issue; the sole non-test change is a cosmetic CI step
rename).

## What changed

- **`cmd/llm-wiki/acceptance_phase3_test.go`** *(new)* — six
  `TestAcceptanceCriterion...` tests, each doc-commented with the criterion/ADR
  it proves, each a full journey (init → plan → apply → verify). Reuses only
  existing in-package harness/helpers/fixtures (`exec`, `decodeEnvelope`,
  `initBundle`, `writeContentFile`, `snapshotAll`, `changedPaths`, `allUnder`,
  `hashFile`, and the `authoringDraft` const kept byte-aligned with
  `skills/wiki-authoring/SKILL.md`). Adds three small local page-content consts
  for the criterion-6 and criterion-13 seeds; no new helpers.
- **`.github/workflows/test.yml`** — the named acceptance step is renamed
  `Acceptance corpus (Phase 2 gate evidence)` → `Acceptance corpus (Phase 2+3
  gate evidence)` and its comment updated to note it now also carries the Phase 3
  authoring/staged-mutation criterion evidence. **No command change** — the
  existing `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` auto-selects
  the new tests.
- **`prompts/issue-039-phase3-acceptance-corpus.md`** *(new)* — the session prompt.
- **`notes/eval-issue-039.md`** *(new)* — this evidence artefact.

## Criteria → test mapping

| Criterion | Test | Journey | Proves |
|---|---|---|---|
| **6** — unknown fields survive authoring | `TestAcceptanceCriterion6UnknownFieldsRoundTripPlanApply` | seed page w/ `custom_field`,`x-tool-meta` → edit `plan`→`apply`; also new-page direction | both unknown fields and the edited known field are present in the **committed live file** (delta vs. the plan-layer unit test, which checks only the staged postimage); committed hash = `plan.stagedHash`, both directions (ADR-006). |
| **9** — sourced claims resolve offline | `TestAcceptanceCriterion9SourcedClaimCitationsResolveOffline` | `authoringDraft` (`LLM_WIKI_EVIDENCE_SECTIONS=Evidence`) → `inspect --content` → `plan`→`apply` → `validate` | zero `core-citation-*` findings on inspect-content and on the committed bundle (in-bundle + https citation classes both resolve, fully offline). **Negative control:** an in-bundle citation to a missing target (`wiki/nope.md`) raises `core-citation-unresolved` (warning — asserted on findings, not exit code), proving resolution is non-vacuous (ADR-008). |
| **10** — repeated input is a no-op | `TestAcceptanceCriterion10RepeatedUnchangedInputIsNoOp` | `plan`+`apply` once → re-`plan` identical draft | `plan.noOp`, empty `transaction`, empty `affectedPaths`, exactly one `wiki/photosynthesis*.md`, byte-identical whole tree (staging incl.) (ADR-006 fixed point). |
| **11** — edit previewed before apply | `TestAcceptanceCriterion11EditPreviewPrecedesApply` | commit base → `plan` an edit → decline → `apply` | non-empty `plan.diff` carrying the edited text; on the decline path the live hash is unchanged and only `.llm-wiki/` changed; on apply the committed hash = the previewed `plan.stagedHash` (ADR-006). |
| **12** — writes only via staged plan/apply | `TestAcceptanceCriterion12WritesOnlyThroughStagedPlanApply` | base snapshot → `inspect`+`plan` → `apply` | after inspect/plan only `.llm-wiki/` changed and the target is absent; after apply the sole non-staging change is exactly the target page; staging cleaned (ADR-005/006). |
| **13** — stale plan rejected, zero mutation | `TestAcceptanceCriterion13StalePlanRejectedZeroMutation` | seed → `plan` edit → out-of-band rewrite → `apply` | exit 4, `invalid-invocation`, six-field envelope, no `apply` payload; **whole-tree** snapshot around the rejected apply is byte-identical (stronger than the unit test's single-file check) (ADR-006 base binding). |

## Verification

Evidence is separated by category. **This PR does not by itself prove "all five
platforms green" or "full matrix green."** Read each category on its own terms.

### 1. Local validation

Go 1.24.13, `GOTOOLCHAIN=local`, `/Users/hermes/sdk/go1.24.13/bin/go`, on a Unix
(developer macOS/arm64) host:

```
gofmt -l cmd/llm-wiki/acceptance_phase3_test.go             # empty (formatted)
go build ./...                                              # clean
go vet ./...                                                # clean
go test ./...                                               # all packages pass on this host
go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v   # 12/12 PASS (6 Phase 2 + 6 Phase 3)
```

The six new Phase 3 tests each pass immediately against merged `main` — no engine
change was needed, consistent with the test-only scope. `go test ./...` passes
locally because the host is Unix; the file-permission-mode assertions that fail on
Windows (category 4) hold here. Local validation is single-host and is **not**
evidence about any other platform.

### 2. Named acceptance-step evidence (CI)

The renamed `Acceptance corpus (Phase 2+3 gate evidence)` CI step runs
`go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` before the full suite,
per matrix leg, and now name-selects both the Phase 2 corpus and the new Phase 3
corpus. The step targets only `./cmd/llm-wiki` and is ordered **before** the full
`go test ./...`, so it emits per-platform acceptance evidence on every leg even
where an unrelated package's suite is red (see category 4). _Per-platform PASS
lines for this PR's run are to be read from the PR's own CI run once it finalizes._

### 3. Known CI caveats carried forward (NOT introduced by #39)

- **#29 — windows-amd64 full `go test ./...` red.** The Windows full-suite job is
  red only in `internal/fsafe` and `internal/txn`, on POSIX file-permission-mode
  assertions (`mode = -rw-rw-rw-, want -rw-r-----`, etc.) that cannot hold on
  Windows. These predate #39, live in `internal/` (out of this issue's scope), and
  fail identically on `main`. Because the acceptance step targets only
  `./cmd/llm-wiki` and runs before the full suite, it still emits green Windows
  acceptance evidence even though that job ends red.
- **#30 — macos-amd64 (`macos-13`) runner may not dispatch.** In prior phases this
  leg stayed queued/never-ran in this environment, producing **no** acceptance and
  **no** full-suite evidence. If it remains undispatched on this PR's run,
  macos-amd64 has produced no observed evidence for #39 either; its code path is
  identical to macos-arm64 and `cross-compile-smoke` builds `darwin/amd64` cleanly,
  which is supporting **inference**, not observed per-platform CI.

Both caveats are for Hermes to weigh at the Phase 3 gate; neither is a #39
regression, and the milestone-level per-platform verdict is **#40's** job, not
this PR's.

## Non-goals honored

- **No new engine behavior.** All six new tests pass against merged `main`
  unchanged; no corpus-exposed engine gap was found, and none would have been
  fixed here (it would go to the owning issue / a follow-up).
- **No CI logic change** — only the step name/comment refresh; the command is
  unchanged.
- **`design/state.md` and `knowledge/` untouched** — the phase-status/exit-criteria
  refresh and milestone close are **#40 (Phase 3 closeout)**, which consumes this
  corpus and note as gate inputs.
- **No skill/adapter/enrichment changes**; no enrichment-half of criterion 6
  (Phase 5), no academic-research profile (Phase 4), no criterion outside
  {6, 9, 10, 11, 12, 13}.
- **No fixes for #29 or #30.**

## Pointer to closeout

Phase 3 closeout (milestone close + `design/state.md` / `knowledge/` refresh +
the milestone-level per-platform CI accounting) is **#40**. This corpus and
evidence note are the gate inputs; #40 consumes them.

## Commands to reproduce

```bash
export GOTOOLCHAIN=local
GO=/Users/hermes/sdk/go1.24.13/bin/go
cd /Users/hermes/llm-wiki-kit
"$GO" version   # go1.24.13
"$GO" build ./... && "$GO" vet ./... && "$GO" test ./...
"$GO" test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v
```
