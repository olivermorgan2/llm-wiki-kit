---
name: milestone-summary
description: Generate design/milestones/N-summary.md for a closed or near-closing milestone — what shipped, ADRs adopted, deferred work — from git log, the GitHub milestone, and accepted ADRs in the phase range. Use when writing a retrospective for a closed milestone; for the close itself use /complete-milestone; for readiness checks use /audit-milestone.
permission-category: 1  # substitutable — writes design/milestones/N-summary.md locally; reads gh + git log non-mutating, per workflow-guide §7
inputs:
  - name: "milestone"
    required: true
    description: "Milestone number or title"
  - name: "--dry-run"
    required: false
    description: "Preview without writing the file"
  - name: "--overwrite"
    required: false
    description: "Replace an existing summary (lessons zone preserved)"
outputs:
  - artefact: "design/milestones/N-slug.md"
    description: "Milestone summary; lessons zone preserved on re-run"
next:
  - skill: complete-milestone
    when: "the summary is reviewed"
---

# milestone-summary

Render a retrospective summary for a milestone into
`design/milestones/<N>-<slug>.md` using
[`templates/milestone-summary-template.md`](../../templates/milestone-summary-template.md).
Sources: `git log` between phase-boundary tags, the GitHub milestone
(via `gh`), and accepted ADRs in the milestone's date range.

This skill implements the summary half of
ADR-037. It is
**write-once-by-default** and refuses to overwrite an existing
summary unless `--overwrite` is passed. The `lessons` zone is
preserved verbatim across re-runs (the user authors it).

## When to use this skill

- After `/audit-milestone <N>` reports pass (or after the user
  accepts the gaps), and before `/complete-milestone <N>` closes
  the milestone — so the closing PR can include the summary.
- Mid-milestone, with `--dry-run`, to preview the retrospective
  shape and tune the lessons zone manually.
- After a release ceremony, to backfill a missed summary for a
  closed milestone.

If you only need release notes (a flat list of merged PRs grouped
by commit prefix), that is `/changelog`. This skill is the broader
retrospective: it pulls in ADRs, deferred work, and a lessons
section, and writes a per-milestone artefact under `design/`.

## What this skill does not do

- Does not close the GitHub milestone. That is `/complete-milestone`.
- Does not write release notes for `gh release`. That is
  `/changelog`. The two outputs are complementary; one is for
  shipped users, the other is for the project team.
- Does not invent lessons. The `lessons` zone is left as a TODO
  pointer for the user to author; subsequent runs preserve it.
- Does not modify `design/state.md`. That is owned by `/pause`,
  `prepare-issue`, `claude-issue-executor`, `pr-review-packager`,
  and `/complete-milestone`.

## Inputs

- **Required:** the milestone identifier — number or title
  (resolved via `gh api repos/:owner/:repo/milestones?state=all`,
  same as `/audit-milestone`).
- **Implicit:** `gh` authenticated; `git` history present;
  `templates/milestone-summary-template.md` present.

## Output

`design/milestones/<N>-<slug>.md` — one file per milestone, where
`<N>` is the GitHub milestone number and `<slug>` is the
kebab-cased title (lowercased, non-alphanumeric collapsed to `-`,
truncated to ~30 chars at a `-` boundary).

Examples:
- Milestone `v-next` (#3) → `design/milestones/3-v-next.md`
- Milestone `Phase 2 — Ingest` (#5) → `design/milestones/5-phase-2-ingest.md`

## Invocation

```
/milestone-summary <N>
                   [--dry-run]
                   [--overwrite]
```

Flags:

- `--dry-run` — render the would-be content to stdout; make zero
  filesystem changes. Useful for preview and for tuning the
  lessons placeholder before the first write.
- `--overwrite` — replace an existing summary file. The `lessons`
  zone is **still preserved** (it is the only zone the skill never
  rewrites); other zones are regenerated from current sources.
  Without this flag, an existing summary file aborts the run.

## Execution protocol

1. **Validate input.** Argument is a positive integer or
   non-empty string. If neither, stop with usage.
2. **Confirm template exists.** Read
   `templates/milestone-summary-template.md`. If missing, stop
   with `"Template not found at templates/milestone-summary-template.md.
   Run Issue #46's work first."`
3. **Resolve the milestone.** Same logic as `/audit-milestone`:
   `gh api repos/:owner/:repo/milestones?state=all --paginate`
   then match by number or title. Capture `number`, `title`,
   `description`, `closed_at` (or `null`), `created_at`.
4. **Determine the tag boundary.**
   - If `design/build-out-plan.md` exists and the matching
     `## Phase` block has a `**Exit criteria:**` line that names
     a tag (e.g. `released v0.3.0`), use the previous phase's
     tag as the start and this phase's tag as the end.
   - Else, fall back to `git tag --list --sort=-creatordate` and
     pick the two tags that bracket the milestone's `created_at`
     and `closed_at` dates. If `closed_at` is null, use `HEAD`
     as the end.
   - If no tags exist at all, use the repo's first commit as
     the start and `HEAD` as the end. Note this in the overview
     zone.
5. **Pull merged PRs in the milestone.** Run
   `gh pr list --search "milestone:\"<title>\"" --state merged --limit 200 --json number,title,url,labels,body,mergedAt`.
   Group by conventional-commit prefix in title:
   - `feat(...)` / `feat:` → Features
   - `fix(...)` / `fix:` → Fixes
   - `docs`, `chore`, `refactor`, `test`, `style`, `ci` → Docs / chores
   - Anything else → Docs / chores (catch-all)
   Render each as `#NN — ADR-NNN — <one-line summary>` (omit
   `ADR-NNN —` if no ADR token in the title or body).
6. **List ADRs accepted in the date range.** Glob
   `design/adr/adr-*.md` and filter by the `**Date:**` field in
   each file's frontmatter (or first 10 lines). Include only ADRs
   whose date is between the milestone's `created_at` and
   `closed_at` (inclusive). For each, extract the title and the
   first sentence of `## Decision`.
7. **List deferred work.** Run
   `gh issue list --milestone "<title>" --state open --limit 200 --json number,title,milestone`
   for issues still open in the milestone. Plus
   `gh issue list --milestone "<title>" --state closed --limit 200 --json number,title,closedAt`
   filtered to issues whose `closedAt` is after the milestone's
   `closed_at` and whose milestone has changed (moved out). Both
   sets render as `#NN — <title> — moved / open / descoped`.
8. **Compose the file content.** Read the template, fill each
   marker-fenced zone except `lessons`. The `lessons` zone gets
   the template's literal default (TODO pointer + placeholder
   paragraph). Substitute the header fields:
   `{{MILESTONE_TITLE}}`, `{{MILESTONE_NUMBER}}`,
   `{{YYYY-MM-DD}}` (from `closed_at` or today's date if open).
9. **Resolve the target path.**
   `design/milestones/<N>-<slug>.md`. Compute `<slug>` per the
   "Output" section above.
10. **Handle file existence.**
    - File absent → proceed.
    - File present and `--overwrite` not set → stop with
      `"Summary already exists at design/milestones/<...>.md.
      Pass --overwrite to regenerate (lessons zone is preserved
      regardless)."`
    - File present and `--overwrite` set → read the existing file,
      extract the verbatim bytes between
      `<!-- summary:lessons:start -->` and
      `<!-- summary:lessons:end -->`, and substitute that block
      into the new content before writing. Other zones are
      replaced fresh.
11. **Show the diff in chat** as a fenced markdown block. Ask
    explicitly: *"Write this to design/milestones/<...>.md?
    (yes / edit / cancel)"*. Default no.
12. **On `yes`, write the file.** Create
    `design/milestones/` if it does not exist. Report the
    absolute path and a one-line summary of what was filled
    (PR count, ADR count, deferred count) plus a reminder to
    edit the `lessons` zone before committing.
13. **On `--dry-run`,** render the would-be content to stdout
    and stop. No filesystem changes; no approval prompt.

## Edge cases

- **Tag range is empty** (no merged PRs in range) — the `shipped`
  zone is filled with one line per group: *"none"*. The summary
  still writes, since ADRs and deferred work may still be
  populated.
- **Milestone has no closed_at** (still open) — proceed; use
  today's date in the header. The summary is "as of today";
  re-run with `--overwrite` after close to get the final version.
- **`design/build-out-plan.md` absent** — skip the phase-tag
  detection; rely on date-based tag inference.
- **No tags at all** — use first commit and HEAD. Note in the
  overview that the project has not cut a release yet.
- **Existing summary file with broken `lessons` fences** — stop;
  do not overwrite. Tell the user which fence is malformed and
  ask them to repair manually.
- **Slug collision** (same `<N>-<slug>.md` path computed for two
  different milestones — should be impossible since `<N>` is
  unique, but guard anyway) — stop and report the existing file's
  contents.
- **`gh` rate-limited** — surface verbatim and stop. Partial
  output is not written.

## Invariants

- The `lessons` zone is **never** overwritten. Even with
  `--overwrite`, the existing block is read and re-substituted.
- Read-only against GitHub. No `gh` mutators.
- Read-only against `design/state.md`, `design/adr/`, and
  `design/build-out-plan.md`. Only `design/milestones/<N>-...md`
  is written.
- The marker fences in the template are preserved verbatim in the
  output — `/milestone-summary` can re-run and only the zones it
  owns change.

## Self-check before writing

- [ ] The milestone was resolved by `gh api`, not assumed.
- [ ] The tag range was determined explicitly (phase tags, date
  tags, or first-commit fallback) and named in the output.
- [ ] Every marker-fenced zone in the template is filled (or
  intentionally empty per the edge cases).
- [ ] The `lessons` zone in the output matches either the
  template default (first run) or the existing file's verbatim
  block (re-run with `--overwrite`).
- [ ] The user explicitly confirmed the write (or `--dry-run`
  was set).
- [ ] No `design/state.md` write was attempted.

## Handoff

`/milestone-summary` is a write-once leaf for the summary file.
Natural next steps after a successful write:

- The user edits the `lessons` zone with their own retrospective.
- `/complete-milestone <N>` closes the GitHub milestone and
  archives `design/state.md` — the summary file is the artefact
  that closing PR includes.
- `/release` cuts the next tag if the milestone-to-release
  cadence is one-to-one.

See [`example.md`](../audit-milestone/example.md) for a
walkthrough that exercises both `/audit-milestone` and
`/milestone-summary` against a phased example project — the two
skills are designed to chain. (This skill ships without its own
`example.md` by design — see
[`docs/skills.md` §5](../../docs/skills.md#skills-without-a-worked-example-sidecar).)
