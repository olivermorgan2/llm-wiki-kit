# Project State

_Last updated: 2026-07-08_

## Current status

- **2026-07-08 — Phase 5 / Enrichment + index maintenance: ADR-011 drafted (`proposed`); Qwen3.7 Max adversarial review next.**
  Phase 4 is the last closed phase; ratification debt is 0 (Oliver ratified
  ADR-006/007/008/009/010 on 2026-07-08). Phase 5 **milestone #5** and gate-blocking
  issue **#76** (`design: Draft + accept ADR-011 — deterministic index maintenance`)
  are filed; **[ADR-011](adr/adr-011-deterministic-index-maintenance.md) is now
  drafted as `proposed`** on branch `docs/adr-011-index-maintenance` (Option A —
  fenced generated regions in the OKF-reserved `index.md`, standalone ADR-006 staged
  write, `core-index-stale` warning; all 9 issue-#76 scope items decided or deferred
  with rationale). **No review has run and no acceptance is claimed.** Next action is
  the **adversarial review — Qwen3.7 Max via OpenRouter** (substituted for OpenAI
  Codex; overlay-legal, recorded on issue #76 and in the log), **not** implementation;
  ADR-011 remains the prerequisite before Phase 5 implementation issues I2–I7 open,
  unblocking only on Qwen `READY` + Oliver acceptance. The provisional build-out plan
  reserved ADR-010 for index maintenance, but ADR-010 is now the Phase 4 profile-data
  schema ADR; the index-maintenance ADR takes the next free number, ADR-011. Details
  in the 2026-07-08 entries of [`../knowledge/log.md`](../knowledge/log.md). `main` was
  previously rolled back to `051590f` (Phase 4 closeout) and the hermes-workflow-overlay
  was adopted (`main` now guard-gated and branch-protected, PRs only).
- **Phase 4 / Academic-research profile: complete.** The `academic-research`
  profile shipped end-to-end per addendum 003 on the ADR-007 profile system:
  the profile-data schema + citation vocabulary ADR-010 (#53), the data-backed
  loader with one-level `extends: core` (#54), the type-conditional structural
  rules (#55), the citation obligations + cross-page evidence rules that land
  ADR-008's two Codex carry-ins (#56), the full profile + per-type fixture
  corpus (#57), `init --profile academic-research` + per-type authoring
  templates (#58), and the named, criterion-mapped Phase 4 acceptance corpus
  (#59). This closeout (#60) refreshes state + knowledge. `main` = `502c717`.
  **Milestone #4 closes after this PR merges** (0 open issues once #60 lands) —
  not closed in this PR.
- **Current issue: #60 — Phase 4 closeout: state and knowledge update.**
  Docs-only. Brings `design/state.md` + `knowledge/` to final Phase 4 truth. No
  engine, test, or CI changes.
- **Next phase: Phase 5 / Enrichment + index maintenance.** Next action: file
  the Phase 5 backlog per `design/build-out-plan.md` §Phase 5. The build-out
  plan provisionally reserved **ADR-010 for index maintenance**; that number was
  taken by the Phase-4 schema ADR (adr-alloc allocates sequentially), so the
  index-maintenance ADR takes the next free number when Phase 5 lands — recorded
  in `knowledge/log.md` and the ADR-010 numbering note.
- **Phase 3 / Authoring + staged mutation: complete; milestone #3 closed**
  (closeout PR #50). **Phase 2 / Install-init: complete; milestone #2 closed**
  (closeout PR #31). **Phase 1 / Foundation: complete** (ADR-001–005 accepted).
  Oliver **ratified** ADR-006/007/008/009 — **and ADR-010** — on 2026-07-08
  (via the ratification PR closing #74); no ratification debt outstanding.

## Phase 4 — Academic-research profile — complete

Goal: the initial domain value proposition — a researcher can `init` a wiki with
the academic-research profile and author each profiled page type
(`source`, `claim`, `method`, `question`, `synthesis`) through the staged
workflow, with domain rules (required fields, enums, list-min, required sections,
evidence/citation obligations) enforced and reported **separately** from OKF
(criterion 5). Governing decision: ADR-007 (accepted); the schema surface and
citation vocabulary it deferred are fixed by **ADR-010** (accepted this phase);
ADR-008's two carry-ins land here.

### Issues + PRs (all merged; closeout #60 via this PR)

| Issue | Title (abbrev.) | ADR | PR | Merge commit | Merged |
|---|---|---|---|---|---|
| #53 | ADR-010 profile-data schema + rule/citation vocabulary | 010 | #61 | `7b6e20b` | 2026-07-04 |
| #54 | data-backed profile loader + one-level extends | 007/010 | #62 | `e6c3cf3` | 2026-07-04 |
| #55 | type-conditional structural rules (profile-*) | 010 | #63 | `d55a554` | 2026-07-04 |
| #56 | citation obligations + cross-page evidence (carry-ins) | 008/010 | #64 | `9497179` | 2026-07-04 |
| #57 | ship academic-research contract + fixture corpus | 007/010 | #65 | `8419c2b` | 2026-07-04 |
| #58 | `init --profile academic-research` + per-type templates | 007 | #66 | `f5cc82a` | 2026-07-04 |
| #59 | Phase 4 acceptance corpus (criterion 4) | 006/008/010 | #67 | `502c717` | 2026-07-04 |
| #60 | Phase 4 closeout — state + knowledge | — | (this PR) | — | — |

### ADR adopted this phase

- **ADR-010 — Declarative profile-data schema and profile rule / citation
  vocabulary** (accepted 2026-07-04). A **closed, per-type declarative schema**
  (`required`/`recommended`/`enums`/`listMin`/`recommendedAnyOf`/
  `requiredSections`/`evidenceSections`/`citation.{requireWhen,forbiddenTargetTypes}`
  + a `severities` override map) the engine owns rule kinds for; core `okf-*`/
  `core-*` rules stay **engine-code and byte-identical** (golden parity, a full
  declarative-core migration deferred). Both ADR-008 carry-ins decided:
  `profile-citation-required` fires on **absence** (a present-but-unresolved
  citation satisfies it and only raises the promotable `core-citation-unresolved`);
  repo-path resolution class 3 gated on `isIntraWiki`. Finding-code namespace:
  generic `profile-*` (`ruleset: profile`). Codex second-opinion `NEEDS_REVISION`
  → revised in-PR (added the `severities` key per ADR-008 5(d), corrected the
  carry-in-1 wording, sharpened the extends-core semantics, specified the
  field-presence one-finding precedence). Accepted under the 2026-07-03
  autonomous-phase mandate and flagged for Oliver's async ratification alongside
  ADR-006/007/008/009; **ratified by Oliver 2026-07-08**.

## Validation evidence (2026-07-04, repo at `main` = `502c717`)

Toolchain: `go1.24.13 darwin/arm64` (`GOTOOLCHAIN=local`,
`/Users/hermes/sdk/go1.24.13/bin/go`). Evidence is recorded in five categories,
mirroring the Phase 2/3 closeout discipline. **Unlike Phase 3, the full CI matrix
is green on all five platforms** — the two carried caveats (#29 windows perm
bits, #30 macos-amd64 runner) were resolved and merged before Phase 4 (#52/#51).

**1. Local validation** (single Unix host):

- `go build ./...` — **PASS** (exit 0).
- `go vet ./...` — **PASS** (exit 0).
- `go test ./...` — **PASS** (all 12 test packages; the new `profiles/` package
  is embed-only, no test files): `cmd/gen-checksums`, `cmd/llm-wiki`,
  `internal/contract`, `internal/fsafe`, `internal/manifest`, `internal/plan`,
  `internal/platform`, `internal/profile`, `internal/scaffold`, `internal/txn`,
  `internal/validate`, `internal/yamladapter`.
- `go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v` — **PASS** (Phase 2
  + Phase 3 + Phase 4 criteria, incl. the three criterion-4 academic-research
  journeys and their negative-control subtests).

**2. Named acceptance-step evidence (CI)** — the `Acceptance corpus (Phase 2+3+4
gate evidence)` step runs `go test ./cmd/llm-wiki -run '^TestAcceptance'` before
the full suite, per leg. Primary source is **run 28715363791** (PR #67 head
`67c85ca`, byte-identical to merged `main` `502c717`):

| Platform (ADR-002) | Acceptance step |
|---|---|
| linux-amd64 | **PASS** |
| linux-arm64 | **PASS** |
| macos-arm64 | **PASS** |
| macos-amd64 | **PASS** (macos-15-intel, restored by #51) |
| windows-amd64 | **PASS** |

The three criterion-4 journeys —
`TestAcceptanceCriterion4AcademicResearchAuthorEachProfiledType`,
`…NegativeControls` (source-missing-authors / method-missing-section /
question-bad-status), `…EvidenceObligationJourney` — are green on **all five
platforms**.

**3. Full CI matrix — GREEN on all five platforms** (PR #67 run 28715363791):

| Platform (ADR-002) | full `go test ./...` |
|---|---|
| linux-amd64 | green |
| linux-arm64 | green |
| macos-arm64 | green |
| macos-amd64 | green |
| windows-amd64 | green |
| `build bundle + checksums` (5 platforms) | green |
| ADR-002 selfcheck smoke (per-platform) | green on all five |

The `main`-tip CI run **28715417217** (`502c717`, the #67 merge) **completed
success** — the full five-platform matrix is green on `main` itself, not only on
the PR branch. Each Phase-4 PR (#61–#67) also merged only after its own
five-platform matrix went green. This is a cleaner posture than the Phase 2/3
closeouts, which merged over red/pending matrices.

**4. Known pre-existing failures: none.** The Phase-3 caveats are resolved:
windows-amd64 full suite is green (#29/PR #52 asserts POSIX perm bits on unix
only); macos-amd64 is observed on `macos-15-intel` (#30/PR #51). Residual: the
`macos-15-intel` Aug 2027 sunset stays tracked in `knowledge/risks.md`.

**5. Golden parity (core behavior unchanged):** every pre-Phase-4 test passes
byte-identical across all seven PRs. Core `okf-*`/`core-*` findings, severities,
and messages are engine-code and untouched (ADR-010 sub-decision 2); the
data-driven profile layer fires only for profiled types, and a `core` (or zero)
profile runs no per-type rules. The core scaffold output is byte-identical
(scaffold golden tests unchanged).

## Phase 4 exit criteria

Exit gate (`design/build-out-plan.md` §Phase 4): acceptance criterion **4
(academic-research)** plus the addendum-003 fixture table pass. Verdict:
**observed green on all five platforms** via the named acceptance step and the
`internal/validate` fixture corpus.

| Criterion / gate | Status | Evidence |
|---|---|---|
| **4** — init academic-research + author each profiled type + validate clean | pass (5/5) | `TestAcceptanceCriterion4AcademicResearchAuthorEachProfiledType` (init → staged plan/apply of source/claim/method/question/synthesis → validate clean under the profile) |
| **addendum-003 fixtures** — each invalid trips exactly one rule; valid clean | pass | `internal/validate/academic_fixtures_test.go` (`TestAcademicInvalidExamplesFailExactlyOneRule` over the 9-fixture corpus; `TestAcademicValidExamplesHaveNoFindings`) |
| **non-vacuous gate** — rules actually fire end-to-end | pass (5/5) | `TestAcceptanceCriterion4AcademicResearchNegativeControls` (profile-required-field / profile-required-section / profile-field-enum fail validation) |
| **evidence obligation** — supported claim needs a citation | pass (5/5) | `TestAcceptanceCriterion4AcademicResearchEvidenceObligationJourney` (uncited → `profile-citation-required`; cited → clear) |
| **5** — OKF vs profile findings reported separately | pass | new codes all tagged `ruleset: profile`; fixture/unit tests assert the split |

## Notes / deferrals

- **ADR-010 ratification** — accepted under the 2026-07-03 autonomous-phase
  mandate, flagged for Oliver's async ratification alongside ADR-006/007/008/009;
  **ratified by Oliver 2026-07-08** (via the ratification PR closing #74).
- **Codex second-opinion reviews** — Codex reviewed I1–I5 (#53–#57): ADR-010
  (`NEEDS_REVISION` → revised), the loader (`NEEDS_REVISION`: non-mutating deep
  copy + honest inheritance docs), the structural rules (`NEEDS_REVISION`:
  empty-string precedence fix), the citation obligations (`READY`), and the
  fixture corpus (`NEEDS_REVISION`: added enum/list-min fixtures). Each finding
  was addressed in-PR. **I6/I7 (#58/#59) reviews were deferred** — the Codex CLI
  hit its usage limit (resets 2026-07-07); both are covered by strong end-to-end
  test evidence (scaffold-validates-clean-under-profile; the CLI init→validate
  and plan/apply journeys) and are flagged on the PRs for a later Codex pass.
- **Declarative-core migration** — core rules stay engine-code in Phase 4 (golden
  parity). A full migration of the `core` profile-layer rules into
  `profiles/core/profile.yaml` is a separate, golden-guarded future issue
  (ADR-010 sub-decision 2), out of MVP scope.
- **LLM_WIKI_EVIDENCE_SECTIONS demoted** — evidence contexts now come primarily
  from the active profile's per-type `evidenceSections` (read from
  `llm-wiki.yaml`); the env var remains an explicit global override (#56).
- **Q4 (research-profile templates / conditional-section syntax)** stays
  assumption-locked to addendum 003, but its **schema surface is now fixed** by
  ADR-010 (a closed vocabulary, no conditional-section syntax). A different
  template set supersedes addendum 003 with a new addendum; conditional-section
  syntax would need a new rule kind (new ADR).
- **Provenance risk** narrowed: ADR-008's two carry-ins are implemented and
  fixture/acceptance-proven for the academic-research profile; the risk's
  remaining scope is cross-surface (skills/hooks/CI/CLI) parity (Phase 6).
  Supply-chain signing stays a separate deferred ADR.
- **CI on closeout PRs** — as with the Phase 2/3 closeouts, #60 may merge over a
  still-queued `main`-tip matrix; the docs-only gate call is Hermes's, and the
  byte-identical code matrix (PR #67) is already fully green on all five
  platforms.
