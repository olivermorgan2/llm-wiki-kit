---
name: pr-review-packager
description: Draft a pull-request body from templates/pr-template.md, auto-fill the Closes line and ADR references from branch and commit history, show the draft for approval, then open the PR via gh pr create. Use when packaging a feature branch into a PR after implementation commits land.
permission-category: 3  # non-substitutable — gh pr create is public, hard-to-reverse, per workflow-guide §7
inputs:
  - name: "--label / --milestone / --reviewer"
    required: false
    description: "Pass-through flags to gh pr create"
  - name: "--draft"
    required: false
    description: "Open the PR as a draft"
  - name: "--base"
    required: false
    description: "Override the base branch (default main)"
outputs:
  - artefact: "(GitHub PR)"
    description: "PR opened via gh pr create"
  - artefact: "design/state.md"
    description: "Updated (in-flight cleared, recent prepended)"
next: []
---

# pr-review-packager

Package a feature branch into a well-structured pull request. The skill
reads [`templates/pr-template.md`](../../templates/pr-template.md),
fills in `Closes #N` and ADR references from the branch name and commit
history, derives a change summary from `git log main..HEAD`, shows the
finished body in chat for approval, and only then calls
`gh pr create`.

This skill completes the implementation pipeline:
`/prepare-issue` → `claude-issue-executor` → **`/pr-review-packager`**.
See ADR-015 for
the decision record. The skill is also the **preserver** in the
cross-skill design-question carry-forward chain decided in
ADR-040
— it reads the executor's `notes/eval-issue-NNN.md` and emits
`## Notes for #N` sections in the PR body so the carry-forward
survives in PR history. See
[`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040)
for the canonical schema.

## When to use this skill

Use when an implementation branch is ready for review and the user
wants to open the PR without hand-filling the template. Typical
invocation:

```
/pr-review-packager
```

Optional pass-through flags are supported and forwarded verbatim to
`gh pr create`:

```
/pr-review-packager --label feature --milestone "MVP" --reviewer @alice
```

The skill is **read-only** with respect to `templates/pr-template.md`
— it consumes the file without editing it.

## What this skill does not do

- Does not edit `templates/pr-template.md`. Ever. It is a read-only
  consumer.
- Does not modify the local branch or rewrite commits. If the commit
  history needs cleaning up, the user does that before invoking the
  skill.
- Does not push the branch. Pushing is the user's responsibility; the
  skill fails cleanly if the upstream branch does not exist yet and
  reports the exact `git push -u origin <branch>` command to run.
- Does not create a draft PR by default. If the user wants one they
  can pass `--draft`.
- Does not review the PR content for correctness (that is `/review`,
  a built-in). It only packages the body.
- Does not run CI, tests, or linters. It only reports what `git log`
  shows.
- Does not open the PR without explicit user approval of the rendered
  body. The approval gate from ADR-015's chosen option (Option B) is
  mandatory.

## Auto-mode permission category

This skill is **category 3** — *non-substitutable* — in the kit-wide
auto-mode permission contract. See
[`docs/workflow-guide.md` §7](../../docs/workflow-guide.md#7-auto-mode-permission-contract-adr-041)
for the canonical contract.

`gh pr create` is public-visibility and hard-to-reverse: a PR opened
in error is visible to collaborators, may trigger CI runs and
notifications, and requires explicit close-or-edit follow-up. This
class of operation is **never** substituted by auto-mode regardless
of mode.

The contract requires: **explicit `yes` from the operator before
`gh pr create` is called, regardless of whether auto-mode is
active**. Steps 12–13 of the execution protocol below already
implement this — the user sees the rendered title and body in chat
and must reply `yes` before the skill calls `gh`. ADR-041 does not
change this behaviour; it codifies the existing convention as a
contract so future maintenance cannot regress it.

If a future change ever proposes silencing the explicit-yes gate
under auto-mode, it must be rejected at PR review as a cat-3
violation. The same rule applies to any other operation the skill
might add that touches public state (e.g. labelling, assigning
reviewers as a side-effect of merge): each must gate behind an
explicit-yes prompt that auto-mode does not satisfy.

## Inputs

- **Required:** none — the skill reads everything it needs from the
  current git worktree, the repo, and `gh`.
- **Optional flags** (passed through to `gh pr create`):
  - `--label <name>` (repeatable)
  - `--milestone <name>`
  - `--reviewer <user-or-team>` (repeatable)
  - `--draft`
  - `--base <branch>` (defaults to `main`)
  - `--assignee <user>` (repeatable)

Flags the skill does **not** forward: anything that would suppress the
body (`--fill`, `--body-file`), anything that would bypass the approval
gate, and anything destructive.

## Output

- A new pull request on GitHub, opened via
  `gh pr create --title "..." --body "..." --base <base> --head <current-branch>`
  plus any pass-through flags.
- A one-line confirmation in chat including the PR URL returned by
  `gh`.

## Data sources and how the skill reads them

The deterministic extraction — the issue-number fallback, ADR-token
resolution, commit grouping, the PR-title draft, and the
design-questions carry-forward read — is done by `bin/pr-context`
(`bin/pr-context --base <base> --format json`); the skill consumes its
`outputs` and writes the narrative. The five underlying sources
`pr-context` reads, in priority order, are:

1. **Current branch and git history.**
   - `git symbolic-ref --short HEAD` → current branch name. A detached
     HEAD aborts the skill (see Edge cases).
   - `git rev-parse --abbrev-ref --symbolic-full-name @{u}` → upstream
     ref, used to confirm the branch has been pushed. If no upstream
     is set, the skill stops with a clear message telling the user to
     run `git push -u origin <branch>` first.
   - `git log <base>..HEAD --format="%H%x09%s"` → commits ahead of
     base, used for the change summary and as a secondary source for
     issue and ADR tokens. `<base>` is `main` unless the user passed
     `--base`.
2. **PR template.** Read `templates/pr-template.md` verbatim. If it is
   missing, abort (see Edge cases).
3. **Issue prompt file** (fallback). Glob `prompts/issue-NNN-*.md` and
   pick the most recently modified file. Useful when the branch name
   does not encode the issue number.
4. **ADR files.** For each `ADR-NNN` token found, glob
   `design/adr/adr-<NNN>-*.md` and resolve the full filename.
5. **Eval-summary file** (optional). Once the issue
   number is resolved (step 6 of the execution protocol), read
   `notes/eval-issue-NNN.md` (zero-padded). Parse the
   `### design-questions` YAML block under `## Follow-ups` if
   present. Group entries by `target-issue`; each unique
   `target-issue` becomes one `## Notes for #M` section in the
   rendered PR body. If the file is absent or has no
   `### design-questions` block, skip silently — most issues have
   no entries. Schema source of truth:
   [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040),
   mirrored machine-readably in
   [`schemas/design-questions.v1.yaml`](../../schemas/design-questions.v1.yaml).

`gh` is the only allowed tool for creating the PR. Do not call the
REST/GraphQL API directly.

## Execution protocol

Run these steps in order. Stop on the first failure unless otherwise
noted.

1. **Verify working tree.** Run `git status --porcelain`. If it is
   non-empty, warn the user and ask whether to continue — the
   uncommitted changes will not be included in the PR.
2. **Resolve the branch.** Run `git symbolic-ref --short HEAD`. On a
   detached HEAD, abort with:
   `"HEAD is detached. Check out a branch before running /pr-review-packager."`
3. **Confirm base.** The base branch is `main` unless `--base` was
   passed. If the current branch equals the base, abort with:
   `"Current branch is <base>. Check out a feature branch before packaging a PR."`
4. **Confirm the branch has been pushed.** Run
   `git rev-parse --abbrev-ref --symbolic-full-name @{u}`. If it
   errors, abort with:
   `"No upstream for <branch>. Run: git push -u origin <branch>"`.
5. **List commits ahead of base.** Run
   `git log <base>..HEAD --format="%H%x09%s"`. If the output is empty,
   abort with:
   `"No commits on <branch> ahead of <base>. Nothing to package."`
6. **Gather deterministic context.** Run
   `bin/pr-context --base <base> --format json` and read its `outputs`
   for the rest of this section — it performs the issue/ADR/commit
   extraction the skill used to describe inline. The issue number is
   `outputs.issueNumber` (no `#`); `outputs.issueSource` names which
   priority source won (branch name → commit subjects → newest issue
   prompt → `none`). When `issueSource` is `none`, leave
   `{{ISSUE_NUMBER}}` as `<!-- TODO: fill in issue number -->` and flag
   it during the approval gate rather than blocking.
6.5. **Carry-forward design questions.** Read `outputs.carryForward[]`
    — `pr-context` already read `notes/eval-issue-NNN.md`, parsed the
    `### design-questions` block under `## Follow-ups`, and grouped
    entries by `target-issue` (ascending). Each group is
    `{targetIssue, entries:[{title, context}]}` and becomes one
    `## Notes for #M` section in step 11. An empty list means none
    (skip silently). If `outputs.warnings` reports a malformed block,
    surface it at the approval gate; do not abort. The canonical schema
    and field semantics live in
    [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040).

7. **ADR references.** Read `outputs.adrs[]` — each is
   `{token, path, resolved}`, deduped in first-seen order across the
   branch, commit subjects, and newest prompt. Record each resolved
   `path` as a `Related ADR:` line; flag an unresolved token as a TODO
   at the approval gate. If the list is empty, set the `## ADR` body to
   `none`.
8. **Change summary.** Read `outputs.commitGroups[]` — each is
   `{verb, bullets[]}` with bullets oldest-first and the trailing
   `(ADR-NNN, #N)` already stripped. (`infra` is one of the recognised
   verbs, matching the kit's canonical label set; unrecognised prefixes
   group under `other`.) Render one bullet per entry under its group.
   If there is only one group, omit the group heading and just list the
   bullets.
9. **PR title.** Use `outputs.prTitle` — the newest commit subject with
   the trailing `(ADR-NNN, #N)`/`(#N)` already stripped. The user can
   override during approval.
10. **Draft the Summary paragraph.** One sentence: take the verb and
    object from the PR title, plus the issue number if known. Example:
    `"Adds the pr-review-packager skill that packages feature
    branches into well-structured PRs (closes #17)."` This is a
    best-effort draft — the user will edit if needed.
11. **Fill the template.** Read `templates/pr-template.md` and
    substitute each placeholder:

    | Placeholder | Filled with |
    |---|---|
    | `{{One-paragraph description of what this PR changes and why.}}` | the Summary draft from step 10 |
    | `{{ISSUE_NUMBER}}` | issue number from step 6 (no `#`) |
    | `` `design/adr/adr-{{NNN}}-{{short-title}}.md` `` line | one resolved ADR path per line, each as `Related ADR: design/adr/adr-NNN-short-title.md`. If no ADRs, the whole line becomes `Related ADR: none`. |
    | `{{Bullet the substantive changes, not every file touched.}}` block | the grouped bullets from step 8 |
    | `{{Paste test-runner output: total / passed / failed / skipped.}}` | left as a placeholder `<!-- TODO: paste test-runner output or delete the code fence and write "no code changes — docs only" -->` |
    | `{{Steps a reviewer should run to convince themselves the change works.}}` | `<!-- TODO: list verification steps, or write "none needed" -->` |

    **After the rendered template body, append `## Notes for #M`
    sections (per ADR-040)**, one per unique `target-issue` from
    step 6.5, in ascending issue-number order. Each section has the
    form:

    ```markdown
    ## Notes for #<M>

    Carried forward from this issue's eval summary
    (`notes/eval-issue-NNN.md`). The next executor session for #<M>
    should address these in its plan.

    - **<title 1>**: <context paragraph 1>
    - **<title 2>**: <context paragraph 2>
    ```

    If step 6.5 produced no entries, no `## Notes for #M` sections
    are emitted (skip silently — most PRs have none). The section
    title format `## Notes for #<M>` is the kit-canonical anchor
    that `prepare-issue` later scans for; do not vary it.

    Strip the template's HTML comments that instruct the author (they
    are not part of the rendered body the skill produces).

12. **Show the filled body to the user** as a fenced markdown block
    plus the derived title on a line above it. Flag any TODO markers
    and any unresolved tokens (e.g. ADR file missing, no issue
    number). Ask explicitly:
    `"Open PR with this title and body? (yes / edit / cancel)"`.
    **Do not call `gh pr create` before this confirmation.**
13. **Handle the response.**
    - `yes` → proceed to step 14.
    - `edit` → ask the user what to change, apply it, re-render, re-ask.
      Loop until `yes` or `cancel`.
    - `cancel` → stop without calling `gh`.
14. **Open the PR.** Call:
    ```
    gh pr create \
      --title "<title>" \
      --body  "<body>" \
      --base  "<base>" \
      --head  "<current-branch>" \
      [pass-through flags]
    ```
    Pass-through flags are the `--label`, `--milestone`, `--reviewer`,
    `--assignee`, `--draft` values the user provided at invocation.
15. **Update `design/state.md` if present.** Per
    ADR-035,
    close out the issue in the state file:
    - prepend a one-line entry to the `state:recent` zone of the
      form `#<PR> — ADR-NNN — <commit-summary first sentence>`
      (use `none` for the ADR token if no ADR was resolved). If the
      zone now has more than five entries, drop the oldest so the
      rolling list stays at five.
    - rewrite the `state:in-flight` zone to `Issue: none`,
      `Prompt: n/a`, `Branch: n/a`, `Status: none`.
    - rewrite the `state:continue-here` zone to one short paragraph
      pointing at the just-opened PR, e.g. `"Review and merge #<PR>;
      then pick the next issue from the queue."`.
    - rewrite the `state:next-action` zone (ADR-048;
      [`docs/workflow-control.md` §4](../../docs/workflow-control.md#4-finding-the-next-step))
      to `skill: none`, `args: n/a`, `preconditions: []`,
      `blocked-by: "PR #<PR> awaiting review/merge"` — the next action
      is the human review of the open PR, not a skill. Skip this zone
      if the file predates it (no `state:next-action` fences).
    Rewrite each zone with `bin/fence replace --file design/state.md
    --dialect state --zone <zone> --body-file <tmp>` rather than editing
    markers in the prompt; it replaces only the bytes between a zone's
    fences. If `design/state.md` is absent, skip silently. If marker
    fences are broken, `bin/fence` exits 1 instead of writing; surface
    the broken zone in the final report and suggest `/pause`.
16. **Report.** Print the PR URL returned by `gh`, plus a one-line
    summary of what was filled vs. what was left as TODO. Done.

## Reference: extraction rules, edge cases, self-check

Detailed rules and tables live in [`reference.md`](reference.md):

- **Issue-link extraction** — branch-name → commit-subject → prompt-file priority order; leading-zero strip rule.
- **ADR-link extraction** — token regex, glob resolution, multi-ADR handling.
- **Change-summary derivation** — conventional-commit prefix grouping; group ordering with `infra` adjacent to `build`.
- **Edge cases** — detached HEAD, no upstream, missing template, multiple issues, existing PR.
- **Self-check before calling `gh`** — ADR-023 index sync and ADR-040 Notes-for-#M emission gates.
- **Relationship to other skills** — `/review`, `claude-issue-executor`, `/changelog`, `/release`.

## Review-before-create checkpoint

The user always sees the rendered title and body in chat before
`gh pr create` is called. This preserves the approval gate from
ADR-015's chosen option. Never skip this step — not even when the
draft looks clean, not even with a `--yes` flag (the skill does not
accept one).

## Receipts

Opening a PR is a mutating (cat-3) action, so the skill records an
idempotency receipt keyed by the **issue (or PR) number**, per
[`docs/receipts.md`](../../docs/receipts.md):

- **Before** creating the PR, check for an existing receipt
  (`bin/write-receipt --find --skill pr-review-packager --key <issue>`,
  or read `.claude/receipts/pr-review-packager__<issue>.json`). A
  `completed` receipt naming an open PR means one already exists — surface
  it instead of opening a duplicate.
- **On completion**, write a `completed` receipt with the PR number in
  `outputs`. Writing it is best-effort and never blocks the handoff;
  receipts are local, gitignored state.

## Handoff

The PR is the handoff. Once opened, the review and merge are
human-driven (or handled by `/review`). This skill is done when
`gh pr create` returns a URL and the URL is reported to the user.

See [`example.md`](example.md) for a worked invocation on this very
issue.
