---
name: start
description: Inspect project state and recommend (or, for local cat-1 steps, invoke) the next skill to run. The kit's front door and "what now?" router; also answers to /next. Use at the start of a session, after finishing a step, or any time the next move is unclear. Reads design/state.md (esp. the next-action zone) and project artefacts; never mutates GitHub.
permission-category: 1  # substitutable — reads state and recommends; any skill it invokes enforces its own category, per workflow-guide §7
inputs:
  - name: "goal"
    required: false
    description: "Optional one-line hint about what the user wants to do next (e.g. 'plan a feature', 'ship')"
outputs:
  - artefact: "(routing recommendation)"
    description: "The next skill to run, why, and the active operating mode; may directly invoke a cat-1 target"
next:
  - skill: idea-to-prd
    when: "the project has only an idea (no PRD)"
  - skill: prd-to-mvp
    when: "a normalized PRD exists but no MVP"
  - skill: prepare-issue
    when: "an issue is selected but has no prepared prompt"
  - skill: claude-issue-executor
    when: "a prepared prompt exists on disk"
  - skill: pr-review-packager
    when: "a branch is ahead of main and ready to ship"
  - skill: feature-prd
    when: "the user wants a major feature update on a mature project"
---

# start

The kit's front door. `/start` (and its alias `/next`) answers one
question — **what do I run now?** — by inspecting project state and
recommending the next appropriate skill, rather than making the user
remember the full skill inventory. It is the router half of the
unified control model in
ADR-048; the
canonical model lives in
[`docs/workflow-control.md`](../../docs/workflow-control.md).

## When to use this skill

- **At the start of a session**, to be told where the project stands
  and what to do next.
- **After finishing a step**, instead of recalling the handoff.
- **Any time the next move is unclear.**

`/start` and `/next` are the same skill. Use `/start` on a fresh or
empty project (it doubles as onboarding); use `/next` mid-flow. Either
name works at any time.

If you already know the exact skill you want, just run it — `/start` is
a convenience, not a required gateway.

## What this skill does not do

- Does not mutate GitHub. It reads state and recommends; it never opens
  PRs, creates issues, pushes, or closes milestones.
- Does not duplicate `/resume`. `/resume` *briefs* you from
  `design/state.md`; `/start` *routes* you to the next skill. On a
  project with a healthy `state.md`, `/start` reads the same file —
  especially the `next-action` zone — and turns it into a recommended
  action. `/start` works even when `state.md` is absent by inspecting
  artefacts directly.
- Does not run cat-3 skills for you. When the next step is cat-3
  (`/issue-planner`, `/pr-review-packager`, `/release`,
  `/complete-milestone`), `/start` recommends it and explains the
  explicit-approval gate; it does not invoke it.

## Inputs

- **Optional:** a one-line `goal` hint ("plan a feature", "ship what I
  have", "close the milestone"). When given, it biases the
  recommendation toward the matching verb in the
  [verb layer](../../docs/workflow-control.md#3-the-human-facing-verb-layer).
- **Read (never modified):** `design/state.md` (especially the
  `next-action` zone), `design/prd.md`, `design/prd-normalized.md`,
  `design/mvp.md`, `design/build-out-plan.md`, `prompts/issue-*.md`,
  and — read-only — `git status` / `git log main..HEAD` and open PRs
  via `gh`.

## Output

A short brief naming:

1. **Active operating mode** — `interactive` (default), `assisted`, or
   `autonomous`, per
   [`docs/workflow-control.md` §1](../../docs/workflow-control.md#1-operating-modes).
   State it so the session's pre-authorisation is visible in one place.
2. **Where the project stands** — one line.
3. **Recommended next step** — the verb and its underlying skill, with a
   one-line reason.
4. **What it will do next** — for a cat-1 target, offer to invoke it
   directly; for cat-2/cat-3, recommend it and name the approval gate.

## Routing protocol

Resolve the **first** matching rule, top to bottom. Prefer the
`next-action` zone when it is present and actionable.

1. **`next-action` zone present and `skill` ≠ `none`.** If
   `design/state.md` has a `state:next-action` zone with a concrete
   skill and its `preconditions` hold and `blocked-by` is `none`,
   recommend that skill with those args. If a precondition is unmet or
   `blocked-by` is set, report the blocker instead and fall through to
   the artefact rules below for an alternative.
2. **No `design/prd.md` and no scoping artefacts.** Project is just an
   idea → recommend `/idea-to-prd` (the `/start`-as-onboarding case).
3. **A PRD exists but not normalized** (`design/prd.md` without
   `design/prd-normalized.md`) → recommend `/prd-normalizer`.
4. **Normalized PRD but no MVP** (`design/prd-normalized.md` without
   `design/mvp.md`) → recommend `/prd-to-mvp`.
5. **MVP and build-out plan but no GitHub backlog** → recommend
   `/backlog` (`/issue-planner`, cat-3 — recommend, do not invoke).
6. **A branch is ahead of `main`** (`git log main..HEAD` non-empty) and
   no open PR for it → recommend `/ship` (`/pr-review-packager`, cat-3).
7. **A prepared prompt exists** (`prompts/issue-NNN-*.md`) with no
   branch yet → recommend `/work` (`/claude-issue-executor`).
8. **An open issue is selected but has no prompt** → recommend
   `/prepare-issue NNN`.
9. **A milestone's issues are all merged** → recommend
   `/finish-milestone`.
10. **Mature project, user hint names a major feature** → recommend
    `/feature` (`/feature-prd`, see
    ADR-049).

### When the next step is ambiguous

If two or more rules plausibly apply and the `goal` hint does not break
the tie, **ask exactly one clarifying question** — e.g. *"You have a
prepared prompt for #95 and an open branch for #94 — continue #94's PR,
or start #95?"* — rather than dumping the full skill list. One question,
then route.

## Invoking vs. recommending

- **cat-1 target** (`/idea-to-prd`, `/prd-normalizer`, `/prd-to-mvp`,
  `/prepare-issue`, `/resume`, `/feature-prd`, …): offer to invoke it
  directly in the same session, honouring the active operating mode.
- **cat-2 target** (`/claude-issue-executor`): recommend it; it runs its
  own plan-mode rhythm and approval gate.
- **cat-3 target** (`/issue-planner`, `/pr-review-packager`, `/release`,
  `/complete-milestone`): recommend it and name the explicit-approval
  gate. Never invoke a cat-3 skill from `/start`.

See [`example.md`](example.md) for worked routings.

## Handoff

`/start` ends by naming the recommended skill. The user runs it (or
approves `/start` invoking a cat-1 target). After that skill finishes,
running `/start` again advances to the next step.
