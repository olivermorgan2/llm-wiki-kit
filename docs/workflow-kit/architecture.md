# Architecture

This document describes how the Claude Code Workflow Kit is organized and
why it is shaped the way it is. It is a current, practical reference — not a
history of how the kit got here. For the design-decision history, see
[Public vs. internal docs](#9-public-vs-internal-docs) at the end.

---

## 1. Purpose

The kit gives a **new, structured project** a disciplined, GitHub-first path
from a rough idea to a tagged release. It is a set of Claude Code skills,
templates, and docs that you install into your own project; from then on the
kit drives the flow through a PRD, recorded decisions, GitHub issues, and pull
requests. It also maintains a current architecture document so the project's
shape stays visible as it evolves.

The goal is explicit decisions and incremental, reviewable work without
reinventing a workflow each time. The primitives — define the problem, scope
the first cut, record decisions, break work into issues, review via PR — fit
any structured project: software, research, writing, curricula, design
systems, internal policy.

## 2. Scope

**In scope.** New projects that live in a git repository, use GitHub, and are
driven with Claude Code. The kit installs **per project** (no global state)
and covers the whole lifecycle: scoping, decisions, issue-by-issue execution,
releases, and session continuity.

**Out of scope.** Existing-repo migration (the kit is new-projects-only),
non-GitHub providers, and non-Claude tooling. The kit does not try to abstract
over every VCS host or assistant — it commits to one well-supported path and
makes it good.

## 3. System model

The kit has four moving parts:

| Part | What it is |
|---|---|
| **Skills** (`skills/`) | The unit of behavior. Each skill is a directory with a `SKILL.md` (plus optional `example.md`, `reference.md`). A skill is invoked as a slash command inside Claude Code and performs one workflow step. |
| **Templates** (`templates/`) | Starter documents the skills render into your project — `CLAUDE.md`, PRD, architecture, ADR, issue, PR, changelog, and continuity templates. |
| **Agent contract** (`kit.json`, `schemas/`, `bin/`) | A machine-readable index of every skill — its permissions, inputs, outputs, and the next step it hands off to — plus the validators that keep that index honest. This is what lets the kit reason about itself. |
| **Docs** (`docs/`) | Human documentation about installing and running the kit. |

At install time the **skills** are copied into the target project's
`.claude/skills/`, and selected **templates** are rendered into project files.
Nothing points back at the kit afterward: the installed project is
self-contained and evolves independently. There is no shared global install
and no runtime dependency on the kit being present.

## 4. Workflow lifecycle

The kit is a small set of slash commands. The core loop —
**scope → backlog → implement → ship** — repeats for the life of the project:

| Stage | Commands | Output |
|---|---|---|
| Scope | `/idea-to-prd`, `/prd-normalizer`, `/prd-to-mvp` | PRD, MVP, build-out plan |
| Decide | `/adr-writer` | ADRs in the project's `design/adr/` |
| Backlog | `/issue-planner` | GitHub issues + project board |
| Implement | `/prepare-issue`, `/claude-issue-executor` | feature branch, tests, commits |
| Ship | `/pr-review-packager`, `/changelog`, `/release`, `/workflow-docs` | PR, release notes, refreshed README/architecture/AI summary, tagged release |

Work is scoped **one GitHub issue per branch and PR**, following
[GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow).

A **verb layer** sits in front of the full command set so you don't have to
memorize it: `/start`, `/decide`, `/backlog`, `/work`, `/ship`, `/release`.
Of these, only `/start` (with its `/next` alias) is an installed router
skill; `/decide`, `/backlog`, `/work`, and `/ship` are **documentation
aliases** that map to existing skills rather than separate command folders
(see [workflow control](workflow-control.md) for the verb→skill map).
Type `/start` (or `/next`) and the kit inspects the project state and tells
you the next step. Three **operating modes** — interactive, assisted,
autonomous — set how much the assistant does before pausing for you.

## 5. Design principles

These are the load-bearing choices that explain the kit's shape.

- **Project-local install.** Each project gets its own copy of the skills
  under `.claude/skills/`. No global install, no central skills directory, no
  version skew between projects.
- **New projects only.** The kit assumes a clean start, which keeps every
  skill simple — there is no migration or reconciliation logic to carry.
- **GitHub-first.** One host, deeply supported (issues, projects, PRs,
  releases), instead of a shallow abstraction over many.
- **Plan-first execution.** Skills propose a plan and wait for approval before
  acting. Approval is the canonical gate; the operating mode only changes how
  often it triggers.
- **Issue-by-issue work.** Each unit of work is one issue, one branch, one PR
  — small, reviewable, revertible.
- **Continuously documented architecture.** Target projects maintain
  `design/architecture.md` as the current human-readable system shape. ADRs
  capture rationale; the architecture doc captures today's design and is
  refreshed by `/workflow-docs` after meaningful changes and before releases.
- **Self-describing.** Every skill declares its permissions, inputs, outputs,
  and handoff in `kit.json`, and validators (`bin/`) fail the build if a
  skill and its declaration drift apart.
- **Lightweight.** No premature automation and no speculative abstractions —
  the kit ships only what a real project needs.

## 6. Repository layout

```text
claude-workflow-kit/
  README.md          ← overview and quick start
  LICENSE            ← MIT (covers the kit itself)
  CLAUDE.md          ← rules for Claude Code working in this repo
  kit.json           ← machine-readable index of skills + bin surfaces
  skills/            ← the skills the kit ships (one dir per skill)
  templates/         ← starter docs rendered into your project
  schemas/           ← JSON schemas for the agent contract
  bin/               ← install/bootstrap + validators
  docs/              ← documentation about the kit (incl. this file)
  examples/          ← worked end-to-end example projects
  ai-review/         ← optional AI PR-review prompt pack + config example
  .github/           ← issue / PR templates + CI checks
```

The split between **what ships and what is generated** matters: `skills/` and
selected `templates/` are copied into your project; `docs/`, `examples/`, and
the kit's own `CLAUDE.md` stay in the kit and are not copied. See
[`repo-structure.md`](repo-structure.md) for the full kit-vs-target map.

## 7. Safety boundaries

The kit is built to be safe to run with an AI assistant:

- **Approval gate.** Plan-first means destructive or outward-facing steps
  (pushing, creating issues, tagging releases) are proposed and wait for your
  approval. Autonomous mode widens what runs without a prompt but never
  removes the gate from identity-sensitive actions.
- **Permission declarations.** Each skill declares the permission categories
  it needs in `kit.json`; the validators in `bin/` enforce that a skill cannot
  silently acquire broader scope than it declares.
- **Idempotent install.** The bootstrap installer skips files that already
  exist, so re-running it is safe.
- **No global footprint.** Because install is project-local, running the kit
  in one project cannot affect another or your machine's global state.
- **License clarity.** The MIT license covers the kit itself, not the work you
  produce with it — your PRDs, decisions, and source stay yours.

## 8. Extension model

You extend the kit by **adding or editing skills**:

1. A skill is a directory under `skills/<name>/` containing a `SKILL.md` with
   frontmatter (name, description) and a body that describes the step.
2. Register the skill in `kit.json` with its permissions, inputs, outputs, and
   the step it hands off to, so the rest of the kit can reason about it.
3. Run the validators in `bin/` (also wired into CI under `.github/`) to
   confirm the skill and its `kit.json` declaration agree.

Templates are extended the same way — add a file under `templates/` and have a
skill render it. Because the agent contract is validated, the kit stays
coherent as it grows: a skill that is added, renamed, or changed without
updating its declaration fails the build rather than shipping broken.

## 9. Public vs. internal docs

This repository is the **public distribution** of the kit. It centers on a
small, stable surface: this `README` and `docs/` (including this architecture
doc), the reusable `skills/`, `templates/`, and `examples/`, and the standard
project files (`LICENSE`, contribution guidance, security policy).

The kit's **detailed decision history — its Architecture Decision Records
(ADRs), evaluation reports, roadmaps, and superseded designs — is maintained
internally and is not part of this public distribution.** Those records
capture *why* past decisions were made and are not required to install, run,
or extend the kit. Kit documentation refers to these records by bare
identifiers such as `ADR-006`; they are internal design records of the
source repository, and their outcomes are summarized in this document.
This document is the public, current statement of the
architecture; when a decision changes, this doc is what gets updated.

If you are contributing and need the rationale behind a specific design choice,
raise it in an issue and the maintainers can share the relevant internal
context.
