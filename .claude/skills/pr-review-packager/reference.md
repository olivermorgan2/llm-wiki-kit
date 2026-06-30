# pr-review-packager — reference

This file contains the issue-link and ADR-link extraction rules,
the change-summary derivation algorithm, edge cases, the pre-`gh`
self-check, and the relationship-to-other-skills map referenced
from [`SKILL.md`](SKILL.md). The 16-step execution protocol and
the review-before-create approval gate stay in SKILL.md.

## Issue-link extraction — detailed rules

Priority order (first hit wins):

1. **Current branch name.** Regexes, tried in order:
   - `^issue-0*(\d+)-` (e.g. `issue-17-pr-review-packager` → `17`)
   - `-#?(\d+)$` (e.g. `add-auth-17` → `17`)
   - `#(\d+)` anywhere in the branch name.
2. **Commit subjects from `git log <base>..HEAD`.** Walk
   newest-first. For each subject, extract all `#(\d+)` tokens.
   Accept the first one that is not also preceded by `ADR-`
   (guards against false positives from things like
   `ADR-015#section`). Typical hits come from the repo's
   `feat(skills): add X (ADR-NNN, #17)` convention.
3. **Most recent issue prompt file.** Glob `prompts/issue-0*(\d+)-*.md`,
   sort by mtime descending, take the top match, parse the leading
   digits.

Strip leading zeros before writing into the PR body — GitHub's
`Closes #17` does not want `#017`.

## ADR-link extraction — detailed rules

Sources, in the same priority order as issue extraction: branch name,
then commit subjects, then the most recent prompt file.

Parse tokens with the regex `ADR-0*(\d+)` case-insensitive. Preserve
first-seen order, deduplicate.

For each unique `NNN`, glob `design/adr/adr-0*<NNN>-*.md`. If exactly
one file matches, use it. If multiple match (should never happen,
ADR numbers are unique), use the shortest filename. If none match,
keep the `ADR-NNN` token in the body and append
`<!-- TODO: ADR file not found in design/adr/ -->`.

If multiple ADRs are found, emit one `Related ADR:` line each, in the
order they were discovered:

```
Related ADR: design/adr/adr-015-pr-review-packager-skill.md
Related ADR: design/adr/adr-005-template-architecture.md
```

If zero ADRs are found anywhere, emit the single line
`Related ADR: none`.

## Change-summary derivation

Given commit subjects from `git log <base>..HEAD --format="%s"`
(oldest-first):

1. For each subject, identify the conventional-commit verb prefix by
   matching `^(feat|fix|docs|refactor|chore|test|perf|ci|build|infra|style)(\([^)]*\))?:\s*`.
   Everything after the `:` is the bullet text. If no prefix matches,
   the whole subject is the bullet text and the group is `other`.
   `infra` is included alongside the strict conventional-commit verbs
   to match the kit's canonical label set advertised in
   `templates/claude-md-template.md`.

   The mandatory literal `:` after the optional `(scope)` group is the
   verb-boundary anchor: subjects like `infrastructure: foo` do not
   match because the character at the position immediately after the
   verb (`s`, in this case) is neither `(` nor `:`. Likewise for any
   prefix that happens to start with one of the listed verbs but has
   trailing letters before `(`/`:` (e.g. `featureset: foo`,
   `chorelist: foo`).
2. Strip a trailing ` (ADR-NNN, #N)` or ` (#N)` suffix from each
   bullet.
3. Group bullets by prefix. Group order in the output:
   `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `ci`, `build`,
   `infra`, `chore`, `style`, `other`. (`infra` sits adjacent to
   `build` and before `chore` — operational changes that aren't
   features or fixes.)
4. If only one group has any bullets, omit the group heading and emit
   just the bullets.
5. If multiple groups have bullets, emit each as a short subheading
   in bold (`**Features**`, `**Fixes**`, `**Docs**`, `**Infra**`,
   etc.) followed by the bullets.

One bullet per commit. Do not attempt to merge related commits — the
user can edit during the approval gate.

## Edge cases

- **Detached HEAD** → abort with
  `"HEAD is detached. Check out a branch before running /pr-review-packager."`.
- **Current branch is the base branch** → abort with
  `"Current branch is <base>. Check out a feature branch before packaging a PR."`.
- **No upstream for the branch** → abort with
  `"No upstream for <branch>. Run: git push -u origin <branch>"`.
- **No commits ahead of base** → abort with
  `"No commits on <branch> ahead of <base>. Nothing to package."`.
- **`templates/pr-template.md` missing** → abort with
  `"templates/pr-template.md not found. The kit's PR template must exist before this skill can run."`.
- **Uncommitted changes in the working tree** → warn and ask to
  continue; do not abort automatically.
- **No issue number anywhere** → proceed, but fill
  `Closes #<!-- TODO: fill in issue number -->` and flag it at the
  approval gate. Do not guess.
- **ADR token found but no file** → keep the token in the body and
  append `<!-- TODO: ADR file not found -->`. Flag at the approval
  gate.
- **Multiple issue numbers on the branch** → emit one `Closes #N`
  line per unique number, in first-seen order. GitHub accepts
  multiple.
- **`gh pr create` exits non-zero** (e.g. duplicate PR, auth error) →
  print `gh`'s stderr verbatim and stop. Do not retry.
- **An existing PR for the branch** — `gh pr create` will report this.
  Surface the existing PR's URL from `gh pr view --json url` so the
  user can decide whether to update it manually.

## Self-check before calling `gh`

- [ ] Current branch is not the base branch and is not detached.
- [ ] Upstream exists for the current branch.
- [ ] At least one commit is ahead of the base branch.
- [ ] `templates/pr-template.md` was read successfully.
- [ ] Every placeholder in the rendered body is either filled or
      explicitly marked with a `<!-- TODO: ... -->` comment.
- [ ] Every referenced ADR either resolves to a file in
      `design/adr/` or is explicitly marked TODO.
- [ ] **ADR index is in sync.** Run `bin/sync-adr-index --check`. If
      it reports drift (exit 1), the branch added or modified an
      ADR without updating `design/adr/README.md`. Run
      `bin/sync-adr-index`, commit the index update on the same
      branch, and re-show before opening the PR. (ADR-023.)
- [ ] If `notes/eval-issue-NNN.md` exists and contains a
      `### design-questions` block with entries, the corresponding
      `## Notes for #M` sections were emitted in the PR body, one
      per unique `target-issue` (per ADR-040). When the file is
      absent or has no entries, no Notes sections are emitted —
      verify by checking the rendered body before approval.
- [ ] The user explicitly confirmed the body with `yes`.

If any fail, fix and re-show before calling `gh`.

## Relationship to other skills and commands

- **`/review`** (built-in): reviews an existing PR. This skill
  creates one. They complement each other — package first, review
  after.
- **`claude-issue-executor`** (ADR-014, Issue #16): runs the
  implementation session for an issue. When the implementation is
  done and committed, the user (or the executor's handoff step)
  invokes `/pr-review-packager` to open the PR.
- **`/changelog`** (ADR-016, Issue #18): complementary — consumes
  merged PRs to update `CHANGELOG.md`. This skill is upstream of it.
- **`/release`** (ADR-017, Issue #19): downstream — cuts a release
  from merged PRs. This skill is also upstream of it.
