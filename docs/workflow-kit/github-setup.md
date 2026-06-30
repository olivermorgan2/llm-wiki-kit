# GitHub Setup Guide

How to set up GitHub for a **new target project** that uses the Claude Code
Workflow Kit. The kit is GitHub-first
(ADR-004); this
guide is the practical configuration that makes that decision real.

> **Prerequisites:** Git, GitHub CLI, `gh auth login` working, and SSH
> set up. See [`docs/install.md`](install.md#1-prerequisites) — this guide
> does not repeat that material.

---

## What this guide covers

0. **Credentials and the required `gh` scopes** for every GitHub-integrated skill (read this first)
1. Creating the GitHub repository for your target project
2. Picking `main` as the default branch and using a `main + feature branch` model
3. The six default labels the kit expects
4. A concrete default milestone set for a new project
5. Branch naming and pull-request rules
6. Optional branch protection
7. The `.github/` assets the kit will ship (deferred to Issue #9)
8. A short `gh` command cheat-sheet

## Who this is for

Two audiences, which the guide keeps separate:

- **Primary — you are setting up a new target project.** Almost everything
  below applies to you.
- **Secondary — you are working on the kit itself.** The kit repo already
  has labels and milestones applied. You do not need to re-apply them.
  Where the two differ (milestones especially), the guide calls it out.

---

## GitHub credentials and required scopes

Every GitHub-integrated skill drives the **GitHub CLI** (`gh`) rather
than the REST/GraphQL API directly. They all rely on one authenticated
`gh` session, so getting credentials right once unblocks the whole
workflow. This is the single place that documents what `gh` needs; the
agent contract's preflight points here
([`docs/agent-contract.md`](agent-contract.md#preflight-for-github-integrated-skills)).

### Required scopes

| Scope | Why the kit needs it | Skills |
|---|---|---|
| `repo` | Read/write issues, pull requests, releases, and milestones | `prepare-issue`, `resume`, `audit-milestone`, `changelog` (read); `issue-planner`, `pr-review-packager`, `release`, `complete-milestone` (write) |
| `read:org` | Resolve the org that owns a GitHub **Project** board | `issue-planner` (when it creates/links a Project) |
| `project` | Create and write to the GitHub **Project** board (ADR-012) | `issue-planner` (board creation; skip with `--no-project`) |
| `workflow` | Push files under `.github/workflows/` | Optional AI PR review module only (ADR-046); not needed for the core flow |

`repo` alone covers the entire core planning-to-release loop. The extra
`read:org` + `project` scopes are needed only by `issue-planner`'s
Project-board step — and only the **first** time, when the board is
created. If you run `issue-planner --no-project`, `repo` is sufficient.

### Refresh scopes with one command

If you authenticated with `gh auth login` (web/device flow, token in
the keyring), add the Project-board scopes to your existing token in
place:

```bash
gh auth refresh -s project,read:org
```

If you authenticated with a **Personal Access Token** instead, `gh auth
refresh` cannot widen it — edit the token's permissions in GitHub
**Settings → Developer settings → Tokens** (a fine-grained token needs
*Issues*, *Pull requests*, *Contents*, and *Projects* read/write on the
target repo), then re-run `gh auth login` with the updated token.

### Preflight checklist

Run this before any GitHub-integrated skill (an agent should run it at
the start of a session that will touch GitHub):

```bash
# 1. Authenticated, and as the right account?
gh auth status

# 2. Does the repo resolve from here?
gh repo view --json nameWithOwner -q .nameWithOwner

# 3. Only if you will create a Project board (issue-planner):
#    confirm the Project scopes, and add them if missing.
gh auth status            # look for 'project' and 'read:org' in Token scopes
gh auth refresh -s project,read:org   # run if they are absent
```

`gh auth status` prints the active account and, for classic tokens, the
granted scopes. If step 2 errors, you are either not inside the target
git repo or `origin` does not point at GitHub — fix that before running
any skill, because every GitHub skill infers the repo from the git
remote. If `issue-planner` fails partway through with a Project-board
permission error, step 3 is the fix; the issue-creation step itself
only needs `repo`.

---

## 1. Create the GitHub repository

Pick one of the two paths from [`docs/install.md`](install.md#2-choose-your-target-project).
The short version:

```bash
# Preferred: create on GitHub, then work locally
gh repo create my-project --public --clone   # or --private
cd my-project
```

Use `--private` if the project is commercial or contains anything you
would not want public. You can flip visibility later in the GitHub UI.

---

## 2. Default branch and model

Use **`main` + feature branches**. No develop branch, no release branches,
no gitflow. `main` is always the stable, always-deployable tip.

```
main (stable, always deployable)
  │
  ├── add-feature-x       ← one feature branch per issue
  ├── fix-bug-y
  └── refactor-module-z
```

`gh repo create` already defaults to `main`. If you are working with an
older repo, confirm under **Settings → Branches → Default branch**.

---

## 3. Labels

The kit uses six labels. Keep them consistent across every new target
project — the kit's issue-planning skill will rely on exactly these names.

| Label | Color | Use for |
|---|---|---|
| `feature` | `0E8A16` (green) | New functionality |
| `bug` | `D73A4A` (red) | Something broken |
| `design` | `1D76DB` (blue) | ADR or architecture work |
| `infra` | `FBCA04` (yellow) | CI, deployment, tooling, repo setup |
| `security` | `B60205` (dark red) | Security-related changes |
| `docs` | `0075CA` (blue) | Documentation only |

Create them on a new repo with:

```bash
gh label create feature  --color 0E8A16
gh label create bug      --color D73A4A
gh label create design   --color 1D76DB
gh label create infra    --color FBCA04
gh label create security --color B60205
gh label create docs     --color 0075CA
```

Keep the default GitHub labels (`enhancement`, `good first issue`, etc.)
or delete them — both are fine. The kit only depends on the six above.

> **For the kit repo itself:** these labels already exist on
> `olivermorgan2/claude-workflow-kit`. You do not need to re-create them.

---

## 4. Milestones

Milestones structure delivery into visible phases so you can see progress
and scope creep at a glance.

### Default milestone set for a new target project

Start a new project with **these three milestones**:

| Milestone | Intent | Typical contents |
|---|---|---|
| `Foundation` | Repo setup, project skeleton, first ADRs, CLAUDE.md, CI if needed | Install the kit, draft ADRs, create the initial issue backlog |
| `MVP` | The smallest version of the product that is useful | Feature issues derived from the accepted ADRs; everything needed to call it "v0.1" |
| `Post-MVP` | Polish, the next wave of features, stretch goals | Work that is real but not blocking the MVP |

Create them with:

```bash
gh api repos/:owner/:repo/milestones -f title="Foundation"
gh api repos/:owner/:repo/milestones -f title="MVP"
gh api repos/:owner/:repo/milestones -f title="Post-MVP"
```

(Replace `:owner/:repo` with the actual path, or run from inside the repo
and GitHub CLI will resolve it.)

Each new issue should be assigned to exactly one milestone. If an issue
doesn't have a milestone, that's a signal it hasn't been scoped yet.

### Scaling up later

Add more milestones when you need to — for example, a `Beta` between
MVP and Post-MVP, or one milestone per release (`v0.2`, `v0.3`). Start
with three; split only when the list gets crowded.

> **For the kit repo itself:** milestones are `M1 - Foundation + ADRs`,
> `M2 - Planning Skills`, `M3 - Execution Workflow`, `M4 - Examples +
> Validation`. These are build milestones specific to the kit; do **not**
> copy them into your target project.

---

## 5. Branch naming and pull requests

### Branch rules

| Rule | Why |
|---|---|
| Always branch from `main` | Start from stable, known-good code |
| One branch per issue (or closely related group) | Keeps PRs focused and reviewable |
| Descriptive, kebab-case names | `add-auth-middleware`, not `feature-1` |
| Never commit directly to `main` | All changes go through PRs |
| Delete branches after merge | Prevents branch clutter |

Typical lifecycle:

```bash
git checkout main
git pull
git checkout -b add-feature-x
# … commits …
git push -u origin add-feature-x
gh pr create --fill   # or with a filled PR body
# after merge:
git checkout main && git pull
git branch -d add-feature-x
```

### Pull-request rules

- Every PR links to its issue using `Closes #N` in the body. GitHub will
  auto-close the issue on merge.
- Keep PRs small and focused on one issue. If a PR grows to cover
  multiple issues, split it.
- Every PR references the governing ADR when relevant.
- Required sections in the PR body — Summary, Closes, ADR, Changes,
  Test results, Manual verification. The template defined below (Issue
  #9) enforces this.

---

## 6. Optional branch protection

For solo developers this is optional; for small teams it is recommended.
Enable under **Settings → Branches → Branch protection rules**, target
branch `main`:

| Rule | Effect |
|---|---|
| Require a pull request before merging | Forces all changes through PR review |
| Require approvals (≥ 1 reviewer) | Useful once you have a second contributor |
| Require status checks to pass | CI must be green before merge |
| Require branches to be up to date | Rebase / merge `main` before merging PR |
| Do not allow force pushes to `main` | Protects history |

Solo developers can safely skip approvals and status checks until CI
exists. The first two rules (require PR, disallow force push) are worth
turning on early even solo — they prevent accidental direct commits to
`main`.

---

## 7. `.github/` assets (shipped by Issue #9)

The `.github/` directory is where GitHub looks for issue and PR templates.
The kit ships three files; this section is the spec they were built
against and continues to be the contract if they're ever revised.

### [`.github/pull_request_template.md`](../.github/pull_request_template.md)

**Required sections:**

- `## Summary` — one-paragraph description of the change
- `## Closes` — `Closes #N` line(s)
- `## ADR` — `Related ADR: design/adr/adr-NNN-...md` (or "none" if N/A)
- `## Changes` — bulleted list of the substantive edits
- `## Test results` — paste of the test-runner output or "no code changes"
- `## Manual verification` — steps the reviewer should run, or "none needed"

### [`.github/ISSUE_TEMPLATE/feature-request.md`](../.github/ISSUE_TEMPLATE/feature-request.md)

**Required frontmatter:** `name`, `about`, `title`, `labels: feature`

**Required body sections:**

- `## Summary`
- `## ADR` — reference to the ADR this issue implements (or "none")
- `## Goal`
- `## Why it matters`
- `## Tasks` — checkbox list
- `## Acceptance criteria`
- `## Notes` — labels and milestone

### [`.github/ISSUE_TEMPLATE/docs-task.md`](../.github/ISSUE_TEMPLATE/docs-task.md)

**Required frontmatter:** `name`, `about`, `title`, `labels: docs`

**Required body sections:**

- `## Summary`
- `## Scope` — which docs are affected
- `## Files affected`
- `## Acceptance criteria`
- `## Notes` — milestone

Both issue templates are picked up by the GitHub "New issue" chooser
automatically once they exist in `.github/ISSUE_TEMPLATE/` on the
default branch. Inline guidance in each template uses HTML comments so
the rendered issue/PR view is clean even when an author submits without
deleting the prompts.

---

## 8. `gh` command cheat-sheet

```bash
# Auth
gh auth status
gh auth login

# Repo
gh repo create my-project --public --clone
gh repo view --web

# Labels
gh label create <name> --color <hex>
gh label list

# Milestones (no short-form command; use the API)
gh api repos/:owner/:repo/milestones -f title="<name>"
gh api repos/:owner/:repo/milestones --jq '.[] | {number, title, state}'

# Issues
gh issue create --title "..." --body "..." --label feature --milestone "MVP"
gh issue list --state open
gh issue view <n> --web
gh issue close <n> --comment "..."

# PRs
gh pr create --fill
gh pr view --web
gh pr checks
gh pr merge --squash --delete-branch
```

---

## What's next

- [`docs/workflow-guide.md`](workflow-guide.md) — end-to-end flow from idea to release
- [`docs/skills.md`](skills.md) — how to use the installed skills, including `/adr-writer` for writing ADRs
- [`docs/claude-code-guide.md`](claude-code-guide.md) — using Claude Code with the kit

## Alignment

Every section of this guide traces back to
ADR-004, which
names the GitHub assumptions in one line each: repositories, issues,
labels, milestones, pull requests, `main + feature branch`, and optional
branch protection.
