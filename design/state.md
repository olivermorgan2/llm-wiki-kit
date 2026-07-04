# Project State

_Last updated: 2026-07-04_

## Current status

- **Phase 3 / Authoring + staged mutation: complete.** The `page` command
  group shipped end-to-end on the ADR-006 transaction substrate: `page inspect`
  read-only report (#42), `page plan` hash-bound staged whole-page preview
  (#34), `page apply` journaled commit with stale-plan rejection (#35), the
  ADR-008 citation mechanism — evidence contexts, offline three-class resolver,
  `core-citation-*` findings (#36) — the plan-time citation-loss approval gate
  (#37), the authoring skill adapter + `page inspect --content` overlay
  validation (#38), and the named, criterion-mapped Phase 3 acceptance corpus
  (#39). This closeout (#40) refreshes state + knowledge and records the gate
  evidence. `main` = `05b1b9a`. **Milestone #3 closes after this PR merges**
  (0 open issues once #40 lands) — not closed in this PR.
- **Current issue: #40 — Phase 3 closeout: state and knowledge update.**
  Docs-only. Brings `design/state.md` + `knowledge/` to final Phase 3 truth
  with honest CI caveats (#29/#30 carried forward). No engine, test, or CI
  changes.
- **Next phase: Phase 4 / Academic-research profile.** Next action: file the
  Phase 4 backlog per `design/build-out-plan.md` §Phase 4; ADR-007
  (profile-system boundary) is already accepted and covers it, and ADR-008's
  profile citation vocabulary + two Codex carry-ins land there.
- **Phase 2 / Install-init: complete; milestone #2 closed** (closeout PR #31,
  milestone closed manually 2026-07-03). **Phase 1 / Foundation: complete**
  (issues #1–#6, PRs #7–#13, ADR-001–005 accepted). Oliver's async ratification
  of ADR-006/007/008/009 remains the one outstanding flag — not a merge
  blocker.

## Phase 3 — Authoring + staged mutation — complete

Goal: the thinnest end-to-end "create real value" path — author a new page
through the staged, preview-before-write workflow (`page inspect` / `page plan`
/ `page apply` with hash-bound, stale-rejecting plans, plus a thin authoring
skill adapter), with sourced-claim citations resolved offline and unknown
frontmatter fields preserved across the round trip.

### Issues + PRs (all merged; closeout #40 via this PR)

| Issue | Title (abbrev.) | ADR | PR | Merge commit | Merged |
|---|---|---|---|---|---|
| #32 | ADR-008 provenance & citation model (accepted) | 008 | #33 | `1d6ca79` | 2026-07-03 |
| #42 | `page inspect` + `page` subcommand dispatch | 006/008 | #43 | `02fa777` | 2026-07-03 |
| #34 | `page plan` — staged plan, hash-bound manifest, diff, no-op | 006 | #44 | `84a08e7` | 2026-07-04 |
| #35 | `page apply` — journaled commit + stale-plan rejection | 003/006 | #45 | `04c358b` | 2026-07-04 |
| #36 | citation mechanism — evidence contexts, offline resolver, findings | 008 | #46 | `7a5f17a` | 2026-07-04 |
| #37 | citation-loss approval gate for page plans | 003/008 | #47 | `68abc1a` | 2026-07-04 |
| #38 | authoring skill adapter + `page inspect --content` | 005/006/008 | #48 | `c3a8b9e` | 2026-07-04 |
| #39 | Phase 3 acceptance corpus (criteria 6, 9–13) | 006/008 | #49 | `05b1b9a` | 2026-07-04 |
| #40 | Phase 3 closeout — state + knowledge | — | (this PR) | — | — |

Issue **#41** is a **closed duplicate of #40** (`state=CLOSED`,
`stateReason=DUPLICATE`, no PR) — recorded here for the ledger; not reopened,
not otherwise touched.

### ADR adopted this phase

- **ADR-008 — Provenance & citation model** (accepted 2026-07-03). Citations
  are ordinary inline Markdown links; evidentiary status comes from a
  profile-designated evidence context; "resolvable" is a total, deterministic,
  offline three-class predicate sharing one resolver with `core-broken-link`;
  preservation is enforced by a plan-time citation-loss gate routed through the
  existing ADR-003 `approval` member. Codex-re-reviewed **`READY` (5/5)** and
  flipped `proposed` → `accepted` under the 2026-07-03 autonomous-phase mandate,
  **flagged for Oliver's async ratification** alongside ADR-006/007/009. Two
  Codex non-blocking carry-ins deferred to Phase 4 (gate resolution class 3 on
  `isIntraWiki`; whether a present-but-unresolved citation satisfies a
  require-citation obligation).

## Validation evidence (2026-07-04, repo at `main` = `05b1b9a`)

Toolchain: `go1.24.13 darwin/arm64` (`GOTOOLCHAIN=local`,
`/Users/hermes/sdk/go1.24.13/bin/go`).

This is a docs-only closeout — **no product/source code changed.** Evidence is
recorded in five categories, mirroring the Phase 2 closeout discipline
([`../notes/eval-issue-039.md`](../notes/eval-issue-039.md) is the gate input).
It does **not** claim "all five platforms green" or "full matrix green."

**1. Local validation** (single Unix host — evidence about this host only):

- `go build ./...` — **PASS** (exit 0).
- `go vet ./...` — **PASS** (exit 0).
- `go test ./...` — **PASS** (all 12 packages ok): `cmd/gen-checksums`,
  `cmd/llm-wiki`, `internal/contract`, `internal/fsafe`, `internal/manifest`,
  `internal/plan`, `internal/platform`, `internal/profile`, `internal/scaffold`,
  `internal/txn`, `internal/validate`, `internal/yamladapter`. (Phase 3 added
  `internal/plan`; the citation resolver lives in `internal/validate`, not a
  separate package.)
- `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` — **12/12 PASS**
  (6 Phase 3 criteria 6, 9, 10, 11, 12, 13 + 6 Phase 2 criteria 1, 2, 3×3,
  4-core).

`go test ./...` passes here **because the host is Unix**; the file-permission
assertions that fail on Windows (category 4) hold on this host. Local
validation is not evidence about any other platform.

**2. Named acceptance-step evidence (CI)** — the `Acceptance corpus (Phase 2+3
gate evidence)` step runs `go test ./cmd/llm-wiki -run '^TestAcceptance'` before
the full suite, per leg. Primary source is **run 28700138687** (PR #49 head
`92c0e78`); the `92c0e78..29ce581` branch tip was **docs-only**
(`notes/eval-issue-039.md`), so this run's code is identical to merged `main`
(`05b1b9a`):

| Platform (ADR-002) | Acceptance step |
|---|---|
| linux-amd64 | **PASS** (full job green) |
| linux-arm64 | **PASS** (full job green) |
| macos-arm64 | **PASS** (full job green) |
| windows-amd64 | **PASS** — named acceptance step `success`; the job then failed only at the full `go test` (category 4, #29) |
| macos-amd64 | **not run** — job cancelled, runner unavailable (category 5, #30) |

Green on **all four platforms where the acceptance step executed (4 of 5),
including Windows.** Not observed on macos-amd64.

**3. Full CI matrix — NOT green:**

| Platform (ADR-002) | full `go test ./...` |
|---|---|
| linux-amd64 | green |
| linux-arm64 | green |
| macos-arm64 | green |
| windows-amd64 | **RED** — pre-existing `internal/` permission tests (category 4) |
| macos-amd64 | **did not run** (category 5) |
| `build bundle + checksums` (5 platforms) | green |
| ADR-002 selfcheck smoke (per-platform) | green on linux-amd64, linux-arm64, macos-arm64, windows-amd64; macos-amd64 cancelled |

The `main`-tip CI run **28700401723** (`05b1b9a`) and the branch-tip run
**28700336201** (`29ce581`) were still **pending/queued with no jobs dispatched**
at closeout time — the same "merged over a pending matrix" situation as the
Phase 2 closeout (PR #31). Recorded honestly; the docs-only gate call is
Hermes's.

The **closeout PR #50 run 28700768953** (PR #50 head `67114b9`) reproduces the
same posture on the PR that carries this closeout: linux-amd64, linux-arm64, and
macos-arm64 **pass**; windows-amd64's named acceptance step **PASSES** and the
full suite fails **only** on the known `internal/fsafe`/`internal/txn` POSIX
mode-bit assertions (#29); macos-amd64 test + selfcheck are **queued/unobserved**
(#30); `build bundle + checksums` and the observed per-platform selfcheck smokes
**pass**. Same caveats as above, now named on the closeout PR itself.

**4. Known pre-existing failure (not a Phase 3 regression):** windows-amd64
full `go test ./...` is red **only** in `internal/fsafe` and `internal/txn`, on
Unix file-permission-mode assertions (`mode = -rw-rw-rw-, want -rw-r-----`).
These predate Phase 3 and are unchanged by it (the Phase 3 train touched no
permission code). Tracked as follow-up issue **#29** (deferral, not fixed here).

**5. macos-amd64 no evidence (runner unavailable):** the `macos-13` leg was
cancelled/never-dispatched on branch and `main` alike. macos-amd64 has produced
**no** acceptance evidence and **no** full-suite evidence for Phase 3; its code
path is identical to green macos-arm64 and the `build bundle + checksums` job
builds `darwin/amd64` cleanly — **inference, not observed CI.** Tracked as
follow-up issue **#30**.

## Phase 3 exit criteria

Exit gate (`design/build-out-plan.md` §Phase 3): acceptance criteria **6, 9, 10,
11, 12, 13** pass. Verdict: **observed on 4 of 5 platforms (including Windows via
the named acceptance step); macos-amd64 closed on inference + standing follow-up
#30.** The closeout does **not** claim 5/5 — the gap is stated for Hermes/Oliver
to weigh. Per-criterion journeys are in
[`../notes/eval-issue-039.md`](../notes/eval-issue-039.md) (§"Criteria → test
mapping").

| Criterion | Status | Evidence (test in `cmd/llm-wiki/acceptance_phase3_test.go`) |
|---|---|---|
| **6** — unknown frontmatter fields survive authoring round trip | pass (4/5 observed) | `TestAcceptanceCriterion6UnknownFieldsRoundTripPlanApply` — both unknown fields + edited known field present in the committed live file, both directions (ADR-006) |
| **9** — sourced-claim citations resolve offline | pass (4/5 observed) | `TestAcceptanceCriterion9SourcedClaimCitationsResolveOffline` — zero `core-citation-*` on inspect/commit; negative control raises `core-citation-unresolved` (ADR-008) |
| **10** — repeated unchanged input is a no-op | pass (4/5 observed) | `TestAcceptanceCriterion10RepeatedUnchangedInputIsNoOp` — `plan.noOp`, empty transaction, byte-identical whole tree (ADR-006 fixed point) |
| **11** — edit previewed before apply | pass (4/5 observed) | `TestAcceptanceCriterion11EditPreviewPrecedesApply` — non-empty `plan.diff`; decline leaves live hash unchanged; apply commits `plan.stagedHash` (ADR-006) |
| **12** — writes only through staged plan/apply | pass (4/5 observed) | `TestAcceptanceCriterion12WritesOnlyThroughStagedPlanApply` — inspect/plan touch only `.llm-wiki/`; apply's sole non-staging change is the target page (ADR-005/006) |
| **13** — stale plan rejected, zero mutation | pass (4/5 observed) | `TestAcceptanceCriterion13StalePlanRejectedZeroMutation` — exit 4, `invalid-invocation`, whole-tree byte-identical around the rejected apply (ADR-006 base binding) |

Full accounting in [`../notes/eval-issue-039.md`](../notes/eval-issue-039.md).

## Notes / deferrals

- **Windows permission-mode tests** — full suite red on windows-amd64
  (`internal/fsafe` + `internal/txn`); pre-existing, predates Phase 3, not a
  regression. Follow-up issue **#29** (still open, unfixed by Phase 3).
- **macos-amd64 runner gap** — `macos-13` leg never dispatched; no amd64-macOS
  CI evidence. Phase 3 closed with this leg unobserved, as Phase 2 did.
  Follow-up issue **#30** (still open).
- **ADR-008 Codex carry-ins → Phase 4** — gate resolution class 3 on
  `isIntraWiki`; whether a present-but-unresolved citation satisfies a
  require-citation obligation. Both are profile-vocabulary work, deferred to the
  Phase 4 academic-research profile.
- **Enrichment half of criterion 11 → Phase 5**; **profile citation vocabulary
  → Phase 4**; **cross-surface (skills/hooks/CI/CLI) parity → Phase 6**.
- **Signing / provenance** — content-provenance (citations) is decided by
  ADR-008; **supply-chain** signing/attestation stays deferred to a dedicated
  supply-chain ADR (out of ADR-002/009 scope); residual supply-chain risk stays
  `open`.
- **Oliver's async ratification** of ADR-006/007/008/009 remains outstanding —
  recorded as a standing flag under the 2026-07-03 autonomous-phase mandate; not
  a merge blocker.
- **CI on closeout PRs** — the Phase 2 closeout PR #31 merged over a red-Windows
  / pending-macos matrix as the docs-only gate call; #40 sits in the same
  posture (windows full-suite red #29, macos-amd64 undispatched #30, main-tip
  run pending) and is Hermes's gate call.
