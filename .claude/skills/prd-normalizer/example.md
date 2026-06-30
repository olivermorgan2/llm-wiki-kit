# prd-normalizer — worked examples

Two end-to-end transformations: one from a standard PRD, one from a
custom document. Reference material, not runtime artifacts.

---

## Example 1 — Standard PRD input

The input here is byte-for-byte the output of
[`idea-to-prd/example.md`](../idea-to-prd/example.md). This example
demonstrates that `idea-to-prd` and `prd-normalizer` chain cleanly.

### Input — `design/prd.md`

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
- Units — metric only, or both?
- What happens with a GPX that has pauses or GPS dropouts?
```

### Elicitation by the skill

Identity fields are missing — the PRD has a title but no explicit
product name field, and no one-line description. One batched question:

1. Confirm product name as "Pace Drift"?
2. One-line description — does *"A web app that shows where my race
   pace drifted from target splits, from a single GPX file"* work?

User confirms both. The input also mentions "web … on a laptop" and
"no mobile-optimized view in the first release" — enough to populate
"constraints and preferences" without another question.

### Output — `design/prd-normalized.md`

```markdown
# Pace Drift — Normalized PRD

## Product name
Pace Drift

## One-line description
A web app that shows where my race pace drifted from target splits,
from a single GPX file.

## Problem
After a race, I want to see how my pace drifted across the course
relative to the splits I was targeting. Strava shows pace but does not
cleanly compare against a target, so the answer takes manual work.

## Target users

### Primary user
Me — a solo runner analyzing my own races after the fact on a laptop.

### Secondary user
[TBD]

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

## Constraints and preferences
- Web-only, laptop-first. No mobile-optimized view yet.
- Single-user, local-session use. No backend persistence implied.

## Success signals
- A GPX from a recent race loads in under a few seconds.
- The drift chart matches what I can verify by hand from the same file.
- I stop opening Strava to answer the "where did I fade?" question.

## Open questions
- Target input format — single pace vs. per-kilometer vs. per-mile?
- Units — metric only, or both?
- What happens with a GPX that has pauses or GPS dropouts?
```

Note the transformation: "No mobile-optimized view in the first
release" was an MVP scope decision disguised as a non-goal. Moved to
"constraints and preferences" as a product-level shape note.
`prd-to-mvp` will decide the in-scope cut.

### Self-check trace

- [x] Product name is present and not `[TBD]`.
- [x] Problem is present and not `[TBD]`.
- [x] A primary user is named and not `[TBD]`.
- [x] At least one core capability is listed (four are).
- [x] At least one user story is present (four are).
- [x] Every one of the 11 canonical fields appears. Secondary user is
      `[TBD]` — a soft field, acceptable.

All six pass. Write the file and point the user at `prd-to-mvp` and
`adr-writer` as the next steps.

---

## Example 2 — Custom PRD input

A different project presented in the user's own brain-dump style — no
conventional headings, mixed prose and bullets, some musings.

### Input — `design/brief.md` (user's original, preserved)

```markdown
# Stand-up Notes Bot — what I'm thinking

So the daily stand-up in our team Slack is a mess. Everyone types
their updates into the channel, but by 10am it's buried, and the
manager (Priya) spends 15 minutes every morning scrolling back to
piece it together for her own report. I want to fix this.

Idea: a Slack bot that listens for "standup" posts in #team-eng each
morning, collects them into a single tidy summary, and posts that
summary into #team-leads at 10:00 with everyone's update grouped by
person.

Things I care about:
- It has to work in Slack. No "come to our website to see it."
- It should run on the cheap infra we already have (AWS, Lambda).
- Needs to respect that people sometimes post updates late — maybe a
  cutoff of 9:55am and anything after that is just… in the channel.
- No AI rewriting of people's updates. Priya wants the verbatim text.

Not doing:
- Not building a whole standup app. Slack-only.
- No analytics, no trends, no who-is-behind-schedule stuff.
- No DM reminders (for now anyway).

Stuff I haven't figured out:
- What happens if someone posts in a thread vs. the main channel?
- Should it handle multiple teams or just us to start? Probably just us.
- What about time zones? Half the team is in London, half in NYC.
```

### Elicitation by the skill

The skill reads the whole document and identifies fields by meaning:

- Problem → "daily stand-up is buried, Priya spends 15 minutes
  scrolling back"
- Target users → Priya (manager); team engineers post, but the
  *consumer* is Priya
- Goal → clean 10am summary in #team-leads
- User stories → implied but not written — ask
- Core capabilities → Slack listener, collector, formatter, poster
- Non-goals → "Not doing" block maps directly
- Constraints → Slack, AWS Lambda, no AI rewriting
- Success signals → implied ("Priya stops scrolling") — ask to confirm
- Open questions → "Stuff I haven't figured out" block maps directly

Batched questions (≤ 5):

1. Confirm product name as "Stand-up Notes Bot"?
2. One-line description — *"A Slack bot that collects a channel's
   daily stand-up posts into a tidy summary in the managers'
   channel"* — work?
3. Primary user is Priya (reads the summary). Any secondary user worth
   naming (e.g. the team engineers who post)?
4. Can you turn the workflow into 2–3 user stories? For example:
   *"As a manager, I see a single tidy summary at 10am so I don't
   scroll back through the channel."*
5. Success signal — is "Priya stops manually scrolling to write her
   report" the right measure, or is there another?

User answers. The skill drafts user stories from the answer and fills
the remaining fields.

### Output — `design/prd-normalized.md`

```markdown
# Stand-up Notes Bot — Normalized PRD

## Product name
Stand-up Notes Bot

## One-line description
A Slack bot that collects a channel's daily stand-up posts into a tidy
summary in the managers' channel.

## Problem
The daily stand-up in #team-eng is posted as free-form messages that
get buried by 10am. Priya, the manager, spends around 15 minutes every
morning scrolling back to piece the updates together for her own
report.

## Target users

### Primary user
Priya — engineering manager who reads the summary and produces her own
upward report.

### Secondary user
Team engineers posting the original updates in #team-eng.

## Goal
Produce a single, tidy summary of the team's stand-up updates at 10:00
each morning, posted in #team-leads, grouped by person. Priya stops
scrolling back through #team-eng to assemble her report.

## User stories / scenarios
- As a manager, I see one tidy summary at 10am in #team-leads so I do
  not scroll back through #team-eng.
- As a manager, I see each engineer's update verbatim so I trust the
  summary without re-checking the source.
- As an engineer, I post my update in #team-eng however I normally do
  so — no new ritual.

## Core capabilities
- Listen for stand-up posts in #team-eng up to a cutoff time.
- Collect the posts and group them by author.
- Post a single formatted summary to #team-leads at 10:00.
- Respect a cutoff (default 09:55); anything later stays in the
  source channel only.

## Non-goals
- Not a standalone stand-up app. Slack only.
- No analytics, trend reports, or who-is-behind-schedule features.
- No DM reminders.
- No AI rewriting of engineer updates — verbatim text only.

## Constraints and preferences
- Slack-native. No web UI.
- Runs on existing AWS Lambda infrastructure.
- Must preserve posts verbatim; no paraphrasing.
- Single-team only.

## Success signals
- Priya stops manually scrolling to assemble her morning report.
- The 10:00 summary arrives on time every working day.
- Engineers do not change how they post their updates.

## Open questions
- Threaded replies vs. main-channel posts — collect both, or only
  main-channel?
- Time zones — London and NYC split. Does the 10:00 cutoff mean
  10:00 UTC, 10:00 London, or per-recipient local?
```

Note the transformation:

- The user's brain-dump headings ("Things I care about", "Not doing",
  "Stuff I haven't figured out") were remapped to canonical fields
  without preserving the original heading names.
- MVP-scope statements in the input ("Probably just us" for a single
  team) were moved to "constraints and preferences" as a product-shape
  note, not retained under non-goals.
- The constraints section captures the Slack/Lambda/verbatim
  requirements together — the primary hint source for `adr-writer`.

### Self-check trace

- [x] Product name is present and not `[TBD]`.
- [x] Problem is present and not `[TBD]`.
- [x] A primary user is named (Priya).
- [x] At least one core capability is listed (four are).
- [x] At least one user story is present (three are).
- [x] Every one of the 11 canonical fields appears. After the batched
      questions, nothing was left `[TBD]` — the input plus answers
      covered every field. (Example 1 is where the `[TBD]` marker is
      exercised, on "Secondary user".)

All six pass. The user's original `design/brief.md` is untouched. The
skill writes `design/prd-normalized.md` and points the user at
`prd-to-mvp` and `adr-writer` as the next steps.
