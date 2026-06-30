# claude-issue-executor — worked example

A walk-through of a single end-to-end session driven by the prompt at
`notes/issue17-prompt.md` (legacy `notes/` location — equally applies to
a `prompts/issue-017-pr-review-packager.md` from `/prepare-issue`).

The goal of this example is to show how the skill enforces the plan-first
discipline, how it names the branch, what the commits look like, and what
the evaluation summary includes at the end.

---

## 1. Invocation

User types:

```
/claude-issue-executor notes/issue17-prompt.md
```

### Sidebar — Plan-mode rhythm (per ADR-039)

Before stepping into preflight, the executor classifies this session
against the **significance checklist** in `SKILL.md`. Issue #17 ships
a new skill (`skills/pr-review-packager/SKILL.md`) — that's a
`skills/*/SKILL.md` edit, so the session is **clearly-significant**.
The hybrid path runs:

1. Assistant reports the classification and *requests* plan-mode entry:
   *"Significant — touches `skills/pr-review-packager/SKILL.md`. Toggle
   plan mode (`shift+tab shift+tab`) when ready."*
2. User toggles plan mode.
3. Assistant proposes the plan inside plan mode (the chat plan-gate's
   8-step rule operates here, with the harness lock providing
   belt-and-braces enforcement). See step 4 below.
4. User approves and exits plan mode (`shift+tab`); optionally enables
   auto-accept edits (`shift+tab` again) for the execution phase.
5. Assistant executes (steps 6–9 below). If a new significant boundary
   appears mid-session — say, an unforeseen `templates/*` change — the
   assistant pauses and re-flags before crossing it.

A trivial session (single typo, ADR status flip, `feature-ideas.md`
status flip) skips this sidebar entirely; the chat plan-gate alone
suffices. Borderline sessions trigger a one-line *"Significant?"*
question instead.

## 2. Preflight

The skill checks, in order:

- Does the file exist? Yes.
- Is the working tree clean? Yes.
- Is the current branch `main`? Yes.

Any "no" here would stop the session with an explicit question for the
user — it would not silently fix the problem.

## 3. Prompt validation

The skill reads `notes/issue17-prompt.md` and checks for the required
sections:

- Context — present
- ADR (`design/adr/adr-015-pr-review-packager-skill.md`) — present
- GitHub Issue (`Number: #17`) — present
- Goal — present
- Requirements — present
- Scope and constraints — present
- Instructions for you — present

Extracted fields:

- **ADR number:** `015`
- **Issue number:** `17`

## 4. Plan proposal (the plan gate)

The skill now proposes the plan in a single message. **No writes happen
yet.** Between this message and the user's approval, the skill invokes
only read-only tools.

> Here is the plan for issue #17 (ADR-015, pr-review-packager skill):
>
> **Branch:** `pr-review-packager-skill` off `main`.
>
> **Files to create:**
>
> - `skills/pr-review-packager/SKILL.md` — prose-for-Claude orchestration
>   covering template read, commit-history summary, Closes/ADR extraction,
>   the user-approval gate, and the `gh pr create` invocation on approval.
> - `skills/pr-review-packager/example.md` — worked walk-through of a
>   filled PR body.
>
> **Files to modify:**
>
> - `skills/README.md` — flip the `pr-review-packager/` row from "later"
>   to "Issue #17".
>
> **Commit sequence:**
>
> 1. `feat(skills): add pr-review-packager skill (ADR-015, #17)` —
>    `SKILL.md` and `example.md`.
> 2. `docs(skills): wire pr-review-packager into skills/README.md (ADR-015, #17)` —
>    the README table update.
>
> **Tests:** this repo has no test runner (doc-only kit); verification is
> a walk-through trace in `example.md`.
>
> **Verification at the end:**
>
> - confirm the two files exist and the README row is updated,
> - mentally run the skill against `notes/issue17-prompt.md` and confirm
>   the commit/branch/message shape is correct,
> - re-read `design/adr/adr-015-pr-review-packager-skill.md` to confirm no
>   drift from the decision.
>
> Approve to proceed?

## 5. The plan gate in action

User replies: `go ahead`.

Only now does the skill start writing files and running git. If the user
had replied with a question — "why two commits instead of one?" — the
skill would answer, re-present the plan if the answer changed anything,
and wait again.

## 6. Branch creation

```bash
git checkout main
git pull
git checkout -b pr-review-packager-skill
```

(`git pull` is skipped if the user's workflow never syncs `main` from a
remote — for local-only kits the skill notices the absence of a tracking
branch and skips the pull.)

## 7. Incremental implementation

### Commit 1 — add the skill

Files written:

- `skills/pr-review-packager/SKILL.md`
- `skills/pr-review-packager/example.md`

Commit:

```
feat(skills): add pr-review-packager skill (ADR-015, #17)
```

### Commit 2 — wire into the skills README

File modified:

- `skills/README.md` (row flipped from "later" to "Issue #17")

Commit:

```
docs(skills): wire pr-review-packager into skills/README.md (ADR-015, #17)
```

Each commit touches a single logical concern. The commit message cites
both the ADR (`ADR-015`) and the issue (`#17`) verbatim from the prompt.

## 8. Verification

The repo has no runnable tests. The skill runs a trace-style check:

- Walks through `skills/pr-review-packager/SKILL.md` with the example
  branch's commit messages as input, confirming the skill would extract
  `Closes #17` and `Related ADR: design/adr/adr-015-...md`.
- Re-reads `templates/pr-template.md` and confirms every required
  section is addressed by the skill's drafting protocol.
- Re-reads ADR-015 and confirms the approval-gate language in the
  skill matches the decision.

## 9. Evaluation summary

The skill prints:

> **What changed**
>
> - Added `skills/pr-review-packager/SKILL.md` (orchestration, approval
>   gate, drafting protocol).
> - Added `skills/pr-review-packager/example.md` (worked walk-through).
> - Updated `skills/README.md` (row for `pr-review-packager/` now marked
>   "Issue #17").
>
> **Commits**
>
> - `<sha1>` feat(skills): add pr-review-packager skill (ADR-015, #17)
> - `<sha2>` docs(skills): wire pr-review-packager into skills/README.md (ADR-015, #17)
>
> **Verification performed**
>
> - Walk-through trace of the drafting protocol against this branch's own
>   commits — passes.
> - ADR-015 re-read — no drift.
> - `templates/pr-template.md` re-read — every section covered.
>
> **Follow-ups**
>
> - `#18` (changelog skill) can consume the commit history produced by
>   this session directly.
> - None of the scope was deferred.
>
> **Commands to reproduce verification**
>
> ```bash
> git log --oneline main..HEAD
> cat skills/pr-review-packager/SKILL.md
> cat skills/README.md | grep pr-review-packager
> ```
>
> **Next step:** run `/pr-review-packager` to draft the PR from this
> branch. This skill does not auto-invoke it.

## 10. What the plan gate prevented

If the skill had skipped the plan gate and jumped straight into writing
files, several things would have gone wrong or been harder to catch:

- The branch name would not have been reviewed by the user in advance.
- A decision to split the work into two commits vs. one would have been
  silent.
- Any scope disagreement (e.g. "please also add a `--dry-run` flag")
  would have surfaced only after the code was written.

The plan gate is cheap: one message, one reply. It is the single
mechanism that turns "plan-first" from a convention into an enforceable
contract.
