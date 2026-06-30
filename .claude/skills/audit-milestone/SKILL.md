---
name: audit-milestone
description: Verify a GitHub milestone is complete — all issues closed, all referenced ADRs linked to merged PRs, all phase exit criteria met. Returns pass/fail with a gap list. Use when checking milestone readiness before closing; for the close itself use /complete-milestone; for a post-close retrospective use /milestone-summary.
permission-category: 1  # substitutable — read-only gh queries; advisory pass/fail report, per workflow-guide §7
inputs:
  - name: "milestone"
    required: true
    description: "Milestone number or title to audit"
outputs:
  - artefact: "(report)"
    description: "Pass/fail audit with a gap list; advisory, no files written"
next:
  - skill: milestone-summary
    when: "audit passes"
  - skill: complete-milestone
    when: "audit passes and summary is reviewed"
  - skill: release
    when: "the milestone delivers a release"
---

# audit-milestone

Read a GitHub milestone, verify it is finishable, and return a
pass/fail report with a concrete gap list. Three checks:

1. **Issues:** every issue assigned to the milestone is closed.
2. **ADRs:** every ADR token referenced in milestone issues has at
   least one merged PR that mentions the same token.
3. **Phases (when applicable):** every `## Phase` block in
   `design/build-out-plan.md` whose milestone matches has its exit
   criterion satisfied (PRs merged, tag cut, status row updated).

This skill implements the audit half of
ADR-037. It is
**advisory** — the report goes to the user; `/complete-milestone`
chains it but does not gate on the result.

## When to use this skill

- Before invoking `/complete-milestone` — confirm there is no
  loose end.
- Mid-phase, to see how close a milestone is to finishable.
- During a release ceremony, to spot-check a milestone before
  `/release` cuts the tag.

If you only want a list of merged PRs since the last tag, that is
`/changelog --since-last-release`. If you want to *write* a
retrospective summary, that is `/milestone-summary`. This skill
**only reads**.

## What this skill does not do

- Does not modify the milestone, the issues, the ADRs, or the
  working tree. Read-only against GitHub and the local repo.
- Does not block `/complete-milestone` — audit warns, the user
  decides.
- Does not invent phase exit criteria. If `design/build-out-plan.md`
  is absent, the phase check degrades to skipped (kit-only mode).
- Does not infer a milestone from "current work." The argument is
  required.

## Inputs

- **Required:** the milestone identifier — either a number
  (`/audit-milestone 3`) or a title
  (`/audit-milestone v-next`). Both forms resolve to the same
  GitHub milestone object via
  `gh api repos/:owner/:repo/milestones?state=all`.
- **Implicit:** `gh` authenticated against the repo whose `origin`
  remote points at the milestone's repository.

## Output

A single pass-or-fail report:

- **On pass:** `✅ audit-milestone PASS — <title> (#<number>)`
  followed by a one-line summary of the counts (issues closed, ADRs
  linked, phases complete).
- **On fail:** `❌ audit-milestone FAIL — <title> (#<number>)`
  followed by one bullet per gap, grouped by check
  (issues / ADRs / phases). Each bullet includes a concrete fix
  pointer (e.g. *"Issue #47 still open — close it or move it to a
  later milestone."*).
- **On invocation error:** standard `gh` error verbatim and stop.

The report is plain markdown. No code fences. No banners.

## Execution protocol

Run these steps in order. Stop on the first hard error unless
otherwise noted.

1. **Validate input.** Argument must be a positive integer (number)
   or a non-empty string (title). If neither, stop with usage.
2. **Resolve the milestone.**
   - Run `gh api repos/:owner/:repo/milestones?state=all --paginate`
     and pick the entry whose `number` or `title` matches the
     argument.
   - If no match, stop with `"Milestone <arg> not found in <repo>.
     Run `gh api repos/:owner/:repo/milestones?state=all --jq '.[].title'`
     to list available milestones."`
   - Capture: `number`, `title`, `state`, `description`, `closed_at`.
3. **List issues in the milestone.** Run
   `gh issue list --milestone "<title>" --state all --limit 200 --json number,state,title,labels,body`.
   Partition into open vs. closed. Open issues are gaps.
4. **Extract ADR tokens.** From every issue's title and body, grep
   `ADR-[0-9]+` (case-insensitive, deduplicated). For each unique
   token:
   - Confirm `design/adr/adr-NNN-*.md` exists. If the file is
     missing, record a *referenced-but-missing* gap.
   - Confirm at least one merged PR mentions the token in title or
     body: `gh pr list --search "ADR-NNN" --state merged --limit 5 --json number,title`.
     If no merged PR is found, record an *ADR-without-merged-PR*
     gap.
5. **Check phase status (optional).** If
   `design/build-out-plan.md` exists:
   - Find any `## Phase N: <name>` block whose body or
     `**Milestone:**` field references the milestone (by title or
     number).
   - For each matched phase, read its `**Exit criteria:**` line
     and check the `**Status:**` row. A status of
     `released v<X.Y.Z>` (per `/release --milestone-phase`) is the
     only state that passes; `planned` or `in-progress` is a gap.
   If `design/build-out-plan.md` is absent, skip this check and
   note it in the report (kit-only mode).
6. **Triage results.**
   - 0 gaps across all three checks → pass.
   - ≥1 gap in any check → fail.
7. **Render the report.** Print the pass/fail header, the count
   summary, and (on fail) the grouped gap bullets. Stop.

## Gap categories and fix hints

Use these fixed phrases so the report is parseable:

| Category | Fix hint |
|---|---|
| Open issue in milestone | *"Issue #NN still open — close it, merge its PR, or move it to a later milestone."* |
| ADR referenced but missing | *"ADR-NNN referenced in issue #NN but `design/adr/adr-NNN-*.md` not found. Add the ADR or correct the reference."* |
| ADR without merged PR | *"ADR-NNN referenced but no merged PR mentions it. Confirm the implementation PR was merged with `(ADR-NNN, #NN)` in its title or body."* |
| Phase still in-progress | *"Phase N (<name>) status is `<status>` in `design/build-out-plan.md`. Run `/release --milestone-phase=N` to mark it `released vX.Y.Z`, or update the row manually."* |
| Phase missing exit criterion | *"Phase N (<name>) has no `**Exit criteria:**` line in `design/build-out-plan.md`. Add one before auditing."* |

## Edge cases

- **Milestone has no issues.** Treat as a gap: *"Milestone has no
  issues — was this intentional? Audit cannot pass an empty
  milestone."* Do not fail silently.
- **Issue body is empty.** Skip ADR extraction for that issue;
  do not error.
- **`gh` rate-limited.** Surface the rate-limit notice verbatim
  and stop. Re-run once the limit resets.
- **`design/adr/` directory absent.** Skip the ADR-file existence
  check (record a one-line note); still run the merged-PR check
  against `gh pr list --search "ADR-NNN"`.
- **`design/build-out-plan.md` absent.** Skip the phase check;
  record `phases: skipped (no build-out-plan.md)` in the count
  summary.
- **Milestone already closed.** The skill still runs — useful for
  retrospective audits. The pass/fail logic is unchanged.
- **Argument matches multiple milestones (rare title collision).**
  Stop with the list of candidates and ask the user to disambiguate
  by number.
- **Token `ADR-NNN` appears in a closed-but-unmerged PR.** Treat as
  a gap — only merged PRs satisfy the ADR-linked check.

## Invariants

- Read-only. Never run `gh` mutators (`gh api -X PATCH`,
  `gh issue close`, `gh milestone edit`, etc.).
- Never modify the working tree.
- Never block downstream skills. The report is the output; the
  caller decides what to do.
- Never paginate past 200 issues in a single milestone — kit
  milestones are small. If a milestone exceeds 200 issues, stop
  with a notice; the user is past the kit's design envelope.

## Self-check before returning a result

- [ ] The milestone was resolved by number or title via `gh api`,
  not assumed.
- [ ] Every open issue in the milestone is reported (or zero, if
  none).
- [ ] Every unique ADR token from issue bodies was checked for both
  file presence and merged-PR linkage.
- [ ] If `design/build-out-plan.md` was present, the phase check
  ran; if absent, the report says so explicitly.
- [ ] The pass/fail decision matches the gap count: 0 → pass,
  ≥1 → fail.
- [ ] No `gh` mutator was invoked.

## Handoff

`/audit-milestone` is a leaf — its output is the user's decision
input. Two natural next steps:

- **Pass:** run `/milestone-summary <N>` to draft the retrospective,
  then `/complete-milestone <N>` to close out.
- **Fail:** address the gaps (close the open issue, merge the
  ADR-linked PR, run `/release --milestone-phase=N`), then re-run
  this skill. Or, if the user explicitly accepts the gaps, proceed
  to `/complete-milestone <N>` — that skill chains this audit but
  does not gate on its result.

See [`example.md`](example.md) for a worked walkthrough covering
both pass and fail cases on a phased example project.
