# knowledge/ — index

Front door to the `llm-wiki-kit` knowledge layer. Conventions live in
[`SCHEMA.md`](SCHEMA.md); this file is the live map.

## Current phase

**PRD gate complete — READY for `/prd-to-mvp`.** Bootstrapped 2026-06-30
from `claude-workflow-kit` v5.0.1. Source PRD normalized; adversarial Codex
review ran (`NEEDS_REVISION`) and its findings landed as PRD addenda
001–005. `/prd-to-mvp` has **not** been started. See [`log.md`](log.md) for
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
| [`design/adr/`](../design/adr/) | ADR index — no project ADRs yet (kit ADRs are documented under `docs/`). |
| `design/mvp.md`, `design/build-out-plan.md` | Not yet created — produced by `/prd-to-mvp`. |

## Next action

Run `/prd-to-mvp` against `design/prd-normalized.md` **plus** addenda
001–005 (the addenda carry the acceptance criteria, planning assumptions,
domain contract, slice order, and scope boundary the review required).
Owner decisions that could revise MVP scope are tracked as
assumption-locked items in [`open-questions.md`](open-questions.md).
