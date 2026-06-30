---
name: issue-planner
description: Turn design/mvp.md and design/build-out-plan.md into a reviewed batch of GitHub issues, plus a Project board. Use when filing the issue backlog after the MVP and build-out plan have been finalized.
permission-category: 3  # non-substitutable — gh issue create is public-visibility; existing approval gate is correct, per workflow-guide §7
inputs:
  - name: "--dry-run"
    required: false
    description: "Draft the issue batch without creating it"
  - name: "--no-project"
    required: false
    description: "Skip GitHub Project board creation"
  - name: "--force"
    required: false
    description: "Recreate issues even if a backlog exists"
outputs:
  - artefact: "(GitHub issues)"
    description: "Batch of issues created via gh issue create"
  - artefact: "(GitHub Project board)"
    description: "Project board for the issues"
next:
  - skill: prepare-issue
    when: "the backlog is created"
  - skill: workflow-docs
    when: "documenting the project"
---

# issue-planner

Read `design/mvp.md` and `design/build-out-plan.md`, draft a batch of
GitHub issues with titles, bodies, labels, milestones, and ADR
references, show the full batch for human approval, then create the
approved issues via `gh issue create`. On successful creation, create
a GitHub Project board and add every new issue to it.

This skill comes from
ADR-011 (issue-planner
hybrid draft-approve-create flow) and
ADR-012
(GitHub Projects board baked into issue creation).

## When to use this skill

Use after `prd-to-mvp` (Issue #7) has produced `design/mvp.md` and
`design/build-out-plan.md`, and after `adr-writer` (Issue #7) has
drafted the ADRs surfaced by that plan. The backlog this skill creates
is the direct input to `prepare-issue` (Issue #15) and, through it, the
executor skill.

If `design/mvp.md` or `design/build-out-plan.md` does not exist, stop
cleanly and tell the user which prerequisite skill to run.

## What this skill does not do

- Does not write ADRs — that is `adr-writer`.
- Does not open PRs, create branches, or run the executor — that is the
  executor skill and `pr-review-packager` (later issues).
- Does not modify `design/mvp.md`, `design/build-out-plan.md`, or any
  ADR file in place. If the plan needs changing, re-run `prd-to-mvp`.
- Does not silently bulk-create issues. Human approval is mandatory.
- Does not work against an arbitrary repo. It targets the current
  working repo as resolved by `gh` (user must be inside a checkout with
  `gh` authenticated).

## Inputs

- **Required:** `design/mvp.md` — phased feature list and product
  principles, rendered from `templates/mvp-template.md`.
- **Required:** `design/build-out-plan.md` — structured work items,
  phases, milestone recommendation, and initial issue backlog, rendered
  from `templates/build-out-plan-template.md`.
- **Required:** an authenticated `gh` CLI (`gh auth status` must
  succeed). For Project board creation, the token needs the `project`
  OAuth scope; `gh auth refresh -s project,read:project` will prompt
  for it.
- **Optional:** `design/adr/` — used to attach ADR references to
  issues when a draft matches an ADR title or topic.
- **Optional:** `design/planning.md`
  — when present, the sequencing-rationale section informs phase
  ordering of the issue backlog and the requirement IDs (`R1`,
  `R2`, …) are referenced in issue bodies for traceability. When
  absent, behaviour is unchanged — phase order comes from
  `build-out-plan.md` alone.

## Outputs

- **GitHub issues** in the current repository, one per work item in
  the build-out plan's "Initial issue backlog" section (or equivalent,
  see parsing protocol).
- **One GitHub Project board** named `<repo> — <milestone>` (ADR-012
  naming), with default columns **Todo**, **In Progress**, **Review**,
  **Done**. Every created issue is added to the board.
- **A summary message** back to the user, listing created issue
  numbers, titles, milestones, and the Project URL.
- **No local files are written** by this skill. The source of truth is
  GitHub.

## Invocation

- `/issue-planner` — full flow: parse, draft, show batch, prompt for
  approval, create issues, create board, add issues, report.
- `/issue-planner --dry-run` — parse and draft, show the full batch,
  but make zero `gh` API calls that mutate state. Read-only `gh`
  commands (listing existing issues, milestones, labels) are still
  permitted so the preview is accurate.
- `/issue-planner --force` — skip the idempotency check (see below).
  Use only when the user explicitly wants duplicates, e.g. after
  deleting a batch manually.
- `/issue-planner --no-project` — create issues but skip the Project
  board. Useful when the `project` OAuth scope is unavailable and the
  user wants to proceed anyway.

Flags compose: `/issue-planner --dry-run --no-project` is valid.

## Parsing protocol

The skill parses by **heading hierarchy**, not regex or YAML. Humans
write these files; tolerate minor formatting drift.

### From `design/build-out-plan.md`

1. **Locate the `## Initial issue backlog` section.** This is the
   canonical backlog source (see `templates/build-out-plan-template.md`).
2. Inside that section, each `### {{Milestone}}` subheading is a
   milestone grouping. The heading text is the milestone name the
   issues should be filed against.
3. Every bullet (`- {{title}}`) under that subheading is one issue
   title.
4. **Phase-driven path.** When the build-out plan
   contains `### Phase N: <name>` blocks (the post-ADR-032 shape) or
   `### Phase N — <name>` blocks (legacy shape), create **one
   GitHub milestone per phase** named "Phase N — <name>" by default,
   and assign each issue to its phase's milestone. The phase a given
   issue belongs to is determined by:
   - the `**Phase:**` frontmatter on its linked ADR, if any,
   - else the phase whose **Scope** or **Deliverables** bullets best
     match the issue title,
   - else the user's confirmation at the approval step.
   The `## Milestone recommendation` table, when present, can override
   the default per-phase milestone name.
5. **Single-phase fallback.** A plan with exactly one `### Phase`
   block — or with no `### Phase` headings at all — is treated as
   one implicit phase. Behaviour is identical to the pre-ADR-032
   flat path: one milestone for the whole project (named from the
   `## Milestone recommendation` table or prompted from the user).
6. **Backlog-section path (legacy).** If the backlog section is
   missing or empty, fall back to deriving issues from each phase's
   **Deliverables** bullets. Warn the user that this fallback is in
   use.
7. Capture the `## Milestone recommendation` table if present — it maps
   milestone short names (M1, M2, ...) to a focus description used in
   the milestone's GitHub description field.

### From `design/mvp.md`

1. Pull the **product name** from the top-level heading
   (`# <name> — MVP`) for the Project board name.
2. Pull **product principles** (`## Product principles`) to mention in
   issue bodies when a principle is directly relevant to a work item —
   keep this terse, do not copy the whole list into every issue.
3. Pull the **In scope / Out of scope** lists to sanity-check the
   backlog. If a drafted issue title maps to something on the "Out of
   scope" list, warn the user during the approval step.

### From `design/adr/`

1. For each ADR file, parse the first-line heading (`# ADR-NNN: Title`).
2. When a drafted issue title or body mentions an ADR number, topic, or
   a close match of the ADR title, attach a link to that ADR file in
   the issue body's **ADR** section.
3. If no ADR applies, write `none` and one line on why (mirrors
   `templates/issue-template.md`).

### Tolerated drift

- Milestone short names may be prefixed (`M1`, `M1 - name`, `Milestone
  1`). Normalise to the form used in the `## Milestone recommendation`
  table if present, otherwise use the text as-is.
- Bullets may use `-` or `*`. Accept both.
- A backlog subheading with zero bullets is skipped with a warning.
- A malformed heading (e.g. `###Phase 1` with no space) is skipped and
  reported at the end as a parsing warning.

## Issue draft format

Each drafted issue follows `templates/issue-template.md` and the
convention observed in this kit's own issues #11–#25:

- **Title:** concise, ends with the related ADR reference in
  parentheses when one applies, e.g. `CLAUDE.md starter template
  (ADR-007)`. No leading `[type]` prefix; labels carry that metadata.
- **Body:** sections in this order:
  1. **Summary** — one paragraph, what the issue is about.
  2. **ADR** — link to `design/adr/adr-NNN-short-title.md`, or `none`
     with a one-line reason.
  3. **Goal** — 1–2 sentences, the concrete outcome.
  4. **Why it matters** — link back to the ADR or a user need.
  5. **Tasks** — checkbox list derived from the build-out plan's
     deliverables or the skill's own decomposition; 2–6 items.
  6. **Acceptance criteria** — observable, not a restatement of tasks.
  7. **Notes** — labels and milestone echoed for human readers. Add a
     `Prompt:` line pointing at `prompts/issue-NNN-short-title.md`
     **only if** that file already exists; do not invent one.
- **Labels:** default to `feature` + a kind label (`design`, `docs`,
  `infra`, `chore`, `test`). Infer the kind from the work item's text
  and the plan's phase context. Never invent labels that do not exist
  in the repo; query `gh label list` first and only use labels that
  exist (see "Label handling" below).
- **Milestone:** exact milestone name as known to GitHub (see
  "Milestone handling" below).

## Label handling

1. Query `gh label list --json name` to get the set of labels that
   already exist in the repo.
2. Only attach labels from that set. If a draft wants a label that
   does not exist (e.g. `design` is absent), list the missing labels
   at the approval step and ask the user whether to:
   - create the missing labels (via `gh label create <name>`), or
   - drop them from the affected drafts, or
   - abort.
3. Never silently create labels. This is an explicit approval gate.

## Milestone handling

1. Query `gh api repos/:owner/:repo/milestones?state=all --jq '.[].title'`
   to list existing milestones.
2. For every milestone referenced by a drafted issue:
   - If the milestone exists, use its title verbatim.
   - If it does not, show the list of missing milestones at the
     approval step and ask whether to create them. On approval, create
     each via `gh api repos/:owner/:repo/milestones -f title=... -f description=...`.
     Pull the description from the build-out plan's milestone
     recommendation table if available.
3. If the build-out plan contains no milestone recommendation and no
   per-issue milestone can be inferred, prompt the user for a single
   default milestone name and reuse it for the whole batch.

## Idempotency

Creating issues is an irreversible action; duplicates are painful to
clean up. Before the approval step, the skill runs:

```
gh issue list --state all --limit 500 --json number,title,milestone
```

For each drafted issue, it looks for an existing issue with the same
title **or** a title that differs only in punctuation and case. Matches
are classified as **already exists**.

At the approval step:

- Drafts that already exist are shown with their existing issue number
  and are **excluded from creation by default**.
- The user can re-include them by running with `--force`, which skips
  this check entirely and creates duplicates.

The live `gh issue list` query above is the primary duplicate guard. As a
complement, the skill records an idempotency **receipt** keyed by the
plan/milestone id after a creation run, per
[`docs/receipts.md`](../../docs/receipts.md) — a local audit trail of what
this plan already filed, readable on a resumed session. Writing it is
best-effort and never blocks creation.

## Approval flow

The skill MUST present the full batch **before** making any mutating
`gh` call. The approval message has three parts:

1. **Preflight summary** — counts of drafts, per-milestone breakdown,
   missing milestones, missing labels, already-exists collisions, any
   parsing warnings.
2. **Full drafts** — every issue's title, labels, milestone, ADR ref,
   and body, numbered. Long; that is intentional.
3. **Explicit prompt** — "Reply `yes` to create these N issues, the
   missing milestone(s), and the Project board. Reply `no` or suggest
   edits to revise."

Only a literal `yes` (case-insensitive, no leading punctuation) is
treated as approval. Anything else is treated as "revise" — the skill
asks what to change and re-drafts.

On `--dry-run`, the approval prompt is replaced with a message saying
"Dry-run: no changes made. Re-run without `--dry-run` to create."

## Creation flow

On approval:

1. **Create missing labels** (if the user approved them).
2. **Create missing milestones** (if the user approved them).
3. **Create issues** in batch order (same order as the approval list).
   For each:
   ```
   gh issue create \
     --title "<title>" \
     --body-file <tmpfile> \
     --label "<l1>,<l2>" \
     --milestone "<milestone>"
   ```
   Pipe the body through a temp file to preserve newlines and special
   characters. Capture the returned issue URL/number.
4. **Create the Project board** (unless `--no-project`):
   ```
   gh project create --owner "@me" --title "<repo> — <milestone>"
   ```
   Use the milestone of the first created issue for the board name.
   If the batch spans multiple milestones, use the build-out plan's
   product name: `<product-name> — backlog`. The default views and
   columns (Todo / In Progress / Review / Done) are created by GitHub;
   if the kit later needs custom columns, that is a follow-up.
5. **Add issues to the Project** — for each created issue URL:
   ```
   gh project item-add <project-number> --owner "@me" --url <issue-url>
   ```
   `<project-number>` comes from the `gh project create` response.
6. **Report back** with: created issue numbers and titles, Project URL,
   any drafts that were skipped (already exists / label mismatch), and
   exact commands the user can run to inspect the result.

If any step fails, stop at that point and report what succeeded, what
failed, and the `gh` error. Do not attempt automatic rollback — the
user is better placed to decide whether to delete a partially-created
batch.

## OAuth scope for Projects

`gh project create` and `gh project item-add` require the `project`
scope. If `gh auth status` does not show it, the skill:

1. Warns the user at the preflight step.
2. Offers to run `gh auth refresh -s project,read:project` (which will
   open a browser).
3. If the user declines, suggests re-running with `--no-project`.

Never silently downgrade — the user asked for a Project and should get
one or a clear explanation of why they did not.

## Edge cases

- **Missing `design/mvp.md` or `design/build-out-plan.md`:** stop
  cleanly. Print which file is missing and the prerequisite skill
  (`prd-to-mvp`).
- **Plan files exist but backlog section is empty:** fall back to
  parsing `## Phases` deliverables (see parsing protocol), and warn.
- **Malformed headings:** skip the malformed section, collect as
  warnings, and show them at the approval step.
- **No milestones parsable:** prompt the user for a single default
  milestone name and reuse it for the batch.
- **`gh` not authenticated:** stop and print `gh auth login`
  instructions.
- **Not inside a GitHub-backed repo:** `gh repo view` will fail — stop
  cleanly and ask the user to check their remote.
- **User is in this kit repo itself** (not a target project): the
  plan files `design/mvp.md` / `design/build-out-plan.md` do not
  exist here — only ADRs and the kit's own issues do. The skill will
  correctly stop at the "missing plan files" check. This is expected
  behaviour; the skill is designed to run in target projects.
- **Draft maps to an "Out of scope" item:** warn at the approval step
  but allow creation — the user may intentionally be planning a
  future milestone.
- **`--force` on a repo with pre-existing matches:** proceed, but
  remind the user in the preflight summary that duplicates will be
  created.

## Self-check before calling `gh`

Before any mutating `gh` call, confirm:

- [ ] User has typed literal `yes` in response to the approval prompt
  (unless `--dry-run`, in which case no mutating call is made).
- [ ] Every drafted issue has a non-empty title, body, and milestone.
- [ ] Every label attached to a draft exists in the repo (or the user
  approved creating it).
- [ ] Every milestone referenced exists in the repo (or the user
  approved creating it).
- [ ] No `{{...}}` template placeholders remain in any body.

If any fail, stop and report.

## Handoff

The backlog this skill creates is the input to:

- **`prepare-issue`** (Issue #15) — reads one issue at a time and
  produces a per-issue prompt under `prompts/issue-NNN-short-title.md`.
- **`workflow-docs`** (Issue #20) — may cross-reference the issue
  backlog and Project board in generated README / workflow-guide
  content.

See [`example.md`](example.md) for a worked run on a small Pace Drift
build-out plan.
