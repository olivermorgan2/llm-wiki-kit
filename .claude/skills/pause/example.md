# pause — worked example

A walk-through of `/pause --handoff` invoked mid-session on the same
fictional `acme-platform` project used in
[`skills/resume/example.md`](../resume/example.md).

---

## 1. Pre-existing state

`design/state.md` was last touched at the start of the session:

```markdown
**Last updated:** 2026-04-29

<!-- state:in-flight:start -->

## In-flight issue

- **Issue:** #44
- **Prompt:** `prompts/issue-044-design-state-md-plus-resume-and-pause-skills.md`
- **Branch:** `feat/v-next-adr-035-state-md-resume-pause`
- **Status:** prepared

<!-- state:in-flight:end -->
```

The session has since:

- created the branch and made three commits,
- discovered a sibling concern (ADR-038 boundary) and decided to
  defer it,
- noted that `notes/` was missing on a fresh checkout (mild blocker).

The user invokes `/pause --handoff` because the context window is
near exhaustion.

## 2. Skill behaviour

### Step 3 — Phase

`design/build-out-plan.md` shows three phases; Phase 2 (Editor) is
the earliest unfinished. The skill writes `Phase 2 — Editor` to the
`phase` zone (no change from prior run).

### Step 4 — In-flight issue

The user did not pass an explicit number. The most recent prompt
file is `prompts/issue-044-...md`, mtime today, branch
`feat/v-next-adr-035-state-md-resume-pause` exists. The skill keeps
issue `#44` and updates `Status` from `prepared` → `executing`.

### Step 5 — Recent work

`gh pr list --state merged --limit 5 --json ...` returns:

- #54 (2026-04-30) — `/clarify` skill producing `design/decisions.md`
- #53 (2026-04-30) — plan-mode rhythm in claude-issue-executor
- #51 (2026-04-30) — ADR-039 plan-mode-for-significant-tasks
- #50 (2026-04-30) — implementation phases (ADR-032)
- #49 (2026-04-29) — `/planning` skill (ADR-031)

ADR numbers are extracted from titles. The zone is rewritten with
the five entries.

### Step 6 — Blockers

The user types: "notes/ was missing on fresh checkout — minor".
Skill writes one line; the prior `none` is replaced.

### Step 7 — Continue here

Skill proposes:

> Resume executing
> `prompts/issue-044-design-state-md-plus-resume-and-pause-skills.md`
> on branch `feat/v-next-adr-035-state-md-resume-pause`. Three
> commits landed (prompt, state-template, /resume); next is commit 4
> wiring state.md updates into prepare/execute/review.

User accepts.

### Step 9 — Line-cap check

`wc -l design/state.md` → 64 lines. Under cap. Silent.

### Step 10 — `--handoff`

`notes/handoff-2026-04-30.md` is written, modelled on
`notes/handoff-2026-04-30-v-next-batch-resume.md`. Sections:

- **Date** and **Where we left off** — one paragraph
- **State of v-next** — three tables (merged / open / queued)
- **Recommended next step on resume** — numbered list
- **Standing rules** — pointer to auto-memory
- **Working tree state at handoff** — branch name, untracked files
- **What NOT to do on resume**

## 3. Output diff (excerpt)

```diff
-**Last updated:** 2026-04-29
+**Last updated:** 2026-04-30
@@ -- state:in-flight zone
-- **Status:** prepared
+- **Status:** executing
@@ -- state:recent zone (5 entries rewritten)
@@ -- state:blockers zone
-none
+notes/ was missing on fresh checkout — minor
@@ -- state:continue-here zone
-Run `/prepare-issue 44` to write the prompt.
+Resume executing prompts/issue-044-... [as above]
```

Plus a new file: `notes/handoff-2026-04-30.md` (~80 lines).

## 4. Report

```
Wrote design/state.md (4 zones changed: in-flight, recent, blockers, continue-here)
Wrote notes/handoff-2026-04-30.md
File is 64 lines (cap 100).
```

The user can now `/clear` safely. The next session runs `/resume`
to pick up exactly where this one left off.
