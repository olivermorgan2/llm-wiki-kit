---
name: planning
description: Capture deeper planning context — requirements decomposition, risks, assumptions, sequencing rationale, open research questions — into design/planning.md. Use when project scope warrants deeper preparation before ADR drafting; for resolving specific open questions use /clarify; for committing a decision to an ADR use /adr-writer.
permission-category: 1  # substitutable — writes design/planning.md locally, per workflow-guide §7
inputs:
  - name: "--granularity"
    required: false
    description: "Decomposition depth: coarse | standard | fine"
outputs:
  - artefact: "design/planning.md"
    description: "Decomposition, risks, sequencing (marker-fenced)"
next:
  - skill: adr-writer
    when: "decisions surface that need ADRs"
  - skill: issue-planner
    when: "the sequencing feeds the backlog"
---

# planning

Take `design/prd-normalized.md` and `design/mvp.md` and produce a
single planning artefact at `design/planning.md` that decomposes the
in-scope MVP capabilities into requirements, surfaces risks and
assumptions, justifies the build-out plan's phase order, and lists
open research questions that need resolving before specific ADRs can
be drafted.

This skill is **opt-in**. Small projects can skip it entirely and go
from `prd-to-mvp` straight to `adr-writer`. Larger projects use it
to harden ambiguity into decisions before ADRs are drafted, so
`adr-writer` captures decisions and not exploration.

See ADR-031
for the rationale and the contract this skill ships against.

## When to use this skill

- After `prd-to-mvp` has produced `design/mvp.md` and
  `design/build-out-plan.md`.
- Before `adr-writer`, when the project is large enough that ADRs
  drafted directly from the MVP would mix decisions with discovery.
- When a fresh planning round is needed mid-project — e.g. after a
  pivot, or before starting a new milestone.

If the project is small enough that one or two ADRs cover all
architectural questions, skip this skill. Running it on a tiny
project produces a doc with more ceremony than content.

## What this skill does not do

- Does not draft ADRs — that is `adr-writer`. This skill surfaces
  a "Decisions needing ADRs" list, but does not author the ADRs
  themselves.
- Does not modify `design/prd-normalized.md` or `design/mvp.md`. If
  scope is wrong, fix it upstream.
- Does not modify `design/build-out-plan.md`. The planning doc
  *justifies* the phase ordering recorded there; it does not change
  it.
- Does not auto-resolve open research questions. Owners and target
  dates are recorded; the resolutions land in follow-up ADRs or in
  `design/decisions.md` (once ADR-033 ships).
- Does not duplicate content the PRD or MVP already captures.
  `planning.md` is for *new* context — decomposition, risk,
  sequencing rationale, open questions — not a restatement.

## Inputs

- **Required:** `design/prd-normalized.md`.
- **Required:** `design/mvp.md`.
- **Optional:** `design/build-out-plan.md` — used to derive the
  sequencing-rationale section. If absent, the skill prompts the user
  for the phase order before justifying it.
- **Optional:** an existing `design/planning.md` — used as the
  starting point for re-runs. The skill preserves editorial text
  outside the marker fences.
- **Optional flag:** `--granularity={coarse|standard|fine}` — phase-
  count target band.
  Same precedence as `prd-to-mvp`: explicit flag > `**Granularity:**`
  line in `design/build-out-plan.md` > default `standard`. The flag
  affects this skill only when no build-out-plan exists yet — the
  declared phase order in an existing build-out-plan is canonical
  so the flag does not re-decompose phases. If the
  flag and the stored value disagree, the skill prints a one-line
  warning that the new tier will only take effect on the next
  `prd-to-mvp` run.

## Output

- **`design/planning.md`** — rendered from
  [`templates/planning-template.md`](../../templates/planning-template.md).

The output is a single self-contained markdown file. Marker fences
(e.g. `<!-- planning:requirements:start -->`) wrap each
auto-generated section so re-runs can refresh sections in place
without clobbering hand-written commentary.

## Execution protocol

1. **Validate inputs.** Confirm `design/prd-normalized.md` and
   `design/mvp.md` exist. If either is missing, stop and tell the
   user to run `prd-normalizer` and / or `prd-to-mvp` first.
2. **Resolve granularity tier.** If `--granularity=<tier>` was
   passed, validate the value is one of `coarse|standard|fine`. Reject
   invalid tiers with a one-line error and stop. If
   `design/build-out-plan.md` exists and contains a `**Granularity:**`
   line:
   - When no flag was passed, read the stored value and use it.
   - When a flag was passed and disagrees with the stored value,
     prefer the flag for this run and print a one-line warning that
     the build-out-plan still records the old tier and will need a
     `prd-to-mvp` re-run to update on disk.
   Else default to `standard`. The resolved tier informs sequencing
   rationale (step 7) but does not re-decompose phases when a
   build-out-plan already declares them — ADR-031 makes the
   build-out-plan canonical for phase order.
3. **Detect re-run.** If `design/planning.md` already exists, read
   it and use the marker-fenced sections as starting points. Editorial
   text outside the fences is preserved verbatim.
4. **Read the inputs end to end.** PRD, MVP, and build-out plan if
   present. Note in-scope capabilities, product principles, and the
   declared phase order.
5. **Decompose requirements.** For each in-scope capability, draft
   2–5 concrete requirements with stable IDs (`R1`, `R2`, …). Group
   by capability heading so traceability is obvious.
6. **Surface risks.** Walk through the requirements and identify
   risks that could derail delivery. Each risk gets impact,
   likelihood, and a one-line mitigation. Default to honesty over
   completeness — risks that don't matter are easy to drop later.
7. **Record assumptions.** List things the plan depends on (tools,
   access, prior decisions, external services). Each assumption
   carries an "if this is wrong" pointer so it's obvious when one
   breaks.
8. **Justify sequencing.** For each phase in the build-out plan,
   write a short paragraph explaining why it has to come before the
   next. This is the section `issue-planner` reads when deciding
   issue order. If no build-out plan exists yet, ask the user for
   the phase order before writing this section. Reference the
   resolved granularity tier (from step 2) inline when explaining
   why this project landed at this phase count.
9. **Capture open research questions.** List questions whose
   answers are needed before specific ADRs can be drafted. Each
   question carries an owner and a target answer date.
10. **List decisions needing ADRs.** Synthesise the architectural
    questions surfaced by the planning round into a flat list of
    decision topics. This list is the input to `adr-writer`.
11. **Self-check** (see below).
12. **Render** `design/planning.md` from
    `templates/planning-template.md`. Substitute every
    `{{PLACEHOLDER}}`. Preserve any editorial text outside marker
    fences from the prior version.
13. **Print** the path to the rendered file and the count of items
    per section so the user can spot gaps quickly.

## How re-runs work

The skill is idempotent in the sense that re-running it on a project
where `planning.md` already exists refreshes the marker-fenced
sections in place. Editorial commentary outside the fences (a
section header annotated with project context, a hand-written aside,
a link to a Slack thread) is preserved.

If the prior version is malformed (missing fences, unclosed
sections), the skill stops and asks the user before regenerating —
clobbering hand-written context silently is the failure mode this
guards against.

## Edge cases

- **PRD or MVP missing.** Stop. Tell the user to run the upstream
  skills first. Do not invent inputs.
- **Build-out plan missing.** Continue, but prompt the user for the
  phase order before writing the sequencing-rationale section. The
  resolved granularity tier (default `standard` if no flag) informs
  the prompt — e.g. "5–8 phases for `standard`, 1–3 for `coarse`".
- **Existing `planning.md` malformed.** Stop. Show the user which
  marker fences are missing or unclosed and ask whether to
  regenerate from scratch (with confirmation) or fix manually.
- **No in-scope capabilities to decompose.** Stop. The MVP is
  empty or the PRD is too thin — flag this back upstream.
- **No risks worth recording.** Allowed. Leave the section with a
  single bullet "No material risks identified at planning time."
  Do not pad.
- **`--granularity` invalid.** Stop in step 2 with a one-line error
  naming the three valid values (`coarse`, `standard`, `fine`). Do
  not write any output.
- **`--granularity` disagrees with the build-out-plan's stored tier.**
  The flag wins for this run; print a one-line warning that the
  build-out-plan still records the old tier and a `prd-to-mvp` re-run
  is needed to update it on disk.

## Self-check before writing

- [ ] Every in-scope MVP capability has at least one requirement in
  the decomposition section.
- [ ] Every risk has an impact, likelihood, and mitigation.
- [ ] Every assumption has an "if this is wrong" pointer.
- [ ] Sequencing rationale covers every phase in the build-out plan.
- [ ] Open research questions name an owner and a target date.
- [ ] Decisions needing ADRs are architectural questions, not feature
  descriptions.
- [ ] No `{{...}}` template placeholders remain in the rendered file.
- [ ] Marker fences are intact and balanced (each `:start` has a
  matching `:end`).

## Handoff

The "Decisions needing ADRs" list at the end of `planning.md` is the
direct input to `adr-writer`. Hand it the list as a batch.
`adr-writer` reads `design/planning.md` (when present) for context
and produces one ADR per decision topic.

`issue-planner` also reads `design/planning.md` (when present) and
uses the sequencing-rationale section to order the issue backlog.

See [`example.md`](example.md) for a worked PRD + MVP → planning.md
walkthrough.
