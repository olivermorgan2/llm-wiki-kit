---
name: release
description: Tag a semver release, generate release notes via /changelog, and publish a GitHub Release. Use when ready to cut, tag, and publish a release; for release notes alone use /changelog instead.
permission-category: 3  # non-substitutable — git tag, git push, gh release create — maximum public visibility, per workflow-guide §7
inputs:
  - name: "--version / --bump"
    required: false
    description: "Explicit X.Y.Z, or major | minor | patch"
  - name: "--draft / --prerelease"
    required: false
    description: "Release visibility flags"
  - name: "--dry-run"
    required: false
    description: "Preview without tagging or publishing"
outputs:
  - artefact: "(git tag + GitHub Release)"
    description: "Annotated tag pushed; Release published via gh"
next: []
---

# release

Orchestrate the release ceremony for a project: pick the next semver,
generate release notes via the `/changelog` skill, create an annotated
git tag, push it, and publish a GitHub Release. Every mutating step is
gated behind a single explicit approval.

This skill comes from ADR-017.

## When to use this skill

- When `main` is at a state worth cutting a release from, and the user
  wants a tagged, published GitHub Release.
- After a feature or fix has been merged and the user wants to ship it.
- As the terminal step in the delivery chain: issue → PR → merge →
  `/release`.

Do not use this skill to draft changelog content in isolation — that is
`/changelog`. `/release` is the orchestration around it.

## What this skill does not do

- Does not generate changelog content itself. It invokes `/changelog`.
  If `/changelog` is not installed, the skill aborts with a pointer to
  install it (see Issue #18).
- Does not modify existing tags. Never force-pushes tags.
- Does not create releases without explicit user approval.
- Does not merge branches, push commits, or alter `main`. It only tags.
- Does not decide semver unilaterally; it proposes and the user
  confirms.

## Invocation

```
/release [--version=X.Y.Z]
         [--bump=major|minor|patch]
         [--branch=main]
         [--draft]
         [--prerelease]
         [--dry-run]
         [--force-product-shape | --force-workflow-shape]
         [--milestone-phase=N]
```

Flags:

- `--version=X.Y.Z` — explicit target version. Overrides any bump
  suggestion. Must be valid semver without the `v` prefix; the tag
  itself is `vX.Y.Z`.
- `--bump=major|minor|patch` — bump level. Applied to the last
  detected tag. If there is no prior tag and `--version` is not given,
  the first release defaults to `0.1.0` (minor bump treated as the
  seed release).
- `--branch=main` — release branch. Defaults to `main`. Configurable
  for projects on a non-default primary branch.
- `--draft` — create the GitHub Release as a draft (passes
  `--draft` to `gh release create`).
- `--prerelease` — mark the GitHub Release as a prerelease
  (passes `--prerelease` to `gh release create`).
- `--dry-run` — walk through every step up to the approval gate and
  stop. Render the preview, print each command that would run, make
  no mutations.
- `--force-product-shape` — force the product-shape release-body
  framing regardless of what project-shape detection (see below)
  would have inferred. Documented for operators whose project trips
  the workflow-shape heuristic but is genuinely a software product.
  Mutually exclusive with `--force-workflow-shape`; passing both is
  an invocation error.
- `--force-workflow-shape` — force the workflow-shape release-body
  framing on a project whose detection signals were below threshold
  but the operator wants the workflow-shape clarifier anyway.
  Symmetric to `--force-product-shape`. Mutually exclusive with it.
- `--milestone-phase=N` — optional. If set, after a successful release
  the skill updates the matching phase row in
  `design/build-out-plan.md` from `in-progress` (or `planned`) to
  `released <tag>`. Kept minimal; if the file or phase row is not
  found, the skill reports and continues.

If neither `--version` nor `--bump` is supplied, the skill computes a
**suggestion** (see below) and presents it for confirmation.

## Default release boundary

The default release unit is **one phase**. When
`design/build-out-plan.md` has multiple `### Phase N: <name>` blocks
and `--milestone-phase` was not passed explicitly, infer the target
phase from:

1. The set of GitHub milestones closed since the previous tag — if
   exactly one phase milestone is newly closed, that phase is the
   release boundary.
2. The set of merged PRs in the range — group by their milestone; if
   they all belong to one phase milestone, use it.
3. Otherwise, prompt the user once for the phase number (or `none`
   to release as a multi-phase bundle).

Confirm the inferred phase explicitly in the plan output. Single-phase
projects (one Phase block, or no Phase headings) skip this inference
entirely and behave as before — one release covers the whole project.

To group multiple phases into one release (e.g. v1.0.0 covers Phases
1, 2, and 3), pass `--milestone-phase=` blank or accept the
"multi-phase bundle" option at the prompt. The build-out-plan rows
for every phase in the bundle are updated to `released <tag>`.

## Suggested-version heuristic

When no version is supplied the skill proposes major / minor / patch.
The deterministic signal collection runs in `bin/release-suggest` (in
target projects: `.claude/bin/release-suggest`), not in this prompt:

```
bin/release-suggest --since-last-release --format json
```

The helper reads the commit range, reuses `bin/changelog-collect` for
the categorized commit tally, and scans for BREAKING-CHANGE markers,
`breaking` PR labels, ADR supersede/status-change markers, and new
ADRs. It emits a JSON envelope whose `outputs` carry `suggestedBump`,
`confidence`, `signals`, and `warnings`. It is **advisory only**: it
writes nothing, tags nothing, and publishes nothing.

This skill consumes that output as the starting point and keeps the
judgment: present the suggested bump, its confidence, and the signals
to the user, then let `--version` / `--bump` or the user's confirmation
override it. Treat a `low`/`medium` confidence, any `warnings`, or an
empty range (`suggestedBump: null`) as a cue to look closer before
settling — never tag on the suggestion alone. See
[`reference.md`](reference.md#suggested-version-heuristic) for the
signal-to-tier rules and how to read the envelope.

## Prerequisites check

Before anything else, verify:

- `gh` is installed and authenticated (`gh auth status`).
- `git` working tree is clean (`git status --porcelain` is empty).
- Current branch equals `--branch` (default `main`).
- Local branch is in sync with `origin/<branch>` — run
  `git fetch origin <branch>` then compare `HEAD` to
  `origin/<branch>`. Refuse if ahead, behind, or diverged.
- `/changelog` skill is available at `.claude/skills/changelog/` (or
  `skills/changelog/` when running in the kit repo). Abort if missing
  with: *"/release requires the /changelog skill. Install it first
  (ADR-016, Issue #18)."*
- **ADR index is in sync.** If `design/adr/` exists, run
  `bin/sync-adr-index --check` (or `.claude/bin/sync-adr-index --check`
  in target projects). Refuse if it reports drift (exit 1) — the
  release should not capture a stale index. The user must run
  `bin/sync-adr-index`, commit the update, and retry.

If any check fails, stop and report the specific failure. Do not
attempt to fix the environment.

## Project-shape detection

Per ADR-042.
After Prerequisites check passes, scan the project for non-product
indicators and classify the release as either *product-shape* (the
default) or *workflow-shape*. The classification gates the framing of
the release-body content.

The kit applies to any structured project — software or otherwise. On
non-product projects (research projects, books, curricula, content
projects, design system docs, internal-policy documents), defaulting
to product-shape framing — *"first tagged release of …"*, semver-shaped
language, software-flavoured copy — actively misleads users. This
detection step is the structural enforcement point that matches the
release surface to the project's actual shape.

The detection signals, threshold rule, and outcome semantics live in
[`reference.md`](reference.md#project-shape-detection--signals-threshold-outcome).
The signal list there is the single source of truth — cross-references
in `docs/workflow-guide.md` §2.i point back to it.

### Overrides

Two operator flags override the auto-detected `shape` value:

- `--force-product-shape` — forces `shape = product` regardless of
  signal count. Use on a project that trips the workflow-shape
  heuristic but is genuinely a software product (e.g. a software
  project whose PRD prose includes the word "workflow" enough times
  to score the PRD-language signal, plus a docs-only sub-project
  layout that scores the package-manifest signal).
- `--force-workflow-shape` — forces `shape = workflow` regardless of
  signal count. Use on a non-product project whose detection signals
  are too sparse to trigger but the operator wants the workflow-
  shape clarifier (e.g. a research-shaped project that does carry a
  `requirements.txt` for analysis tooling, suppressing the package-
  manifest signal).

Override flags take effect immediately after detection runs and
before the release plan is rendered. The release plan reports the
override prominently so the user sees that auto-detection was
overridden.

The two flags are mutually exclusive. Passing both is a usage error
— `/release` aborts with *"--force-product-shape and
--force-workflow-shape are mutually exclusive."*

## Preceding-tag detection

Detect the last tag with:

```
git describe --tags --abbrev=0
```

- If this succeeds, use the result as the `--since` reference for
  `/changelog` and as the base version for `--bump`.
- If it fails (exit non-zero, "No names found"), treat this as the
  **first release**. Default suggestion is `0.1.0`. Pass the repo's
  initial commit as the range start to `/changelog`:
  `git rev-list --max-parents=0 HEAD | tail -1`.

## Release flow

Everything below happens in one conversational turn before any
mutation. The user sees the full plan and types `yes` exactly once to
execute every step.

1. **Run prerequisites check.** Stop on any failure.
2. **Run project-shape detection** (see "Project-shape detection"
   section above). Score the four
   signals; record `shape = product` (default) or `shape = workflow`
   (when ≥2 signals fire). Apply `--force-product-shape` or
   `--force-workflow-shape` if set, recording the override for the
   release plan to surface. Reject mutually-exclusive flag combos
   here with the documented error message.
3. **Detect the last tag.** Record it (or note "first release").
4. **Determine the target version.**
   - If `--version` is set, use it.
   - Else if `--bump` is set, apply it to the last tag.
   - Else run `bin/release-suggest --since-last-release --format json`,
     read `suggestedBump` / `confidence` / `signals` / `warnings` from
     its `outputs`, and present the suggestion (with its signals and
     confidence) for confirmation. The suggestion is advisory — the
     user confirms or overrides it.
5. **Refuse if `vX.Y.Z` already exists** as a local or remote tag
   (`git tag -l vX.Y.Z` or `git ls-remote --tags origin vX.Y.Z`).
   Suggest the next patch and exit; do not prompt for overwrite.
6. **Invoke `/changelog` for release notes:**
   ```
   /changelog --since-last-release --output=- --github-release=vX.Y.Z
   ```
   Capture the markdown from stdout. If `/changelog` reports no
   commits since the last tag, stop cleanly with "No changes since
   `<last-tag>`. Nothing to release." and do not create a tag.
7. **Assemble the release plan and render it for review:**
   - Target version and tag (`vX.Y.Z`).
   - **Project shape:** `product` or `workflow`. When the value came
     from an override flag rather than auto-detection, label it
     `<shape> (overridden from <auto-detected-shape>)` so the user
     sees auto-detection was bypassed.
   - Last tag (or "first release").
   - Suggested-bump rationale, if applicable.
   - **Release-notes preview** (the `/changelog` output verbatim).
     When `shape = workflow`, the preview is preceded by the
     workflow-shape clarifier banner — the banner ships with the
     notes into the temp file consumed by `git tag` and
     `gh release create`. The literal banner text:

     ```
     > This is a workflow tag for documentation drift-tracking; the
     > project is not a software product (see PRD for project shape).
     > The version number is for snapshot ordering, not semantic
     > versioning of an API.
     ```

     For `shape = product`, no banner; the changelog renders verbatim
     as it always has (existing behaviour, unchanged).
   - Annotated tag message:
     ```
     Release X.Y.Z

     <first non-heading line from release notes>
     ```
     For `shape = workflow`, the first non-heading line is the
     opening line of the clarifier banner — that's intentional, it
     labels the tag itself as a workflow-shape release for anyone
     reading `git show vX.Y.Z` later.
   - GitHub Release config: draft? prerelease? title = `vX.Y.Z`.
   - The exact commands that will run, in order.
8. **Approval gate.** Ask: *"Type `yes` to create and push the tag
   and publish the release. Any other response cancels."* Accept only
   the literal string `yes` (case-insensitive, trimmed). Any other
   input — including `y`, `ok`, `sure` — cancels.
9. **On `yes`, execute in order** (see [Execution sequence](#execution-sequence)).
10. **On any other input, cancel** and report "Release cancelled. No
    changes made."

## Execution sequence

After approval (skip in `--dry-run`):

```bash
# 1. Write release notes to a temp file so `gh` and the tag share text.
#    For shape=workflow, prepend the clarifier banner so the banner
#    ships with the notes into both the annotated tag and the
#    GitHub Release body. For shape=product, the file contains only
#    the rendered changelog (existing behaviour).
NOTES=$(mktemp -t release-notes)
/changelog --since-last-release --output="$NOTES" --github-release=vX.Y.Z
if [ "$shape" = "workflow" ]; then
  CLARIFIER=$(mktemp)
  cat > "$CLARIFIER" <<'BANNER'
> This is a workflow tag for documentation drift-tracking; the
> project is not a software product (see PRD for project shape).
> The version number is for snapshot ordering, not semantic
> versioning of an API.

BANNER
  cat "$NOTES" >> "$CLARIFIER"
  mv "$CLARIFIER" "$NOTES"
fi

# 2. Annotated tag. First line of notes becomes the tag summary.
SUMMARY=$(head -n 1 "$NOTES")
git tag -a vX.Y.Z -m "Release X.Y.Z

$SUMMARY"

# 3. Push the tag only (never the branch).
git push origin vX.Y.Z

# 4. Publish the GitHub Release.
gh release create vX.Y.Z \
    --title "vX.Y.Z" \
    --notes-file "$NOTES" \
    [--draft] [--prerelease]

# 5. Upload bootstrap-workflow-kit as a release asset (per ADR-029).
# Only if bin/bootstrap-workflow-kit exists in the repo.
if [ -f bin/bootstrap-workflow-kit ]; then
  gh release upload vX.Y.Z bin/bootstrap-workflow-kit
fi

# 6. Optional: update build-out plan phase status.
# Only if --milestone-phase=N was passed and design/build-out-plan.md exists.
```

Then:

- Report the release URL from `gh release view vX.Y.Z --json url -q .url`.
- Clean up the temp notes file.
- Remind the user nothing else was pushed; `main` is untouched.

## Dry-run mode

With `--dry-run`:

- Run steps 1–7 of the release flow (including project-shape
  detection, so the dry-run preview shows whether the workflow-shape
  clarifier would have been emitted).
- Present the plan as normal.
- Instead of the approval prompt, print:
  *"Dry-run. Would execute:"* followed by the full command sequence
  with resolved values (actual version, resolved tag name, resolved
  notes file path placeholder).
- Make zero mutations. Do not create the temp file, do not invoke
  `git tag`, `git push`, or `gh release create`.

## Reference: edge cases, invariants, self-checks

See [`reference.md`](reference.md) for: the edge-case behaviour table
(no prior tag, dirty tree, /changelog empty, gh failures,
force-shape flag conflicts, PRD/build-out-plan missing), the invariant
list (never force-push, never mutate in dry-run), and the pre-plan
and post-execution self-check checklists.

## Receipts

Cutting a release (git tag + GitHub Release) is a mutating (cat-3)
action, so the skill records an idempotency receipt keyed by the
**release tag**, per [`docs/receipts.md`](../../docs/receipts.md):

- **Before** tagging, check for an existing receipt
  (`bin/write-receipt --find --skill release --key <tag>`, or read
  `.claude/receipts/release__<tag>.json`). A `completed` receipt means the
  tag was already cut — stop rather than creating a duplicate tag/Release.
- **On completion**, write a `completed` receipt with the tag and Release
  URL in `outputs`. Writing it is best-effort and never blocks the
  handoff; receipts are local, gitignored state.

## Handoff

`/release` is the terminus of the delivery chain. After it succeeds,
there is no downstream skill to invoke. The user's next action is
external: share the release URL, notify stakeholders, start the next
issue.

See [`example.md`](example.md) for a worked walkthrough.
