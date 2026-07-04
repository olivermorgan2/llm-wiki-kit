# Project State

_Last updated: 2026-07-04_

## Current status

- **Phase 3 / Authoring + staged mutation: in progress.** The `page`
  command group is building out issue-by-issue on the ADR-006 transaction
  substrate. Landed so far: ADR-008 (provenance & citation model) accepted
  (#32); `page inspect` read-only report (#42, PR #43); `page plan` staged
  whole-page preview (#34, PR #44); `page apply` journaled commit with
  stale-plan rejection + generic ADR-003 approval-refusal plumbing (#35,
  PR #45); citation mechanism — evidence contexts, offline three-class
  resolver, `core-citation-{malformed,unresolved,duplicate}` findings (#36,
  PR #46). `main` = `7a5f17a`.
- **Current issue: #37 — citation-loss approval gate for page plans.**
  Implements ADR-008 sub-decision 6: `page plan` computes the citation-loss
  diff (normalized evidence-context citation-target sets, source vs staged),
  emits a `core-citation-loss` finding, and writes the ADR-003 approval
  sidecar so `page apply` refuses the unapproved removal with exit 3 until
  `--approve`. Builds on #36's resolver and #44/#45's generic approval
  plumbing (the trigger that writes the sidecar). Out of scope: authoring
  adapter (#38), acceptance fixtures (#39), Phase 3 closeout (#40), Phase 4
  profile vocabulary, new envelope fields, and all network I/O. Phase 3 stays
  **open**.
- **Phase 2 / Install-init: complete; milestone #2 closed** (closeout PR #31,
  milestone closed manually 2026-07-03). Oliver's async ratification of
  ADR-006/007/009 and ADR-008 remains the one outstanding flag — not a merge
  blocker.

## Phase 2 / Install-init — complete

Phase 2 made the kit installable. It landed the three ADR prerequisites
(ADR-006 cross-file transaction model, ADR-007 profile-system boundary, ADR-009
install/upgrade/uninstall ownership), then built on them: the `internal/txn`
cross-file transaction commit over the `internal/fsafe` staging area (ADR-006);
`llm-wiki init` with the core profile + wiki bundle scaffold (ADR-007); the
`install` lifecycle for new and non-empty repos with `--dry-run`, silent-
overwrite refusal, and a `.llm-wiki/manifest.json` version record (ADR-009);
multi-platform release builds + selection wiring with a per-platform selfcheck
smoke (ADR-002); a five-platform CI test matrix; and a named, criterion-mapped
install/init acceptance corpus that serves as the Phase 2 gate evidence.

### Issues (all closed; closeout #21 closed via PR #31)

| Issue | Title | ADR |
|-------|-------|-----|
| #14 | Draft + accept ADR-006/007/009 — Phase 2 architectural prerequisites | ADR-006, ADR-007, ADR-009 |
| #16 | Cross-file transaction commit on the fsafe staging area | ADR-006 |
| #18 | `llm-wiki init` with the core profile + wiki bundle scaffold | ADR-007 |
| #19 | Install into new + non-empty repos: `--dry-run`, overwrite refusal, version record | ADR-009 |
| #15 | Stand up cross-platform CI test matrix (5 platforms) | — |
| #20 | New-repo + non-empty-repo install/init acceptance corpus | ADR-002, ADR-003, ADR-005, ADR-006, ADR-009 |
| #17 | Multi-platform release builds + selection shim wiring | ADR-002 |
| #21 | Phase 2 closeout state and knowledge update | — (this PR) |

### PRs (all merged)

| PR | Merged | Closes | Merge commit |
|----|--------|--------|--------------|
| #22 | 2026-07-03 | #14 | `4136e70` |
| #23 | 2026-07-03 | #16 | `9077a6a` |
| #24 | 2026-07-03 | #18 | `c8c40d2` |
| #25 | 2026-07-03 | #19 | `d5bc4cd` |
| #26 | 2026-07-03 | #15 | `fb65639` |
| #27 | 2026-07-03 | #20 | `a078007` |
| #28 | 2026-07-03 | #17 | `33dd78a` |
| #31 | 2026-07-03 | #21 | `30cfbac` |

### ADRs adopted

- ADR-006 — Staged mutation / cross-file transaction model
- ADR-007 — Profile-system boundary (one inheritance level; core)
- ADR-009 — Install/upgrade/uninstall asset-ownership + version-record manifest

All three were Codex-re-reviewed **`READY` (5/5, no blockers)** and flipped
`proposed` → **`accepted`** on **2026-07-03** under the autonomous-phase
mandate. They are **flagged for Oliver's async ratification** — recorded as a
standing flag, not a merge blocker. Signing / provenance attestation is **not**
covered by ADR-009; it is re-deferred to a dedicated supply-chain ADR, so the
residual supply-chain risk stays `open` in
[`../knowledge/risks.md`](../knowledge/risks.md).

## Validation evidence (2026-07-03, repo at `main` = `33dd78a`)

Toolchain: `go1.24.13 darwin/arm64` (`GOTOOLCHAIN=local`,
`/Users/hermes/sdk/go1.24.13/bin/go`).

This is a docs-only closeout — **no product/source code changed.** Evidence is
recorded in five categories, mirroring the discipline of
[`../notes/eval-issue-020.md`](../notes/eval-issue-020.md). It does **not**
claim "all five platforms green" or "full matrix green."

**1. Local validation** (single Unix host — evidence about this host only):

- `go build ./...` — **PASS** (exit 0).
- `go vet ./...` — **PASS** (exit 0).
- `go test ./...` — **PASS** (all 11 packages ok): `cmd/gen-checksums`,
  `cmd/llm-wiki`, `internal/contract`, `internal/fsafe`, `internal/manifest`,
  `internal/platform`, `internal/profile`, `internal/scaffold`, `internal/txn`,
  `internal/validate`, `internal/yamladapter`.
- `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` — **6/6 PASS**
  (criteria 1, 2, 3×3, 4-core).

`go test ./...` passes here **because the host is Unix**; the file-permission
assertions that fail on Windows (category 4) hold on this host. Local
validation is not evidence about any other platform.

**2. Named acceptance-step evidence (CI)** — the `Acceptance corpus (Phase 2
gate evidence)` step runs `go test ./cmd/llm-wiki -run '^TestAcceptance'` before
the full suite, per leg:

| Platform (ADR-002) | Acceptance step |
|---|---|
| linux-amd64 | **6/6 PASS** |
| linux-arm64 | **6/6 PASS** |
| macos-arm64 | **6/6 PASS** |
| windows-amd64 | **6/6 PASS** |
| macos-amd64 | **not run** — runner unavailable (category 5) |

Green on **all four platforms where a runner executed it (4 of 5), including
Windows.** Not observed on macos-amd64.

**3. Full CI matrix — NOT green:**

| Platform (ADR-002) | full `go test ./...` |
|---|---|
| linux-amd64 | green |
| linux-arm64 | green |
| macos-arm64 | green |
| windows-amd64 | **RED** — pre-existing `internal/` tests (category 4) |
| macos-amd64 | **did not run** (category 5) |
| `cross-compile-smoke` (5 binaries + checksum manifest) | green |
| ADR-002 bundle selfcheck smoke (per-platform) | green |

**4. Known pre-existing failure (not a Phase 2 regression):** windows-amd64
full `go test ./...` is red **only** in `internal/fsafe` and `internal/txn`, on
Unix file-permission-mode assertions (`mode = -rw-rw-rw-, want -rw-r-----`).
Same failures on `main`'s own run at `fb65639`
([run 28646677901](https://github.com/olivermorgan2/llm-wiki-kit/actions/runs/28646677901)).
Tracked as follow-up issue **#29** (deferral, not fixed here).

**5. macos-amd64 no evidence (runner unavailable):** the `macos-13` leg never
dispatched — queued/never-ran on branches and `main` alike. The `main` tip run
at `33dd78a`
([run 28651421487](https://github.com/olivermorgan2/llm-wiki-kit/actions/runs/28651421487))
was still **pending with no jobs dispatched** at closeout time. macos-amd64 has
produced **no** acceptance evidence and **no** full-suite evidence; its code
path is identical to green macos-arm64 and `cross-compile-smoke` builds
`darwin/amd64` cleanly — **inference, not observed CI.** Tracked as follow-up
issue **#30**.

## Phase 2 exit criteria

Exit gate (`design/build-out-plan.md` §Phase 2): acceptance criteria **2, 3,
4 (core)** pass on all five platforms + criterion **1** (install half)
demonstrated. Verdict: **observed on 4 of 5 platforms; macos-amd64 closed on
inference + follow-up #30.** The closeout does **not** claim 5/5 — the gap is
stated for Hermes/Oliver to weigh.

| Criterion | Status | Evidence |
|---|---|---|
| **1** — install half (dry-run full plan, no-op) | pass (4/5 observed) | `TestAcceptanceCriterion1InstallDryRunFullPlanNoOp`; acceptance step |
| **2** — CLI runs on every platform | pass (4/5 observed) | `TestAcceptanceCriterion2VersionRunsOnHostPlatform`; `cross-compile-smoke` + selfcheck smoke build/select the 5 binaries |
| **3** — install into new + non-empty repos, no file loss, refusal zero-mutation | pass (4/5 observed) | `TestAcceptanceCriterion3InstallNewRepoThenValidateClean`, `…NonEmptyRepoNoFileLoss`, `…CollisionRefusalIsZeroMutation` |
| **4 (core)** — init with core profile | pass (4/5 observed) | `TestAcceptanceCriterion4InitCoreThenValidateClean` |

Full accounting in [`../notes/eval-issue-020.md`](../notes/eval-issue-020.md).

## Current phase — Phase 3 / Authoring + staged mutation (in progress)

Goal: the thinnest end-to-end "create real value" path — author a new page
through the staged, preview-before-write workflow (`page inspect` / `page plan`
/ `page apply` with hash-bound, stale-rejecting plans; authoring skill adapter).

Progress: `page inspect` (#42), `page plan` (#34), `page apply` (#35), and the
citation mechanism (#36) — ADR-008's core resolver + evidence contexts +
`core-citation-*` findings — are merged. The citation-loss approval gate (#37)
— ADR-008 sub-decision 6's plan-time loss diff wired through ADR-003's approval
member — is the current issue. Remaining Phase 3 work after #37: authoring
skill adapter (#38), acceptance-corpus expansion (#39), Phase 3 closeout (#40).

**ADR dependency note:** ADR-006 (staged mutation) is accepted — Phase 3
consumes its `inspect/plan/apply` UX + hash-bound stale-plan rejection half.
**ADR-008 (provenance & citation model) is now drafted and accepted**
(2026-07-03, issue #32, branch `docs/adr-008-provenance-citation-model`; Codex
`READY` 5/5 under the autonomous-phase mandate, flagged for Oliver's async
ratification; the `design/adr/` index 008 gap is filled). Phase 3 authoring is
**unblocked**. Phase 3 consumes ADR-008's core mechanism (context-based
citations, the total offline three-class resolver, the plan-time
citation-loss/`approval` gate) plus the criterion-9 fixtures; Phase 4 consumes
its profile citation vocabulary. Next: **file Phase 3 issues** (`page
inspect`/`plan`/`apply` + authoring adapter + provenance/citation fixtures).

## Notes / deferrals

- **Windows permission-mode tests** — full suite red on windows-amd64
  (`internal/fsafe` + `internal/txn`); pre-existing, predates #20. Follow-up
  issue **#29**.
- **macos-amd64 runner gap** — `macos-13` leg never dispatched; no amd64-macOS
  CI evidence. Phase 2 closed with this leg unobserved. Follow-up issue **#30**.
- **Signing / provenance** — still deferred to a dedicated supply-chain ADR
  (ADR-009 explicitly out of scope); residual supply-chain risk stays `open`.
- **Oliver's async ratification** of ADR-006/007/009 remains outstanding —
  recorded as a flag, consistent with the 2026-07-03 autonomous-phase mandate;
  not a merge blocker.
- **CI on the closeout PR #31** showed the windows full-suite job red
  (pre-existing) and macos-amd64 undispatched — the PR was docs-only, and
  Hermes merged over the red/pending matrix as the closeout gate call.
