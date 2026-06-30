# Your first PR in 15 minutes

This is the shortest path from a new, kit-installed project to one
merged pull request. It uses **five slash commands** and a single toy
idea. No ADRs, no flags, no reference reading required — just follow
the steps in order.

By the end you will have:

- a scoped MVP and a one-phase build-out plan,
- one GitHub issue,
- one feature branch with a small, tested change,
- one pull request you merge yourself.

## Before you start (2 min)

You need the kit already installed in a fresh GitHub project. If you
have not done that yet, follow the [Quick start](../README.md#quick-start)
first, then come back here. Confirm you are ready:

```bash
ls .claude/skills     # lists idea-to-prd, prd-to-mvp, … (the kit is installed)
gh auth status        # "Logged in to github.com" (GitHub CLI is ready)
```

Open Claude Code in the project root (`claude`). Run every command
below **inside Claude Code**, not in a plain shell.

> The example idea used throughout: **"A tiny command-line tool that
> prints a friendly greeting."** Swap in your own one-line idea if you
> like — the steps are identical.

## Step 1 — Shape the idea into a PRD (3 min)

```text
/prd-normalizer
```

Paste your one-line idea when asked. The skill writes a normalized PRD
to `design/prd-normalized.md`. That single file is what the next step
reads, so you never have to repeat the idea.

> **Only have a vague thought, not even a sentence?** Run
> `/idea-to-prd` first to draft `design/prd.md`, then run
> `/prd-normalizer`. For this tutorial the one-liner above is enough,
> so we skip it.

**Checkpoint (~3 min in):** `design/prd-normalized.md` exists.

## Step 2 — Scope the MVP (3 min)

```text
/prd-to-mvp
```

This produces `design/mvp.md` (what's in scope, what's out, what
success looks like) and `design/build-out-plan.md` (the work split
into phases). For a toy idea you'll get a single small phase — exactly
what you want for a first PR. Approve the output when prompted.

**Checkpoint (~6 min in):** `design/mvp.md` and
`design/build-out-plan.md` exist.

## Step 3 — Create the backlog on GitHub (2 min)

```text
/issue-planner
```

The skill drafts a small batch of issues from the build-out plan and
shows them for your approval. It will **not** create anything on
GitHub until you reply `yes`. Approve, and it files the issues and
sets up a project board. Note the number of the first issue (say
`#1`).

**Checkpoint (~8 min in):** `gh issue list` shows your new issue(s).

## Step 4 — Implement one issue (4 min)

```text
/claude-issue-executor 1
```

Use the issue number from step 3. The executor prepares a prompt
automatically (you do **not** run `/prepare-issue` yourself), proposes
a short plan, waits for your approval, then creates a feature branch
and implements the change in small commits with a test alongside.
Approve the plan and let it work.

**Checkpoint (~12 min in):** a feature branch exists with one or two
commits.

## Step 5 — Open the pull request (2 min)

```text
/pr-review-packager
```

This fills the PR body (`Closes #1`, a change summary), shows it for
approval, and — after you reply `yes` — opens the PR with
`gh pr create`. Open the PR link it prints, give it a final look, and
merge:

```bash
gh pr merge --squash --delete-branch
```

**Done (~15 min in):** you have created and merged your first PR with
the kit. 🎉

## The five commands, at a glance

| # | Command | Produces |
|---|---|---|
| 1 | `/prd-normalizer` | `design/prd-normalized.md` |
| 2 | `/prd-to-mvp` | `design/mvp.md`, `design/build-out-plan.md` |
| 3 | `/issue-planner` | GitHub issues + project board |
| 4 | `/claude-issue-executor <N>` | feature branch + commits + test |
| 5 | `/pr-review-packager` | the pull request |

Deferred on purpose, so the happy path stays at five commands:

- `/idea-to-prd` — only if you don't yet have a one-line idea written.
- `/adr-writer` — only when a real architectural decision comes up.
  A toy greeting tool doesn't need one.
- Flags like `--granularity` — useful on bigger projects; ignore them
  here.

## If something fails

- A slash command doesn't autocomplete, skills aren't found, the
  project board fails to create, or a re-run changes files
  unexpectedly → see
  [`docs/troubleshooting.md`](troubleshooting.md), organized by the
  exact symptom you see.
- You want the full, non-toy walkthrough (releases, milestones,
  ongoing development) → see
  [`docs/workflow-guide.md`](workflow-guide.md).

## What next

You've exercised the core loop the kit runs many times over a
project's life: **scope → backlog → implement → ship.** From here:

- File and work the next issue with `/claude-issue-executor <N>`.
- Capture a real decision with `/adr-writer` when one arises.
- Cut a release with `/release` when `main` is worth shipping.

The end-to-end reference for all of it is
[`docs/workflow-guide.md`](workflow-guide.md).
