# workflow-docs — worked example

A single run against the **Pace Drift** target project after `prd-to-mvp`
has produced `design/mvp.md` and `adr-writer` has produced two ADRs.
This example shows (1) a first run from a clean repo, (2) section
omission when a source is missing, and (3) a re-run that preserves a
manual edit outside markers.

---

## 1. Target project state before the run

```
pace-drift/
├── CLAUDE.md
├── design/
│   ├── adr/
│   │   ├── adr-001-gpx-parsing-location.md    (accepted)
│   │   └── adr-002-test-framework.md          (accepted)
│   ├── mvp.md
│   └── build-out-plan.md
└── .claude/skills/workflow-docs/SKILL.md
```

No `design/prd.md` (the user went straight from a rough idea through
`idea-to-prd` → `prd-normalizer`, producing `design/prd-normalized.md`
which was then consumed by `prd-to-mvp`). No `README.md` yet.

Relevant snippets the skill will read:

**`design/mvp.md`** (relevant lines only):

```markdown
# Pace Drift — MVP

## Product name
Pace Drift

## One-line description
A browser tool that shows where a runner's pace drifted during a
single race, from a GPX file.

## Product goal
Let a runner load one GPX and immediately see pace-vs-target drift,
with no sign-up and no data leaving their machine.

## Target users
### Primary user
A recreational runner who wants post-race pace insight in under a minute.

## Product principles
1. Local session only — no server-side persistence.
2. Verbatim data — no AI smoothing or interpretation.
3. One runner, one race, one view.

## MVP scope
### In scope
- Load a GPX from disk.
- Compute pace drift against a target pace.
- Render a single-page drift visualisation.

### Out of scope
- Multi-race comparison.
- Saved history across sessions.
- Sharing / export.
```

**`CLAUDE.md`** (relevant lines only):

```markdown
# Pace Drift
> A browser tool that shows where a runner's pace drifted during a race.

## Current phase
MVP — Phase 1: GPX parsing and drift computation.

Active milestone: `MVP`.

## Technology stack
- Runtime / language: Node.js 22 + TypeScript
- Framework: Vite (browser app)
- Data layer: none (local session only)
- Key libraries: gpx-parse, d3

## How to run
\```bash
pnpm install
pnpm dev
pnpm test
\```
```

**ADRs**: `adr-001-gpx-parsing-location.md` (Parse GPX files in the
browser, not on a server) and `adr-002-test-framework.md` (Use Vitest
as the test framework).

## 2. First run

```
User: /workflow-docs
```

The skill:

1. Confirms `.git/` exists.
2. Reads `design/mvp.md`, `design/adr/adr-001-*.md`, `design/adr/adr-002-*.md`,
   and `CLAUDE.md`. No `design/prd.md` → that field stays empty but MVP
   covers the same ground.
3. Builds the context dict:
   - `PROJECT_NAME = "Pace Drift"`
   - `PROJECT_TAGLINE = "A browser tool that shows where a runner's pace drifted during a single race, from a GPX file."`
   - `PROJECT_DESCRIPTION = "Let a runner load one GPX and immediately see pace-vs-target drift, with no sign-up and no data leaving their machine."`
   - `PRIMARY_USER = "A recreational runner who wants post-race pace insight in under a minute."`
   - `CURRENT_PHASE = "MVP — Phase 1: GPX parsing and drift computation."`
   - `CURRENT_MILESTONE = "MVP"`
   - `IN_SCOPE_BULLETS = ["Load a GPX from disk.", "Compute pace drift against a target pace.", "Render a single-page drift visualisation."]`
   - `OUT_OF_SCOPE_BULLETS = ["Multi-race comparison.", "Saved history across sessions.", "Sharing / export."]`
   - `INSTALL_COMMAND = "pnpm install"`
   - `DEV_COMMAND = "pnpm dev"`
   - `TEST_COMMAND = "pnpm test"`
   - `KEY_ADR_BULLETS = ["- ADR-002: Use Vitest as the test framework.", "- ADR-001: Parse GPX files in the browser, not on a server."]` (newest first)
4. Omission rules: every README section has source data → all included.
   For the architecture doc and AI summary, `extension-points` has no
   source → omitted.
5. Renders all three files in memory and shows them in chat as fenced
   blocks, then asks:
   > Write these files?
   > - README.md (new, 50 lines)
   > - design/architecture.md (new, 55 lines)
   > - design/ai-summary.md (new, 38 lines)
   > (yes / edit / cancel)
6. User replies `yes`.
7. Skill writes all three files and reports paths.

Generated `README.md`:

```markdown
# Pace Drift

<!-- workflow-docs:start:tagline -->
> A browser tool that shows where a runner's pace drifted during a single race, from a GPX file.
<!-- workflow-docs:end:tagline -->

<!-- workflow-docs:start:overview -->
## Overview

Let a runner load one GPX and immediately see pace-vs-target drift,
with no sign-up and no data leaving their machine.
<!-- workflow-docs:end:overview -->

<!-- workflow-docs:start:who-for -->
## Who this is for

A recreational runner who wants post-race pace insight in under a minute.
<!-- workflow-docs:end:who-for -->

<!-- workflow-docs:start:status -->
## Status

Current phase: MVP — Phase 1: GPX parsing and drift computation.
Active milestone: `MVP`.
<!-- workflow-docs:end:status -->

<!-- workflow-docs:start:scope -->
## What is in scope

- Load a GPX from disk.
- Compute pace drift against a target pace.
- Render a single-page drift visualisation.

### Out of scope

- Multi-race comparison.
- Saved history across sessions.
- Sharing / export.
<!-- workflow-docs:end:scope -->

<!-- workflow-docs:start:how-to-run -->
## How to run

\```bash
pnpm install
pnpm dev
pnpm test
\```
<!-- workflow-docs:end:how-to-run -->

<!-- workflow-docs:start:key-decisions -->
## Key decisions

- ADR-002: Use Vitest as the test framework.
- ADR-001: Parse GPX files in the browser, not on a server.

See [`design/adr/`](design/adr/) for the full list.
<!-- workflow-docs:end:key-decisions -->

<!-- workflow-docs:start:more -->
## More

- [`CLAUDE.md`](CLAUDE.md) — project rules for Claude Code
- [`design/mvp.md`](design/mvp.md) — MVP statement
- [`design/architecture.md`](design/architecture.md) — current architecture/design reference
- [`design/ai-summary.md`](design/ai-summary.md) — AI-readable project summary
- [`design/adr/`](design/adr/) — decision history and rationale
<!-- workflow-docs:end:more -->
```

Generated `design/architecture.md` (abridged):

```markdown
# Pace Drift — Architecture

**Last updated:** 2026-04-17

<!-- workflow-docs:start:overview -->
## Overview

Pace Drift is a local-only browser app that loads one GPX file and shows where
a runner drifted from target pace.
<!-- workflow-docs:end:overview -->

<!-- workflow-docs:start:major-components -->
## Major components

- GPX file loader and parser
- Pace-drift computation module
- Single-page visualisation layer
<!-- workflow-docs:end:major-components -->

<!-- workflow-docs:start:key-constraints -->
## Key constraints

- No server-side persistence.
- No AI smoothing or interpretation of running data.
- One runner, one race, one view for the MVP.
<!-- workflow-docs:end:key-constraints -->
```

Generated `design/ai-summary.md` (abridged — the same marker pattern
wraps each section; `extension-points` is absent because it had no
source):

```markdown
# Pace Drift — AI Summary

**Last updated:** 2026-04-17

<!-- workflow-docs:start:objectives -->
## Objectives

- Let a runner load one GPX and see pace-vs-target drift in under a minute.
- Keep the session local — no sign-up, no server-side persistence.
- Render a single-page drift visualisation from verbatim data (no AI smoothing).
<!-- workflow-docs:end:objectives -->

<!-- workflow-docs:start:tech-stack -->
## Tech stack

- Runtime / language: Node.js 22 + TypeScript
- Framework: Vite (browser app)
- Data layer: none (local session only)
- External services / APIs: none
- Deployment target: static site
<!-- workflow-docs:end:tech-stack -->

<!-- workflow-docs:start:constraints -->
## Constraints

- Local session only — no server-side persistence.
- Verbatim data — no AI smoothing or interpretation.
- One runner, one race, one view.
<!-- workflow-docs:end:constraints -->

<!-- workflow-docs:start:current-status -->
## Current status

- Current phase: MVP — Phase 1: GPX parsing and drift computation.
- Active milestone: `MVP`.
<!-- workflow-docs:end:current-status -->

<!-- workflow-docs:start:key-decisions -->
## Key decisions

- ADR-002: Use Vitest as the test framework.
- ADR-001: Parse GPX files in the browser, not on a server.
<!-- workflow-docs:end:key-decisions -->
```

Note that sections with no source data are omitted from each generated
file — heading, body, and markers are all removed cleanly.

## 3. Re-run after a manual edit

The user adds a paragraph under `## How to run` in `README.md`
**outside** the markers:

```markdown
<!-- workflow-docs:end:how-to-run -->

### Troubleshooting

If `pnpm dev` fails with "GPX parse error: empty track", your GPX file
may have no `<trkpt>` elements. Re-export from your watch app with GPS
data included.
```

Then lands ADR-003 (switch from Vite to esbuild for the bundler) and
re-runs `/workflow-docs`.

The skill:

1. Reads the existing generated docs and detects their
   `start:/end:` marker pairs are intact.
2. Regenerates `key-decisions` (now three ADRs, newest first) and
   re-checks every other section for changes. Only `key-decisions`
   has a real diff.
3. Shows the diff:
   ```diff
     <!-- workflow-docs:start:key-decisions -->
     ## Key decisions

   + - ADR-003: Switch from Vite to esbuild for bundling.
     - ADR-002: Use Vitest as the test framework.
     - ADR-001: Parse GPX files in the browser, not on a server.

     See [`design/adr/`](design/adr/) for the full list.
     <!-- workflow-docs:end:key-decisions -->
   ```
4. Notes the manual `### Troubleshooting` section is outside markers
   and will be preserved.
5. Asks to write. User confirms.
6. Writes — the `### Troubleshooting` block is untouched.

## 4. Section omission — a sparse target project

If the user runs `/workflow-docs` before `prd-to-mvp` (so there is a
`CLAUDE.md` and two ADRs but no `design/mvp.md` and no PRD), the skill
emits only:

`README.md`:

```markdown
# <repo basename>

<!-- workflow-docs:start:status -->
## Status

Current phase: <from CLAUDE.md>.
Active milestone: `<from CLAUDE.md>`.
<!-- workflow-docs:end:status -->

<!-- workflow-docs:start:how-to-run -->
## How to run

\```bash
<commands from CLAUDE.md>
\```
<!-- workflow-docs:end:how-to-run -->

<!-- workflow-docs:start:key-decisions -->
## Key decisions

- ADR-002: ...
- ADR-001: ...

See [`design/adr/`](design/adr/) for the full list.
<!-- workflow-docs:end:key-decisions -->

<!-- workflow-docs:start:more -->
## More

- [`CLAUDE.md`](CLAUDE.md) — project rules for Claude Code
- [`design/mvp.md`](design/mvp.md) — MVP statement
- [`design/architecture.md`](design/architecture.md) — current architecture/design reference
- [`design/ai-summary.md`](design/ai-summary.md) — AI-readable project summary
- [`design/adr/`](design/adr/) — decision history and rationale
<!-- workflow-docs:end:more -->
```

The `tagline`, `overview`, `who-for`, and `scope` sections are all
omitted entirely — headings and bodies gone, no stray blank lines. The
skill also warns the user that running `prd-to-mvp` first would
produce a much richer README.

## 5. Self-check trace

For the first-run README:

- [x] Every `{{PLACEHOLDER}}` is filled (no section was omitted).
- [x] All 8 sections wrapped in matching `start:/end:` markers.
- [x] No omitted heading / body / marker left behind.
- [x] (N/A on first run) No content outside markers to preserve.
- [x] User explicitly confirmed the write.

For the first-run architecture doc and AI summary:

- [x] Every `{{PLACEHOLDER}}` in included sections is filled.
- [x] `extension-points` section omitted — heading, body, markers all
      gone.
- [x] Remaining five sections wrapped in matching markers.
- [x] User explicitly confirmed the write.

The skill reports back:

> Wrote 3 files:
> - /Users/.../pace-drift/README.md (new, 50 lines)
> - /Users/.../pace-drift/design/architecture.md (new, 55 lines)
> - /Users/.../pace-drift/design/ai-summary.md (new, 38 lines)
> Sections omitted: design/architecture.md#extension-points, design/ai-summary.md#extension-points (no source data).
