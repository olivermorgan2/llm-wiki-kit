---
name: prd-to-mvp
description: Scope a normalized PRD into an MVP statement (design/mvp.md) and a build-out plan (design/build-out-plan.md). Use when scoping the next milestone after the PRD has been normalized.
permission-category: 1  # substitutable — local doc creation (design/mvp.md, design/build-out-plan.md), per workflow-guide §7
inputs:
  - name: "--granularity"
    required: false
    description: "Decomposition depth: coarse | standard | fine"
outputs:
  - artefact: "design/mvp.md"
    description: "Scoped MVP capabilities and principles"
  - artefact: "design/build-out-plan.md"
    description: "Phased plan plus the 'Decisions needing ADRs' list"
next:
  - skill: adr-writer
    when: "the surfaced decisions need ADRs"
---

# prd-to-mvp

Take `design/prd-normalized.md` and produce two artifacts: a scoped
**MVP statement** at `design/mvp.md` and a sequenced **build-out plan**
at `design/build-out-plan.md`. Both are rendered from existing
templates in this kit.

## When to use this skill

Use after `prd-normalizer` (Issue #6) has produced
`design/prd-normalized.md`. If that file does not exist, run
`prd-normalizer` first.

This skill is the entry point from product planning into implementation
planning. Its outputs feed `adr-writer` (this issue) for architectural
decisions and `issue-planner` (later issue) for the GitHub backlog.

## What this skill does not do

- Does not draft ADRs — that is `adr-writer`. This skill surfaces a
  list of decisions that need ADRs, but does not write them.
- Does not execute the build-out plan — that is the executor skill.
- Does not revise the PRD in place. If the PRD is wrong, re-run
  `prd-normalizer`.
- Does not invent capabilities not present in the normalized PRD. If
  scope feels thin, that is information about the PRD, not licence to
  pad.

## Inputs

- **Required:** `design/prd-normalized.md`.
- **Optional:** user preference on project size — weekend project,
  one-month, multi-month — to guide how aggressive the scope cut is.
- **Optional flag:** `--granularity={coarse|standard|fine}` — phase-
  count target band.
  `coarse` aims for 1–3 phases, `standard` (default) for 5–8, `fine`
  for 8–12. Bands are targets, not hard caps; the skill picks the
  actual count and justifies it inline. Precedence (highest first):
  explicit flag > `**Granularity:**` line in an existing
  `design/build-out-plan.md` > default `standard`.

## Outputs

- **`design/mvp.md`** — rendered from
  [`templates/mvp-template.md`](../../templates/mvp-template.md).
- **`design/build-out-plan.md`** — rendered from
  [`templates/build-out-plan-template.md`](../../templates/build-out-plan-template.md).

Both files are produced from a single scoping conversation. The two
artifacts are tightly coupled: phase 1 of the build-out plan must
cover every in-scope capability of the MVP.

## Scoping protocol

1. **Resolve granularity tier.** If `--granularity=<tier>` was passed,
   validate the value is one of `coarse|standard|fine` and use it.
   Else, if `design/build-out-plan.md` already exists and contains a
   `**Granularity:** <tier>` line with a valid value, read it. Else
   default to `standard`. Reject invalid tiers with a one-line error
   naming the three valid values and stop. The resolved tier sets
   the target phase-count band for step 6: `coarse` 1–3, `standard`
   5–8, `fine` 8–12.
2. Read `design/prd-normalized.md` end to end.
3. Propose an in-scope / out-of-scope split for every core capability
   and user story. Default to aggressive cuts; over-scoping is the
   more common failure.
4. Ask the user to confirm or adjust the split — one batched turn
   (≤ 5 questions). Include the project-size question if the user
   has not stated a preference.
5. Render `design/mvp.md` from `templates/mvp-template.md`. Derive
   3–5 product principles from the normalized PRD's constraints,
   non-goals, and goal — principles that will resolve future scope
   arguments.
6. Draft `design/build-out-plan.md` from
   `templates/build-out-plan-template.md`. Write the resolved tier
   into the `**Granularity:**` line under the header. Sequence into
   N phases per the resolved band — `coarse` 1–3, `standard` 5–8
   (default), `fine` 8–12. The band is a target, not a hard cap;
   pick the actual count for *this* project and include a one-line
   inline justification in the rendered plan (e.g. *"3 phases
   chosen — within the standard band — because the GitHub-integration
   work cleanly splits from the storage and ingestion layers"*).
   Each phase has: goal, scope bullets, ADR dependencies,
   deliverables, exit criterion. For small projects where the whole
   MVP fits in one delivery cut (typically `coarse` with 1 phase),
   emit a single `## Phase 1` block — every downstream skill treats
   a single-phase plan identically to a flat plan, so this is the
   back-compat path. Ask the user once whether the phase count and
   shape feel right.
7. Surface a **"Decisions needing ADRs"** list as the final section
   of `build-out-plan.md` (or a short separate note). Each item is one
   architectural question with a one-line context pointer. This list
   is the input to `adr-writer`.

## Single-phase fallback

A weekend or one-week project usually does not benefit from
multi-phase decomposition. In that case, emit one `## Phase 1`
block whose scope covers the entire MVP and exit criterion is "MVP
ships." Downstream skills (`issue-planner`, `workflow-docs`,
`/release`) treat a single-phase plan as one implicit phase and
behave identically to the pre-ADR-032 flat-plan path. The decision
is the user's; surface the option explicitly during scoping rather
than defaulting silently.

## How `[TBD]` fields are handled

The normalized PRD may contain `[TBD]` markers (see `prd-normalizer`).

- If a hard-required field is `[TBD]` (product name, problem, primary
  user, ≥1 capability, ≥1 story), stop and send the user back to
  `prd-normalizer`. MVP scoping cannot proceed without these.
- If a soft field is `[TBD]` (e.g. constraints and preferences), note
  it as an open item in `mvp.md`'s acceptance section and continue.

## Self-check before finishing

Do not write either file until all six hold:

- [ ] Every core capability from the normalized PRD is classified
  (in scope or out of scope). None left ambiguous.
- [ ] The MVP success criteria are end-to-end user outcomes, not
  implementation tasks.
- [ ] The "Out of scope" list combines (a) product-level non-goals from
  the normalized PRD and (b) capabilities this skill deferred. The two
  sources are not conflated — readers can see why each item is out.
- [ ] Build-out-plan phase 1's exit criteria collectively cover every
  in-scope capability.
- [ ] The `**Granularity:**` line in the rendered `build-out-plan.md`
  is set to one of `coarse | standard | fine` — never left as the
  `{{GRANULARITY}}` placeholder — and the actual phase count is
  inside the band.
- [ ] No `{{...}}` template placeholders remain in either rendered
  file.

## Handoff to `adr-writer`

The "Decisions needing ADRs" list at the end of the build-out plan is
the direct input to `adr-writer`. Each item should be a single
architectural question (e.g. *"GPX parsing: browser-side vs.
server-side"*), not a feature description. `adr-writer` accepts a
batch of these in one run.

See [`example.md`](example.md) for a worked PRD → MVP + build-out plan
walkthrough.
