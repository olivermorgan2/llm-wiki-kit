---
name: idea-to-prd
description: Turn a rough idea into a lightweight PRD draft suitable for prd-normalizer. Use when starting from a rough idea before formal PRD intake.
permission-category: 1  # substitutable — drafts a PRD locally, per workflow-guide §7
inputs:
  - name: "idea"
    required: true
    description: "A rough idea in any form (paragraph, notes, conversation)"
outputs:
  - artefact: "design/prd.md"
    description: "Lightweight PRD draft"
next:
  - skill: prd-normalizer
    when: "the PRD draft is ready"
---

# idea-to-prd

Produce a lightweight PRD draft at `design/prd.md` from whatever rough
idea material the user has. The draft is intentionally minimal — just
enough to hand off to `prd-normalizer` (Issue #6).

## When to use this skill

Use when the user is starting a project with only a rough idea and no
PRD. If the user already has a standard PRD, skip this skill and go
straight to `prd-normalizer`. If the user has custom planning notes in
their own format, also skip this skill — `prd-normalizer`'s custom path
handles that case directly.

If the user wants to draft a PRD offline before involving Claude Code —
by hand, or by giving an external LLM a canonical shape to fill — point
them at [`templates/prd-template.md`](../../templates/prd-template.md).
That template's 11 sections match `prd-normalizer`'s canonical fields
one-to-one, so a faithfully-filled file flows through
`prd-normalizer` as a pass-through. Use this skill only when there is
no draft PRD at all.

This skill exists to cover the "no PRD yet" path from
ADR-003.

## What this skill does not do

- Does not normalize into the internal canonical planning format — that
  is `prd-normalizer` (Issue #6).
- Does not scope to an MVP — that is `prd-to-mvp` (Issue #7).
- Does not draft ADRs — that is `adr-writer` (Issue #7).
- Does not aim for PRD completeness. Terse is correct.

## Inputs

- **Required:** a rough idea in any form — a paragraph, bullets, a
  problem statement, scratch notes.
- **Optional:** anything the user already knows about target users,
  constraints, or references.

## Output

- **File:** `design/prd.md` in the target project.
- **Shape:** the eight sections below. This shape intentionally mirrors
  a conventional PRD so that `prd-normalizer`'s standard-PRD path
  consumes it without special-casing.

## PRD draft structure

1. **Problem** — one or two sentences: what hurts, for whom.
2. **Target users** — who this is for. Primary user is required;
   secondary is optional.
3. **Goal** — two or three sentences on what success looks like.
4. **User stories / scenarios** — 2–5 concrete "user does X so that Y"
   stories. No epics, no acceptance criteria yet.
5. **Core capabilities** — the handful of things the product must be
   able to do. Bulleted. No prioritization yet.
6. **Non-goals** — things explicitly out of scope. Include anything the
   user commonly assumes is in scope but is not.
7. **Success signals** — how the user will know the first release is
   working. Keep concrete; avoid vanity metrics.
8. **Open questions** — what the user does not yet know. It is fine —
   and expected — for this section to be non-empty.

## Interview protocol

1. Read whatever rough-idea material the user has provided.
2. Draft a first-pass PRD filling every section; mark unknowns as
   `[TBD — ask user]`.
3. Batch the TBDs into at most five questions per turn. Skip any TBD
   where a reasonable default is obvious from the material.
4. Update the draft from the answers. Repeat steps 3–4 until no TBDs
   remain or the user says "good enough".
5. Run the self-check below.
6. Write `design/prd.md` and tell the user the next step is
   `prd-normalizer`.

## Self-check before finishing

Do not write the file until all five hold:

- [ ] The problem is named in one or two sentences.
- [ ] A primary user is named.
- [ ] At least one concrete user story is present.
- [ ] At least one non-goal is listed.
- [ ] No `[TBD — …]` placeholders remain.

If any fail, keep asking.

## Handoff to the next skill

The output of this skill is the "standard PRD" input to
`prd-normalizer`. Do not try to canonicalize, score, or scope the draft
here — that is the next skill's job, and duplicating it would make the
two skills step on each other.

See [`example.md`](example.md) for a worked input → output pair.
