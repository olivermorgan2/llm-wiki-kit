# audit-milestone — worked example

A walkthrough of the milestone lifecycle chain
(`/audit-milestone` → `/milestone-summary` → `/complete-milestone`
→ optional `/release`) on the
[`phased-podcast-pipeline`](../../examples/projects/phased-podcast-pipeline/)
example project. Two scenarios are covered:

1. **Pass:** Phase 1 (Foundation) is fully done and ready to close.
2. **Fail:** Phase 2 (Publishing) has loose ends — one open issue
   and one ADR with no merged PR.

The same chain runs in both cases; only the audit result differs,
and that difference shapes the user's decision to proceed.

---

## Setup

`phased-podcast-pipeline` ships three GitHub milestones (one per
phase, per ADR-032 / `/issue-planner`):

| Milestone | Number | State |
|---|---|---|
| Phase 1 — Foundation | #1 | open, all issues closed |
| Phase 2 — Publishing | #2 | open, one issue still open |
| Phase 3 — Distribution | #3 | open, no work started |

For the worked example, the user is closing out Phase 1 and
sanity-checking Phase 2.

---

## Scenario 1 — Phase 1 closes cleanly (pass)

### Step 1: `/audit-milestone "Phase 1 — Foundation"`

```
$ /audit-milestone "Phase 1 — Foundation"
```

Skill output:

> Resolving milestone… Phase 1 — Foundation (#1), state: open.
>
> Issues: 3 closed, 0 open.
> ADRs referenced: ADR-001 (storage format).
>   ✓ design/adr/adr-001-storage-format.md exists.
>   ✓ Merged PR found: #14 — feat(ingest): add SQLite storage (ADR-001, #4).
> Phase status: Phase 1 row in build-out-plan.md → `released v0.1.0`.
>
> ✅ audit-milestone PASS — Phase 1 — Foundation (#1)
> 3/3 issues closed · 1/1 ADR linked · 1/1 phase released.

The user proceeds.

### Step 2: `/milestone-summary 1`

```
$ /milestone-summary 1
```

The skill resolves the tag range (`<initial>..v0.1.0`), pulls
merged PRs and accepted ADRs, and renders the would-be content of
`design/milestones/1-phase-1-foundation.md`. After the user types
`yes`, the file is written:

```markdown
# Phase 1 — Foundation — Summary

**Milestone:** Phase 1 — Foundation (#1)
**Closed:** 2026-04-30
...

<!-- summary:overview:start -->

## Overview

Phase 1 settles the storage format and the ingest CLI. Range:
initial commit → v0.1.0. Headline deliverable: `pipeline ingest`
round-trips 50 sample notes without loss.

<!-- summary:overview:end -->

<!-- summary:shipped:start -->

### Features
- #14 — ADR-001 — feat(ingest): add SQLite storage
- #15 — feat(ingest): pipeline ingest CLI
- #16 — feat(schema): note structure validator

### Fixes
- none

### Docs / chores
- #17 — docs(schema): document note structure

<!-- summary:shipped:end -->

<!-- summary:adrs:start -->
- [`adr-001-storage-format.md`](../adr/adr-001-storage-format.md) —
  Use SQLite for the note store; one row per note; FTS5 for search.
<!-- summary:adrs:end -->

<!-- summary:deferred:start -->
- none
<!-- summary:deferred:end -->

<!-- summary:lessons:start -->
<!-- TODO: authored by the user. /milestone-summary skips this zone... -->
{{One or two short paragraphs...}}
<!-- summary:lessons:end -->

<!-- summary:next:start -->
Phase 2 — Publishing kicks off next. State pointer:
`design/state.md`. First issue: TTS pipeline (#18).
<!-- summary:next:end -->
```

The user opens the summary file in their editor, replaces the
lessons-zone placeholder with one paragraph of real retrospective,
and saves.

### Step 3: `/complete-milestone 1 --release`

```
$ /complete-milestone 1 --release
```

The skill chains `/audit-milestone 1` (already passed), reads the
summary file (mtime confirms recent edit), and renders the close
plan:

> Plan:
> - Close milestone Phase 1 — Foundation (#1) on GitHub.
> - Update design/state.md:
>   - in-flight → none
>   - recent → prepend "Milestone closed: Phase 1 — Foundation (#1) — 2026-04-30"
>   - continue-here → "Run `/release --milestone-phase=1`"
> - Chain `/release --milestone-phase=1`.
>
> Type `yes` to close milestone Phase 1 — Foundation (#1)…

The user types `yes`. The skill runs:

```
gh api repos/:owner/:repo/milestones/1 -X PATCH -f state=closed
[updates design/state.md zones]
/release --milestone-phase=1
```

`/release` has its own approval gate; the user types `yes` again
inside it. Tag `v0.1.0` is cut, the GitHub Release is published,
and `design/build-out-plan.md` Phase 1 row updates to
`released v0.1.0`.

Final report:

> ✅ Milestone Phase 1 — Foundation closed.
> Summary: design/milestones/1-phase-1-foundation.md
> State: in-flight cleared, recent prepended, continue-here updated.
> Release: https://github.com/.../releases/tag/v0.1.0

---

## Scenario 2 — Phase 2 has gaps (fail)

### Step 1: `/audit-milestone 2`

```
$ /audit-milestone 2
```

Skill output:

> Resolving milestone… Phase 2 — Publishing (#2), state: open.
>
> Issues: 4 closed, 1 open.
> ADRs referenced: ADR-002 (publishing pipeline).
>   ✓ design/adr/adr-002-publishing-pipeline.md exists.
>   ✗ No merged PR found mentioning ADR-002.
> Phase status: Phase 2 row in build-out-plan.md → `in-progress`.
>
> ❌ audit-milestone FAIL — Phase 2 — Publishing (#2)
>
> Issues:
> - Issue #20 still open — close it, merge its PR, or move it to a later milestone.
>
> ADRs:
> - ADR-002 referenced but no merged PR mentions it. Confirm the
>   implementation PR was merged with `(ADR-002, #NN)` in its title or body.
>
> Phases:
> - Phase 2 (Publishing) status is `in-progress` in `design/build-out-plan.md`.
>   Run `/release --milestone-phase=2` to mark it `released vX.Y.Z`,
>   or update the row manually.

### Step 2: user resolves the gaps

Three concrete actions:

1. Open issue #20 — the user finishes the work, opens a PR with
   the title `feat(metadata): extract from frontmatter (ADR-002, #20)`,
   and merges it.
2. The merge of #20's PR satisfies the ADR-002 linkage gap (same
   token in the title).
3. The phase row will update to `released v0.2.0` automatically
   when `/release --milestone-phase=2` runs at close time — no
   pre-action needed.

The user re-runs `/audit-milestone 2` and gets a pass.

### Alternative: proceed despite gaps

If the user wants to close Phase 2 with the gaps still present (e.g.
descope #20 and move it to Phase 3), they can run
`/complete-milestone 2` — the audit chains and surfaces the gaps,
the close plan shows the gap count, and the approval gate accepts
the explicit `yes`. The milestone closes and #20 stays open in
Phase 3's milestone (the user moves it manually with
`gh issue edit 20 --milestone "Phase 3 — Distribution"`).

ADR-037 is explicit on this: audit warns, the user decides. The
skill never blocks the close.

---

## What this example illustrates

- The chain's three skills are **composable**: each runs standalone,
  and `/complete-milestone` chains the previous two as a convenience.
- The chain is **advisory**, not gating: gaps surface clearly but
  the user has the final say.
- The summary file's `lessons` zone is **user-authored** and
  preserved across re-runs — even with `--overwrite`, the lessons
  block survives.
- `/release --milestone-phase=N` is the **single integration point**
  with the existing release ceremony. No new logic in `/release`;
  the milestone skills compose on top of it.
- The state.md archive on close means the next session's `/resume`
  brief is clean: in-flight cleared, recent prepended, continue-here
  pointing at the next concrete action.

For the skill specs themselves, see
[`SKILL.md`](SKILL.md) (audit),
[`../milestone-summary/SKILL.md`](../milestone-summary/SKILL.md),
and [`../complete-milestone/SKILL.md`](../complete-milestone/SKILL.md).
