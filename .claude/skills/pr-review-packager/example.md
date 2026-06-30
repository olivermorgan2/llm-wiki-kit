# pr-review-packager — worked example

A single end-to-end walkthrough showing how `pr-review-packager`
turns a feature branch into a pull request. This is reference
material, not a runtime artifact.

---

## 1. Setup

The user has just finished implementing Issue #17 on a feature
branch:

```
Branch:   issue-17-pr-review-packager
Base:     main
Upstream: origin/issue-17-pr-review-packager (pushed)
Commits ahead of main:
  abc1234  feat(skills): add pr-review-packager skill (ADR-015, #17)
```

Working tree is clean. The user runs:

```
/pr-review-packager
```

No flags.

## 2. Data the skill gathers

**From git:**

- `git symbolic-ref --short HEAD` → `issue-17-pr-review-packager`
- `git rev-parse --abbrev-ref --symbolic-full-name @{u}` → `origin/issue-17-pr-review-packager`
- `git status --porcelain` → empty
- `git log main..HEAD --format="%H%x09%s"` →
  ```
  abc1234	feat(skills): add pr-review-packager skill (ADR-015, #17)
  ```

**From the filesystem:**

- `templates/pr-template.md` exists and is read.
- `design/adr/adr-015-pr-review-packager-skill.md` exists (matched by
  glob `design/adr/adr-0*15-*.md`).

## 3. Derived values

- **Issue number:** `17`, extracted from the branch name
  (`issue-17-pr-review-packager` → `17`). The branch-name source hit
  first; the skill did not need to fall through to commit messages
  or prompt files.
- **ADR references:** `ADR-015`, extracted from the commit subject's
  `(ADR-015, #17)` tail. Resolved to
  `design/adr/adr-015-pr-review-packager-skill.md`.
- **Change summary groups:** one commit, prefix `feat`. Because there
  is only one group, the skill omits the group heading and emits a
  single bullet:
  ```
  - add pr-review-packager skill
  ```
- **PR title:** `feat(skills): add pr-review-packager skill`
  (derived from the newest commit subject, with the trailing
  `(ADR-015, #17)` stripped).
- **Summary draft:** `Adds the pr-review-packager skill that
  packages feature branches into well-structured PRs (closes #17).`

## 4. Rendered body (shown to user for approval)

The skill prints the following to chat and asks
`"Open PR with this title and body? (yes / edit / cancel)"`.

> **Title:** `feat(skills): add pr-review-packager skill`

````markdown
## Summary

Adds the pr-review-packager skill that packages feature branches into well-structured PRs (closes #17).

## Closes

Closes #17

## ADR

Related ADR: design/adr/adr-015-pr-review-packager-skill.md

## Changes

- add pr-review-packager skill

## Test results

```
<!-- TODO: paste test-runner output or delete the code fence and write "no code changes — docs only" -->
```

## Manual verification

<!-- TODO: list verification steps, or write "none needed" -->
````

The skill flags the two TODO comments so the user can fill them
during the `edit` loop if desired.

## 5. User confirmation

The user replies `edit` and types in test results and verification
steps. The skill re-renders the body with the user's text and asks
again. This time the user replies `yes`.

## 6. PR creation

The skill runs:

```
gh pr create \
  --title "feat(skills): add pr-review-packager skill" \
  --body  "<rendered body>" \
  --base  main \
  --head  issue-17-pr-review-packager
```

`gh` returns:

```
https://github.com/olivermorgan2/claude-workflow-kit/pull/42
```

The skill reports:

> Opened PR #42:
> `https://github.com/olivermorgan2/claude-workflow-kit/pull/42`
> (Closes #17, ADR-015; no TODOs remaining.)

## 7. What would change for other scenarios

### Multiple commits, mixed types

Branch `issue-22-workflow-guide` with three commits:

```
feat(docs): add workflow guide skeleton (ADR-020, #22)
fix(docs): fix heading levels (#22)
docs: add cross-links to ADRs (#22)
```

The change summary becomes:

```
**Features**
- add workflow guide skeleton

**Fixes**
- fix heading levels

**Docs**
- add cross-links to ADRs
```

### Branch name does not encode the issue

Branch `add-auth-middleware` with one commit
`feat(auth): add middleware (ADR-030, #31)`.

- Step 6.1 (branch-name regex) misses — no `issue-N` pattern, no
  trailing `-N`, no `#N`.
- Step 6.2 (commit subjects) matches `#31`.
- `Closes #31` is filled, no TODO needed.

### No issue number anywhere

Branch `chore-tidy-notes`, commit `chore: tidy notes`.

- Step 6.1, 6.2, 6.3 all miss.
- The skill emits
  `Closes #<!-- TODO: fill in issue number -->`
  and flags it at the approval gate. The user fills it during the
  `edit` loop.

### No ADR referenced

Same branch as above — no `ADR-NNN` tokens anywhere.

- The `## ADR` section becomes `Related ADR: none`.

### Detached HEAD

The user ran `/pr-review-packager` while rebasing.

- `git symbolic-ref --short HEAD` fails.
- The skill aborts with
  `"HEAD is detached. Check out a branch before running /pr-review-packager."`

### Branch not pushed

User forgot `git push -u origin <branch>`.

- `git rev-parse --abbrev-ref --symbolic-full-name @{u}` errors.
- The skill aborts with
  `"No upstream for <branch>. Run: git push -u origin <branch>"`.

### Optional flags forwarded

```
/pr-review-packager --draft --label feature --reviewer @alice
```

The skill passes `--draft --label feature --reviewer @alice` through
to `gh pr create` in step 14. Everything else is identical.

### Design-questions round-trip (per ADR-040)

This variant shows the cross-skill carry-forward end-to-end —
producer (executor) → preserver (this skill) → consumer
(`prepare-issue`) — using a synthetic two-issue pair (#100 and
#101). It is the worked example referenced by acceptance criterion
3 of issue #69. See [`docs/workflow-guide.md` §6](../../docs/workflow-guide.md#6-cross-skill-carry-forward-adr-040)
for the canonical schema.

**Scenario:** issue #100 implements a depth-tier note template
(`shallow`, `standard`, `deep`). Mid-session the executor realises
issue #101's archival sweep will also need to honour the same
tiering — but issue #101 is not yet open and #100's scope does not
cover it. The executor records the question for #101 instead of
pre-deciding it.

#### Stage 1 — producer (`claude-issue-executor`)

The executor session for #100 ends. Per ADR-040, the eval summary
is printed to chat **and** persisted to `notes/eval-issue-100.md`.
The persisted file's `## Follow-ups` section contains:

````markdown
## Follow-ups

- The depth-tier vocabulary (`shallow`, `standard`, `deep`)
  becomes load-bearing for downstream skills. See the
  design-questions block below.

### design-questions

```yaml
- title: Depth-tier vocabulary alignment for archival sweep
  target-issue: "#101"
  context: |
    Issue #100's templates emit `tier: shallow|standard|deep` in
    each note's frontmatter. Issue #101's archival sweep should
    honour the same vocabulary — confirm whether the sweep matches
    tier names exactly (`tier == "shallow"`) or whether it should
    re-derive tiers from depth measurement. The latter would let
    the sweep work on legacy notes pre-dating the template, but
    risks drift if the template's tier-from-depth logic ever
    changes.
```
````

#### Stage 2 — preserver (`pr-review-packager`)

The user runs `/pr-review-packager` on the issue-#100 branch. The
skill's step 6.5 reads `notes/eval-issue-100.md`, parses the
`### design-questions` YAML block, and groups by `target-issue`
(one entry, target `#101`). Step 11 appends a `## Notes for #101`
section to the rendered PR body, after the standard template body:

````markdown
## Summary

Adds depth-tier note templates with `shallow|standard|deep`
classification (closes #100).

## Closes

Closes #100

## ADR

Related ADR: design/adr/adr-099-depth-tier-templates.md

## Changes

- add depth-tier note templates
- wire tier-from-depth derivation into template rendering

## Test results

```
PASS  templates/notes
  ✓ shallow tier renders correctly
  ✓ standard tier renders correctly
  ✓ deep tier renders correctly
```

## Manual verification

Render a note at each tier, confirm frontmatter `tier` field.

## Notes for #101

Carried forward from this issue's eval summary
(`notes/eval-issue-100.md`). The next executor session for #101
should address these in its plan.

- **Depth-tier vocabulary alignment for archival sweep**: Issue
  #100's templates emit `tier: shallow|standard|deep` in each
  note's frontmatter. Issue #101's archival sweep should honour
  the same vocabulary — confirm whether the sweep matches tier
  names exactly (`tier == "shallow"`) or whether it should
  re-derive tiers from depth measurement. The latter would let
  the sweep work on legacy notes pre-dating the template, but
  risks drift if the template's tier-from-depth logic ever
  changes.
````

The PR is approved by the user, opened via `gh pr create`, and
later merged.

#### Stage 3 — consumer (`prepare-issue`)

The user runs `/prepare-issue 101`. The skill's step 6.5 runs
`gh pr list --state merged --limit 30 --json number,body,mergedAt`,
finds the merged PR for issue #100, and matches its
`## Notes for #101` section. The skill surfaces the scan window at
the review-before-write checkpoint:

> Scanned 30 most-recent merged PRs (#42 → #105); found 1 'Notes
> for #101' section in PR #102 (issue #100, merged
> 2026-05-12T14:22:18Z).

The rendered prompt for issue #101 inserts a
`## Design questions carried forward from PR #102` subsection
immediately before `Requirements`, embedding the matched Notes
verbatim with an explicit instruction to the executor:

```markdown
## Design questions carried forward from PR #102

The following questions were raised by issue #100's eval summary
and preserved in PR #102's body. Address each in the plan you
propose:

- **Depth-tier vocabulary alignment for archival sweep**: Issue
  #100's templates emit `tier: shallow|standard|deep` in each
  note's frontmatter. Issue #101's archival sweep should honour
  the same vocabulary — confirm whether the sweep matches tier
  names exactly (`tier == "shallow"`) or whether it should
  re-derive tiers from depth measurement. The latter would let
  the sweep work on legacy notes pre-dating the template, but
  risks drift if the template's tier-from-depth logic ever
  changes.
```

When the executor session for #101 starts, the question is part of
the prompt — the executor will address it in the plan it proposes,
not discover it mid-session. Audit-trail-grade visibility: anyone
reading the PR sequence (#102 then #103) can trace the
design-coherence work without opening the eval files.

**Empty case.** When the executor for #100 raises no cross-issue
design questions, the eval file omits the `### design-questions`
block entirely; the packager emits no `## Notes for #N` sections;
the prompt for #101 has no
`## Design questions carried forward from PR #M` subsection. Most
issues are this case — the carry-forward chain produces zero
artefact noise on the common path.
