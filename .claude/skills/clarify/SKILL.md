---
name: clarify
description: Surface unresolved implementation questions ("gray areas") and conduct deep-dive resolution per question, appending settled decisions to design/decisions.md. Use when specific open questions need resolution between MVP scoping and ADR drafting; for broader project context use /planning; for committing a settled decision use /adr-writer.
permission-category: 1  # substitutable — local conversation; appends to design/decisions.md, per workflow-guide §7
inputs: []
outputs:
  - artefact: "design/decisions.md"
    description: "Settled below-ADR decisions appended (marker-fenced)"
next:
  - skill: adr-writer
    when: "a decision graduates to an ADR"
---

# clarify

Take `design/prd-normalized.md`, `design/mvp.md`, and (when present)
`design/planning.md`, scout the codebase, surface a checklist of
unresolved implementation questions, conduct deep-dive resolution
per question, and append settled decisions to `design/decisions.md`.

This skill is **opt-in**. Small projects can skip it entirely and go
from `prd-to-mvp` straight to `adr-writer`. Larger projects use it to
harden ambiguity into decisions before ADRs are drafted, so
`adr-writer` captures decisions and not exploration.

The output sits **below ADR weight by design**. It captures
informal-but-settled context that downstream agents (planner,
`adr-writer`, executor) can rely on without re-asking — but not
architectural commitments. See the **Graduate-to-ADR criterion**
below for the line.

See ADR-033 for the
rationale and the contract this skill ships against.

## When to use this skill

- After `prd-to-mvp` has produced `design/mvp.md`, when the project
  has enough scope that drafting ADRs straight from the MVP would
  mix decisions with discovery.
- After `/planning` has produced `design/planning.md`, when a fresh
  pass over the open research questions would benefit from
  settle-or-promote resolution.
- Mid-project when new gray areas surface — e.g. a sub-system whose
  shape became clear only during execution and that downstream work
  depends on.

If the project has only a handful of obvious decisions (a single
deployment target, one library pick), skip this skill — `adr-writer`
alone is sufficient.

## What this skill does not do

- Does not draft ADRs — that is `adr-writer`. This skill *flags*
  decisions that should graduate to ADRs (per the criterion below)
  but does not author them.
- Does not modify `design/prd-normalized.md`, `design/mvp.md`, or
  `design/planning.md`. Inputs only.
- Does not modify accepted ADRs in place. If a decision conflicts
  with an accepted ADR, the conflict surfaces during the gray-area
  scan and the user is asked whether to draft a superseding ADR.
- Does not delete or rewrite earlier entries in `design/decisions.md`.
  The log is append-only; re-runs only add new entries or skip
  already-settled questions.
- Does not duplicate content already locked by an accepted ADR. The
  scan reads `design/adr/*.md` and skips topics already settled
  there.

## Inputs

- **Required:** `design/prd-normalized.md`.
- **Required:** `design/mvp.md`.
- **Optional:** `design/planning.md`
  — when present, the open-research-questions section is the
  canonical source of gray areas to scan.
- **Optional:** existing `design/decisions.md` — used as the starting
  point for re-runs. Earlier entries are preserved verbatim; the
  skill skips topics already covered.
- **Optional:** `design/adr/*.md` — accepted ADRs lock topics. The
  scan parses titles and statuses; topics whose decisions are
  recorded in an accepted ADR are skipped.

## Output

- **`design/decisions.md`** — rendered from
  [`templates/decisions-template.md`](../../templates/decisions-template.md).

The output is a single self-contained markdown file. Each gray-area
entry is wrapped in marker fences (`<!-- decisions:NN:start -->` /
`<!-- decisions:NN:end -->`) so re-runs preserve every earlier entry
verbatim and append new ones at the end.

## Graduate-to-ADR criterion

The single test for "should this be a decision or an ADR":

> **Would superseding this decision later require a new ADR?**
>
> - **Yes** → it should be an ADR. Hand it to `adr-writer` instead
>   of recording it here.
> - **No** → it lives in `design/decisions.md`.

Reversal-cost is the heuristic. ADRs document choices whose reversal
is structural — a different storage backend, a different deployment
target, a different auth model. Decisions document choices whose
reversal is a routine code change — a default value, a naming
convention, a per-feature configuration default.

When the line is unclear, default to the ADR. Mis-promoting a
decision to an ADR has low cost; under-promoting an architectural
choice to a decision has compounding cost.

## Execution protocol

1. **Validate inputs.** Confirm `design/prd-normalized.md` and
   `design/mvp.md` exist. If either is missing, stop and tell the
   user to run `prd-normalizer` and / or `prd-to-mvp` first.
2. **Detect re-run.** If `design/decisions.md` already exists, read
   it and parse out every existing entry's topic. Earlier entries
   are preserved verbatim; the skill will not add a new entry for a
   topic already present.
3. **Read accepted ADRs.** Glob `design/adr/adr-*.md`. For each ADR
   with `**Status:** accepted`, extract its title and topic. The
   scan will skip gray areas whose topic maps to an accepted ADR —
   those decisions are already locked.
4. **Read PRD, MVP, planning.md (if present), end to end.** Note
   in-scope capabilities, product principles, declared phase order,
   and any open research questions captured in `planning.md`.
5. **Scout the codebase.** Look for reusable patterns or already-made
   informal decisions visible in existing modules. A decision the
   code already encodes is a decision; record it here rather than
   re-deciding.
6. **Surface gray areas.** Identify implementation questions whose
   answers are needed before downstream work can proceed but whose
   reversal would be a routine code change (per the
   graduate-to-ADR criterion). Filter out: (a) topics already in
   `decisions.md`, (b) topics locked by accepted ADRs, (c) topics
   that should be ADRs not decisions.
7. **Show the checklist to the user.** Number the gray areas and
   ask which to resolve in this round. The user can pick a subset;
   unselected gray areas remain open and surface again on the next
   re-run.
8. **Conduct deep-dive resolution.** For each selected gray area,
   ask the user (or surface from the codebase): the question, the
   options worth weighing, the chosen option, and the rationale.
   Keep each deep-dive tight — three exchanges max per question;
   if a decision needs more, it's probably an ADR.
9. **Append to `decisions.md`.** Render each resolved gray area as
   a new section using the marker-fenced shape from
   `templates/decisions-template.md`. Append at the end; never
   rewrite earlier entries.
10. **Self-check** (see below).
11. **Report.** Print the path to the rendered file, the count of
    new entries added, and the count of gray areas left unresolved
    for next time.

## How re-runs work

The skill is idempotent. Running `/clarify` on a project where
`decisions.md` already exists:

- Preserves every earlier entry verbatim (marker fences guard them).
- Skips gray areas whose topic already appears in an existing entry.
- Skips gray areas whose topic is locked by an accepted ADR.
- Surfaces only genuinely new gray areas for the user to choose from.

If `decisions.md` is malformed (missing fences, unclosed sections),
the skill stops and asks the user before regenerating — clobbering
hand-written context silently is the failure mode this guards
against.

## Edge cases

- **PRD or MVP missing.** Stop. Tell the user to run upstream skills
  first. Do not invent inputs.
- **Existing `decisions.md` malformed.** Stop. Show the user which
  marker fences are missing or unclosed and ask whether to fix
  manually before re-running.
- **No gray areas to resolve.** Allowed. Report "No new gray areas
  surfaced; `decisions.md` unchanged" and stop. Do not pad.
- **A gray area conflicts with an accepted ADR.** Stop the deep-dive
  for that area and ask the user whether to draft a superseding ADR
  via `adr-writer`. Do not record a decision that contradicts an
  accepted ADR.
- **A gray area should be an ADR.** Apply the graduate-to-ADR
  criterion. If the answer is "yes", surface it in the report as a
  candidate ADR topic for `adr-writer` rather than recording it
  here.

## Self-check before writing

- [ ] Every selected gray area has a question, options weighed, a
  decision, and a rationale.
- [ ] No new entry duplicates a topic already in `decisions.md` or
  in an accepted ADR.
- [ ] No new entry would, on its own merits, warrant an ADR (apply
  the graduate-to-ADR criterion).
- [ ] All earlier entries are preserved verbatim — diff against the
  pre-run state of `decisions.md` shows additions only.
- [ ] Marker fences are intact and balanced (each `:start` has a
  matching `:end`).
- [ ] No `{{...}}` template placeholders remain in the rendered
  file.

## Handoff

`design/decisions.md` is consumed by `adr-writer` (when present, as
an Optional input — additive context for ADR Context sections). It
is also a useful read for `/claude-issue-executor` when briefing a
new session — though the executor reads it transitively via the
prompt rather than directly.

If the report flags any gray area as a graduate-to-ADR candidate,
hand the topic list to `adr-writer` as the next step.

See [`example.md`](example.md) for a worked PRD + MVP + planning.md →
decisions.md walkthrough.
