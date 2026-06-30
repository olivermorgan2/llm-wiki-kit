# prepare-issue — reference

This file contains the short-title derivation rules, the
template-filling slot map, the carry-forward subsection rules, edge
cases, and the pre-write self-check referenced from
[`SKILL.md`](SKILL.md). The 13-step execution protocol and the
review-before-write approval gate stay in SKILL.md.

## Short-title derivation

Derive the short title from the issue title:

1. Lowercase the title.
2. Strip a trailing `(adr-nnn)` or `(#nn)` suffix if present.
3. Strip leading verb boilerplate ("build ", "add ", "create ",
   "write ") only if the remaining title is still descriptive —
   otherwise keep the verb.
4. Replace any run of non-alphanumeric characters with a single `-`.
5. Trim leading and trailing `-`.
6. Truncate to ~50 characters at the nearest `-` boundary.

Examples:

| Issue title                                                    | Short title                       |
|---------------------------------------------------------------|-----------------------------------|
| "Build pr-review-packager skill (ADR-015)"                    | `pr-review-packager-skill`        |
| "/prepare-issue skill (ADR-013)"                              | `prepare-issue-skill`             |
| "Write docs/workflow-guide.md — end-to-end workflow guide"    | `docs-workflow-guide-end-to-end`  |
| "Create repository skeleton and project-local installation"   | `repository-skeleton-project-local` |

The resulting filename is
`prompts/issue-NNN-<short-title>.md` with `NNN` zero-padded to three
digits.

## Template-filling rules

The template uses `{{PLACEHOLDER}}` slots. Map each slot as follows.

| Slot                               | Source                                                                 |
|------------------------------------|------------------------------------------------------------------------|
| `{{PROJECT_NAME}}`                 | Git repo name (from `git remote get-url origin` or `basename $PWD`).   |
| `{{ONE_LINE_PROJECT_DESCRIPTION}}` | First line of the project's `README.md` if available; else TODO.       |
| `{{WORKFLOW_DOC_PATH}}`            | `docs/workflow-guide.md` if it exists; else TODO.                      |
| `{{ADR_FILE}}`                     | Filename of the first resolved ADR (e.g. `adr-013-prepare-issue-skill.md`). |
| `{{ADR_ONE_LINE_SUMMARY}}`         | First sentence of the `## Decision` section of that ADR.               |
| `{{ISSUE_TITLE}}`                  | Issue title verbatim.                                                  |
| `{{ISSUE_NUMBER}}`                 | The input number (not zero-padded here — the template already has `#`). |
| `{{MILESTONE}}`                    | Milestone title from `gh`, or "none".                                  |
| `{{LABELS}}`                       | Comma-separated label names.                                           |
| `{{ONE_OR_TWO_SENTENCES}}`         | The "Goal" section of the issue body if present; else first paragraph. |
| `{{ONE_PARAGRAPH}}`                | The "Why it matters" section of the issue body if present; else TODO. |
| `{{REQUIREMENT_N}}`                | Bullets from the issue body's "Requirements" section.                  |
| `{{CRITERION_N}}`                  | Bullets from the issue body's "Acceptance criteria" section.           |
| `{{PRIMARY_FOLDERS}}`              | From the issue body's "Scope" section, if present; else TODO.          |
| `{{AVOID_FOLDERS}}`                | From the issue body's "Scope" section, if present; else TODO.          |
| `{{PROJECT_SPECIFIC_CONSTRAINT_OR_DELETE_THIS_LINE}}` | Delete line if no project-specific constraint; else fill. |
| `{{EVALUATION_REQUIREMENT_N}}`     | Bullets from the issue body's "Evaluation" section.                    |

If the issue body closely mirrors the template (as most issues
produced by `issue-planner` will), the mapping is near-direct. If the
body is freeform, do best-effort extraction and mark gaps as TODOs
rather than inventing content.

**Multiple ADRs:** if the issue references more than one ADR, repeat
the `- File:` / `- Decision:` pair for each, in the order they first
appear in the issue body.

## Carry-forward subsection

When step 6.5 found one or more matching `## Notes for #<NNN>`
sections in recently-merged PRs, insert a `## Design questions
carried forward from PR #M` subsection into the rendered prompt
**immediately before the `Requirements` section**. One subsection per
source PR (newest-first by `mergedAt`). The subsection embeds the
matched Notes verbatim and adds an explicit instruction to the
executor:

```markdown
## Design questions carried forward from PR #<M>

The following questions were raised by issue #<source-issue>'s eval
summary and preserved in PR #<M>'s body. Address each in the plan
you propose:

<verbatim ## Notes for #<NNN> body from PR #M>
```

When step 6.5 returned zero matches, **omit the subsection
entirely** — most issues will not have any. Do not emit an empty
heading. The schema and field semantics of the underlying
`design-questions` entries live in
[`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040)
— this skill consumes the rendered Notes-for-#N format only.

## Edge cases

- **Invalid or non-numeric argument** → ask for the issue number and
  stop.
- **Issue not found or auth error** → print `gh`'s error verbatim and
  stop. Do not retry.
- **No ADR referenced** → write
  `ADR: none — no ADR referenced in this issue.` per the template
  comment. Do not block on this.
- **ADR referenced but file missing** → include the reference token
  verbatim with ` <!-- TODO: ADR file not found in design/adr/ -->`.
- **`design/build-out-plan.md` missing** → skip silently; the kit
  repo itself does not have this file.
- **`prompts/_template.md` missing** → stop with a clear message
  pointing at Issue #12.
- **File already exists at target path** → show a diff and ask
  before overwriting. Default no.
- **Multi-line label list** → join with `, `.
- **No milestone** → fill with the string `none`.
- **`design/state.md` missing** → skip the state-update step
  silently. Consistent with how the build-out-plan read is handled.
- **`design/state.md` marker fences broken** → do not rewrite. Tell
  the user which zone is malformed and suggest `/pause` to refresh
  the file in place.
- **`/check-plan` gate fails** → the gate runs *after* user-confirm
  but *before* disk write, so a failed check leaves the working
  tree clean. Iterate up to 3 rounds in step 10, then yield without
  writing.
- **`/check-plan` gate yields** (3 rounds exceeded) → stop. Surface
  remaining failures to the user as items to fix manually. The
  in-memory prompt is discarded; nothing is written.
- **`--skip-check`** → bypass the gate entirely for this invocation.
  Append the documented breadcrumb to the prompt body before
  writing, and proceed to step 11. The flag is single-use; future
  `/prepare-issue` invocations re-enable the gate.
- **No PR-scan matches** → the
  `## Design questions carried forward from PR #M` subsection is
  omitted entirely from the rendered prompt. This is the common
  case; do not emit an empty heading. Surface "no carry-forward
  notes found" in the review-before-write output.
- **`gh pr list` errors during scan (auth, rate limit, network)** →
  surface the error verbatim, ask the user whether to proceed
  without the scan or to abort. Log the skip in the
  review-before-write output so the operator knows the prompt was
  rendered without the scan. Do not retry silently.
- **Multiple PRs contain `## Notes for #<NNN>` sections for the same
  issue** → emit one `## Design questions carried forward from PR #M`
  subsection per source PR, in newest-first `mergedAt` order. This
  is rare but not pathological; it can happen when an upstream issue
  was split across two PRs both of which raised carry-forward notes.

## Self-check before writing

- [ ] The issue number was fetched successfully via `gh`.
- [ ] Every `{{PLACEHOLDER}}` in the template is either filled or
  explicitly marked `<!-- TODO: fill in -->`.
- [ ] The ADR section names a file that exists in `design/adr/`, or
  explicitly says "none".
- [ ] The short title is kebab-case, ≤ 50 chars, and starts with an
  alphanumeric.
- [ ] The filename is `prompts/issue-NNN-<short-title>.md` with
  `NNN` zero-padded to three digits.
- [ ] The user explicitly confirmed the write.
- [ ] `/check-plan` passed (deterministic criteria), or
  `--skip-check` was explicitly set with a recorded breadcrumb in
  the prompt body.
- [ ] PR scan ran (or was skipped with the reason logged in the
  review-before-write output). When matches were found, the
  `## Design questions carried forward from PR #M` subsection(s)
  were inserted immediately before `Requirements`, in newest-first
  `mergedAt` order. When no matches were found, no
  carry-forward subsection was emitted.

If any fail, fix and re-show before writing.
