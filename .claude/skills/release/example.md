# release — worked example

A typical second-release run for the *Pace Drift* project. The last
release was `v0.2.0`; since then one new ADR (ADR-007: rolling-window
smoothing) was accepted, two features were merged, and one bug was
fixed. No breaking changes. The user wants to publish `v0.3.0`.

---

## 1. Invocation

```
/release
```

No flags — the user wants the skill to suggest the version.

## 2. Prerequisites check

The skill runs, silently in the background:

- `gh auth status` — OK.
- `git status --porcelain` — empty, clean tree.
- `git rev-parse --abbrev-ref HEAD` — `main`.
- `git fetch origin main && git rev-list --left-right --count HEAD...origin/main` — `0 0`.
- `.claude/skills/changelog/SKILL.md` — exists.

All pass.

## 3. Last-tag detection

```
git describe --tags --abbrev=0
```

Returns `v0.2.0`.

## 4. Suggested-bump rationale

The skill scans commits and ADR diffs in `v0.2.0..HEAD`:

- Commits: 2 `feat(` + 1 `fix(` + 3 `chore(`.
- ADRs: `design/adr/adr-007-rolling-window-smoothing.md` is new,
  status `accepted`.
- No `BREAKING CHANGE` markers, no `breaking` PR label, no ADR
  status change from `accepted` to `superseded`.

Decision tree:

- Major? No breaking signals.
- Minor? Yes — new accepted ADR **and** new `feat(` commits.
- Patch? N/A, minor wins.

Suggestion: **minor** → `v0.2.0` → `v0.3.0`.

## 5. `/changelog` invocation

```
/changelog --since-last-release --output=- --github-release=v0.3.0
```

Returns (abbreviated):

```markdown
## v0.3.0

### Added
- Rolling-window smoothing for pace series (ADR-007, #42)
- CSV export from the drift chart (#45)

### Fixed
- GPX parser tolerates missing elevation nodes (#47)

### Other
- Bump dev dependencies (#46, #48)
```

## 6. Plan presented to the user

```
Release plan

  Last tag:        v0.2.0
  Next version:    v0.3.0  (minor bump — new ADR + feat commits)
  Branch:          main (in sync with origin)
  Draft:           no
  Prerelease:      no

Release notes preview:

  ## v0.3.0
  ### Added
  - Rolling-window smoothing for pace series (ADR-007, #42)
  - CSV export from the drift chart (#45)
  ### Fixed
  - GPX parser tolerates missing elevation nodes (#47)
  ### Other
  - Bump dev dependencies (#46, #48)

Annotated tag message:

  Release 0.3.0

  Rolling-window smoothing for pace series (ADR-007, #42)

Commands to run on approval:

  /changelog --since-last-release --output=<tmp> --github-release=v0.3.0
  git tag -a v0.3.0 -m "Release 0.3.0\n\nRolling-window smoothing for pace series (ADR-007, #42)"
  git push origin v0.3.0
  gh release create v0.3.0 --title "v0.3.0" --notes-file <tmp>

Type `yes` to create and push the tag and publish the release.
Any other response cancels.
```

## 7. Approval

User types `yes`.

## 8. Execution

Commands run in order. `gh release create` returns:

```
https://github.com/pacedrift/pacedrift/releases/tag/v0.3.0
```

## 9. Report

```
Released v0.3.0.

  Tag:     v0.3.0 (pushed to origin)
  Release: https://github.com/pacedrift/pacedrift/releases/tag/v0.3.0
  Branch:  main (unchanged)
```

---

## Dry-run variant

If the user had run `/release --dry-run` instead, the output stops at
step 6 and the approval prompt is replaced with:

```
Dry-run. Would execute:

  git tag -a v0.3.0 -m "Release 0.3.0\n\nRolling-window smoothing for pace series (ADR-007, #42)"
  git push origin v0.3.0
  gh release create v0.3.0 --title "v0.3.0" --notes-file <tmp>

No changes made.
```

---

## Edge-case variant — tag already exists

If the user had passed `--version=0.2.0` (the existing tag), the skill
stops at step 4:

```
Tag v0.2.0 already exists. Releases do not modify existing tags.
Suggested next version: v0.2.1 (patch). Re-run with --version=0.2.1
or --bump=patch if that's what you want.
```

No further steps run.

---

## Workflow-shape variant — research-shaped project, first release

A first release run for `lit-review-2026`, a research project tracking
a literature review. The repo is markdown-only — `design/prd.md`,
`design/build-out-plan.md`, `notes/`, `sources/` — with no package
manifest. No prior tag exists. The user wants to publish the first
snapshot, `v0.1.0`.

This walkthrough exercises the project-shape detection added in
ADR-042:
four signals fire, the threshold is met, and the workflow-shape
clarifier is emitted.

### 1. Invocation

```
/release
```

No flags. The user wants the skill to detect the shape and suggest
the version automatically.

### 2. Prerequisites check

All checks pass (clean tree, on `main`, in sync with origin,
`/changelog` available, ADR index in sync). Identical to the
product-shape walkthrough's preflight.

### 3. Project-shape detection

The skill scans the project:

- **PRD language signal — fires.** `design/prd.md` problem statement
  reads *"I'm not shipping a product. This is a folder of markdown
  for tracking literature review work."* The substring
  *"not shipping a product"* matches case-insensitive. ✓
- **Build-strategy signal — fires.** `design/build-out-plan.md` Build
  strategy section reads *"There is no compile / build / deploy
  step. Work is markdown."* ✓
- **Success-criteria-shape signal — fires.** PRD Success criteria
  bullets read *"A researcher can find any source by topic."* and
  *"A researcher can see thematic synthesis at any point."* The
  pattern `^a researcher can ` matches both bullets, and the
  section contains zero `test`/`pass`/`coverage`/`build` mentions. ✓
- **Package-manifest signal — fires.** Repo root: `README.md`,
  `design/`, `notes/`, `sources/`. No `package.json`,
  `pyproject.toml`, `go.mod`, `Cargo.toml`, `Gemfile`,
  `requirements.txt`, `setup.py`, `Pipfile`, `mix.exs`, `pom.xml`,
  or `build.gradle`. ✓

**Score: 4/4. Threshold (≥2) met. `shape = workflow`.**

### 4. Last-tag detection

```
git describe --tags --abbrev=0
```

Exits non-zero ("No names found") → first release. Default
suggestion: `v0.1.0`. Range for `/changelog` starts at the repo's
initial commit.

### 5. `/changelog` invocation

```
/changelog --since-last-release --output=- --github-release=v0.1.0
```

Returns (abbreviated):

```markdown
## v0.1.0

### Added
- Initial PRD and build-out plan (#1)
- First-pass source extraction from 14 papers (#3)
- Thematic codebook v1 (#5)

### Other
- Bootstrap repo scaffold (#2)
```

### 6. Plan presented to the user

```
Release plan

  Last tag:        (none — first release)
  Next version:    v0.1.0  (first-release default)
  Project shape:   workflow  (auto-detected — 4/4 signals)
  Branch:          main (in sync with origin)
  Draft:           no
  Prerelease:      no

Release-notes preview (with workflow-shape clarifier):

  > This is a workflow tag for documentation drift-tracking; the
  > project is not a software product (see PRD for project shape).
  > The version number is for snapshot ordering, not semantic
  > versioning of an API.

  ## v0.1.0
  ### Added
  - Initial PRD and build-out plan (#1)
  - First-pass source extraction from 14 papers (#3)
  - Thematic codebook v1 (#5)
  ### Other
  - Bootstrap repo scaffold (#2)

Annotated tag message:

  Release 0.1.0

  > This is a workflow tag for documentation drift-tracking; the

Commands to run on approval:

  /changelog --since-last-release --output=<tmp> --github-release=v0.1.0
  # prepend workflow clarifier to <tmp>
  git tag -a v0.1.0 -m "Release 0.1.0\n\n> This is a workflow tag …"
  git push origin v0.1.0
  gh release create v0.1.0 --title "v0.1.0" --notes-file <tmp>

Type `yes` to create and push the tag and publish the release.
Any other response cancels.
```

The clarifier banner is rendered in the preview (so the user can see
exactly what the release body will say) and is annotated as the
first non-heading line of the tag message — consistent with the
product-shape walkthrough's first-line-of-notes rule, just applied
to the workflow banner instead of the changelog's first entry.

### 7. Approval + execution + report

User types `yes`. Commands run in order. `gh release create` returns
the URL. Final report:

```
Released v0.1.0.

  Tag:          v0.1.0 (pushed to origin)
  Release:      https://github.com/lit-review-2026/lit-review-2026/releases/tag/v0.1.0
  Project shape: workflow
  Branch:       main (unchanged)
```

---

## Override variants

### `--force-workflow-shape` on a project that did not trip detection

A project mostly markdown but carrying `requirements.txt` for one
analysis script. PRD doesn't say "not a product" explicitly. Only
the build-strategy signal and success-criteria-shape signal fire —
2/4, threshold met. But suppose only one signal fires (say, the
operator reorganised PRD prose so the language signal misses). The
operator wants the workflow framing anyway:

```
/release --force-workflow-shape
```

The plan reports `Project shape: workflow (overridden from product)`,
emits the clarifier banner, and proceeds otherwise identically.

### `--force-product-shape` on a project that tripped detection

A software product whose PRD prose happens to use the word
"workflow" several times (PRD-language signal fires) and whose
docs-only sub-project layout has no `package.json` at the repo
root (package-manifest signal fires) — 2/4, threshold met,
`shape = workflow` auto-detected. But the project genuinely is a
product. The operator overrides:

```
/release --force-product-shape
```

The plan reports `Project shape: product (overridden from workflow)`,
emits no clarifier banner, and proceeds with standard product
framing.

### Mutually-exclusive flags rejected

Passing both:

```
/release --force-product-shape --force-workflow-shape
```

Aborts immediately with:

```
--force-product-shape and --force-workflow-shape are mutually exclusive.
```

No further steps run.
