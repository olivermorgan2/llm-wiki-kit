# release — reference

This file contains the suggested-version heuristic, project-shape
detection signals + threshold + outcome bodies, edge cases,
invariants, and the pre/post self-checks referenced from
[`SKILL.md`](SKILL.md). The release flow and execution sequence
(both load-bearing orchestration) stay in SKILL.md.

## Suggested-version heuristic

When no version is supplied, the deterministic signal collection runs in
`bin/release-suggest` (target projects: `.claude/bin/release-suggest`),
not inline in this prompt:

```
bin/release-suggest --since-last-release --format json
```

The helper inspects everything since the last tag (or since the first
commit, if no tag exists) and emits a JSON envelope whose `outputs`
carry four fields the skill consumes verbatim:

- `suggestedBump` — `major` | `minor` | `patch`, or `null` for an empty
  range.
- `confidence` — `high` | `medium` | `low` | `none`.
- `signals` — the explicit `{signal, tier, detail}` list behind the
  suggestion.
- `warnings` — caveats (e.g. unconventional commits that could hide a
  feature, or `gh` unavailable so a `breaking` label could not be
  verified).

The signal-to-tier rules the helper applies (highest tier present
wins):

1. **major** if any of:
   - a commit message or body contains a `BREAKING CHANGE` /
     `BREAKING-CHANGE` marker, or a conventional `!` breaking bang
     (`feat!:`);
   - the PR carries a `breaking` label (looked up via `gh`, or
     `--gh-mock DIR` offline);
   - an ADR supersede / status-change marker appears in the range (a
     commit that supersedes an `accepted` ADR, or adds a `Supersedes:`
     field). This tier alone yields `medium` confidence — confirm
     against the ADR diffs before treating it as `major`.
2. **minor** if one or more new ADRs are introduced in the range, or
   `bin/changelog-collect` (which the helper reuses) categorizes one or
   more commits as features.
3. **patch** otherwise (only `fix` / `docs` / `chore` / `refactor`
   commits and no higher signal).

The helper is conservative and **advisory only** — it writes nothing
and proposes; the skill and the user decide. When the suggestion is
`low`/`medium` confidence, carries `warnings`, or the range is empty
(`suggestedBump: null`), look closer before settling, and explain the
chosen bump in the plan. `--version` / `--bump` always override the
suggestion.

For the byte-stable fixture coverage of these signals (patch-only,
feature/minor, breaking/major, ADR-supersede/major, and empty range),
see `bin/release-suggest-fixtures/` and `bin/collect-eval`.

## Project-shape detection — signals, threshold, outcome

### Detection signals

`/release` scans for the following four indicators, in any order. Each
satisfied signal contributes to the threshold count.

1. **PRD language signal.** `design/prd.md` or
   `design/prd-normalized.md` contains any of *"not [shipping|building]
   a product"*, *"workflow"*, *"folder of markdown"*, or equivalent
   language in the project's problem statement or success criteria.
   Match is substring, case-insensitive, scanned over the whole PRD
   body.
2. **Build-strategy signal.** `design/build-out-plan.md` "Build
   strategy" section (or equivalent heading) contains *"There is no
   compile / build / deploy step"* or equivalent. Match is substring,
   case-insensitive.
3. **Success-criteria-shape signal.** The PRD's "Success criteria"
   section (or equivalent) contains user-outcome strings (e.g.
   *"a researcher can …"*, *"a reader can …"*) rather than test-result
   strings (e.g. *"all tests pass"*, *"100% coverage"*). Heuristic: at
   least one bullet matches `^(a|an) [a-z]+ can `, or the section
   contains zero `test` / `pass` / `coverage` / `build` mentions.
4. **Package-manifest signal.** Repo root has *none* of:
   `package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`, `Gemfile`,
   `requirements.txt`, `setup.py`, `Pipfile`, `mix.exs`, `pom.xml`, or
   `build.gradle`. Match is filesystem presence, case-sensitive.

This signal list is the **single source of truth** for project-shape
detection. Cross-references elsewhere (e.g. `docs/workflow-guide.md`
§2.i) point back here rather than maintaining a parallel list. New
signals are added here as the kit's PRD and build-out templates
evolve; changes go through the normal ADR-amendment path if material.

### Threshold

**Two or more satisfied signals** trigger the workflow-shape path.
Single-signal cases stay on the product-shape path; the threshold is
deliberately conservative to avoid false-positives on borderline
projects (e.g. a software project whose PRD happens to use the word
"workflow" in passing).

### Outcome

After scanning, `/release` records one of two values:

- `shape = product` — the default; standard product-release framing
  applies (this is the existing behaviour, unchanged).
- `shape = workflow` — when ≥2 signals fire; the release surface
  classifies the project as non-product, and subsequent release-body
  framing uses this value.

The `shape` value is presented to the user in the release plan, so
the classification is visible before the approval gate.

If `design/prd.md` / `design/prd-normalized.md` / `design/build-out-plan.md`
are missing (e.g. on a fresh project that hasn't run `/idea-to-prd`
or `/prd-to-mvp` yet), those signals score zero — they neither fire
nor block. Only the package-manifest signal can fire on a project
without PRD/build-out artefacts; one signal is below threshold, so
such projects default to product-shape.

## Edge cases

| Situation | Behaviour |
|---|---|
| No prior tag | Treat as first release. Suggest `0.1.0`. Range for `/changelog` starts at the repo's initial commit. |
| Tag `vX.Y.Z` already exists (local or remote) | Refuse. Suggest the next patch (`vX.Y.(Z+1)`). Never overwrite. |
| Dirty working tree | Refuse. Report `git status --porcelain` output. Tell the user to commit or stash. |
| Wrong branch | Refuse. Report current branch vs. expected. Suggest `git switch <branch>`. |
| Branch out of sync with origin | Refuse. Report ahead/behind counts. Suggest `git pull --ff-only` or `git push`. |
| `/changelog` not installed | Abort with: *"/release requires the /changelog skill. Install it first (ADR-016, Issue #18)."* |
| `/changelog` returns empty output | Stop cleanly: *"No changes since `<last-tag>`. Nothing to release."* No tag created. |
| `gh` not authenticated | Abort. Tell the user to run `gh auth login`. |
| User declines approval | Cancel. Report "Release cancelled. No changes made." Exit 0. |
| `gh release create` fails after tag is pushed | Report failure. The tag is already pushed. Tell the user the tag stands and they can retry the release with `gh release create vX.Y.Z --notes-file …` manually. Do **not** delete the tag. |
| Both `--force-product-shape` and `--force-workflow-shape` passed | Abort with *"--force-product-shape and --force-workflow-shape are mutually exclusive."* No mutations. |
| Project-shape detection scores 0 signals | `shape = product`. Standard product-release framing. No clarifier banner. |
| Project-shape detection scores ≥2 signals but operator passed `--force-product-shape` | Use `shape = product`. The plan reports the override as `product (overridden from workflow)` so the user sees auto-detection was bypassed. |
| Project-shape detection scores 0 or 1 signals but operator passed `--force-workflow-shape` | Use `shape = workflow`. The plan reports the override as `workflow (overridden from product)`. The clarifier banner is emitted. |
| `design/prd.md` and `design/build-out-plan.md` both missing | Three of the four signals score zero (PRD-language, build-strategy, success-criteria-shape). Only the package-manifest signal can fire. Threshold not met → `shape = product`. The user can pass `--force-workflow-shape` if the project genuinely is non-product but lacks the kit's PRD/build-out artefacts. |

## Invariants

- Never force-push tags. Never run `git push --force` or `git tag -f`.
- Never modify or delete existing tags.
- Never mutate without the literal `yes` approval.
- Never mutate in `--dry-run`.
- Never touch files outside the working tree except the GitHub Release
  (and the optional `design/build-out-plan.md` update).

## Self-check before presenting the plan

- [ ] Prerequisites passed.
- [ ] **Project-shape detection ran**, and the resulting `shape`
      value (and any override) is reported in the plan.
- [ ] Target version is valid semver and the tag does not already
      exist.
- [ ] `/changelog` produced non-empty notes.
- [ ] **For `shape = workflow`**, the clarifier banner is shown in
      the release-notes preview and will be prepended to the temp
      file consumed by the tag and the GitHub Release.
- [ ] The annotated tag message, GitHub Release config, and command
      sequence are all shown in the plan.
- [ ] The approval prompt is present and explicit.

## Self-check after execution

- [ ] The tag exists locally (`git tag -l vX.Y.Z`) and on origin
      (`git ls-remote --tags origin vX.Y.Z`).
- [ ] `gh release view vX.Y.Z` returns a release with the expected
      title, notes, and draft/prerelease flags.
- [ ] Working tree is still clean.
- [ ] `main` has not moved.
