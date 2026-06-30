---
name: prd-normalizer
description: Normalize a standard or custom PRD into a single canonical form (design/prd-normalized.md). Use when preparing a PRD before MVP scoping — after /idea-to-prd or against a hand-authored PRD.
permission-category: 1  # substitutable — local doc rewrite, per workflow-guide §7
inputs:
  - name: "prd"
    required: true
    description: "A PRD-like document (standard or custom shape)"
outputs:
  - artefact: "design/prd-normalized.md"
    description: "Canonical 11-field normalized PRD"
next:
  - skill: prd-to-mvp
    when: "scoping the MVP"
  - skill: adr-writer
    when: "decisions are already clear"
---

# prd-normalizer

Produce a canonical **Normalized PRD** at `design/prd-normalized.md`
from either a standard-shaped PRD or the user's custom planning notes.
Downstream skills (`prd-to-mvp`, `adr-writer`) read this artifact only,
so they do not have to branch on input shape.

## When to use this skill

Use when the user has *any* PRD-like input:

- the `design/prd.md` produced by `idea-to-prd` (Issue #5),
- a user-written PRD in a conventional shape,
- a custom document — notes, a product brief, a spec, mixed prose and
  bullets.

If the user does not have any input at all, run `idea-to-prd` first.
This skill exists to cover the "standard PRD" and "custom PRD" paths
from ADR-003.

## What this skill does not do

- Does not scope to in/out of scope — that is `prd-to-mvp` (Issue #7).
- Does not draft ADRs — that is `adr-writer` (Issue #7).
- Does not rewrite the user's prose for tone or polish.
- Does not invent content. Missing fields are marked `[TBD]`, not
  filled in by guess.
- Never overwrites the user's original input file.

## Inputs

- **Required:** a PRD-like document in any shape.
- **Optional:** a product name if the input does not contain one.

## Output

- **File:** `design/prd-normalized.md` in the target project.
- **Shape:** the 11 fields below. Every field is always present; fields
  with no source content and no user answer are marked `[TBD]`.

## Normalized PRD structure

1. **Product name** — short, memorable.
2. **One-line description** — "a thing that does X for Y so they can Z."
3. **Problem** — one or two sentences.
4. **Target users** — primary required; secondary optional.
5. **Goal** — two or three sentences of intended outcome.
6. **User stories / scenarios** — 2–5 concrete "user does X so that Y"
   stories.
7. **Core capabilities** — bulleted, un-prioritized.
8. **Non-goals** — product-level non-goals. Do not promote these to
   MVP-scope decisions; that is `prd-to-mvp`'s job.
9. **Constraints and preferences** — technical constraints, platform
   preferences, known stack decisions, timing constraints. This is the
   primary hint source for `adr-writer`.
10. **Success signals** — how the user will know the product is working.
11. **Open questions** — what the user does not yet know. Non-empty is
    fine and expected.

## Intake protocol — standard PRD path

Use this path when the input is recognisable as an 8-section-shaped PRD
(e.g. output of `idea-to-prd`) or a user PRD with conventional section
names.

1. Map each input section onto the matching canonical field.
2. Fill the two identity fields (product name, one-line description) if
   missing — usually one or two questions.
3. Scan non-goals for items that are really MVP-scope decisions ("not
   in the first release"). Leave them here only if they are
   product-level; otherwise mark them for `prd-to-mvp` and drop from
   non-goals.
4. Extract a "constraints and preferences" section if the input's prose
   mentions platform, stack, or timing hints; otherwise mark `[TBD]`.
5. Run the self-check and write the file.

If the input was authored from
[`templates/prd-template.md`](../../templates/prd-template.md) — whether
filled by hand or via an external LLM — its 11 sections already match
this skill's canonical fields one-to-one. Step 1 is a
pass-through; step 2 may not be needed; the remaining steps confirm
content and write the file. This is the fastest path through the
normaliser.

## Intake protocol — custom PRD path

Use this path when the input does not match a conventional PRD shape.

1. Read the whole document.
2. Identify sections by **semantic meaning**, not heading name.
   ("Who it's for" = target users. "What it does" = capabilities.
   "What it's not" = non-goals.)
3. Map every recognisable piece of content onto one of the 11 fields.
4. Mark unfilled fields `[TBD]`.
5. Batch the gaps into ≤ 5 questions per turn. Skip any gap with an
   obvious default from the input. Repeat until the user has been
   asked about every `[TBD]` once.
6. Run the self-check and write the file. Preserve the user's original
   input at its original path.

## Fallback behaviour

`[TBD]` is a first-class output marker meaning *"the user did not say,
and the normalizer did not invent."* Downstream skills must handle it.

The self-check treats fields as either **hard-required** (never `[TBD]`
in the final output) or **soft** (`[TBD]` is acceptable):

- Hard-required: product name, problem, primary user, ≥ 1 core
  capability, ≥ 1 user story.
- Soft: one-line description, secondary user, goal, non-goals,
  constraints and preferences, success signals, open questions.

If a hard-required field has no source and the user declines to answer,
stop and tell the user — do not write the file.

## Self-check before finishing

Do not write `design/prd-normalized.md` until all six hold:

- [ ] Product name is present and not `[TBD]`.
- [ ] Problem is present and not `[TBD]`.
- [ ] A primary user is named and not `[TBD]`.
- [ ] At least one core capability is listed.
- [ ] At least one user story is present.
- [ ] Every one of the 11 canonical fields appears in the output (even
  if the content is `[TBD]`).

## Handoff to the next skills

Both `prd-to-mvp` and `adr-writer` read `design/prd-normalized.md` only.
Neither re-parses the user's raw input. `adr-writer` pays particular
attention to "constraints and preferences".

See [`example.md`](example.md) for two worked transformations — one
from a standard PRD, one from a custom document.
