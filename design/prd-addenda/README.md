# PRD addenda

This directory holds **additive, non-destructive** revisions and addenda
to the normalized PRD ([`../prd-normalized.md`](../prd-normalized.md)).

## Why addenda instead of in-place edits

The Codex adversarial PRD review (and any later review loops) record their
accepted fixes here rather than rewriting `design/prd-normalized.md` in
place. This keeps the normalized PRD's provenance clean and makes every
revision auditable as its own change. The source PRD
(`design/prd.md`) is never edited.

Exception: if a kit command itself must regenerate `prd-normalized.md`
(e.g. re-running `prd-normalizer`), that regeneration is allowed — it is
the tool's documented output, not a manual in-place edit.

## Convention

- One file per addendum: `NNN-short-title.md` (zero-padded, incrementing —
  `001-...`, `002-...`).
- Each addendum states: the source finding (e.g. Codex review item), the
  decision (accept / reject / defer), and the precise change to the
  normalized PRD's field(s).
- Cross-reference the matching entry in
  [`../../knowledge/log.md`](../../knowledge/log.md); open
  items spawned by a finding go to
  [`../../knowledge/open-questions.md`](../../knowledge/open-questions.md)
  and risks to [`../../knowledge/risks.md`](../../knowledge/risks.md).

## Index

The Codex adversarial PRD review (2026-06-30, verdict `NEEDS_REVISION`,
archived at [`../../knowledge/reviews/2026-06-30-codex-prd-review.md`](../../knowledge/reviews/2026-06-30-codex-prd-review.md))
produced these accepted addenda:

- [`001-mvp-acceptance-criteria.md`](001-mvp-acceptance-criteria.md) —
  carry §17 acceptance criteria into the canonical planning input; label
  success signals as release-gate vs measurement-only.
- [`002-mvp-planning-assumptions.md`](002-mvp-planning-assumptions.md) —
  assumption-lock CI scope, Claude Code version floor, binary selection,
  and JSON-contract compatibility (Q3/Q6/Q7/Q8).
- [`003-academic-research-profile-contract.md`](003-academic-research-profile-contract.md) —
  minimum MVP research-profile contract: page types, fields, sections,
  per-type fixtures (Q4).
- [`004-mvp-slice-order-and-fixture-plan.md`](004-mvp-slice-order-and-fixture-plan.md) —
  slice order separating the must-pass first-release spine from follow-on
  hardening inside MVP, plus the fixture plan.
- [`005-custom-profile-boundary.md`](005-custom-profile-boundary.md) —
  MVP custom profiles are local-file only; registry/trust is Phase 3 (Q5).
