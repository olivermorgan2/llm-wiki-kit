---
name: prepare-issue
description: Auto-fill an implementation prompt file from a GitHub issue number, linked ADRs, and the build-out plan, written to prompts/issue-NNN-short-title.md. Use when preparing to work on an issue before implementation; use /claude-issue-executor after prompts/issue-NNN-*.md exists.
permission-category: 1  # substitutable — reads gh non-mutating; writes prompt file locally, per workflow-guide §7
inputs:
  - name: "issue-number"
    required: true
    description: "GitHub issue number (no #)"
  - name: "--skip-check"
    required: false
    description: "Opt out of the /check-plan pre-write gate"
  - name: "--pr-scan-limit"
    required: false
    description: "Merged-PR scan window for carry-forward notes (default 30)"
outputs:
  - artefact: "prompts/issue-NNN-short-title.md"
    description: "Filled implementation prompt"
  - artefact: "design/state.md"
    description: "Updated (in-flight issue, continue-here)"
next:
  - skill: claude-issue-executor
    when: "the prompt file is written"
---

# prepare-issue

Take a GitHub issue number, pull the issue's title, body, labels,
milestone, and linked ADR references via `gh`, read any relevant
project context (most importantly `design/build-out-plan.md`), fill
`prompts/_template.md`, and write the result to
`prompts/issue-NNN-short-title.md` for the user to paste into a fresh
Claude Code session.

This skill automates the biggest point of friction between "issue
exists on GitHub" and "Claude Code session is briefed" — see
ADR-013. It is
also the **consumer** in the cross-skill design-question
carry-forward chain decided in
ADR-040
— it scans recently-merged PRs for `## Notes for #N` sections and
embeds them in the generated prompt. See
[`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040)
for the canonical schema.

## When to use this skill

Use when the user has a GitHub issue ready to work on and wants a
filled prompt file in `prompts/`. Typical invocation:

```
/prepare-issue 17
```

The argument is a single GitHub issue number (no `#`). The skill is
**read-only** with respect to GitHub — it never creates, edits, or
comments on issues.

If the user does not yet have a GitHub issue backlog, run
`issue-planner` first. If the filled prompt already exists
and is still accurate, skip this skill and re-use the existing file.

## What this skill does not do

- Does not create or modify GitHub issues. Use `issue-planner` for
  that.
- Does not start the implementation session. The filled prompt is a
  handoff artifact; the user (or `claude-issue-executor`, ADR-014)
  runs it in a fresh session.
- Does not modify `prompts/_template.md`. It reads it only.
- Does not invent content when issue metadata is missing. Missing
  fields are left as clearly-marked TODOs.
- Does not overwrite an existing prompt file without explicit user
  confirmation.

## Inputs

- **Required:** a single GitHub issue number, passed as the skill's
  argument (e.g. `/prepare-issue 17`).
- **Implicit repo context:** the current working directory must be
  inside a git repo whose `origin` is the GitHub repo containing the
  issue. `gh` infers the repo from git remotes.
- **Optional flag:** `--skip-check` — opt out of the `/check-plan`
  pre-write gate. Default is on (gate runs); the flag is documented as opt-out for
  rapid iteration on known-good drafts only. When set, the skill
  writes the prompt despite any deterministic-criteria failures and
  appends a one-line breadcrumb to the prompt body —
  `<!-- /check-plan was skipped via --skip-check per ADR-034 -->` —
  so the bypass is auditable.
- **Optional flag:** `--pr-scan-limit <N>` — number of recently-merged
  PRs to scan for cross-issue carry-forward notes. Default is 30,
  which covers ~2-3 milestones at typical kit
  cadence. Override only when working in a repo with unusually high
  or low PR cadence. The scanned window is surfaced to the user at
  the review-before-write checkpoint.

## Output

- **File:** `prompts/issue-NNN-short-title.md` in the target project.
- **Shape:** exactly the section order of `prompts/_template.md`:
  Context, ADR, GitHub Issue, Goal, Why it matters, Requirements,
  Acceptance criteria, Scope, Evaluation, Instructions.
- **Filename convention:** zero-padded to 3 digits, kebab-case short
  title derived from the issue title (see "Short-title derivation"
  below).

## Data sources and how the skill reads them

The skill consults four sources, in priority order:

1. **GitHub issue** (required) — via
   `gh issue view <N> --json title,body,labels,milestone,url`.
   This is the primary source for almost every field.
2. **Linked ADRs** (optional) — any `ADR-NNN` token or
   `design/adr/adr-NNN-*.md` link mentioned anywhere in the issue
   body, title, or labels. For each referenced ADR, read the file
   from `design/adr/` and extract its title and one-line decision
   summary.
3. **Build-out plan** (optional) — `design/build-out-plan.md`, if it
   exists. Grep it for mentions of the issue number or the issue's
   core noun phrase and pull a short contextual paragraph.
4. **Recently-merged PRs** (optional) — via
   `gh pr list --state merged --limit <N> --json number,body,mergedAt`
   where `<N>` is the `--pr-scan-limit` value (default 30). For each
   PR's body, scan for `^## Notes for #<NNN>` sections where `NNN`
   matches the new issue's number. Collect matches as
   `(pr-number, notes-body, merged-at)` tuples. Most issues have
   zero matches; a small minority have one. The scanned window is
   surfaced to the user at the review-before-write checkpoint so
   they can extend the limit if a known carry-forward source fell
   outside the default. Schema source of truth:
   [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040).

`gh` is the only allowed tool for GitHub reads. Do not call the
REST/GraphQL API directly.

## Execution protocol

Run these steps in order. Stop on the first failure unless otherwise
noted.

1. **Validate input.** Confirm the argument is a positive integer.
   If not, ask the user for the issue number and stop.
2. **Check the template exists.** Read `prompts/_template.md`. If it
   is missing, tell the user to run Issue #12's work first and stop.
3. **Fetch the issue.** Run
   `gh issue view <N> --json title,body,labels,milestone,url`. If
   `gh` exits non-zero (issue not found, no auth, wrong repo), print
   the error message verbatim and stop. Do not attempt to guess or
   proceed with partial data.
4. **Parse the issue JSON.** Extract:
   - `title` → issue title string
   - `body` → markdown body
   - `labels[].name` → comma-separated label list
   - `milestone.title` → milestone name (or "none")
   - `url` → canonical GitHub URL (not used in the template but
     useful for logging)
5. **Resolve ADR references.** Scan the issue title + body for both
   forms:
   - `ADR-NNN` tokens (case-insensitive, with or without zero-pad).
   - Explicit path links like `design/adr/adr-NNN-short.md` or
     `adr-NNN-short.md`.
   For each unique `NNN`, glob `design/adr/adr-<NNN>-*.md`. If a
   match is found, read the file and extract:
   - the filename (for the `ADR.File` line),
   - the ADR title from the `# ADR-NNN: Title` heading,
   - a one-line decision summary from the `## Decision` section
     (first sentence is usually enough).
   If no ADR references are found, mark the ADR section with
   "ADR: none — no ADR referenced in this issue." per the template
   comment.
6. **Read build-out plan if present.** If
   `design/build-out-plan.md` exists, grep it for the issue number
   and for noun phrases from the issue title. Capture any short
   matching paragraph as supplementary context to include in the
   "Context" section or the "Why it matters" section where it fits
   naturally. If the file does not exist, skip silently — this is
   normal in a freshly-initialized kit repo.
6.5. **Scan recently-merged PRs for carry-forward notes.** Run
   `gh pr list --state merged --limit <N> --json number,body,mergedAt`
   where `<N>` is the `--pr-scan-limit` value (default 30). For each
   returned PR's body, look for sections matching
   `^## Notes for #<NNN>` where `NNN` is the new issue's number.
   Collect matches as `(pr-number, notes-body, merged-at)` tuples,
   ordered newest-first by `mergedAt`. Most issues have zero
   matches and the step is effectively a no-op; the rare match
   feeds the "Design questions carried forward from PR #M"
   subsection emitted by the template-fill rules. If `gh pr list`
   fails (auth, rate limit), surface the error verbatim, ask
   whether to proceed without the scan, and log the skip in the
   review-before-write output. Surface the scanned window to the
   user at the approval gate, e.g. *"Scanned 30 most-recent merged
   PRs (#42 → #71); found 1 'Notes for #69' section in PR #65."*
7. **Derive the short title.** See "Short-title derivation" below.
8. **Fill the template.** Read `prompts/_template.md` and substitute
   every `{{PLACEHOLDER}}` with the extracted value. For placeholders
   where data is not available, leave the placeholder in place and
   append ` <!-- TODO: fill in -->` so the user sees the gap
   immediately.
9. **Show the user the filled prompt in the chat** as a fenced
   markdown block. Ask explicitly: "Write this to
   `prompts/issue-NNN-short-title.md`? (yes / edit / cancel)".
   **Do not write the file before this confirmation.**
10. **Pre-write check.**
    Unless `--skip-check` was passed, after the user confirms with
    `yes`, pipe the in-memory filled prompt into the kit's
    programmatic check-plan surface:
    ```
    <rendered-prompt> | bin/check-plan --criteria-set prompt --input - --format json
    ```
    Parse the JSON envelope. On exit 0, proceed to step 11 (file-
    exists check). On exit 1, surface the failures (each citing
    its criterion ID — e.g. `PROMPT-C1` — and the `remediation`
    field from the JSON) to the user, ask how to revise, apply
    the revision in memory, re-show the updated prompt for
    confirmation, and re-invoke `bin/check-plan` for round 2.
    After 3 failed rounds, yield: surface the remaining failures
    and stop without writing the file. Warnings (`status: warn`)
    are surfaced but do not block. `--skip-check` short-circuits
    the gate and the skill proceeds to step 11 with a one-line
    breadcrumb appended to the prompt body.

    The programmatic surface is what skills with
    chain points invoke — slash-commands aren't invokable from
    inside another skill's execution. The slash-command surface
    `/check-plan` remains for direct operator use; both share the
    same eval module via `skills/check-plan/criteria.md` as the
    single canonical criteria list.
11. **Handle the file-exists case.** Before writing, check whether
    `prompts/issue-NNN-short-title.md` already exists. If it does,
    show a diff between the existing file and the new content and
    ask explicitly whether to overwrite. Default to "no".
12. **Write the file** only after explicit confirmation. Report the
    absolute path and a one-line summary of what was filled vs. left
    as TODO.
13. **Update `design/state.md` if present.** Per
    ADR-035,
    rewrite two zones:
    - `state:in-flight` → set `Issue: #NNN`, `Prompt:` to the just-
      written path, `Branch: n/a` (executor will set it later),
      `Status: prepared`.
    - `state:continue-here` → one short paragraph naming the next
      action: `"Run /claude-issue-executor prompts/issue-NNN-…md"`.
    - `state:next-action` → the structured complement (ADR-048;
      [`docs/workflow-control.md` §4](../../docs/workflow-control.md#4-finding-the-next-step)):
      `skill: claude-issue-executor`, `args: "<prompt path or NNN>"`,
      `preconditions: ["prompt prompts/issue-NNN-…md exists"]`,
      `blocked-by: none`. Skip this zone if the file predates it
      (no `state:next-action` fences) — do not add it here.
    Rewrite each zone with `bin/fence` rather than editing markers in
    the prompt — it replaces only the bytes between a zone's fences:
    `bin/fence replace --file design/state.md --dialect state --zone <zone> --body-file <tmp>`.
    Other zones (`phase`, `recent`, `blockers`) are left untouched. If
    `design/state.md` is absent, skip silently — this is normal in a
    kit repo or a project that has not adopted ADR-035. If the file
    exists but its marker fences are broken, `bin/fence` exits 1
    instead of writing; tell the user and suggest `/pause` to refresh.

## Reference: derivation, template-filling, edge cases, self-check

See [`reference.md`](reference.md) for:

- **Short-title derivation** — kebab-case algorithm with worked examples.
- **Template-filling rules** — full `{{PLACEHOLDER}}`-to-source slot map and the multi-ADR repetition rule.
- **Carry-forward subsection** — format, placement (immediately before `Requirements`), and the schema source-of-truth pointer to [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040) (machine-readable mirror: [`schemas/design-questions.v1.yaml`](../../schemas/design-questions.v1.yaml)).
- **Edge cases** — invalid argument, gh auth/lookup failures, missing ADR file, missing build-out-plan, file-exists handling, /check-plan gate yields, --skip-check breadcrumb, PR-scan errors, multi-PR carry-forward.
- **Self-check before writing** — the 8-item pre-write checklist.

## Review-before-write checkpoint

The user always sees the filled prompt as chat output before the
file is written. This preserves the review gate from ADR-013's
Option A (the chosen option). Never skip this step, even if the
filled prompt looks clean.

## Handoff

The filled `prompts/issue-NNN-*.md` is the handoff point into a
Claude Code implementation session. The user either pastes it into a
fresh session manually, or (once ADR-014 lands) runs
`claude-issue-executor` to consume it. This skill is done once the
file is written and the path is reported.

See [`example.md`](example.md) for a worked invocation on a real
issue.
