---
name: adr-writer
description: Draft one or more ADRs from architectural decision topics, using the repo's ADR template, written to design/adr/adr-NNN-short-title.md. Use when committing a settled architectural choice; for pre-decision context use /planning or /clarify first.
permission-category: 1  # substitutable — drafts ADRs locally; user accepts manually, per workflow-guide §7
inputs:
  - name: "decision-topics"
    required: true
    description: "One or more architectural decision topics (often the 'Decisions needing ADRs' list from prd-to-mvp)"
  - name: "--skip-check"
    required: false
    description: "Opt out of the /check-plan pre-write gate"
outputs:
  - artefact: "design/adr/adr-NNN-short-title.md"
    description: "One ADR per decision, status: proposed"
  - artefact: "design/adr/README.md"
    description: "ADR index re-synced via bin/sync-adr-index"
next: []
---

# adr-writer

Draft ADRs from a list of architectural decision topics — typically
the "Decisions needing ADRs" list surfaced by `prd-to-mvp`, but also
usable for any decision that comes up later. Output: one file per
decision in `design/adr/`, each rendered from
[`templates/adr-template.md`](../../templates/adr-template.md), with
status `proposed`. Acceptance is a human act and is not done by this
skill.

## When to use this skill

- After `prd-to-mvp` has produced its "Decisions needing ADRs" list,
  to draft them in one batch.
- During implementation, when a new architectural decision needs to
  be recorded.

If the decision has already been made informally and you only need
the ADR to document it, this skill still works — the options analysis
just summarises the alternatives that were considered.

## What this skill does not do

- Does not perform MVP scoping — that is `prd-to-mvp`.
- Does not accept the ADR on the user's behalf. Status starts as
  `proposed`. Promotion to `accepted` is a human review act.
- Does not rewrite or supersede existing ADRs in place. To change
  an accepted decision, draft a new ADR and mark the old one
  superseded.
- Does not invent constraints. Constraints in the Context section
  come from `design/prd-normalized.md`, `design/mvp.md`, or the user.

## Inputs

- **Required:** one or more decision topics. Each topic is a short
  problem statement (e.g. *"GPX parsing: browser-side vs.
  server-side"*).
- **Recommended:** `design/prd-normalized.md` for the constraints
  context.
- **Recommended:** `design/mvp.md` for product principles and MVP
  scope context — useful when an option's pros/cons depend on what
  the MVP includes.
- **Optional:** `design/planning.md`
  for requirements decomposition, risks, assumptions, and the
  "Decisions needing ADRs" list. When present, use it as the input
  batch source and reference its requirement IDs (`R1`, `R2`, …) in
  ADR Context sections where they apply. When absent, behaviour is
  unchanged — small projects skip this artefact and the skill works
  from PRD + MVP alone.
- **Optional:** `design/decisions.md`
  for informal-but-settled context surfaced by `/clarify` before ADR
  drafting. When present, read it for additional Context-section
  background and to avoid re-deciding questions already settled
  there. If a topic in the input batch already appears as a decision,
  prefer drafting an ADR that supersedes it (per the
  graduate-to-ADR criterion in `skills/clarify/SKILL.md`) over
  re-litigating. When absent, behaviour is unchanged.
- **Optional flag:** `--skip-check` — opt out of the `/check-plan`
  pre-write gate. Default is on (gate runs); the flag is documented as opt-out for
  rapid iteration on known-good drafts only. When set, the skill
  writes each ADR despite any deterministic-criteria failures and
  appends a one-line breadcrumb to the commit message —
  `chore: ADR drafted with --skip-check per ADR-034` — so the
  bypass is auditable.

## Output

- **Files:** `design/adr/adr-NNN-short-title.md`, one per decision
  topic, rendered from `templates/adr-template.md`.
- **Numbering:** sequential, never reused. Do not count files by
  hand — call `bin/adr-alloc --title "<topic>" --format json` and read
  `outputs.number`, `outputs.date`, `outputs.slug`, `outputs.filename`,
  and `outputs.indexRowStub`. For a batch, the first call gives the
  starting number; assign that number to the first topic and increment
  by one for each subsequent topic in the confirmed order (the files are
  not yet on disk), reusing each topic's own `--title` call for its slug
  and index-row stub.
- **Status:** `proposed` for every ADR this skill produces.
- **Date:** today's date in `YYYY-MM-DD`.

## Alignment with existing ADRs

The skill follows the conventions of the kit's own ADR-001 through
ADR-006:

- Top-level heading `# ADR-NNN: Title`.
- `**Status:**` and `**Date:**` lines immediately after the heading.
- Optional `**Phase:** N` line when the ADR belongs to a specific
  build-out-plan phase. Omit the line on cross-cutting
  ADRs and on single-phase projects.
- Four top-level sections: `## Context`, `## Options considered`,
  `## Decision`, `## Consequences`.
- Options as `### Option A: Name` etc., each with `- Pros:` and
  `- Cons:` bullets.
- `## Decision` is one short paragraph naming the chosen option and
  why.
- `## Consequences` is a bulleted list — what becomes easier, what
  becomes harder, what the team has to maintain, what is deferred.

If a future ADR amends the template, this skill follows the template.

## Drafting protocol — single decision

Repeat for every topic in the input batch:

1. Restate the decision topic as a one-line problem statement, and
   call `bin/adr-alloc --title "<topic>" --format json` to allocate
   the number, date, slug, and filename (see **Numbering** above).
2. Pull relevant constraints from `design/prd-normalized.md`'s
   "Constraints and preferences" field. Pull product principles from
   `design/mvp.md` if it exists. These populate `## Context`.
3. Propose 2 or 3 options with Pros/Cons. Use 3 only when there is a
   genuine third alternative; otherwise 2.
4. Write `## Decision` naming the chosen option. If the choice is
   non-obvious, ask the user once.
5. Draft `## Consequences` as four bullets covering what becomes
   easier / harder / has to be maintained / is deferred. If a bullet
   has nothing to add, write "None new" rather than removing it.
6. **Phase tagging (optional).** If the input topic
   names a phase, or if `design/build-out-plan.md` makes the ADR's
   phase obvious, write `**Phase:** N` immediately after the
   `**Date:**` line. Omit the line for cross-cutting ADRs and for
   single-phase projects. The line is opt-in — when set,
   `bin/sync-adr-index` surfaces a Phase column in
   `design/adr/README.md`.
7. **Pre-write check.**
   Unless `--skip-check` was passed, pipe the in-memory rendered
   ADR into the kit's programmatic check-plan surface:
   ```
   <rendered-adr> | bin/check-plan --criteria-set adr --input - --format json
   ```
   Parse the JSON envelope. On exit 0, proceed to disk write. On
   exit 1, surface the failures (each citing its criterion ID —
   e.g. `ADR-C3` — and the `remediation` field from the JSON) to
   the user, ask how to revise, apply the revision in memory, and
   re-invoke `bin/check-plan` for round 2. After 3 failed rounds,
   yield: surface the remaining failures and stop without writing
   the ADR. Warnings (`status: warn`) are surfaced but do not
   block. `--skip-check` short-circuits the gate for the current
   invocation only and writes despite failures, leaving a
   breadcrumb in the commit message.

   The programmatic surface is what skills with
   chain points invoke — slash-commands aren't invokable from
   inside another skill's execution. The slash-command surface
   `/check-plan` remains for direct operator use; both share the
   same eval module via `skills/check-plan/criteria.md` as the
   single canonical criteria list.

## Drafting protocol — batch

When the input is multiple topics:

1. Confirm the list with the user in one batched turn — show the
   numbered topics and ask for any to be removed, reframed, or added.
2. For each topic in confirmed order, run the single-decision
   protocol — including the per-ADR `/check-plan` gate (step 7).
   Each ADR is checked independently; one failure does not block
   the others. ADRs that pass the gate are written; ADRs that fail
   the gate after 3 rounds are surfaced in the batch report and not
   written (unless `--skip-check` was set, in which case all are
   written with breadcrumbs).
3. For trivial decisions (the "obvious answer" case), keep the ADR
   short — two options, one-paragraph Decision, terse Consequences.
   Trivial does not mean skip the ADR; it means write it briefly.
4. Render every file. Report back the list of numbers and titles
   created and any that were yielded by the gate.
5. After all files are rendered, run `bin/sync-adr-index` to refresh
   the index in `design/adr/README.md`. This keeps the index table
   aligned with what is on disk. If the script reports
   changes, include `design/adr/README.md` in the next commit
   alongside the new ADRs.

## How `[TBD]` fields are handled

If `design/prd-normalized.md` contains `[TBD]` on the "Constraints and
preferences" field, ask the user once whether any constraints apply to
the current decision; do not block on it.

## Self-check before finishing each ADR

Per file:

- [ ] `NNN` is the number `bin/adr-alloc` returned (next unused in
  `design/adr/`).
- [ ] Status is `proposed`.
- [ ] At least 2 options are listed, each with both Pros and Cons.
- [ ] Decision names one of the listed options (no hybrid unless the
  user explicitly asked for one).
- [ ] Consequences covers all four bullets ("None new" is allowed).
- [ ] Date is today's date in `YYYY-MM-DD`.
- [ ] No `{{...}}` template placeholders remain.
- [ ] `/check-plan` passed (deterministic criteria), or
  `--skip-check` was explicitly set with a recorded breadcrumb.

After the batch:

- [ ] `bin/sync-adr-index` ran cleanly. `design/adr/README.md` is
  staged alongside the new ADRs.

## Handoff

The user reviews each `proposed` ADR and changes status to `accepted`
when ready. This skill does not auto-accept. Accepted ADRs become
inputs to the executor skill (later issue) when implementing the
related issues.

See [`example.md`](example.md) for a worked batch — one contentious
decision and one near-trivial decision drafted in a single run.
