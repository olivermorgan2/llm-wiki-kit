---
name: claude-issue-executor
description: Drive a disciplined implementation session from a prepared issue prompt — plan-first, branch, incremental commits, tests alongside, evaluation summary. Use when running the implementation session for an issue whose prompt artefact already exists; use /prepare-issue first if no prompt is on disk.
permission-category: 2  # operator-acknowledged-bypass — significant-task plan-mode gate per workflow-guide §7
inputs:
  - name: "prompt-or-issue"
    required: false
    description: "Path to prompts/issue-NNN-*.md, or the issue number"
  - name: "--no-prompt"
    required: false
    description: "Run without a prepared prompt file"
outputs:
  - artefact: "(branch + commits)"
    description: "Feature branch, incremental commits, tests alongside"
  - artefact: "notes/eval-issue-NNN.md"
    description: "Evaluation summary"
next:
  - skill: pr-review-packager
    when: "implementation is complete"
---

# claude-issue-executor

Run a full implementation session for a single GitHub issue, starting from
a prepared prompt file. The skill parses the prompt, proposes a plan and
**blocks for explicit user approval**, creates a feature branch from `main`,
implements in incremental commits that reference the governing ADR and
issue number, writes tests alongside code where the project supports them,
and produces an evaluation summary at the end.

This skill is the enforcement mechanism for the execution model decided
in ADR-006 and
the orchestration contract decided in
ADR-014. It
is also the **producer** in the cross-skill design-question
carry-forward chain decided in
ADR-040
— see [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040)
for the canonical schema.

## When to use this skill

- After `/prepare-issue` has written a prompt file — typically
  `prompts/issue-NNN-short-title.md`.
- When starting an implementation session for a single GitHub issue.
- When the user invokes `/claude-issue-executor <path-to-prompt>`.

If no prompt file exists yet, stop and tell the user to run
`/prepare-issue` first (or to hand-author a prompt that matches
`prompts/_template.md`).

## What this skill does not do

- Does not create prompts. That is `/prepare-issue`.
- Does not create pull requests. That is `/pr-review-packager`.
  This skill **suggests** `/pr-review-packager` in its final summary but
  never auto-invokes it.
- Does not draft, amend, or accept ADRs. ADRs are authored by
  `adr-writer` and accepted by a human.
- Does not modify GitHub issues, labels, milestones, or the issue body.
- Does not duplicate logic from other skills. It is an orchestrator:
  prompt parsing, plan gate, branching, commits, tests, summary, handoff.

## Invocation contract

```
/claude-issue-executor <path-to-prompt-file-OR-issue-number>
```

- **Argument (optional):** either (a) a path to a prepared prompt file
  — preferred location `prompts/issue-NNN-short-title.md`, legacy
  fallback `notes/issueN-prompt.md`; or (b) a GitHub issue number, in
  which case the skill resolves the prompt path itself and may
  auto-chain `/prepare-issue` per the auto-chain protocol below.
- **No argument given:** list candidate prompt files from `prompts/`
  and, as a fallback, `notes/issue*-prompt.md`. Ask the user which to
  use. Do nothing else until they answer.

### Auto-chain `prepare-issue`

When the resolved issue has no `prompts/issue-NNN-*.md` on disk, the
skill auto-invokes `/prepare-issue NNN` and proceeds. The prep step
is logged prominently so the user sees it happen — auto-chain reduces
ceremony, not visibility.

When a prompt exists but is **stale** — its mtime is older than the
issue body's `updatedAt` (via `gh issue view <NNN> --json updatedAt`)
or older than any linked ADR's mtime — the skill asks the user
whether to regenerate. Default action: **regenerate with
confirmation, never silently**. On `n`, proceed with the existing
prompt and surface a one-line note in the eventual evaluation
summary.

The auto-chain runs *before* significance classification, so plan-
mode rhythm gates the regenerated prompt as well.

## Inputs

- **Required:** a prepared prompt file containing the sections the skill
  relies on (see **Prompt validation**).
- **Required in the environment:** a clean working tree on `main` (or on
  a branch that tracks `main`). Nothing uncommitted.
- **Read (not modified):** `CLAUDE.md`, the ADR named in the prompt, and
  any other files the prompt points at.

## Outputs

- A feature branch created from `main`.
- A series of focused commits on that branch, each referencing the ADR
  and issue from the prompt.
- Tests alongside implementation, where the project has a test runner.
- An evaluation summary printed at the end of the session **and
  persisted to `notes/eval-issue-NNN.md`**. The
  persisted file is the canonical input to `/pr-review-packager`
  for the carry-forward chain — see [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040).

This skill does **not** push, does **not** open a PR, and does **not**
merge. Those belong to `/pr-review-packager` and the user's own review.

## Prompt validation

Before doing anything else, read the prompt file and confirm it contains
the sections the rest of the skill depends on. Required sections, matched
case-insensitively by heading or leading bold label:

1. **Context** — what project and what this session is about.
2. **ADR** — the governing ADR, e.g. `design/adr/adr-NNN-...md`.
3. **GitHub Issue** — title and number, e.g. `Number: #16`.
4. **Goal** — what the issue should achieve.
5. **Requirements** — the work to be done.
6. **Scope and constraints** (or "Scope") — bounds of the session.
7. **Instructions for you** (or "Instructions") — the plan-first,
   commit-incrementally, evaluate-at-the-end directive.

If any required section is missing, **stop**. Report which sections were
not found and ask the user whether to fix the prompt (`/prepare-issue`
can regenerate it) or proceed without that section. Do not invent
missing content.

Also extract two fields for later:

- **ADR number** — the `NNN` in the ADR filename. Used in commit messages.
- **Issue number** — the `#N` from the GitHub Issue section. Used in
  commit messages and the branch name.

If either is absent, the prompt is malformed — treat as above.

## Branch naming

Derive the branch name from the prompt filename:

- `prompts/issue-017-pr-review-packager.md` → `pr-review-packager-skill` is
  one acceptable form; so is `issue-017-pr-review-packager`. Prefer the
  short, descriptive form that matches the repo convention in
  [`docs/github-setup.md` section 5](../../docs/github-setup.md#5-branch-naming-and-pull-requests):
  kebab-case, descriptive, no leading `feature-` or numeric-only names.
- If the prompt filename already matches repo convention, use it verbatim
  with any `issue-NNN-` prefix stripped when the short title alone is
  already distinctive.
- When in doubt, show the user the proposed branch name in the plan and
  let them amend it before approval.

Create the branch with `git checkout -b <branch>` from `main`. If the
user is not on `main`, switch to `main` and pull first — but only after
confirming the working tree is clean.

## Commit model

- **One logical change per commit.** Do not batch unrelated edits.
- **Commit message format:**
  `<verb>(<scope>): <what> (ADR-NNN, #issue)`
  where `<verb>` is `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, etc.
  and `<scope>` points at the primary folder touched
  (`skills`, `docs`, `templates`, etc.).
- **ADR-NNN and #issue come from the prompt**, not from the current
  conversation. Do not invent them.
- **Tests live in the same commit as the code they cover**, not in a
  separate "add tests" commit (see **Test-alongside**).
- **ADR index sync.** Before any commit that includes a file under
  `design/adr/adr-*.md` (new, modified, or status-changed), run
  `bin/sync-adr-index` and stage `design/adr/README.md` alongside the
  ADR files. Per ADR-023; keeps the index in sync with the
  filesystem on every commit that touches ADRs.

Example for this very issue:

```
feat(skills): add claude-issue-executor skill (ADR-014, #16)
docs(skills): wire claude-issue-executor into skills/README.md (ADR-014, #16)
```

## Test-alongside

Detect whether the project has a test runner by checking, in order:

- `package.json` → `scripts.test` (Node/JS/TS projects)
- `pytest.ini`, `pyproject.toml` with a `[tool.pytest.ini_options]` section
  (Python)
- `Cargo.toml` with a `[lib]` or `[[test]]` section (Rust)
- `go.mod` (Go — `go test ./...` is always available)
- any file ending in `*test*.*` in a conventional tests folder

If any is present, write tests as part of the logical commit that
introduces the code they cover. If the project has **no** test runner
(common for docs-only kits like this one — see this very repo), note
that in the evaluation summary and rely on documentation-style
verification instead.

Tests for a prose-only skill (a SKILL.md that contains no runnable code)
are usually unnecessary; a walk-through in the evaluation summary is the
accepted substitute.

## The plan gate — how plan-first is enforced

This is the single most important rule of the skill. Plan-first is
enforced by following this exact sequence every time:

1. Read the prompt file.
2. Validate it (see **Prompt validation**).
3. Read `CLAUDE.md` and the ADR referenced in the prompt.
4. **Produce a step-by-step plan in a single message** that includes:
   - the branch name you propose to create,
   - every file you will create or modify, with a one-line reason,
   - the commit sequence you intend (message + which files land in each),
   - whether tests will be written and where,
   - the verification steps you will run at the end.
5. **Stop.** Do not call any tool that writes to disk, runs git, runs
   tests, or otherwise mutates state. Wait for the user to answer.
6. Only proceed once the user replies with an explicit approval — e.g.
   "yes", "go ahead", "approved", "LGTM, proceed". A clarifying reply
   ("why are you doing X?") is **not** approval; address the question
   and re-present the plan.
7. If the user asks for changes, revise the plan and return to step 4.
8. If the user denies the plan outright, stop and offer to revise or to
   hand back control.

Operationally, when operating inside Claude Code: between steps 4 and 6
the assistant must not invoke Write, Edit, Bash (other than read-only
commands the user has already authorised), or any other mutating tool.
The only tool calls allowed between "plan proposed" and "plan approved"
are read-only ones needed to answer follow-up questions from the user.

This rule is load-bearing. Skipping it violates ADR-006 and defeats the
purpose of the skill.

See **Plan-mode rhythm** below for harness-level
enforcement that runs alongside this rule.

## Plan-mode rhythm and `--no-prompt` mode

The chat plan-gate above is the in-session 8-step rule. Claude Code
also ships a harness-level mechanism (plan mode, `shift+tab shift+tab`)
that locks mutating tools until the user exits. The two run together.

When the executor parses an invocation it classifies the session
against a significance/trivial checklist and routes to one of three
plan-mode branches, then optionally runs in `--no-prompt` mode for
trivial issues. Both flows are governed by
ADR-039
and ADR-038, and
the executor instances the **category-2** rule of the auto-mode
permission contract (see
[`docs/workflow-guide.md` §7](../../docs/workflow-guide.md#7-auto-mode-permission-contract-adr-041)).

See [`plan-mode.md`](plan-mode.md) for the significance and trivial
checklists, the hybrid routing path, auto-mode behaviour, the
alignment-review obligation, and the `--no-prompt` mode rules.

## Session protocol — end to end

1. **Resolve the prompt (auto-chain `prepare-issue`).**
   - If the argument is a *path*, treat as the prompt file directly.
   - If the argument is an *issue number*, resolve the prompt at
     `prompts/issue-NNN-*.md`. If absent, **auto-invoke
     `/prepare-issue NNN`** and log the prep step prominently. If
     present, run the staleness check: compare the prompt's mtime
     against `gh issue view <NNN> --json updatedAt` and against the
     mtime of every linked ADR. If any is newer, ask the user
     whether to regenerate (default: yes; never silent). On `n`,
     proceed with the existing prompt and note this in the
     evaluation summary.
   - If no argument was given, list candidate prompt files from
     `prompts/` and (fallback) `notes/issue*-prompt.md`. Ask the
     user which to use.
2. **Parse invocation and classify.** With the prompt resolved
   (or `--no-prompt` set — see [`plan-mode.md`](plan-mode.md)),
   **classify the session against the significance checklist** (see
   [`plan-mode.md`](plan-mode.md)): if clearly-significant, request the
   user toggle plan mode before continuing; if borderline, ask once;
   if clearly-trivial, proceed with the chat plan-gate alone.
3. **Preflight.** Confirm the working tree is clean. If dirty, stop and
   ask the user to commit, stash, or discard. Confirm the current branch
   is `main` or agree with the user to switch.
4. **Read and validate the prompt.** If malformed, stop and report.
   Skipped when `--no-prompt` is set; the issue body is used directly.
5. **Read the referenced ADR** and `CLAUDE.md`.
6. **Propose the plan** (see **The plan gate**).
7. **Wait for approval.** No edits until explicit yes.
8. **Create the branch** from `main`.
9. **Implement, commit incrementally.** Each commit is one logical
   change with the required message format. Tests land with code. If
   a new significant boundary appears mid-session (scope expands
   beyond what was classified at start, e.g. an unforeseen
   `skills/*/SKILL.md` edit or a new `templates/*` change), pause and
   re-flag for plan-mode entry. Do not silently cross the boundary.
   When `--no-prompt` is set, append the breadcrumb `issue executed
   without prompt per ADR-038` to the first commit's message.
10. **Verify.** Run the project's tests if any exist. Run any
    verification step called out in the prompt's "Evaluation & testing
    requirements" section.
11. **Update `design/state.md` if present.** Per
    ADR-035,
    rewrite the `state:in-flight` zone: keep `Issue` and `Prompt`,
    set `Branch` to the current branch, set `Status: verified`. Also
    set the `state:next-action` zone (ADR-048;
    [`docs/workflow-control.md` §4](../../docs/workflow-control.md#4-finding-the-next-step))
    to `skill: pr-review-packager`, `args: n/a`,
    `preconditions: ["branch <name> is ahead of main"]`,
    `blocked-by: none` — skip that zone if the file predates it. Use
    `bin/fence replace --file design/state.md --dialect state --zone <zone> --body-file <tmp>`
    for each rewrite rather than editing markers in the prompt. The
    `recent` zone is **not** touched here — it tracks merged PRs and
    is the responsibility of `/pr-review-packager`. If the file is
    absent, skip silently. If marker fences are broken, `bin/fence`
    exits 1 instead of writing; report the broken zone in the
    evaluation summary's follow-ups list and suggest `/pause`.
12. **Evaluation summary.** Print the final summary (see **Evaluation
    summary** below) **and write the same content to
    `notes/eval-issue-NNN.md`** (zero-padded issue number). The
    persisted file is what `/pr-review-packager` reads for the
    cross-skill carry-forward. For `--no-prompt` runs,
    the summary explicitly notes the bypass for the audit trail.
13. **Suggest handoff.** Tell the user the next step is
    `/pr-review-packager` to package a PR. Do not invoke it.

## Evaluation summary

Print a single structured message that covers:

- **What changed** — bulleted list of files created or modified, grouped
  by purpose.
- **Commits** — the commit messages in order, with SHAs when known.
- **Verification performed** — test runner output, build output, or a
  walk-through if the project has no runnable tests.
- **Follow-ups** — anything noted during the session that is out of
  scope for this issue but worth capturing (e.g. "this also enables
  issue #18"). Do not silently defer requirements that were in scope.
  When a design question affects an upcoming issue, add a structured
  `### design-questions` subsection here per the rule below.
- **Commands the user should run** to reproduce the verification
  themselves. Concrete shell lines, not prose.
- **Next step** — `/pr-review-packager` once the session is reviewed.

The same content is written to `notes/eval-issue-NNN.md` so
`/pr-review-packager` can read it deterministically.
The on-disk format mirrors the printed format exactly.

### `### design-questions`

When the executor session raises a design question that affects an
upcoming issue, append a structured `design-questions` block under
`## Follow-ups` of the eval summary. The canonical schema, field
semantics, and `target-issue` quoting rule live in
[`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040),
whose machine-readable mirror is
[`schemas/design-questions.v1.yaml`](../../schemas/design-questions.v1.yaml)
— the block this skill writes is what `bin/validate-carry-forward`
checks against that schema. Do not restate the rules here. The
reader's-aid example below shows the on-disk shape only:

````markdown
### design-questions

```yaml
- title: <one-line problem statement>
  target-issue: "#<N>"
  context: |
    <one-paragraph context note>
```
````

See [`reference.md`](reference.md#design-questions-populate-rules) for
the When-to-populate / When-NOT-to-populate / Empty-case rules.

## Edge cases

See [`reference.md`](reference.md#edge-cases) for the full list:
prompt not found, prompt malformed, dirty working tree, branch already
exists, plan denial, test failures, missing ADR/issue numbers.

## Receipts

This is a mutating (cat-2) skill, so it records an idempotency receipt
keyed by the **issue number**, per
[`docs/receipts.md`](../../docs/receipts.md):

- **Before** starting branch/commit work, check for an existing receipt
  (`bin/write-receipt --find --skill claude-issue-executor --key <issue>`,
  or read `.claude/receipts/claude-issue-executor__<issue>.json`). A
  `completed` receipt means the issue's work already landed — stop and
  report rather than re-running it on a resumed session.
- **On completion**, write a `completed` receipt recording the branch and
  any PR in `outputs`. On an interrupted run leave a `partial`/`failed`
  receipt whose `next-action` names what remains. Receipts are local
  state (gitignored); writing one is best-effort and never blocks the
  handoff.

## Handoff

On successful completion, the final message suggests:

> Next step: `/pr-review-packager` to draft a pull request from this
> branch, using the commit history and the governing ADR.

This skill does not invoke `/pr-review-packager` itself. The user opens
the next session deliberately, preserving a human review checkpoint
between implementation and PR publication.

## Alignment check

Before finishing the session, run the seven-item alignment checklist
in [`reference.md`](reference.md#alignment-check). If any box is
unchecked, the skill has drifted — say so in the evaluation summary
rather than hiding it.

See [`example.md`](example.md) for a worked session driven by
`prompts/issue-017-pr-review-packager.md`.
