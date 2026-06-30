<!--
  Template: Milestone summary
  Filled by: /milestone-summary (ADR-037)
  Output in a target project: design/milestones/<N>-<slug>.md

  Purpose. A retrospective artefact for a closed (or near-closing)
  milestone: what shipped, which ADRs were adopted, what's deferred,
  and any lessons worth carrying forward. ADR-037 makes this a
  routine artefact rather than a manual exercise.

  Default boundary. Per ADR-032, one milestone == one phase. Some
  projects bundle multiple phases into one milestone; the overview
  zone names the phase range so readers can tell.

  Marker fences. Each zone is wrapped by HTML-comment fences
  (<!-- summary:<zone>:start --> / <!-- summary:<zone>:end -->).
  /milestone-summary rewrites only the bytes between the fences for
  zones it owns; editorial commentary outside the fences is preserved
  on re-run. The `lessons` zone is intentionally skill-skipped — the
  user authors it.

  Re-run safety. /milestone-summary writes once by default and
  refuses to overwrite an existing summary unless --overwrite is
  passed. The lessons zone is preserved verbatim across re-runs.
-->

# {{MILESTONE_TITLE}} — Summary

**Milestone:** {{MILESTONE_TITLE}} (#{{MILESTONE_NUMBER}})
**Closed:** {{YYYY-MM-DD}}
**Source:** `gh` is canonical; this file is a retrospective snapshot.

<!-- summary:overview:start -->

## Overview

{{One short paragraph: the milestone's goal, the phase or phase range
it covers, the start and end tag references (e.g. `v0.2.0..v0.3.0`),
and the headline deliverable.}}

<!-- summary:overview:end -->

<!-- summary:shipped:start -->

## What shipped

Merged PRs in the milestone's tag range, grouped by conventional-
commit prefix. Each entry is one line: PR number, ADR if any,
one-line summary.

### Features

- {{#PR — ADR-NNN — one-line summary}}

### Fixes

- {{#PR — one-line summary, or `none`}}

### Docs / chores

- {{#PR — one-line summary, or `none`}}

<!-- summary:shipped:end -->

<!-- summary:adrs:start -->

## ADRs adopted

ADRs accepted in the milestone's date range, with file links.

- [`adr-NNN-<slug>.md`](../adr/adr-NNN-<slug>.md) — {{one-line decision summary}}

<!-- summary:adrs:end -->

<!-- summary:deferred:start -->

## Deferred / moved out

Issues that were assigned to this milestone but did not ship — moved
to a later milestone, descoped, or left open. Empty if everything
shipped.

- {{#NN — short title — moved to {{next-milestone}} / descoped / open}}

<!-- summary:deferred:end -->

<!-- summary:lessons:start -->

## Lessons learned

<!-- TODO: authored by the user. /milestone-summary skips this zone
     on every run — even with --overwrite — so notes here survive. -->

{{One or two short paragraphs: what worked, what surprised us, what
to change next milestone. Keep it tight; the goal is a usable
retrospective, not an essay.}}

<!-- summary:lessons:end -->

<!-- summary:next:start -->

## Next

{{One short paragraph: the next milestone (or "release pending"),
the open thread to pick up, and any cross-milestone follow-up. The
canonical pointer is `design/state.md` — this is the human-readable
companion.}}

<!-- summary:next:end -->
