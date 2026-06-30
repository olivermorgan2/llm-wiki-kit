# resume — worked example

Two short walk-throughs: the happy path (state.md exists and is
fresh) and the fallback path (state.md missing or stale).

---

## Happy path

### Input: `design/state.md`

```markdown
# acme-platform — State

**Last updated:** 2026-04-30
**Source of truth:** `gh` is canonical for issue/PR status; this file is a pointer.

<!-- state:phase:start -->

## Phase

Phase 2 — Editor

<!-- state:phase:end -->

<!-- state:in-flight:start -->

## In-flight issue

- **Issue:** #44
- **Prompt:** `prompts/issue-044-design-state-md-plus-resume-and-pause-skills.md`
- **Branch:** `feat/v-next-adr-035-state-md-resume-pause`
- **Status:** executing

<!-- state:in-flight:end -->

<!-- state:recent:start -->

## Recent work

- #54 — ADR-033 — `/clarify` skill producing `design/decisions.md`
- #53 — ADR-039 — plan-mode rhythm wired into executor
- #51 — ADR-039 — plan-mode-for-significant-tasks ADR accepted
- #50 — ADR-032 — implementation phases (Phase column, build-out blocks)
- #49 — ADR-031 — `/planning` skill + planning-template.md

<!-- state:recent:end -->

<!-- state:blockers:start -->

## Blockers

none

<!-- state:blockers:end -->

<!-- state:continue-here:start -->

## Continue here

Resume executing `prompts/issue-044-design-state-md-plus-resume-and-pause-skills.md`
on branch `feat/v-next-adr-035-state-md-resume-pause`. Next concrete step is
commit 4 (wire state.md updates into prepare/execute/review).

<!-- state:continue-here:end -->
```

### Output: `/resume` brief

```
**Phase:** Phase 2 — Editor
**In-flight:** #44 (executing) — prompt prompts/issue-044-design-state-md-plus-resume-and-pause-skills.md, branch feat/v-next-adr-035-state-md-resume-pause
**Recent:**
- #54 — ADR-033 — `/clarify` skill producing `design/decisions.md`
- #53 — ADR-039 — plan-mode rhythm wired into executor
- #51 — ADR-039 — plan-mode-for-significant-tasks ADR accepted
- #50 — ADR-032 — implementation phases (Phase column, build-out blocks)
- #49 — ADR-031 — `/planning` skill + planning-template.md
**Blockers:** none
**Continue here:** Resume executing prompts/issue-044-design-state-md-plus-resume-and-pause-skills.md on branch feat/v-next-adr-035-state-md-resume-pause. Next concrete step is commit 4 (wire state.md updates into prepare/execute/review).
```

No `gh` calls were made. The user has everything needed to pick up
where the prior session left off.

---

## Fallback path

### Input

`design/state.md` does not exist (a fresh project that has not yet
adopted ADR-035 or that has been pruned).

### Skill behaviour

The skill takes the `gh` fallback. It runs:

```
gh pr list --state open --limit 5 --json number,title,url
gh pr list --state merged --limit 5 --json number,title,url,mergedAt
gh issue list --state open --limit 5 --json number,title,url
```

### Output

```
(state.md not found — using gh fallback. Consider /pause to seed design/state.md.)

**Phase:** unknown — no design/state.md
**In-flight:** none observed (no open PR)
**Recent:**
- #54 — `/clarify` skill producing design/decisions.md (merged 2026-04-30)
- #53 — plan-mode rhythm in claude-issue-executor (merged 2026-04-30)
- #51 — ADR-039 plan-mode-for-significant-tasks (merged 2026-04-30)
**Blockers:** unknown — no design/state.md
**Continue here:** open issues: #44, #45, #43, #47, #46. Pick one and run `/prepare-issue NNN`.
```

The user is briefed but warned that the file should be seeded —
the next `/pause` invocation will do that.
