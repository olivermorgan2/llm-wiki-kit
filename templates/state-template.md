<!--
  Template: Session-continuity state
  Filled by: prepare-issue, claude-issue-executor, pr-review-packager
            (and refreshed by /pause; read by /resume)
  Output in a target project: design/state.md

  Purpose. A small, committed pointer to the project's current
  position so a fresh Claude Code session does not have to
  reconstruct context from `gh` and prompt files. ADR-035 caps the
  file at ~100 lines on purpose: it is a pointer, not a mirror.

  Marker fences. Each zone is wrapped by HTML-comment fences
  (<!-- state:<zone>:start --> / <!-- state:<zone>:end -->). Skills
  rewrite only the bytes between the fences for their zone, so
  editorial commentary outside the fences (and zones owned by other
  skills) is preserved across re-runs. The same pattern is used by
  templates/decisions-template.md.

  Conflict-resolution rule. On merge conflict in design/state.md,
  the most recently merged PR's version wins for the conflicting
  zone. /resume re-derives from `gh` if the file looks suspect —
  e.g. the in-flight issue is already closed/merged, or the
  last-updated date is older than HEAD on `main` by more than the
  recent-work window.

  Phase awareness. Projects that adopted ADR-032 implementation
  phases write a phase name into the `phase` zone (e.g. `Phase 2 —
  Ingest`). Projects without phases write the literal `single`.
-->

# {{PROJECT_NAME}} — State

**Last updated:** {{YYYY-MM-DD}}
**Source of truth:** `gh` is canonical for issue/PR status; this file is a pointer.

<!-- state:phase:start -->

## Phase

{{Phase name from design/build-out-plan.md, or the literal `single` for projects without phases.}}

<!-- state:phase:end -->

<!-- state:in-flight:start -->

## In-flight issue

- **Issue:** {{#NNN, or `none` if no issue is currently in flight}}
- **Prompt:** {{`prompts/issue-NNN-short-title.md`, or `n/a`}}
- **Branch:** {{branch name, or `n/a`}}
- **Status:** {{`prepared` / `executing` / `pr-open` / `none`}}

<!-- state:in-flight:end -->

<!-- state:recent:start -->

## Recent work

Rolling list of the last five issues completed (oldest drops off as
new entries land). One line each: PR number, ADR if any, one-line
summary.

- {{#PR — ADR-NNN — one-line summary}}
- {{#PR — ADR-NNN — one-line summary}}
- {{#PR — ADR-NNN — one-line summary}}
- {{#PR — ADR-NNN — one-line summary}}
- {{#PR — ADR-NNN — one-line summary}}

<!-- state:recent:end -->

<!-- state:blockers:start -->

## Blockers

{{One line per blocker, or `none`. Examples: "waiting on review of #53",
"upstream gh release n/a until 2026-05-15".}}

<!-- state:blockers:end -->

<!-- state:continue-here:start -->

## Continue here

{{One short paragraph naming the next concrete action. Typically a
prompt path, a skill invocation, or a `gh issue view` pointer. /resume
reads this verbatim into its brief.}}

<!-- state:continue-here:end -->

<!-- state:next-action:start -->

## Next action

{{Structured, machine-readable complement to `continue-here` (ADR-048).
A fenced YAML block naming the next skill, its args, preconditions, and
any blocker. `skill: none` when nothing is queued. /resume and /start
read this to propose the next action without parsing prose; an unmet
precondition or non-`none` `blocked-by` is reported as a blocker
instead. Use exact skill names from kit.json, not verbs.}}

```yaml
skill: none
args: n/a
preconditions: []
blocked-by: none
```

<!-- state:next-action:end -->
