# llm-wiki-kit

> _TBD_

This file is the project rules for [Claude Code](https://claude.com/product/claude-code).
Claude reads it on every session. Keep it current — stale rules produce
worse output than no rules.

> Any field still showing `_TBD_` is an optional value left to fill in
> later. `_TBD_` means "unknown but acceptable" — safe to defer, not a
> blocker. Replace it with a real answer once the project settles.

## What this is

<!-- FILL: one paragraph describing what this project does, for whom, and roughly how. Link to the PRD, MVP doc, or AI summary once they exist. -->

## First run

New project? Open Claude Code in this repo and run `/start`. It will inspect
the installed kit surface and route you to the next workflow step.

## Current phase

_TBD_

Active milestone: `_TBD_`.

## Technology stack

- Runtime / language: _TBD_
- Framework: _TBD_
- Database / storage: _TBD_
- Key libraries: _TBD_
- Package manager: _TBD_
- Module system: _TBD_
- Deployment target: _TBD_

## How to run

```bash
_TBD_
_TBD_
_TBD_
```

## Testing

- Framework: _TBD_
- Location: _TBD_
- Run: `_TBD_`
- Coverage expectations: new modules include unit tests for the happy
  path, edge cases, and error handling. Existing tests must continue to
  pass on every PR.

## Code style

- Formatter: _TBD_
- Linter: _TBD_
- Commit style: _TBD_
- Secret management: _TBD_

## Project structure

```
llm-wiki-kit/
  src/            <!-- FILL: application/library code, if applicable -->
  test/           <!-- FILL: unit/integration tests, if applicable -->
  design/         architecture, ADRs, and workflow design artefacts
  prompts/        per-issue prompts (`prompts/issue-NNN-short-title.md`)
  notes/          working notes and evaluation logs
  .claude/skills/ installed workflow skills
```

See also:

- `design/architecture.md` — current architecture/design reference, maintained by `/workflow-docs`
- `design/adr/` — ADRs and ADR index, available on day one
- `design/prd.md` — product requirements, created later by the workflow
- `design/mvp.md` — MVP scope, created later by the workflow
- `design/build-out-plan.md` — phased backlog plan, created later by the workflow
- `design/ai-summary.md` — compact project summary, created later by the workflow
- `prompts/` — per-issue prompts (`prompts/issue-NNN-short-title.md`)
- `notes/` — working notes and evaluation logs
- `.claude/skills/` — installed workflow skills (do not edit by hand)

## Workflow rules

This project follows the [Claude Code Workflow Kit](https://github.com/olivermorgan2/claude-workflow-kit)
model. The rules below are load-bearing — Claude Code should treat them
as hard requirements unless a human overrides them in the session.

- **Plan-first execution.** When given a non-trivial task, propose a short
  step-by-step plan and wait for explicit approval before editing files.
- **Issue-by-issue.** Work is scoped to one GitHub issue at a time.
  Per-issue prompts live in `prompts/issue-NNN-short-title.md`.
- **Keep architecture docs current.** If a change materially alters the
  system shape, update or re-run `/workflow-docs` so `design/architecture.md`
  and `design/ai-summary.md` reflect reality.
- **Consult ADRs before changing load-bearing behaviour.** If a change
  touches architecture, installation, or conventions, check
  `design/adr/` first. Never edit an accepted ADR in place — supersede it
  with a new ADR, then refresh `design/architecture.md`.
- **Stay in scope.** Do not refactor unrelated code, rename files, or add
  speculative abstractions while working on an issue. If something is
  out of scope, note it for a follow-up issue.
- **Tests stay green.** Run `_TBD_` before declaring a task
  done. Fix regressions in the same PR that caused them.
- **Ask when ambiguous.** Prefer a clarifying question over a plausible
  guess on anything that affects scope, public API, or data shape.

## GitHub conventions

This project uses GitHub Flow — `main` is always
deployable; all work lands via pull requests.

- **Repository:** `olivermorgan2/llm-wiki-kit`
- **Default branch:** `main`
- **Branch naming:** _TBD_. Branch from `main`,
  never commit to it directly, delete branches after merge.
- **One issue per branch / PR.** Split PRs that grow to cover multiple
  issues.
- **Labels:** `feature`, `bug`, `design`, `infra`, `security`, `docs`.
  Every issue gets exactly one primary label.
- **Milestones:** `Foundation`, `MVP`, `Post-MVP` (add more only when the
  list gets crowded).
- **Pull requests** include the sections defined in
  [`.github/pull_request_template.md`](.github/pull_request_template.md):
  Summary, `Closes #N`, ADR reference, Changes, Test results, Manual
  verification.
- **Commit messages** reference the ADR and issue when the change is
  driven by them. Example: `feat(auth): add session middleware (ADR-003, #15)`.
  Commit style: _TBD_.

## Review expectations

- Every change lands via PR linked to a GitHub issue with `Closes #N`.
- Plan-first: propose a plan, wait for approval, then implement.
- Keep PRs small and focused. A PR should be reviewable in ~15 minutes.
- No direct commits to `main`.
- Existing tests must pass; new behaviour has new tests.

## Definitions of done

A task is **done** when:

1. Code compiles / type-checks cleanly.
2. `_TBD_` passes locally.
3. `_TBD_` passes locally (skip if `none`).
4. The PR body fills in every required section.
5. The ADR (if any) and the issue number are referenced in the commit.
6. A human has approved the PR or explicitly waived review.

## What this file is NOT

- Not the architecture reference — that is `design/architecture.md`.
- Not a spec — architectural decision history lives in `design/adr/`.
- Not a roadmap — phased plans live in `design/build-out-plan.md`.
- Not an AI-readable summary — that is `design/ai-summary.md`.

Keep this file focused on rules and conventions Claude Code needs to do
its job. When in doubt, link out.
