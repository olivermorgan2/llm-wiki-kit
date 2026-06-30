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
  [`../../knowledge/decisions.md`](../../knowledge/decisions.md); open
  items spawned by a finding go to
  [`../../knowledge/open-questions.md`](../../knowledge/open-questions.md)
  and risks to [`../../knowledge/risks.md`](../../knowledge/risks.md).

No addenda exist yet — the Codex review has not been run.
