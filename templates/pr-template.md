<!--
  Template: Pull Request body
  Filled by: the PR author (optionally with the pr-review-packager skill)
  Output in a target project: .github/pull_request_template.md (shipped
    by Issue #9 in the kit repo; copied straight into each target project
    by the install guide).
  GitHub auto-loads this file as the PR body when the file lives at that
  exact path on the default branch. Keep it short — a PR body is signal,
  not a dumping ground.
-->

## Summary

{{One-paragraph description of what this PR changes and why.}}

## Closes

Closes #{{ISSUE_NUMBER}}
<!-- Multiple? List each on its own line: "Closes #12", "Closes #13". -->

## ADR

Related ADR: `design/adr/adr-{{NNN}}-{{short-title}}.md`
<!-- If no ADR applies, write "none". -->

## Changes

- {{Bullet the substantive changes, not every file touched.}}
- {{Call out anything a reviewer might miss by reading the diff alone.}}

## Test results

```
{{Paste test-runner output: total / passed / failed / skipped.}}
```
<!-- If there are no code changes, write "no code changes — docs only" and delete the code fence. -->

## Manual verification

{{Steps a reviewer should run to convince themselves the change works.
Write "none needed" if the change is fully covered by tests.}}
