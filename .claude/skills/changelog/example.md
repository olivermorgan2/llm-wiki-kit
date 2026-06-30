# changelog — worked example

A run of `/changelog` against this kit's own recent history. The
input is the range from commit `dc2869d` (the "Accept ADRs 007-021"
commit on `main`) to `HEAD`. This is the same sort of range the
skill would see if a release tag `v0.1.0` had been cut at `dc2869d`
and the user ran `/changelog --since-last-release`.

---

## 1. Invocation

```
/changelog --from=dc2869d --to=HEAD
```

Equivalent if `dc2869d` were tagged `v0.1.0`:

```
/changelog --since-last-release
```

## 2. Raw git log in the range

With default `--no-merges`:

```
bdfb349 feat(bin): add install-workflow-kit script (ADR-009, #13)
cd8fd3b feat(skills): add /prepare-issue skill (ADR-013, #15)
f06f3cc feat(templates): expand CLAUDE.md starter template (ADR-007, #11)
43dcf65 feat(prompts): add dedicated prompts/ folder and template (ADR-008, #12)
c344cfa Update feature-ideas.md statuses to reflect accepted ADRs (#24)
```

Four of the five commits match the conventional `feat(...)` prefix
and land in **Features**. The last one (`Update feature-ideas.md ...`)
has no verb prefix and lands in **Other**. The two `Merge #N:` commits
that also sit in this range are dropped by `--no-merges`.

## 3. Parsing trace

For `cd8fd3b feat(skills): add /prepare-issue skill (ADR-013, #15)`:

- Subject regex match → verb = `feat`, scope = `skills`.
- Stripped subject: `add /prepare-issue skill (ADR-013, #15)`.
- ADR tokens: `ADR-013`.
- Issue tokens: `#15`.
- Trailing `(#15)` recognised as duplicate of the linked issue →
  stripped from the subject during render.
- Final rendered subject: `add /prepare-issue skill`.

For `c344cfa Update feature-ideas.md statuses to reflect accepted ADRs (#24)`:

- Subject regex does not match a known verb → section = Other.
- ADR tokens: none (literal `ADRs` is not `ADR-<digits>`).
- Issue tokens: `#24`.
- Trailing `(#24)` stripped from rendered subject.

## 4. Output — stdout

```markdown
# Changelog dc2869d..HEAD

## Features

- add install-workflow-kit script ([bdfb349](https://github.com/olivermorgan2/claude-workflow-kit/commit/bdfb349...), [#13](https://github.com/olivermorgan2/claude-workflow-kit/issues/13), ADR-009)
- add /prepare-issue skill ([cd8fd3b](https://github.com/olivermorgan2/claude-workflow-kit/commit/cd8fd3b...), [#15](https://github.com/olivermorgan2/claude-workflow-kit/issues/15), ADR-013)
- expand CLAUDE.md starter template ([f06f3cc](https://github.com/olivermorgan2/claude-workflow-kit/commit/f06f3cc...), [#11](https://github.com/olivermorgan2/claude-workflow-kit/issues/11), ADR-007)
- add dedicated prompts/ folder and template ([43dcf65](https://github.com/olivermorgan2/claude-workflow-kit/commit/43dcf65...), [#12](https://github.com/olivermorgan2/claude-workflow-kit/issues/12), ADR-008)

## Other

- Update feature-ideas.md statuses to reflect accepted ADRs ([c344cfa](https://github.com/olivermorgan2/claude-workflow-kit/commit/c344cfa...), [#24](https://github.com/olivermorgan2/claude-workflow-kit/issues/24))
```

(Full SHAs truncated with `...` in this example for readability;
the real output uses the full 40-character SHA in the URL.)

Followed by a one-line summary on stderr:

```
5 entries in 2 sections (Features: 4, Other: 1) — stdout
```

## 5. Same range with `--include-merges`

```
/changelog --from=dc2869d --to=HEAD --include-merges
```

Adds two more entries under **Other** because `Merge #N:` subjects
do not match the verb map:

```markdown
## Other

- Merge #13: install-workflow-kit script (ADR-009) ([dbe4562](...))
- Merge #15: /prepare-issue skill (ADR-013) ([779ec45](...))
- Update feature-ideas.md statuses to reflect accepted ADRs ([c344cfa](...), [#24](...))
```

In a squash-merge workflow this is usually noise — the default
`--no-merges` behaviour is what you want.

## 6. File target

```
/changelog --from=dc2869d --to=HEAD --output=CHANGELOG.md
```

Writes the exact markdown above to `CHANGELOG.md`, overwriting if
present. Stdout gets:

```
Wrote 5 entries to CHANGELOG.md
```

The user is expected to move that content under a `## [0.1.0] — 2026-04-17`
heading in an existing `CHANGELOG.md` by hand — this skill does not
do merge-into-existing-changelog. That is intentional: editing a
curated changelog file is a review act.

## 7. GitHub Release target

```
/changelog --since-last-release --github-release=v0.2.0
```

Assuming the tag `v0.2.0` exists and the release has been created
(by `/release` or by hand), the skill runs:

```
gh release edit v0.2.0 --notes-file -
```

with the rendered markdown on stdin. On success it prints the
release URL.

If the release does not exist yet, the skill stops and asks:

> Release v0.2.0 does not exist. Create it now via
> `gh release create v0.2.0 --notes-file -`? (yes/no)

Default no. The skill never creates a tag — that is `/release`'s job.

## 8. Self-check trace

- [x] Exactly one ref-selection mode (`--from/--to`).
- [x] Both refs resolved via `git rev-parse`.
- [x] Title line filled (`dc2869d..HEAD`).
- [x] Every entry has a linked short SHA.
- [x] Every `#N` is a linked issue.
- [x] Trailing `(#N)` stripped from subjects where the issue is
      already a linked ref.
- [x] Section order is Features → Other. Empty sections omitted.
- [x] No duplicates within a section.

All checks pass. The skill reports back:

> Rendered 5 entries across 2 sections for range `dc2869d..HEAD`.
> Output written to stdout.

## 9. Label-based categorization + similarity dedup

Two real cases this skill handles that pre-v3.4 verb-only logic
got wrong. Both surfaced during the v3.3.0 baseline eval (md-notes
fixture).

### Squash-merge subject without verb prefix → categorized via PR labels (F29)

Input commit (squash merge of a `feature`-labeled PR):

```
8b51b44 Add config loader with documented precedence (#2) (#13)
```

- Verb prefix: none. Verb-fallback would route to "Other".
- Issue tokens: `#2` (issue), `#13` (PR — second `(#N)` in the
  kit's two-suffix shape).
- Label step: `gh pr view 13 --json labels --jq '.labels[].name'`
  returns `feature`. Label-to-section: `feature` → Features.
- **Final section: Features** (correctly recovered from PR labels).

Rendered:

```markdown
## Features

- Add config loader with documented precedence ([8b51b44](...), [#2](...))
```

Versus pre-v3.4 behavior, which routed it to "Other" because the
PR title doesn't carry a `feat(...)` prefix.

### Two near-identical commits → grouped into one entry (F27)

Input commits in the same range:

```
8c1adef fix(prd-normalizer): handle missing exit criteria field
2b9d33a fix(prd-normalizer): handle missing exit criteria gracefully
```

- Both subjects share the same `fix(prd-normalizer):` verb-scope
  prefix. Verb-fallback routes both to the same section (Fixes), so
  they meet the within-section precondition for dedup.
- Remaining tokens after stripping the prefix —
  `handle missing exit criteria field` vs
  `handle missing exit criteria gracefully` — overlap 4/5 = 80%
  after lowercasing and filler-word removal, above the 75%
  threshold.
- Treated as duplicates per "Duplicate detection" rule 3 (same
  verb/scope prefix + ≥75% remaining-token overlap, within section).
  Newest commit (`8c1adef` by commit date) provides the canonical
  text. Both SHAs listed in the rendered entry.

Rendered:

```markdown
## Fixes

- handle missing exit criteria field ([8c1adef](...), [2b9d33a](...))
```

Versus pre-v3.4 behavior, which emitted both as separate entries in
the same section because exact-match dedup didn't catch the
near-identical pair. The strict same-verb/scope rule (rather than a
noun-phrase fallback) was deliberate: it avoids false-positive
merges between distinct commits that happen to share a noun-phrase
head and similar wording. Direct-to-main commits without verb
prefixes are not grouped by the heuristic — they can still dedup via
exact-match or same-PR rules.

### Combined: label categorization + similarity dedup over a real release range

Putting both behaviors together on md-notes' `v0.1.0` release range
(7 commits across direct-to-main work + 2 squash merges, with
`feature` and `infra` labels on the PRs):

```markdown
# Changelog v0.0.0..v0.1.0

## Features

- scaffold normalized PRD (via /idea-to-prd + /prd-normalizer) ([38bc68e](...), [d9430fc](...))
- scaffold MVP and single-phase build-out plan (via /prd-to-mvp) ([6139a12](...))
- draft 5 proposed ADRs (via /adr-writer) ([4c90ebb](...))
- Add config loader with documented precedence ([8b51b44](...), [#2](...))

## Infra

- Bootstrap Go scaffold and CI ([434bc5e](...), [#1](...))

## Chores

- accept ADR-001..ADR-005 after human review ([f416b81](...), ADR-001, ADR-005)
```

Pre-v3.4, the same range produced "Features" with two PRD-scaffold
duplicates (F27) and an "Other" section holding the two squash-
merged PRs that should have been Features and Infra (F29). After
v3.4: 3 sections, no duplicates, no "Other" inversion.
