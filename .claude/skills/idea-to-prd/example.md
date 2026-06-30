# idea-to-prd — worked example

A single end-to-end walkthrough showing how the `idea-to-prd` skill
turns a rough idea into a PRD draft. This is reference material, not a
runtime artifact.

---

## 1. User's rough idea (input)

> I want a small web app that lets me drop my race GPX files in and see
> how my pace drifted across the course versus my target splits. I've
> been eyeballing it in Strava and it's annoying. I don't care about
> social features or training plans — I just want the pace drift view
> for a single race at a time.

That is all the user has. No users section, no success criteria, no
non-goals written down.

## 2. Elicitation questions asked by the skill

The skill drafts a first pass internally, marks unknowns as `[TBD]`,
and asks the user a batched set of questions. For this idea, the
batch looks like:

1. Who is this for — just you, or other runners too?
2. Web only, or does a mobile view matter for the first release?
3. What would make this feel "done" to you — a single race view that
   loads a GPX and shows a drift chart, or something richer?
4. Anything you want to rule out up front beyond social and training
   plans? (Auth? Saved history? Multi-race comparison?)

The skill does not ask about the problem or core capabilities —
both were clear enough from the rough idea to draft without input.

## 3. User's answers

1. Just me to start; other runners is a Phase 2 idea.
2. Web only. I'll open it on my laptop after a race.
3. Drop a GPX, pick a target pace or target splits, see a drift chart
   and a split-by-split table. That's it.
4. No accounts, no saved history, no multi-race compare in the first release.

## 4. Output — `design/prd.md` in the target project

```markdown
# Pace Drift — PRD

## Problem

After a race, I want to see how my pace drifted across the course
relative to the splits I was targeting. Strava shows pace but does not
cleanly compare against a target, so the answer takes manual work.

## Target users

### Primary user

Me — a solo runner analyzing my own races after the fact on a laptop.

## Goal

Provide a single-race pace-drift view that loads a GPX, takes a target
pace or target splits, and shows where I was ahead or behind. Success
for the first release is "I stop eyeballing it in Strava."

## User stories / scenarios

- As a runner, I drop a GPX file into the app so that I can analyze a
  race I just finished.
- As a runner, I enter a target pace or per-split targets so that the
  app has something to compare against.
- As a runner, I see a drift chart across the course so that I can spot
  where I slowed down or sped up.
- As a runner, I see a split-by-split table so that I can read exact
  numbers, not just the chart.

## Core capabilities

- Load and parse a single GPX file.
- Accept a target pace, or per-split targets, as input.
- Compute actual vs. target drift across the course.
- Render a drift chart and a split table for that race.

## Non-goals

- No user accounts or authentication.
- No saved history of past races.
- No multi-race comparison.
- No social features, sharing, or feeds.
- No training plans or coaching suggestions.
- No mobile-optimized view in the first release.

## Success signals

- A GPX from a recent race loads in under a few seconds.
- The drift chart matches what I can verify by hand from the same file.
- I stop opening Strava to answer the "where did I fade?" question.

## Open questions

- Target input format — single pace vs. per-kilometer vs. per-mile?
  Probably all three eventually, but pick one for the first release.
- Units — metric only, or both? I run in km but some races are marked
  in miles.
- What happens with a GPX that has pauses or GPS dropouts?
```

## 5. Why this output is ready for `prd-normalizer`

- Eight-section standard-PRD shape — `prd-normalizer`'s standard-PRD
  path can consume it without special-casing.
- No MVP scoping language — no "in scope / out of scope" split, no
  prioritization. That is `prd-to-mvp`'s job.
- No ADR-style decisions — only product intent. That is `adr-writer`'s
  job.
- Deliberately small. Open questions remain and are listed as such
  rather than forced into false answers.

## 6. Self-check trace

- [x] Problem is named in one or two sentences.
- [x] A primary user is named.
- [x] At least one concrete user story is present (four are).
- [x] At least one non-goal is listed (six are).
- [x] No `[TBD — …]` placeholders remain.

All five pass, so the skill writes the file and points the user at
`prd-normalizer` as the next step.
