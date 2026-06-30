# knowledge/ — index

Front door to the `llm-wiki-kit` knowledge layer. Conventions live in
[`SCHEMA.md`](SCHEMA.md); this file is the live map.

## Current phase

**MVP scoped — READY for `/adr-writer`.** Bootstrapped 2026-06-30 from
`claude-workflow-kit` v5.0.1. Source PRD normalized; adversarial Codex review
(`NEEDS_REVISION`) landed as PRD addenda 001–005; `/prd-to-mvp` then produced
[`design/mvp.md`](../design/mvp.md) and
[`design/build-out-plan.md`](../design/build-out-plan.md) (7 phases / Slices
0–6 + cross-platform gate, all 21 acceptance criteria mapped). Working name
and Go/YAML toolchain assumption-locked for MVP. See [`log.md`](log.md) for
the full chain.

## Knowledge layer

| File | What you'll find |
|---|---|
| [`SCHEMA.md`](SCHEMA.md) | Layer conventions, file roles, entry formats. |
| [`project-brief.md`](project-brief.md) | Distilled product definition and architectural spine. |
| [`risks.md`](risks.md) | Live risk register (product, cross-cutting, process, review). |
| [`open-questions.md`](open-questions.md) | Q1–Q8 + bootstrap questions, with status. |
| [`log.md`](log.md) | Chronological decision & review log. |
| [`reviews/`](reviews/) | Verbatim review archives — [`2026-06-30-codex-prd-review.md`](reviews/2026-06-30-codex-prd-review.md). |
| [`sources/`](sources/) | Verbatim sources — [`prd-original.md`](sources/prd-original.md). |

## Pointers into `design/` (authoritative, tool-maintained)

| Artifact | Status |
|---|---|
| [`design/prd.md`](../design/prd.md) | Source PRD, verbatim (also archived at `sources/prd-original.md`). |
| [`design/prd-normalized.md`](../design/prd-normalized.md) | Canonical normalized PRD (11-field form). |
| [`design/prd-addenda/`](../design/prd-addenda/) | Additive review fixes 001–005 carried into the PRD gate. |
| [`design/mvp.md`](../design/mvp.md) | MVP statement — scope, principles, success criteria, binding ship gate. |
| [`design/build-out-plan.md`](../design/build-out-plan.md) | 7-phase plan, criteria→phase map, milestones, issue backlog, ADR-001–012 candidates. |
| [`design/adr/`](../design/adr/) | ADR index — no project ADRs yet; ADR-001–012 candidates surfaced by the build-out plan, drafted next by `/adr-writer`. |

## Next action

Run `/adr-writer` against the "Decisions needing ADRs" list at the end of
[`design/build-out-plan.md`](../design/build-out-plan.md) (ADR-001–012
candidates; ADR-001–005 are Phase-1 prerequisites and should be drafted
first). Owner decisions that could revise MVP scope remain tracked as
assumption-locked items in [`open-questions.md`](open-questions.md) — notably
the working-name lock (QB1) and the Go/YAML toolchain lock (Q2, = ADR-001).
