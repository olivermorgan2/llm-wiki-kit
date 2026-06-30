# Workflow self-test — time-to-first-PR

The kit's promise is that it gets a project from an idea to a first
reviewable PR with little friction. This self-test is how we **prove that
the kit is becoming more frictionless release over release** rather than
assuming it. It is governed by
ADR-050 and
has two halves: an automated, non-mutating check and a manual, full-flow
walk.

## 1. Automated self-test (non-mutating) — `bin/self-test`

[`bin/self-test`](../bin/self-test) exercises the kit's read-only
validation surface against the kit itself and a throwaway stub. It touches
no GitHub and writes nothing outside a temp dir, so it is safe to run on
every change and in CI.

```bash
bin/self-test                      # human summary
bin/self-test --format json        # standard envelope
bin/self-test --append-log notes/self-test-log.md
```

It runs and scores:

- the surface checks — `validate-kit-json`, `validate-carry-forward`,
  `check-consistency`, `check-state-cap`, and `bash -n` across all `bin/*`;
- stub self-checks on throwaway data — the carry-forward validator accepts
  a well-formed block and rejects a malformed one, and the receipt
  writer/reader round-trips — which is the "stub project tests the
  non-mutating parts" requirement, and doubles as the regression test for
  the M4 validators and receipt logic;
- a tool-presence check (`python3`, `jq`, `bash` required; `gh` optional).

It reports **elapsed seconds**, the **step count** ("commands" run), and a
**friction** list (any step whose exit code was not what the self-test
expected, or a missing required tool). Exit `0` = ok / `0` with `degraded`
status when only optional tools are missing / `1` on friction.

What it deliberately does **not** do: open a branch, create a PR, or call
`gh`. Measuring the *mutating* path is the manual protocol below.

## 2. Manual self-test (full flow) — time and friction to a first PR

Run this on a throwaway repo when preparing a release, or whenever a
workflow change might have moved the friction. The goal is a number
(elapsed time), a count (commands/skills invoked), and a short list of
friction points.

1. **Start a stopwatch** and create an empty git repo + GitHub remote.
2. Install the kit: `bin/install-workflow-kit` (see
   [`docs/install.md`](install.md)).
3. Walk the happy path with a tiny toy idea, counting each skill you
   invoke:
   `/idea-to-prd` → `/prd-normalizer` → `/prd-to-mvp` → `/adr-writer`
   (one ADR) → `/issue-planner` (one issue) → `/work` (`prepare-issue` →
   `claude-issue-executor`) → `/ship` (`pr-review-packager`).
4. **Stop the stopwatch** when the first PR is open.
5. Record, in `notes/self-test-log.md`:
   - elapsed wall-clock time,
   - number of skills/commands invoked,
   - every point of friction (a step that needed a retry, an unclear
     prompt, a manual workaround, a missing check).

Keep the toy project trivial — the measurement is about the kit's
ceremony, not the project's complexity.

## 3. Release notes

When preparing a release, summarise the **most recent** automated and
manual self-test results (from `notes/self-test-log.md`) in the release
notes, so each release can be compared against the last. The `release`
skill's changelog step is the natural place; see
[`skills/release/SKILL.md`](../skills/release/SKILL.md).

## Pointers

- Script: [`bin/self-test`](../bin/self-test)
- Results log: `notes/self-test-log.md`
- ADR: `adr-050-reliability-validation-self-test.md`
