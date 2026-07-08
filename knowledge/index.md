# knowledge/ — index

Front door to the `llm-wiki-kit` knowledge layer. Conventions live in
[`SCHEMA.md`](SCHEMA.md); this file is the live map.

## Current phase

**2026-07-08 — Phase 5 / Enrichment + index maintenance: issues filed, ready to build.**
Phase 4 is the last closed phase. Oliver ratified ADR-006/007/008/009/010 on
2026-07-08; ratification debt is 0. Phase 5 milestone (**#5**) is active;
**[ADR-011](../design/adr/adr-011-deterministic-index-maintenance.md)
(deterministic index maintenance) is `accepted`** (Option A: fenced generated regions in the
OKF-reserved `index.md`, standalone ADR-006 staged write, `core-index-stale`
warn). The **adversarial review of the ADR-011 draft — Qwen3.7 Max via OpenRouter**
(`qwen/qwen3.7-max`, substituted for Codex; overlay-legal, recorded on issue
[#76](https://github.com/olivermorgan2/llm-wiki-kit/issues/76) and in
[`log.md`](log.md)) returned `READY`, score 4/5, no blockers
([artifact](reviews/2026-07-08-qwen-adr-011-review.md)); its three non-blocking
clarity findings were folded into the draft before acceptance. Oliver then
explicitly accepted ADR-011 on 2026-07-08, flipping status `proposed` →
`accepted` in PR [#79](https://github.com/olivermorgan2/llm-wiki-kit/pull/79),
keeping ratification debt at 0.

**Phase 5 implementation issues are now filed** and work can begin:

- Enrichment track (I1):
  - [#80](https://github.com/olivermorgan2/llm-wiki-kit/issues/80): enrichment skill adapter (reuses page plan/apply machinery)
  - [#81](https://github.com/olivermorgan2/llm-wiki-kit/issues/81): enrichment preview fixtures (core + academic-research)
- Index track (I2-I7):
  - [#82](https://github.com/olivermorgan2/llm-wiki-kit/issues/82): core index generation and fenced-region parsing (Option A)
  - [#83](https://github.com/olivermorgan2/llm-wiki-kit/issues/83): index CLI command with --dry-run and ADR-006 routing
  - [#84](https://github.com/olivermorgan2/llm-wiki-kit/issues/84): index stability tests (byte-idempotency, human section preservation)
  - [#85](https://github.com/olivermorgan2/llm-wiki-kit/issues/85): `core-index-stale` validation finding (ADR-004 framework)
  - [#86](https://github.com/olivermorgan2/llm-wiki-kit/issues/86): init fence scaffolding (Phase 2 backfill)

**Phase 5 plan** ([phase-5-plan.md](../phase-5-plan.md)) passed adversarial
review by Qwen3.7 Max (verdict: READY, score 4/5) after incorporating 9
findings: corrected acceptance criteria references (criterion 11 + PRD §14, not
14/15), clarified sort order (type + bundle-relative path), added validation
finding code scope (`core-index-stale` / `core-index-unmanaged`), and added
post-apply index refresh integration with page apply. Provisional build-out plan
called this ADR-010, but that number is now the Phase 4 profile-data schema ADR;
the index-maintenance ADR takes the next free number, ADR-011. `main` is
guard-gated and branch-protected (PRs only). See the 2026-07-08 entries in
[`log.md`](log.md).

**Phase 4 / Academic-research profile closed (issues #53–#59, PRs #61–#67 +
closeout PR for #60; milestone #4 closes after that PR merges; ADR-010 accepted
2026-07-04 under the autonomous-phase mandate — flagged for Oliver's async
ratification; ratified by Oliver 2026-07-08).** The `academic-research` profile shipped end-to-end: the ADR-010
profile-data schema + citation vocabulary (#53), the data-backed loader with
one-level `extends: core` (#54), type-conditional structural rules (#55),
citation obligations landing ADR-008's two carry-ins (#56), the full profile +
per-type fixture corpus (#57), `init --profile academic-research` + per-type
templates (#58), and the named Phase 4 acceptance corpus (#59); #60 is the
closeout (evidence in [`../design/state.md`](../design/state.md)). The Phase 4
exit gate (criterion **4 (academic-research)** + the addendum-003 fixtures) is
**green on all five platforms** — the Phase-3 CI caveats are resolved
(windows-amd64 full suite green after #52; macos-amd64 observed on
`macos-15-intel` after #51). **Phase 5 / Enrichment + index maintenance** is
next. Core rules stayed engine-code and byte-identical throughout (golden
parity, ADR-010 sub-decision 2).

**Phase 3 / Authoring + staged mutation closed (issues #32/#34–#42, PRs
#33/#43–#49 + closeout #40; milestone #3 closed).** The staged authoring surface
shipped: `page inspect`/`plan`/`apply`, the ADR-008 citation mechanism, the
plan-time citation-loss approval gate, the authoring skill adapter, and the Phase
3 acceptance corpus (criteria 6, 9–13). Issue #41 is a closed duplicate of #40.

**Phase 1 / Foundation closed (issues #1–#6, PRs #7–#13, ADR-001–005 accepted).
Phase 2 / Install-init closed (issues #14–#21, PRs #22–#28 + closeout PR #31;
milestone #2 closed 2026-07-03; ADR-006/007/009 accepted 2026-07-03 under the
autonomous-phase mandate — flagged for Oliver's async ratification; ratified by
Oliver 2026-07-08).**
Bootstrapped 2026-06-30 from `claude-workflow-kit` v5.0.1.
Source PRD normalized; adversarial Codex review (`NEEDS_REVISION`) landed as PRD
addenda 001–005; `/prd-to-mvp` produced [`design/mvp.md`](../design/mvp.md) and
[`design/build-out-plan.md`](../design/build-out-plan.md) (7 phases / Slices
0–6 + cross-platform gate, all 21 acceptance criteria mapped). The five Phase-1
prerequisite ADRs ([`design/adr/`](../design/adr/) 001–005: toolchain, binary
selection, JSON contract, validation model, safe-FS gate) were drafted `proposed`,
revised against the 2026-07-01 Codex blockers, Codex-re-reviewed `READY`, and
**accepted by Oliver on 2026-07-01** — status flipped to `accepted`; open
questions **Q2/Q7/Q8 closed** with ADR + log back-references. See
[`log.md`](log.md),
[`reviews/2026-07-01-codex-adr-001-005-review.md`](reviews/2026-07-01-codex-adr-001-005-review.md), and
[`reviews/2026-07-01-codex-adr-001-005-rereview.md`](reviews/2026-07-01-codex-adr-001-005-rereview.md).

## Knowledge layer

| File | What you'll find |
|---|---|
| [`SCHEMA.md`](SCHEMA.md) | Layer conventions, file roles, entry formats. |
| [`project-brief.md`](project-brief.md) | Distilled product definition and architectural spine. |
| [`risks.md`](risks.md) | Live risk register (product, cross-cutting, process, review). |
| [`open-questions.md`](open-questions.md) | Q1–Q8 + bootstrap questions, with status. |
| [`log.md`](log.md) | Chronological decision & review log. |
| [`reviews/`](reviews/) | Verbatim review archives — [`2026-06-30-codex-prd-review.md`](reviews/2026-06-30-codex-prd-review.md), [`2026-07-01-codex-adr-001-005-review.md`](reviews/2026-07-01-codex-adr-001-005-review.md), [`2026-07-01-codex-adr-001-005-rereview.md`](reviews/2026-07-01-codex-adr-001-005-rereview.md), [`2026-07-08-codex-pr66-retro-review.md`](reviews/2026-07-08-codex-pr66-retro-review.md), [`2026-07-08-codex-pr67-retro-review.md`](reviews/2026-07-08-codex-pr67-retro-review.md), [`2026-07-08-codex-pr75-ratification-review.md`](reviews/2026-07-08-codex-pr75-ratification-review.md), [`2026-07-08-qwen-adr-011-review.md`](reviews/2026-07-08-qwen-adr-011-review.md). |
| [`sources/`](sources/) | Verbatim sources — [`prd-original.md`](sources/prd-original.md). |

## Pointers into `design/` (authoritative, tool-maintained)

| Artifact | Status |
|---|---|
| [`design/prd.md`](../design/prd.md) | Source PRD, verbatim (also archived at `sources/prd-original.md`). |
| [`design/prd-normalized.md`](../design/prd-normalized.md) | Canonical normalized PRD (11-field form). |
| [`design/prd-addenda/`](../design/prd-addenda/) | Additive review fixes 001–005 carried into the PRD gate. |
| [`design/mvp.md`](../design/mvp.md) | MVP statement — scope, principles, success criteria, binding ship gate. |
| [`design/build-out-plan.md`](../design/build-out-plan.md) | 7-phase plan, criteria→phase map, milestones, issue backlog, ADR-001–012 candidates. |
| [`design/adr/`](../design/adr/) | ADR index — ADR-001–005 **accepted** (2026-07-01, human-accepted); ADR-006/007/009 **accepted** (2026-07-03, Phase 2 — transaction model, profile system, install ownership; Codex `READY` 5/5), **ADR-008 accepted** (2026-07-03, Phase 3 — provenance & citation model; Codex `READY` 5/5), **ADR-010 accepted** (2026-07-04, Phase 4 — profile-data schema & rule/citation vocabulary), all under the autonomous-phase mandate and **ratified by Oliver 2026-07-08**; **ADR-011 accepted** (2026-07-08, Phase 5 — deterministic index maintenance; Qwen3.7 Max `READY` 4/5); ADR-012 candidate still surfaced by the build-out plan. |

## Next action

**Phase 5 issues filed; implementation ready to begin.**
ADR-011 (`accepted`) provides the architectural foundation; Phase 5
implementation issues are now filed in milestone #5. The Phase 5 plan
([phase-5-plan.md](../phase-5-plan.md)) passed adversarial review by Qwen3.7 Max
(verdict: READY, score 4/5) after incorporating 9 findings: corrected acceptance
criteria references (criterion 11 + PRD §14, not 14/15), clarified sort order
(type + bundle-relative path), added validation finding code scope
(`core-index-stale` / `core-index-unmanaged`), and added post-apply index refresh
integration with page apply.

Sequencing (two parallel tracks):

**Track 1 — Enrichment (I1):**
1. [#80](https://github.com/olivermorgan2/llm-wiki-kit/issues/80): enrichment skill adapter (reuses page plan/apply machinery)
2. [#81](https://github.com/olivermorgan2/llm-wiki-kit/issues/81): enrichment preview fixtures (core + academic-research)

**Track 2 — Index (I2-I7):**
1. [#82](https://github.com/olivermorgan2/llm-wiki-kit/issues/82): core index generation and fenced-region parsing (Option A)
2. [#83](https://github.com/olivermorgan2/llm-wiki-kit/issues/83): index CLI command with --dry-run and ADR-006 routing
3. [#84](https://github.com/olivermorgan2/llm-wiki-kit/issues/84): index stability tests (byte-idempotency, human section preservation)
4. [#85](https://github.com/olivermorgan2/llm-wiki-kit/issues/85): `core-index-stale` validation finding (ADR-004 framework)
5. [#86](https://github.com/olivermorgan2/llm-wiki-kit/issues/86): init fence scaffolding (Phase 2 backfill)

Both tracks can proceed in parallel; Track 2 should start with #82 (core index
generation) as the foundation. Track 1 can begin immediately with #80 (enrichment
skill adapter).

Phase 4 is the last closed phase. Ratification debt is 0 (Oliver ratified
ADR-006/007/008/009/010 on 2026-07-08). See the 2026-07-08 entries in
[`log.md`](log.md) for the ADR-011 drafting, the Codex → Qwen3.7 Max reviewer
substitution, the Qwen `READY` review outcome, the Phase 5 milestone #5 + issue
#76 filing, the rollback background, overlay adoption, and ratification.

Milestone #5 (`Phase 5 — Enrichment + index maintenance`, link
[milestone/5](https://github.com/olivermorgan2/llm-wiki-kit/milestone/5)) tracks
all Phase 5 work.

Phase 1 issues (all `Foundation`, repo `olivermorgan2/llm-wiki-kit`), closed via
merged PRs #7–#13:

| # | Title (abbrev.) | ADR(s) | Label |
|---|---|---|---|
| [#1](https://github.com/olivermorgan2/llm-wiki-kit/issues/1) | CLI skeleton + versioned JSON-contract spine | 001, 003 | feature |
| [#2](https://github.com/olivermorgan2/llm-wiki-kit/issues/2) | OS/arch detection + checksum verification | 002 | infra |
| [#3](https://github.com/olivermorgan2/llm-wiki-kit/issues/3) | Core-profile `validate`, OKF-vs-profile, 3 severities | 004 | feature |
| [#4](https://github.com/olivermorgan2/llm-wiki-kit/issues/4) | Broken-link detection at configured severity | 004 | feature |
| [#5](https://github.com/olivermorgan2/llm-wiki-kit/issues/5) | Safe filesystem layer (per-file atomic, symlink/traversal) | 005 | security |
| [#6](https://github.com/olivermorgan2/llm-wiki-kit/issues/6) | Core-profile fixtures + traversal/symlink testdata | 004, 005 | infra |

The [`notes/phase-1-first-issue-spec.md`](../notes/phase-1-first-issue-spec.md)
pre-issue spec is now realized as issue #1. **No GitHub Project board was
created** — the authenticated `gh` token lacks the `project` OAuth scope
(ADR-012's board step skipped; issues stand alone).

Explicit carry-forwards to hold in the first issue and beyond:

- **ADR-003 numeric exit-code table** must be published (stable code→meaning) as
  part of Phase 1 before Phase 1 closes — deferred by ADR-003, not reopened.
- **ADR-006** (staged mutation / inspect-plan-apply cross-file transaction model)
  — was the next-needed ADR after 001–005; now **drafted + accepted** (#14,
  2026-07-03) and its **transaction half implemented** in Phase 2 (#16/PR #23).
  The `inspect/plan/apply` UX half is consumed by Phase 3.
- **ADR-002 does not authorize** full installer / asset-ownership work — install/
  upgrade/uninstall ownership and release signing/provenance remain **ADR-009**.

ADR-006–012 remain to be drafted from the "Decisions needing ADRs" list at the
end of [`design/build-out-plan.md`](../design/build-out-plan.md). Owner decisions
that could revise MVP scope remain tracked as assumption-locked items in
[`open-questions.md`](open-questions.md) — notably the working-name lock (QB1).
