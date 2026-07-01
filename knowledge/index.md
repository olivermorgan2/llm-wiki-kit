# knowledge/ — index

Front door to the `llm-wiki-kit` knowledge layer. Conventions live in
[`SCHEMA.md`](SCHEMA.md); this file is the live map.

## Current phase

**ADR-001–005 (`proposed`) — Codex `NEEDS_REVISION` findings applied; awaiting
re-review.** Bootstrapped 2026-06-30 from `claude-workflow-kit` v5.0.1.
Source PRD normalized; adversarial Codex review (`NEEDS_REVISION`) landed as PRD
addenda 001–005; `/prd-to-mvp` produced [`design/mvp.md`](../design/mvp.md) and
[`design/build-out-plan.md`](../design/build-out-plan.md) (7 phases / Slices
0–6 + cross-platform gate, all 21 acceptance criteria mapped). The five Phase-1
prerequisite ADRs ([`design/adr/`](../design/adr/) 001–005: toolchain, binary
selection, JSON contract, validation model, safe-FS gate) are `proposed` and
self-validated; the 2026-07-01 Codex gate's four blocking revision targets have
now been **applied** (see the 2026-07-01 revision entry in [`log.md`](log.md)).
Acceptance remains a human act after the Codex **re-review** returns `READY`. See
[`log.md`](log.md) and
[`reviews/2026-07-01-codex-adr-001-005-review.md`](reviews/2026-07-01-codex-adr-001-005-review.md).

## Knowledge layer

| File | What you'll find |
|---|---|
| [`SCHEMA.md`](SCHEMA.md) | Layer conventions, file roles, entry formats. |
| [`project-brief.md`](project-brief.md) | Distilled product definition and architectural spine. |
| [`risks.md`](risks.md) | Live risk register (product, cross-cutting, process, review). |
| [`open-questions.md`](open-questions.md) | Q1–Q8 + bootstrap questions, with status. |
| [`log.md`](log.md) | Chronological decision & review log. |
| [`reviews/`](reviews/) | Verbatim review archives — [`2026-06-30-codex-prd-review.md`](reviews/2026-06-30-codex-prd-review.md), [`2026-07-01-codex-adr-001-005-review.md`](reviews/2026-07-01-codex-adr-001-005-review.md). |
| [`sources/`](sources/) | Verbatim sources — [`prd-original.md`](sources/prd-original.md). |

## Pointers into `design/` (authoritative, tool-maintained)

| Artifact | Status |
|---|---|
| [`design/prd.md`](../design/prd.md) | Source PRD, verbatim (also archived at `sources/prd-original.md`). |
| [`design/prd-normalized.md`](../design/prd-normalized.md) | Canonical normalized PRD (11-field form). |
| [`design/prd-addenda/`](../design/prd-addenda/) | Additive review fixes 001–005 carried into the PRD gate. |
| [`design/mvp.md`](../design/mvp.md) | MVP statement — scope, principles, success criteria, binding ship gate. |
| [`design/build-out-plan.md`](../design/build-out-plan.md) | 7-phase plan, criteria→phase map, milestones, issue backlog, ADR-001–012 candidates. |
| [`design/adr/`](../design/adr/) | ADR index — ADR-001–005 drafted (`proposed`); ADR-006–012 candidates still surfaced by the build-out plan. |

## Next action

**Rerun the Codex ADR/milestone review gate** over the revised ADR-001–005. The
bounded revision pass is complete (2026-07-01): ADR-001 comment-preservation
traceability, ADR-002 checksum trust root / residual supply-chain risk, ADR-005
per-file vs cross-file atomicity (transaction model deferred to ADR-006), and
ADR-004 baseline vs hard-validation errors are all addressed, plus ADR-003
`--json`/exit-code refinements. Do **not** flip the five ADRs to `accepted`, and
do not close Q2/Q7/Q8, until the re-review returns `READY` and Oliver accepts. ADR-006–012 remain to be
drafted from the "Decisions needing ADRs" list at the end of
[`design/build-out-plan.md`](../design/build-out-plan.md). Owner decisions that
could revise MVP scope remain tracked as assumption-locked items in
[`open-questions.md`](open-questions.md) — notably the working-name lock (QB1).
