# knowledge/ — index

Front door to the `llm-wiki-kit` knowledge layer. Conventions live in
[`SCHEMA.md`](SCHEMA.md); this file is the live map.

## Current phase

**ADR-001–005 human-accepted (2026-07-01); Phase 1 (engine + contract spine)
backlog filed as GitHub issues #1–#6, first prompt prepared — implementation
next.** Bootstrapped 2026-06-30 from `claude-workflow-kit` v5.0.1.
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
| [`reviews/`](reviews/) | Verbatim review archives — [`2026-06-30-codex-prd-review.md`](reviews/2026-06-30-codex-prd-review.md), [`2026-07-01-codex-adr-001-005-review.md`](reviews/2026-07-01-codex-adr-001-005-review.md), [`2026-07-01-codex-adr-001-005-rereview.md`](reviews/2026-07-01-codex-adr-001-005-rereview.md). |
| [`sources/`](sources/) | Verbatim sources — [`prd-original.md`](sources/prd-original.md). |

## Pointers into `design/` (authoritative, tool-maintained)

| Artifact | Status |
|---|---|
| [`design/prd.md`](../design/prd.md) | Source PRD, verbatim (also archived at `sources/prd-original.md`). |
| [`design/prd-normalized.md`](../design/prd-normalized.md) | Canonical normalized PRD (11-field form). |
| [`design/prd-addenda/`](../design/prd-addenda/) | Additive review fixes 001–005 carried into the PRD gate. |
| [`design/mvp.md`](../design/mvp.md) | MVP statement — scope, principles, success criteria, binding ship gate. |
| [`design/build-out-plan.md`](../design/build-out-plan.md) | 7-phase plan, criteria→phase map, milestones, issue backlog, ADR-001–012 candidates. |
| [`design/adr/`](../design/adr/) | ADR index — ADR-001–005 **accepted** (2026-07-01); ADR-006–012 candidates still surfaced by the build-out plan (ADR-006 is the next-needed draft — cross-file transaction model). |

## Next action

**Phase 1 backlog filed as GitHub issues #1–#6 (2026-07-01)** under the
`Foundation` milestone; the first issue's prompt is prepared at
[`../prompts/issue-001-cli-contract-spine.md`](../prompts/issue-001-cli-contract-spine.md).
The next workflow step is to **implement issue #1** via `/claude-issue-executor`
(plan-first, one issue per session), and in parallel **draft ADR-006** before
any cross-file mutation work begins.

Phase 1 issues (all `Foundation`, repo `olivermorgan2/llm-wiki-kit`):

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
  must be drafted **before any cross-file mutation/transaction implementation**;
  it is the next-needed ADR after 001–005.
- **ADR-002 does not authorize** full installer / asset-ownership work — install/
  upgrade/uninstall ownership and release signing/provenance remain **ADR-009**.

ADR-006–012 remain to be drafted from the "Decisions needing ADRs" list at the
end of [`design/build-out-plan.md`](../design/build-out-plan.md). Owner decisions
that could revise MVP scope remain tracked as assumption-locked items in
[`open-questions.md`](open-questions.md) — notably the working-name lock (QB1).
