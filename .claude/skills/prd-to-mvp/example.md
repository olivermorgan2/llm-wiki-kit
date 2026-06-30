# prd-to-mvp — worked example

Continues the Pace Drift scenario. Input is byte-for-byte the
`design/prd-normalized.md` shown in
[`prd-normalizer/example.md`](../prd-normalizer/example.md) (Example 1).

---

## 1. Input — `design/prd-normalized.md`

The Pace Drift normalized PRD: 11 fields covering problem, primary
user (a solo runner analyzing races on a laptop), goal (single-race
pace-drift view from a GPX), four user stories, four core capabilities
(parse GPX; accept target pace or splits; compute drift; render chart
and table), product-level non-goals (no accounts, no history, no
multi-race, no social, no training plans), constraints (web-only,
laptop-first, single-user, local session), three success signals, and
three open questions. See the prd-normalizer example for the full text.

## 2. Elicitation by the skill

Batched, one turn:

1. Project size — weekend, one-month, multi-month?
2. Of the four core capabilities, the in-scope cut feels like all
   four (load GPX, accept target, compute drift, render chart +
   table). That's the whole product. Confirm?
3. Target input format is `[TBD]` in the open questions. For the
   first release, pick one — single overall pace? Per-kilometer
   splits? Per-mile splits?
4. Phase shape — does *foundation → UI → validation* in three phases
   match how you'd build it, or do you want to combine?

User answers: weekend project; yes all four capabilities are in
scope; pick single overall target pace for the first release
(per-split as Phase 2 work); three phases is fine.

## 3. Output — `design/mvp.md`

```markdown
# Pace Drift — MVP

**Last updated:** 2026-04-15

## Product name

Pace Drift

## One-line description

A web app that shows where my race pace drifted from a target pace,
loaded from a single GPX file.

## Product goal

Replace the manual "scroll through Strava and eyeball it" workflow
with a single-race drift view. Success for the first release is "I
stop opening Strava to answer the where-did-I-fade question after a
race."

## Target users

### Primary user

Me — a solo runner analyzing my own races after the fact on a laptop.

## Core problem

Strava shows pace but does not cleanly compare against a target pace,
so working out where I drifted off-target during a race takes manual
arithmetic and channel-scrolling. The answer takes longer than it
should and I usually give up.

## Product principles

1. One race at a time. The product never grows multi-race comparison.
2. Verbatim data. No AI smoothing or interpretation of the GPX.
3. Local session only. No accounts or server-side persistence.
4. Laptop-first. Mobile is a deliberate non-goal.

## MVP scope

### In scope

- Load and parse a single GPX file dropped into the page.
- Accept one target pace (e.g. "4:30/km") as input.
- Compute actual vs. target drift across the course.
- Render a drift chart across the course.
- Render a split-by-split table.

### Out of scope

Product-level non-goals from the PRD:
- User accounts or authentication.
- Saved history of past races.
- Multi-race comparison.
- Social features, sharing, or feeds.
- Training plans or coaching suggestions.

Deferred by this MVP scoping:
- Per-kilometer or per-mile target splits (single overall pace only
  in the first release).
- Mobile-optimized view.
- Handling of GPX files with paused recording or GPS dropouts.

## Primary outputs

A static web app — the user opens it locally or at a URL, drops a GPX,
enters a target pace, and sees the drift chart and split table.

## Success criteria

The MVP succeeds if a user can:

1. Drop a GPX from a recent race and see the drift chart in under
   a few seconds.
2. Read the split table and verify by hand that the numbers match the
   GPX file.
3. Stop opening Strava to answer "where did I fade?" for the rest of
   their race season.

## Deferred to later

- Per-split target input — Phase 2 once the single-pace flow is shipped.
- Mobile view — only if the laptop-first version is actually used
  enough to justify the work.
- Saved history — explicitly a non-goal, not a deferral.

## Acceptance criteria for this document

This MVP statement is acceptable when it:

- names a clear product and user — yes,
- lists what is in and out of scope without ambiguity — yes,
- and can drive the build-out plan, ADRs, and issue backlog without
  further interpretation — yes (see design/build-out-plan.md).

Open items: target input format for the eventual per-split path is
unresolved (Phase 2 decision).
```

## 4. Output — `design/build-out-plan.md`

```markdown
# Pace Drift — Build-Out Plan

**Last updated:** 2026-04-15

## Objective

Ship the MVP of Pace Drift defined in design/mvp.md: a static web app
that renders a drift chart and split table from one GPX file and one
target pace. The plan covers everything from repo bootstrap to a
locally-runnable first release.

## Build strategy

1. Bootstrap the repo with the workflow kit's standard scaffold.
2. Build the GPX-parsing and drift-computation core (no UI).
3. Build the input UI (file drop + target pace).
4. Build the output UI (drift chart + split table).
5. Wire the pieces together and dry-run on real race files.
6. Polish and write the README.

## Scope

- In scope: the five in-scope capabilities from design/mvp.md.
- Out of scope: every item in the "Out of scope" list of design/mvp.md.
- Assumptions: a TypeScript browser app is acceptable; user has a
  local toolchain (Node, browser); user has at least one real GPX
  file to test against.

## Success criteria

The plan is complete when a user can:

1. Open the local app, drop a GPX, enter a target pace, and see a
   drift chart within seconds.
2. Read the split table and verify the numbers by hand against the
   GPX file.
3. Reproduce the result for a second, different race.

## Repository structure

```text
pace-drift/
  src/
    parser/        ← GPX parsing
    drift/         ← actual-vs-target computation
    ui/            ← input form, chart, split table
    main.ts        ← entry point
  test/            ← mirrors src/
  public/          ← index.html, assets
  design/          ← mvp.md, build-out-plan.md, adr/
  README.md
  CLAUDE.md
```

## Phases

### Phase 1 — Foundation

- **Goal:** GPX in, drift numbers out, no UI.
- **Deliverables:** repo scaffold; GPX parser; drift computation
  module; unit tests for both.
- **Exit criteria:** given a real GPX file and a target pace, the
  drift module returns per-segment drift values that match a hand
  calculation.

### Phase 2 — UI

- **Goal:** the drift numbers become a chart and a table the user
  can read.
- **Deliverables:** file-drop input; target-pace input; drift chart
  component; split table component.
- **Exit criteria:** the user can drop a GPX, enter a pace, and see
  both a chart and a table on the page.

### Phase 3 — Validation and polish

- **Goal:** confidence the MVP is correct on real race data.
- **Deliverables:** dry-run on at least two real races; README; basic
  error messages for unparseable GPX.
- **Exit criteria:** MVP success criteria from design/mvp.md hold on
  two distinct GPX files.

## Milestone recommendation

| Milestone | Focus |
|---|---|
| M1 | Foundation — parser and drift computation |
| M2 | UI — input and output components |
| M3 | Validation — real-race dry-runs and README |

## Initial issue backlog

### M1 — Foundation

- Bootstrap repo with workflow kit scaffold
- Implement GPX parser (single file, single track)
- Implement drift computation against a single target pace
- Unit tests for parser and drift computation

### M2 — UI

- File-drop input component
- Target-pace input component
- Drift chart component
- Split-by-split table component
- Wire input → computation → output on the page

### M3 — Validation

- Dry-run on real race GPX (race 1)
- Dry-run on real race GPX (race 2)
- Write README with run instructions and known limitations

## Testing strategy

- Unit tests for parser and drift modules — happy path, edge cases
  (empty file, single point), and known-bad inputs.
- Manual dry-runs on at least two real-race GPX files in Phase 3 —
  the only honest way to validate the MVP success criteria.
- No integration tests are needed for the MVP; there is no backend.

## Risks and mitigations

### Risk 1 — GPX dialect variance

Different recorders produce slightly different GPX. *Mitigation:*
Phase 3 dry-runs on files from at least two different sources
(watch, phone) before declaring the MVP done.

### Risk 2 — Drift math is wrong but plausible-looking

The chart could be wrong in ways that aren't obvious. *Mitigation:*
Phase 1 exit criterion requires a hand-calculation match.

## Acceptance criteria for this document

This build-out plan is acceptable when it:

- matches the MVP statement — yes,
- sequences work in realistic phases — yes (3 phases),
- identifies initial ADRs or decisions — yes (see below),
- and produces a practical milestone and issue structure — yes
  (3 milestones, 12 issues).

## Decisions needing ADRs

Surfaced for handoff to `adr-writer`. Each item is one architectural
question.

1. **GPX parsing location: browser-side vs. server-side.** Affects
   privacy, file-size limits, cost, and the static-site assumption.
2. **Test framework.** Vitest, Jest, or node:test. Affects developer
   ergonomics; mostly a defaults question.
```

## 5. Self-check trace

- [x] Every core capability from the normalized PRD (4 of them) is
      classified in scope; no ambiguity.
- [x] MVP success criteria are user outcomes ("drop a GPX and see the
      chart in seconds"), not implementation tasks.
- [x] The "Out of scope" list separates the five product-level
      non-goals from the three deferred items, with each labelled by
      source.
- [x] Phase 1 exit criteria cover the in-scope capabilities related
      to parsing and computation; Phase 2 covers the UI capabilities;
      together they cover all five.
- [x] No `{{...}}` placeholders remain in either rendered file.

All five pass. Both files are written. The "Decisions needing ADRs"
list at the end of the build-out plan is handed to `adr-writer` next.
