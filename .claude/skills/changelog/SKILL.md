---
name: changelog
description: Parse git log between two refs and emit grouped, readable release notes in markdown to stdout, a file, or a GitHub Release body. Use when generating release notes only; for tagging and publishing use /release instead.
permission-category: 1  # substitutable — renders to stdout / file / Release-body draft; publishing is /release's job, per workflow-guide §7
inputs:
  - name: "--from / --to"
    required: true
    description: "Commit range to render (or --since-last-release)"
  - name: "--output"
    required: false
    description: "Write to FILE instead of stdout"
  - name: "--github-release"
    required: false
    description: "Use the output as the body for GitHub Release TAG"
outputs:
  - artefact: "(stdout)"
    description: "Grouped release notes (or FILE / GitHub Release body)"
next:
  - skill: release
    when: "publishing the release"
---

# changelog

Generate release notes from git history between two refs. The skill
reads `git log`, parses each commit for verb prefix, ADR tokens, and
issue references, groups the commits into sections (Features, Fixes,
Docs, Refactoring, Chores, Other), and emits markdown suitable for a
`CHANGELOG.md` entry or a GitHub Release body.

The skill is read-only with respect to git (it never creates tags,
commits, or branches). The one write it can perform is posting the
rendered markdown to a GitHub Release body via `gh release`. Release
tagging itself is out of scope — that is the `/release` skill.

## When to use this skill

Typical invocations:

```
/changelog --since-last-release
/changelog --from=v1.2.0 --to=HEAD
/changelog --from=v1.2.0 --to=v1.3.0 --output=CHANGELOG.md
/changelog --since-last-release --github-release=v1.3.0
```

Use this skill when:

- You are cutting a release and want draft notes to paste into
  `CHANGELOG.md` or a GitHub Release.
- You want a quick summary of what has landed on `main` since the last
  tag, before deciding whether to ship.
- You want to review what two refs differ by, grouped by area.

Do not use this skill to tag or publish a release — that belongs in
`/release` (future issue #19, which will call this skill internally).

## What this skill does not do

- Does not create, move, or delete git tags.
- Does not create GitHub Releases from scratch with a version bump —
  it can only write the body of a release that already exists or be
  told to create one at a tag that already exists. Tagging and version
  bumping live in `/release`.
- Does not push to any remote.
- Does not edit `CHANGELOG.md` in place beyond what `--output=FILE`
  writes. Merging a new section into an existing changelog is the
  user's call.
- Does not require a specific commit-message style. Commits that do
  not match the repo's `<verb>(<scope>): <what> (ADR-NNN, #issue)`
  convention land in an "Other" section rather than being dropped.

## Inputs

Exactly one of the following ref-selection modes is required:

- `--from=<ref> --to=<ref>` — explicit range. Both refs must resolve
  via `git rev-parse`. Either can be a tag, branch, or SHA.
- `--since-last-release` — auto-detect the most recent tag with
  `git describe --tags --abbrev=0`, use it as `<from>`, and use `HEAD`
  as `<to>`. If no tags exist, fall back to the repo's first commit
  (`git rev-list --max-parents=0 HEAD | tail -1`) as `<from>`.

Optional flags:

- `--output=FILE` — write the rendered markdown to `FILE` instead of
  stdout. Overwrites without prompting; if the user wants to preserve
  an existing file, they pass a different path.
- `--github-release=TAG` — post the rendered markdown as the body of
  the GitHub Release at `TAG` via
  `gh release edit TAG --notes-file -` (stdin). If the release does
  not yet exist, prompt once and offer `gh release create TAG --notes-file -`.
- `--title=STRING` — override the default H1 title
  (`# Changelog <from>..<to>`). Useful when posting to a GitHub
  Release, where the tag is already implied.
- `--include-merges` — include `Merge pull request` and `Merge #N:`
  commits as top-level entries. Default: excluded, because in a
  squash-merge workflow the merge commit duplicates the squashed
  commit it referenced.

If both `--from/--to` and `--since-last-release` are passed, stop and
ask the user which they meant. Do not guess.

## Output

- **Default:** rendered markdown written to stdout.
- **With `--output=FILE`:** the same markdown written to `FILE`, plus
  a one-line summary to stdout (`Wrote N entries to FILE`).
- **With `--github-release=TAG`:** the markdown piped into
  `gh release edit TAG --notes-file -` (or `gh release create` after
  confirmation). On success, print the release URL returned by `gh`.

### Markdown shape

```markdown
# Changelog <from>..<to>

## Features

- <subject> ([abbrev-sha](commit-url), [#N](issue-url), [ADR-NNN](adr-link))
- ...

## Fixes

- ...

## Docs

- ...

## Refactoring

- ...

## Chores

- ...

## Security

- ...

## Other

- ...
```

Section order is fixed as above. A section with no entries is
omitted entirely. If every section is empty, the output is:

```markdown
# Changelog <from>..<to>

_No changes in this range._
```

Within a section, entries are ordered by commit date, newest first.

**Duplicate detection.** `bin/changelog-collect` groups commits that
describe the same underlying change (exact stripped-subject match, the
same trailing `(#N)` PR reference, or a shared `<verb>(<scope>):` prefix
with ≥75% token-set overlap). Each group arrives as one entry in
`outputs.sections[].entries[]`, with the newest commit's subject as
`canonicalSubject` and every contributing short SHA in `shas`. Render
each entry once; do not re-derive the grouping.

## Data source: bin/changelog-collect

Do not parse `git log` in the prompt. Call the deterministic collector
and consume its JSON envelope:

```
bin/changelog-collect --from <ref> --to <ref> --format json
bin/changelog-collect --since-last-release --format json
bin/changelog-collect --from <ref> --to <ref> --include-merges --format json
```

`bin/changelog-collect` parses the `git log` stream, extracts issue and
ADR tokens, categorizes each commit (label-primary via the PR's GitHub
labels, then a conventional-commit verb fallback, then Other), dedupes
within each section, and resolves the origin web URL. Read the standard
envelope's `outputs`:

- `outputs.range` — `{from, to, title, empty}`. When `empty` is true,
  render the "No changes in this range" stub.
- `outputs.origin` — the origin web URL, or null (render bare SHAs).
- `outputs.ghAvailable` — false when `gh` could not supply labels;
  surface that categorization was coarser.
- `outputs.sections[]` — ordered `{id, title, entries[]}`, only the
  non-empty sections, in the fixed order Features, Fixes, Docs,
  Refactoring, Chores, Security, Other.
- each entry — `{canonicalSubject, shas[], fullShas[], issues[],
  adrs[], prRef}`.
- `outputs.warnings[]` — surface verbatim on stderr.

The label-to-section and verb-to-section maps, the squash-merge
patterns (conventional-commit, two-suffix `(#issue) (#PR)`, manual
merge), and robustness against multi-paragraph or non-ASCII bodies all
live in `bin/changelog-collect`. Label-primary exists because GitHub
squash subjects are PR titles that often lack a verb prefix; the
collector reads the PR's labels to recover the right section. The skill
turns `outputs` into prose and decides phrasing and destination — it
does not re-derive the parsing, counting, or grouping.

## Rendering rules

Per entry:

```
- <subject-without-verb-prefix> ([<short-sha>](<commit-url>)[, [<short-sha-2>](<commit-url>)...][, [#N](<issue-url>)][, ADR-NNN])
```

- `<commit-url>`: `<origin-web-url>/commit/<full-sha>` where
  `<origin-web-url>` is derived from
  `git remote get-url origin` (convert `git@github.com:owner/repo.git`
  to `https://github.com/owner/repo`). If no origin is configured,
  omit the link and leave the bare short SHA.
- `<issue-url>`: `<origin-web-url>/issues/<N>`. Multiple issues are
  rendered as `[#1](...), [#2](...)`.
- ADR tokens are rendered as plain `ADR-013`, `ADR-016` — not linked,
  because ADR files live in the repo and their paths are already
  canonical.
- **Grouped commits.** When dedup grouped multiple commits into one
  entry (per "Duplicate detection" above), list each contributing
  short SHA as a linked reference, in commit-date order (newest
  first). The canonical text comes from the newest commit's
  stripped subject.

If `<subject-without-verb-prefix>` ends with `(#N)` and that `#N` is
already rendered as a linked issue, strip the trailing `(#N)` from the
subject so the reference is not duplicated.

### Title line

Default `# Changelog <from>..<to>`. When `<from>` is a tag and
`<to>` is `HEAD`, render as `# Changelog since <from>`. When both are
tags, render as `# Changelog <from> → <to>`. The `--title` flag
overrides all of this verbatim.

## Execution protocol

1. **Collect.** Call `bin/changelog-collect` with the requested ref
   mode and `--format json`, passing `--include-merges` and `--origin`
   through when given. It enforces that exactly one ref mode is set and
   exits 2 on bad invocation or an unresolvable ref; surface that error
   verbatim and stop.
2. **Read `outputs`.** Use `range`, `origin`, `ghAvailable`,
   `sections[]`, and `warnings[]` as described in "Data source". Print
   any `warnings` to stderr.
3. **Render markdown.** Emit the title (see "Title line"), then each
   section in `outputs.sections` in order. Render each entry per
   "Rendering rules", building commit links from `origin` + `fullShas`
   and issue links from `origin` + `issues`. When `range.empty` is
   true, emit the "No changes in this range" stub under the title.
4. **Write to target.** stdout (default), file (`--output`), or
   GitHub Release (`--github-release`). For `--github-release`:
   - Run `gh release view TAG` to check existence.
   - If exists: `gh release edit TAG --notes-file -` with the
     rendered markdown on stdin.
   - If not: show the rendered markdown and ask:
     *"Release TAG does not exist. Create it now via
     `gh release create TAG --notes-file -`? (yes/no)"*. On yes,
     create; on no, stop without writing.
5. **Report.** One-line summary to stdout: the number of entries per
   section and the destination.

## Edge cases

`bin/changelog-collect` owns the parsing edge cases and reports them in
`outputs`: no-tags fallback to the first commit, an empty range
(`range.empty`), unconventional/PR-less commits landing in Other, PRs
with no or multiple labels, `gh` unavailable (`ghAvailable: false` plus
a warning), non-existent PR lookups, merge commits under
`--include-merges`, and a non-git directory (exit 2). Read those from
`outputs`/`warnings` rather than re-deriving them. The skill still owns
the publishing edge cases:

- **`gh` not installed** and `--github-release` was passed → stop with
  a clear error telling the user to install `gh` or remove the flag.
- **`--output=FILE`** whose parent directory does not exist → stop and
  ask, rather than failing mid-write.

## Self-check before writing output

- [ ] `bin/changelog-collect` exited 0 (it validates that exactly one
  ref mode is set and that the refs resolve).
- [ ] The markdown has a title and at least one section (or the
  "No changes" stub when `range.empty`).
- [ ] Every rendered entry has a resolvable short SHA from `shas`.
- [ ] Every `#N` in rendered entries is a linked reference (or
  `origin` was null, in which case it is plain).
- [ ] No `<from>..<to>` placeholders remain in the title.
- [ ] Every section rendered came straight from `outputs.sections` in
  order — the grouping and categorization were not re-derived.
- [ ] When `--output=FILE` is used, the parent directory of `FILE`
  exists.
- [ ] When `--github-release=TAG` is used, `gh` is installed and
  either the release exists or the user confirmed creation.

If any fail, fix and re-render before writing.

## Handoff

The rendered markdown is the deliverable. Typical next steps for
the user:

- Paste into `CHANGELOG.md` under a new `## [version] — date`
  heading.
- Review and tweak the GitHub Release body after this skill posts
  it.
- Run `/release` (future issue #19), which calls this skill with
  `--since-last-release --github-release=<new-tag>` as part of a
  larger tag-and-publish flow.

See [`example.md`](example.md) for a worked run on this repo's own
recent history.
