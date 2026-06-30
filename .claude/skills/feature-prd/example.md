# feature-prd — worked example

A mature CLI tool (`slug-utils`, shipped at v1.0) wants to add a major
feature: **batch slugification from a file**, with streaming for large
inputs. This crosses the public API, performance, and the existing
"single-string in, single-string out" assumption — too big for one
issue. So it enters as a follow-up PRD.

## Input

> "I want `slug-utils` to slugify a whole file of lines at once, and
> stream so it handles a million-line file without loading it all into
> memory. Keep the existing single-string API exactly as-is."

## Output: `design/prd-addenda/001-batch-streaming.md`

```markdown
# PRD Addendum 001: Batch and streaming slugification

## Context

- Original PRD: design/prd.md (single-string slugify/unslugify, v1.0)
- Current milestone/state: v1.0 shipped; single-phase project, no
  in-flight work (design/state.md)
- Trigger for this update: users processing large CSV exports need
  whole-file slugification without an external loop

## Problem

slug-utils only transforms one string per call. Slugifying a large
file means an external loop and, for million-line inputs, holding the
whole file in memory.

## Goals

- Slugify every line of a file in one invocation.
- Stream input and output so memory stays bounded regardless of size.

## Non-goals

- Parallel/multi-core processing.
- Format-aware parsing (CSV columns, JSON fields) — lines only.

## What changes

- New `slugifyStream(readable, writable, opts)` API.
- New `--file` / `--stdin` CLI paths.

## What does not change

- The existing `slugify(string)` and `unslugify(string)` functions —
  identical signatures, identical behaviour. Existing callers are
  untouched.
- The Unicode handling decided in ADR-002.

## Affected assumptions

- Original PRD assumed "one string in, one string out." This addendum
  adds a streaming shape alongside it; the original assumption still
  holds for the single-string API.

## ADR impact

- **Create:** ADR-003 — streaming API shape (back-pressure, error
  propagation, partial-line handling at chunk boundaries).
- **Revise (via supersession):** none.
- **No change but relevant:** ADR-001 (public API surface) — extended,
  not contradicted.

## User stories

- As a data engineer, I run `slug-utils --file export.csv` and get a
  slugified file without writing a loop.

## Requirements

- Bounded memory on inputs larger than RAM.
- A line that fails to slugify reports its line number and does not
  abort the stream.

## Migration and compatibility notes

- Purely additive; no migration. Minor version bump (new capability,
  no break) per the kit's versioning policy.

## Issue decomposition

1. ADR-003 streaming API shape *(no deps; non-goal: CLI)*.
2. `slugifyStream` core *(deps: #1; non-goal: CLI flags)*.
3. `--file` / `--stdin` CLI wiring *(deps: #2)*.

## Success metrics

- A 1M-line file slugifies in bounded memory.
- Existing single-string tests still pass unchanged.

## Open questions

- Default chunk size? (TBD — benchmark in #2.)
```

## Handoff

The addendum names ADR-003 as needed → next step `/adr-writer` for the
streaming API shape, then `/issue-planner` to file the three decomposed
issues. The original `design/prd.md` was read but never modified.
