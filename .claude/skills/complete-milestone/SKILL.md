---
name: complete-milestone
description: Close a GitHub milestone, archive milestone-scoped state in design/state.md, and optionally chain /release with --milestone-phase. Chains /audit-milestone and /milestone-summary; never blocks on audit gaps. Use when closing a completed milestone; pre-flight via /audit-milestone, post-close summary via /milestone-summary.
permission-category: 3  # non-substitutable — closes GitHub milestone via gh; chains /release with --release, per workflow-guide §7
inputs:
  - name: "milestone"
    required: true
    description: "Milestone number or title"
  - name: "--release / --no-release"
    required: false
    description: "Whether to chain /release on close"
  - name: "--dry-run"
    required: false
    description: "Preview without closing the milestone"
outputs:
  - artefact: "(GitHub milestone)"
    description: "Milestone closed via gh"
  - artefact: "design/state.md"
    description: "Updated (in-flight, recent, continue-here zones)"
next:
  - skill: release
    when: "--release is set"
---

# complete-milestone

Close a milestone end-to-end:

1. Run `/audit-milestone <N>` (chained). Report any gaps; require
   explicit `yes` if gaps are present, but **never block**.
2. Run `/milestone-summary <N>` (chained, unless `--skip-summary`).
   Wait for the user to author the lessons zone before continuing.
3. Close the GitHub milestone via `gh api`.
4. Archive `design/state.md` zones — clear `in-flight`,
   prepend the close to `recent`, set `continue-here` to the next
   action.
5. Optionally chain `/release --milestone-phase=N` if the milestone
   maps to a phase boundary.

This skill implements the close half of
ADR-037. It is
the natural endpoint of the milestone chain:
`/audit-milestone` → `/milestone-summary` → `/complete-milestone`
→ optional `/release`.

## When to use this skill

- When a milestone is finishable and the user wants to close it
  out — flip the GitHub milestone to `closed`, archive the state
  pointer, optionally cut the release.
- After `/milestone-summary` has run and the user has edited the
  `lessons` zone of the summary file.
- As the terminal step in a phase-to-release cadence: phase work
  merged → `/audit-milestone` → `/milestone-summary` → user
  authors lessons → `/complete-milestone --release`.

If the user only wants to verify completeness, that is
`/audit-milestone`. If they only want a retrospective without
closing, that is `/milestone-summary`. This skill is for the
**actual close** — it mutates the GitHub milestone state.

## What this skill does not do

- Does not bypass the audit. The audit always runs (chained); its
  result is surfaced to the user. But the audit does not gate the
  close.
- Does not author the `lessons` zone. The user does that between
  the summary step and the close step.
- Does not invoke `/release` without explicit consent. The
  `--release` flag (or the interactive prompt) is required.
- Does not delete the milestone. It only sets `state=closed`.
  Milestones stay in the repo's history for retrospective access.
- Does not modify `design/build-out-plan.md` itself. That is
  `/release --milestone-phase=N`'s job; this skill chains to it.

## Inputs

- **Required:** the milestone identifier (number or title; same
  resolution as `/audit-milestone` and `/milestone-summary`).
- **Implicit:** `gh` authenticated; clean working tree (the close
  may write to `design/state.md`); the matching summary file
  exists at `design/milestones/<N>-<slug>.md` (warn if absent —
  see edge cases).

## Invocation

```
/complete-milestone <N>
                    [--release]
                    [--no-release]
                    [--release-bump=major|minor|patch]
                    [--release-version=X.Y.Z]
                    [--skip-summary]
                    [--dry-run]
```

Flags:

- `--release` — chain `/release --milestone-phase=N` after the
  close. Skips the interactive release prompt.
- `--no-release` — skip the release chain entirely. Skips the
  interactive prompt.
- `--release-bump=<level>` — passed through to `/release`.
- `--release-version=X.Y.Z` — passed through to `/release`.
- `--skip-summary` — skip the `/milestone-summary` chain step.
  Use only when the summary file already exists and was authored
  in a prior session.
- `--dry-run` — walk every step up to the approval gate and
  stop. Render the plan; print each command that would run; make
  no mutations.

If neither `--release` nor `--no-release` is set, the skill
prompts at step 6 below.

## Execution protocol

Every mutating step is gated behind a single `yes` approval. The
user sees the full plan before any mutation.

1. **Validate input.** Argument is a positive integer or
   non-empty string. If neither, stop with usage.
2. **Resolve the milestone.** Same logic as `/audit-milestone`.
   Capture `number`, `title`, `state`, `description`,
   `closed_at`, `created_at`. If `state` is already `closed`,
   stop with `"Milestone <title> (#<N>) is already closed."`.
3. **Prerequisites check.**
   - `gh auth status` succeeds.
   - `git status --porcelain` is empty (clean working tree).
   - Current branch is the project's primary branch (`main` by
     default; configurable via the project's existing
     conventions).
4. **Chain `/audit-milestone <N>`.** Capture the result. On
   fail, surface the gap list to the user verbatim. The audit
   does not gate progress — but step 6's approval prompt
   incorporates the gap count so the user sees it before
   confirming.
5. **Chain `/milestone-summary <N>`** (unless `--skip-summary`).
   Pass `--dry-run` if `--dry-run` is set on this skill. The
   summary skill writes (or previews) `design/milestones/<N>-<slug>.md`.
   After it returns, prompt the user: *"Edit the lessons zone in
   design/milestones/<...>.md, then type `continue` to proceed.
   Type `cancel` to stop."* Wait for the literal `continue` or
   `cancel`. Any other input re-prompts.
6. **Assemble the close plan and render it for review:**
   - Milestone: `<title> (#<N>)`.
   - Audit result: pass / fail (with gap count).
   - Summary file path (with mtime — confirms the user edited it).
   - State.md changes (which zones will be rewritten and to what).
   - Release chain decision: `--release` / `--no-release` /
     interactive prompt result.
   - The exact `gh` and file-write commands that will run, in
     order.
7. **Approval gate.** Ask: *"Type `yes` to close milestone
   <title> (#<N>) and update design/state.md. Any other response
   cancels."* Accept only the literal `yes` (case-insensitive,
   trimmed). Any other input cancels cleanly.
8. **On `yes`, execute:**
   - Close the milestone:
     `gh api repos/:owner/:repo/milestones/<N> -X PATCH -f state=closed`.
   - Update `design/state.md` (only if file exists):
     - `in-flight` zone → set `Issue: none`, `Prompt: n/a`,
       `Branch: n/a`, `Status: none`.
     - `recent` zone → prepend
       `Milestone closed: <title> (#<N>) — <YYYY-MM-DD>`; trim
       to 5 entries (oldest drops).
     - `continue-here` zone → write one short paragraph naming
       the next action: either *"Run `/release --milestone-phase=N`"*
       (if `--release`), *"Open the next milestone"* (if
       `--no-release`), or the literal interactive choice.
     - Preserve `phase` and `blockers` zones verbatim.
   - Run `bin/check-state-cap` (if present) and surface any
     warning.
9. **On any other input,** cancel. Report
   *"Milestone close cancelled. No changes made."* Exit 0.
10. **Optional release chain.** If `--release` (or interactive
    `yes`), invoke `/release --milestone-phase=<phase-num>`,
    passing through `--release-bump` and `--release-version` if
    set. The release skill has its own approval gate; this skill
    does not pre-approve the release. If the user declines the
    release inside `/release`, the milestone stays closed —
    that's fine. The two are decoupled.
11. **Report.** Print:
    - The closed milestone URL
      (`gh api repos/:owner/:repo/milestones/<N> --jq .html_url`).
    - The summary file path.
    - The state.md zones rewritten.
    - The release URL if a release was cut, else
      *"Release skipped (or declined). Run `/release --milestone-phase=N`
      later to cut a tag."*

## Dry-run mode

With `--dry-run`:

- Run steps 1–6 of the close protocol. Render the full plan.
- Step 5 (`/milestone-summary`) runs in `--dry-run` mode too —
  the summary content is rendered to stdout, not written.
- Instead of step 7's approval, print *"Dry-run. Would execute:"*
  followed by the full command sequence with resolved values.
- Make zero mutations. No `gh api -X PATCH`. No `state.md` write.
  No release chain.

## Edge cases

| Situation | Behaviour |
|---|---|
| Milestone already closed | Stop with notice. Do not re-close. |
| Audit reports gaps | Surface verbatim; require explicit `yes` to proceed; do not block. |
| Summary file absent (and `--skip-summary` not set) | Run `/milestone-summary` first; do not skip silently. |
| Summary file absent and `--skip-summary` set | Warn: *"No summary file at design/milestones/<...>.md. Continuing because --skip-summary was set."* Continue. |
| `design/state.md` absent | Skip the state-archive step silently (consistent with `prepare-issue`'s edge case). Other steps proceed. |
| `design/state.md` marker fences broken | Stop before the close. Tell the user which zone is malformed and suggest `/pause` to refresh. Do not partially write. |
| `gh` rate-limited at any step | Surface verbatim. If the milestone close already succeeded but state.md write failed, report the partial state and tell the user what to retry manually. |
| `--release` set but `design/build-out-plan.md` has no matching phase | Surface the warning from `/release` itself; the release skill handles this case. |
| User authors no lessons zone | The `lessons` zone retains the template default. The skill does not enforce; that is editorial discipline. |

## Invariants

- Never close a milestone without the literal `yes` approval.
- Never invoke `/release` without explicit consent (flag or
  interactive `yes`).
- Never write to `design/state.md` if the marker fences are broken;
  refuse and ask the user to refresh first.
- Never modify other state.md zones (`phase`, `blockers`).
- Never push, force-push, or alter `main`. The close is GitHub-side;
  state.md changes are working-tree-only.

## Self-check before the approval gate

- [ ] Milestone resolved by `gh api`; not already closed.
- [ ] Audit ran; gaps surfaced to the user.
- [ ] Summary ran (or was skipped explicitly); summary file exists
  on disk before close.
- [ ] State.md changes preview is shown — exact zones, exact new
  content.
- [ ] Release chain decision is shown (set, skipped, or
  interactive).
- [ ] Every command in the plan is the literal command that will
  run, with resolved values.

## Self-check after execution

- [ ] `gh api repos/:owner/:repo/milestones/<N> --jq .state` returns
  `closed`.
- [ ] `design/state.md` has updated `in-flight`, `recent`, and
  `continue-here` zones; `phase` and `blockers` are unchanged.
- [ ] If `--release` chained, the release URL is reported.
- [ ] Working tree has the state.md change (uncommitted is fine —
  the user decides whether to commit it now or with the next PR).

## Receipts

Closing a milestone is a mutating (cat-3) action, so the skill records an
idempotency receipt keyed by the **milestone number**, per
[`docs/receipts.md`](../../docs/receipts.md):

- **Before** closing, check for an existing receipt
  (`bin/write-receipt --find --skill complete-milestone --key <milestone>`,
  or read `.claude/receipts/complete-milestone__<milestone>.json`). A
  `completed` receipt means the milestone was already closed — report it
  rather than re-running.
- **On completion**, write a `completed` receipt. Writing it is
  best-effort and never blocks the handoff; receipts are local, gitignored
  state.

## Handoff

`/complete-milestone` is the terminus of the milestone chain. After
it succeeds:

- If the release chained, the project has a new tag and a closed
  milestone. The next concrete action is opening the next
  milestone (or moving on).
- If the release did not chain, the user can run
  `/release --milestone-phase=<N>` later — the milestone is
  already closed and the build-out-plan row update happens at
  release time.

The summary file under `design/milestones/` is now the canonical
retrospective for this milestone. Future runs of `/release` can
quote from it; future planning rounds can reference it.

See [`audit-milestone/example.md`](../audit-milestone/example.md)
for a worked walkthrough of the full chain on a phased example
project. (This skill ships without its own `example.md` by design —
see [`docs/skills.md` §5](../../docs/skills.md#skills-without-a-worked-example-sidecar).)
