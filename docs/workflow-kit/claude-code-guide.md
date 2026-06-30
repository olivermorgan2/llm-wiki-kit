# Claude Code guide

How to use Claude Code effectively with the workflow kit: plan mode,
invoking skills via `/skill-name`, the approve-then-implement loop, and
the pitfalls that bite new users.

This guide covers the **interaction patterns**. For the end-to-end
process (idea → ADR → issue → PR → release) see
[`workflow-guide.md`](workflow-guide.md).
For what each skill does, read its own `SKILL.md` under
[`../skills/`](../skills/README.md), or use the functional reference in
[`skills.md`](skills.md).

## What this guide is

A practical reference for anyone running Claude Code inside a target
project that has the kit installed. It assumes you have already
installed the kit per [`install.md`](install.md) and that your working
directory is the target project root.

It is not a tutorial on Claude Code itself — see Anthropic's docs for
that. It is also not a skill-by-skill walkthrough — each skill's
`SKILL.md` covers its own inputs, outputs, and edge cases.

## Prerequisites

Before the rest of this guide makes sense, you need:

- **Claude Code installed** and on your `PATH`. Verify with
  `claude --version`.
- **GitHub CLI authenticated.** Verify with `gh auth status`. The kit's
  skills call `gh` for issue, PR, and release work.
- **The kit installed into the target project.** Verify with
  `ls .claude/skills` — you should see `idea-to-prd`, `adr-writer`,
  `claude-issue-executor`, and the rest. If the directory is empty,
  run through [`install.md`](install.md) first.
- **A clean working tree** before starting an implementation session.
  `git status` should report nothing staged and nothing modified.

If any of these fail, fix that first. The skills assume they all pass.

## Plan mode

**Plan mode** is Claude Code's built-in read-only mode: the assistant
investigates, reads files, and proposes a plan, but refuses to write,
edit, or run mutating commands. You leave plan mode by explicitly
approving the plan — then the assistant implements it.

This is the interaction model that every build-oriented skill in the
kit enforces at the prompt level, per
ADR-006.

### When skills use it

These skills gate on an explicit plan approval before touching disk:

- `/claude-issue-executor` — see the "plan gate" section of its
  [`SKILL.md`](../skills/claude-issue-executor/SKILL.md). This is the
  main implementation skill; it is also the strictest.
- `/prepare-issue` — proposes the filled prompt before writing it to
  `prompts/`.
- `/issue-planner` — proposes the issue batch and Project board layout
  before calling `gh`.
- `/pr-review-packager` — proposes the PR body before calling
  `gh pr create`.
- `/release` and `/changelog` — propose the notes before tagging or
  publishing.

The planning skills (`/idea-to-prd`, `/prd-normalizer`, `/prd-to-mvp`,
`/adr-writer`) are interview-shaped: they ask questions and produce a
draft in one pass, so there is no separate plan gate — the draft
itself is the artifact you review.

### How to trigger it manually

You don't need a skill to use plan mode. You can ask for it directly:

```
Before you edit anything, write a plan. Do not touch files until I approve.
```

Or use Claude Code's `--plan` / plan-mode toggle at session start.
Either approach is fine — the skills do the same thing, just under a
consistent prompt contract.

A good plan includes:

- the branch you'll create,
- every file you'll create or modify, each with a one-line reason,
- the commit sequence (message + which files land in each),
- what tests you'll add, where,
- the verification steps at the end.

If the plan is vague on any of those, ask for a rewrite before
approving.

## Invoking skills

Skills are invoked with a slash command:

```
/claude-issue-executor prompts/issue-021-claude-code-guide.md
```

The general form is `/<skill-name> [args]`. Arguments vary per skill —
most take a path or a GitHub reference, some take none. Each skill's
`SKILL.md` documents its own invocation contract under the
**Invocation contract** section.

### The full skill list

See [`../skills/README.md`](../skills/README.md) for the authoritative
list. As of this writing the kit ships:

- `/idea-to-prd`
- `/prd-normalizer`
- `/prd-to-mvp`
- `/adr-writer`
- `/issue-planner`
- `/prepare-issue`
- `/claude-issue-executor`
- `/pr-review-packager`
- `/changelog`
- `/release`
- `/workflow-docs`

### Tips

- **Run from the target project root.** Project-local skills are
  discovered relative to the working directory (see
  [`install.md`](install.md#troubleshooting)).
- **One skill per session.** Skills are designed to hand off, not chain
  inline — e.g. `/prepare-issue` writes a prompt, you close the
  session, then open a new one for `/claude-issue-executor`.
- **If a slash command doesn't resolve**, check `ls .claude/skills` —
  the directory name must match the command name exactly.

## The approve-then-implement loop

This is the core interaction pattern. It comes from
ADR-006 and is
enforced by every build skill.

The loop has five steps:

1. **Prompt parsed.** You invoke a skill with a prompt file, or paste
   a filled `prompts/issue-NNN-*.md` (copied from
   [`prompts/_template.md`](../prompts/_template.md)) manually. The
   assistant reads `CLAUDE.md`, the referenced ADR, and any files
   called out by the prompt.
2. **Plan proposed.** The assistant produces a single message
   containing a step-by-step plan. It then stops.
3. **User approves** (or asks for changes). Clarifying questions are
   not approval — the assistant should answer them and re-propose.
4. **Implement.** The assistant creates the branch, edits files,
   writes tests alongside code, commits incrementally with messages
   that reference the ADR and issue (e.g. `feat(x): add Y (ADR-003, #15)`).
5. **Evaluate.** The assistant prints an evaluation summary: what
   changed, commits made, tests run, follow-ups, and the commands you
   should run to verify.

### Example round-trip

**You:** `/claude-issue-executor prompts/issue-042-add-auth.md`

**Claude Code (plan):**
> **Plan**
>
> - Branch: `add-auth-middleware` off `main`.
> - Create `src/auth/middleware.ts` — JWT validation wrapper.
> - Modify `src/app.ts` — wire middleware into the router.
> - Create `test/auth/middleware.test.ts` — happy path, expired
>   token, missing header, malformed token.
> - Commits:
>   1. `feat(auth): add JWT middleware (ADR-003, #42)` — middleware + tests
>   2. `feat(auth): wire middleware into router (ADR-003, #42)` — app.ts
> - Verify: `npx vitest run` after each commit.
>
> Approve to proceed.

**You:** `approved, go`

**Claude Code (implement):** creates the branch, makes the edits,
runs tests, commits. Prints the evaluation summary at the end with
SHAs, test output, and the `gh pr create` command for the next step.

### What to do if the plan is wrong

Reply with the correction. The skill revises and re-proposes. Do not
approve a plan you disagree with — once edits start, undoing them is
messier than rewriting the plan.

If the plan is fundamentally off (wrong scope, wrong files, misreads
the ADR), say so plainly: "this plan isn't right, the ADR asks for X
not Y, revise." The skill will replan from scratch rather than patch
around the misread.

## Common pitfalls

Each pitfall below: what it looks like, why it bites, and how to
avoid it.

### Destructive commands on the wrong target

- **What it looks like.** Claude Code proposes `rm -rf <dir>`,
  `git reset --hard origin/main`, `git checkout .`, `git clean -fd`,
  `git push --force`, or `git branch -D <branch>` — often in a
  recovery-from-mistake context.
- **Why it bites.** These commands drop uncommitted work silently.
  On `main` or a shared branch, `--force` push overwrites other
  people's commits.
- **How to avoid.** Read every destructive command before approving.
  If the assistant proposes one, ask what it protects against —
  often a safer alternative exists (e.g. `git stash` instead of
  `git checkout .`, `git reset --mixed` instead of `--hard`). Never
  approve a force-push to `main` or `master`. The kit's
  [`CLAUDE.md`](../CLAUDE.md) template tells Claude Code to avoid
  these by default, but you are the final gate.

### Forgetting to stage files

- **What it looks like.** `git commit` runs cleanly, but a new file
  you expected is missing from the commit. `git status` shows it as
  untracked.
- **Why it bites.** The commit looks successful, the branch looks
  healthy, but the next session (or CI) fails because a file is
  unreachable.
- **How to avoid.** Check `git status` after every commit the
  assistant makes. If a new file is untracked, ask the assistant to
  stage and amend (or to add it in a follow-up commit). Skills like
  `/claude-issue-executor` stage explicitly per file, not via
  `git add -A`, so untracked files surface rather than get swept in
  silently.

### Overwriting existing work

- **What it looks like.** The assistant rewrites a file wholesale
  when the intended change was a one-line edit. Your carefully
  formatted doc comes back reformatted, or your fixture file comes
  back with invented values.
- **Why it bites.** `git diff` gets unreadable, reviewers can't tell
  what actually changed, and unrelated content drifts.
- **How to avoid.** Prefer plans that say "edit" over "rewrite."
  When you see a plan step like "update `src/x.ts`," ask which
  lines. If the assistant does rewrite a file, run `git diff` before
  approving the next step — revert unexpected changes immediately.

### Dirty tree on entry

- **What it looks like.** You start a session while `git status`
  shows modified or untracked files from earlier work.
- **Why it bites.** The session's commits mix in work that wasn't
  part of the issue, scope blurs, and reviewers get surprised in
  the PR.
- **How to avoid.** Commit, stash, or discard before invoking a
  build skill. `/claude-issue-executor` refuses to start with a
  dirty tree — that refusal is the feature, not a bug. Honor it.

### Committing directly to `main`

- **What it looks like.** You start a session forgetting to branch,
  and the first commit lands on `main`.
- **Why it bites.** Breaks [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow),
  bypasses PR review, and conflicts with branch protection if
  enabled.
- **How to avoid.** Every build skill creates a branch from `main`
  before the first edit. If you're hand-driving a session, your
  first command should be
  `git checkout main && git pull && git checkout -b <branch>`.
  If you realize mid-session that you're on `main`, stop and
  `git switch -c <branch>` before continuing — the uncommitted work
  moves with you.

### Chaining skills without a break

- **What it looks like.** You finish `/claude-issue-executor` and
  immediately invoke `/pr-review-packager` in the same session.
- **Why it bites.** The review checkpoint between implementation
  and PR disappears — the whole point of separating the two
  skills. See ADR-015.
- **How to avoid.** Close the session, review the diff yourself
  (`gh pr view --web` or `git log -p`), then open a new session
  for the next skill.

## Where to go next

- [`workflow-guide.md`](workflow-guide.md) — the end-to-end
  walkthrough this guide plugs into (idea → MVP → backlog →
  per-issue loop → release, plus the steady-state iteration loop).
- [`skills.md`](skills.md) — functional reference for every skill in
  the kit, grouped by what they do.
- [`../skills/README.md`](../skills/README.md) — source-tree index of
  all installed skills with links to each `SKILL.md`.
- [`github-setup.md`](github-setup.md) — GitHub labels, branches,
  and PR conventions the skills assume.
