# prepare-issue — worked example

A single end-to-end walkthrough showing how the `prepare-issue`
skill turns a GitHub issue number into a filled prompt file. This is
reference material, not a runtime artifact.

---

## 1. User invocation

```
/prepare-issue 17
```

That is the entire input. The skill infers the repo from the
current git remote and resolves everything else from `gh` and the
filesystem.

## 2. Data the skill fetches

**From `gh issue view 17 --json title,body,labels,milestone,url`:**

- Title: `Build pr-review-packager skill (ADR-015)`
- Labels: `feature`, `design`
- Milestone: `M5 - v-next`
- Body: contains "Goal", "Why it matters", "Requirements",
  "Acceptance criteria", "Scope and constraints",
  "Evaluation & testing requirements" sections.
- Body mentions `ADR-015` and references
  `design/adr/adr-015-pr-review-packager-skill.md`.

**From `design/adr/adr-015-pr-review-packager-skill.md`:**

- Title line: `# ADR-015: pr-review-packager skill`
- Decision (first sentence): "Build a pr-review-packager skill that
  creates PRs with proper template, links, and summaries."

**From `design/build-out-plan.md`:**

- File does not exist in this repo. Skip silently.

## 3. Derived values

- Short title: `pr-review-packager-skill`
  - Lowercase: `build pr-review-packager skill (adr-015)`
  - Strip `(adr-015)` suffix → `build pr-review-packager skill`
  - Strip leading `build ` → `pr-review-packager skill`
  - Kebab-case → `pr-review-packager-skill`
- Output path: `prompts/issue-017-pr-review-packager-skill.md`

## 4. Filled template (shown to user for review)

The skill prints the following to chat and asks "Write this to
`prompts/issue-017-pr-review-packager-skill.md`? (yes / edit /
cancel)".

````markdown
You are working in my `workflow-kit` repository.

Context:
- A GitHub-first, project-local toolkit that takes a new project from idea to MVP to release.
- Follow the rules in `CLAUDE.md`.
- The workflow model is described in `docs/workflow-guide.md`.

ADR:
- File: `design/adr/adr-015-pr-review-packager-skill.md`
- Decision: Build a pr-review-packager skill that creates PRs with proper template, links, and summaries.

GitHub Issue:
- Title: Build pr-review-packager skill (ADR-015)
- Number: #17
- Milestone: M5 - v-next
- Labels: feature, design

Goal
Build the pr-review-packager skill that creates pull requests with proper template formatting, issue and ADR links, and change summaries derived from commit history.

Why it matters
Consistent, well-structured PRs make code review faster and maintain a clear audit trail from ADR decisions through implementation to merge. Automating the PR body from templates and commit history removes manual busywork and reduces the chance of missing links or context.

Requirements
- Draft the PR body from `templates/pr-template.md`
- Fill in `Closes #N` and ADR references automatically from the branch or commit history
- Derive a change summary from the commit history on the current branch
- Present the draft PR to the user for review and approval before creation
- Create the PR via `gh pr create` after approval
- Support setting labels, milestone, and reviewers if configured

Acceptance criteria
- The skill produces well-structured PRs matching the PR template
- Issue and ADR links are correct and present in the PR body
- The change summary accurately reflects the commits on the branch
- The user approves the PR draft before it is created
- The skill handles edge cases (no template file, no commits, detached HEAD) gracefully

Scope and constraints
- Primary folders to touch: `skills/pr-review-packager/`
- Folders to avoid unless absolutely necessary: other skills, core templates
- Keep the skill focused on PR packaging, not on review feedback or CI checks

Evaluation & testing requirements
- Verify that the skill correctly parses the PR template and fills placeholders
- Confirm that ADR and issue references are extracted from commit messages
- Test with a branch that has multiple commits referencing different issues
- Confirm the user approval gate works (no PR created without explicit approval)
- All existing tests must continue to pass.
- If a change cannot be unit tested, document the manual verification.

Instructions for you
1. Read the relevant docs and existing files:
   - `CLAUDE.md`
   - `design/adr/adr-015-pr-review-packager-skill.md`
   - any existing modules under `skills/pr-review-packager/`
   - any existing tests related to the modules you will change
2. Propose a short, step-by-step implementation PLAN for this issue, including:
   - new files or modules to create,
   - existing files to modify,
   - key functions or structures,
   - your verification or test plan.
3. Wait for my approval of the plan before making any edits.
4. After I approve, implement the plan:
   - keep changes focused on this issue's scope,
   - commit incrementally with messages referencing the ADR and issue
     (e.g. "feat(scope): add thing (ADR-NNN, #NN)").
5. At the end, provide an evaluation summary:
   - what changed,
   - verification steps performed,
   - any follow-up work needed for later issues,
   - exact commands I should run to inspect the result myself.

Do not start editing files until I explicitly approve your plan.
````

## 5. User confirmation

User replies `yes`. The skill writes
`prompts/issue-017-pr-review-packager-skill.md` and reports:

> Wrote `prompts/issue-017-pr-review-packager-skill.md` (all
> placeholders filled; no TODOs remain). Open it in a fresh Claude
> Code session to start implementation.

## 6. What would change for other issues

- **Issue #22** ("Write docs/workflow-guide.md — end-to-end workflow
  guide (ADR-020)"):
  - Short title: `docs-workflow-guide-end-to-end-workflow-guide`
    (truncated at 50 chars on a `-` boundary).
  - Output: `prompts/issue-022-docs-workflow-guide-end-to-end.md`
    (after truncation).
  - ADR: `adr-020-end-to-end-workflow-guide.md`.

- **Issue referencing no ADR** (e.g. a chore):
  - ADR section becomes `ADR: none — no ADR referenced in this
    issue.`
  - Everything else is populated as normal.

- **Issue #999 that does not exist:**
  - `gh issue view 999` exits non-zero.
  - Skill prints the error message and stops without writing
    anything.
