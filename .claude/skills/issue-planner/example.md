# issue-planner — worked example

A single run of `/issue-planner` on a small Pace Drift build-out plan.
Shows how the skill parses the plan, drafts issues, presents them for
approval, and creates them along with a Project board.

---

## 1. Input — the plan files

### `design/mvp.md` (excerpt)

```markdown
# Pace Drift — MVP

## Product principles

1. Local session only — no server-side persistence.
2. Verbatim data — no AI smoothing.
3. Laptop-first.

## MVP scope

### In scope

- Upload a GPX file.
- Compute pace drift against a target pace.
- Render a pace drift chart.

### Out of scope

- User accounts.
- Multi-file comparisons.
```

### `design/build-out-plan.md` (excerpt)

```markdown
## Milestone recommendation

| Milestone | Focus |
|---|---|
| M1 — foundations | Repo, ADRs, parser spike |
| M2 — core | Drift calculation + chart |
| M3 — polish | Error handling, docs |

## Initial issue backlog

### M1 — foundations

- Set up repo structure and ADRs
- Browser-side GPX parser spike

### M2 — core

- Pace drift calculation
- Render drift chart
```

### `design/adr/`

- `adr-001-gpx-parsing-location.md` (title: *Parse GPX files in the
  browser, not on a server*)
- `adr-002-test-framework.md` (title: *Use Vitest for tests*)

## 2. The user runs the skill

```
/issue-planner
```

## 3. Preflight — what the skill reports

```
Parsed design/build-out-plan.md and design/mvp.md.

4 issues drafted across 2 milestones:
  M1 — foundations: 2 issues
  M2 — core:        2 issues

ADR matches:
  "Browser-side GPX parser spike" -> ADR-001
  (others: none)

Labels needed: feature, design, infra
  Already exist: feature
  Missing:       design, infra  <-- will prompt

Milestones needed: "M1 — foundations", "M2 — core"
  Already exist: (none)
  Missing:       both  <-- will prompt

Idempotency check: no existing issues with matching titles.

Warnings: none.
```

## 4. The full batch shown for approval

(Abridged — the real output shows all four issues in full.)

```
Issue 1 of 4
  Title:     Set up repo structure and ADRs
  Labels:    feature, infra
  Milestone: M1 — foundations
  ADR:       none (project-setup work; no ADR applies)
  Body:
    ## Summary
    Establish the repo skeleton...
    ## ADR
    none — project-setup work.
    ## Goal
    A repo a contributor can clone and run.
    ## Why it matters
    Every later issue depends on this baseline.
    ## Tasks
    - [ ] Create design/, src/, tests/
    - [ ] Add ADR-000 template
    - [ ] Commit baseline
    ## Acceptance criteria
    - `git clone` + `npm install` succeeds locally.
    ## Notes
    Labels: feature, infra
    Milestone: M1 — foundations

Issue 2 of 4
  Title:     Browser-side GPX parser spike (ADR-001)
  Labels:    feature, design
  Milestone: M1 — foundations
  ADR:       design/adr/adr-001-gpx-parsing-location.md
  Body: ...
```

Final line:

```
Reply `yes` to create these 4 issues, the 2 missing milestones, the 2
missing labels, and the Project board "Pace Drift — M1 — foundations".
Reply `no` or suggest edits to revise.
```

## 5. User approves

User types `yes`.

## 6. Creation sequence

```
$ gh label create design --color d4c5f9
$ gh label create infra --color fbca04

$ gh api repos/:owner/:repo/milestones \
    -f title='M1 — foundations' \
    -f description='Repo, ADRs, parser spike'
$ gh api repos/:owner/:repo/milestones \
    -f title='M2 — core' \
    -f description='Drift calculation + chart'

$ gh issue create \
    --title 'Set up repo structure and ADRs' \
    --body-file /tmp/issue-1.md \
    --label feature,infra \
    --milestone 'M1 — foundations'
# -> https://github.com/user/pace-drift/issues/1

$ gh issue create \
    --title 'Browser-side GPX parser spike (ADR-001)' \
    --body-file /tmp/issue-2.md \
    --label feature,design \
    --milestone 'M1 — foundations'
# -> https://github.com/user/pace-drift/issues/2

# ... issues 3 and 4 ...

$ gh project create --owner @me --title 'Pace Drift — M1 — foundations'
# -> project number 7, URL https://github.com/users/user/projects/7

$ gh project item-add 7 --owner @me --url .../issues/1
$ gh project item-add 7 --owner @me --url .../issues/2
$ gh project item-add 7 --owner @me --url .../issues/3
$ gh project item-add 7 --owner @me --url .../issues/4
```

## 7. Report back to user

```
Created 4 issues:
  #1 Set up repo structure and ADRs              (M1 — foundations)
  #2 Browser-side GPX parser spike (ADR-001)     (M1 — foundations)
  #3 Pace drift calculation                      (M2 — core)
  #4 Render drift chart                          (M2 — core)

Created Project: Pace Drift — M1 — foundations
  https://github.com/users/user/projects/7

Next steps:
  - Run /prepare-issue <N> to generate a per-issue prompt under notes/.
  - Inspect: gh issue list --milestone 'M1 — foundations'
  - Inspect board: gh project view 7 --owner @me --web
```

## 8. What a dry-run looks like

Same as steps 1–4, but step 4's final line is:

```
Dry-run: no changes made. Re-run without --dry-run to create these
4 issues + 2 milestones + 2 labels + the Project board.
```

No `gh create` / `gh project create` calls are made. Read-only calls
(`gh label list`, `gh api .../milestones`, `gh issue list`) are still
made so the preview is accurate.
