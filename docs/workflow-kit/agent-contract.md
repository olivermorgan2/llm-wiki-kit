# Agent Contract

The machine-readable layer of the Claude Code Workflow Kit. An AI
agent (or any script) can use this layer to discover the kit's skills,
permission categories, inputs, outputs, and handoff order — and to
drive the kit's programmatic surfaces — **without reading every
`SKILL.md` body**.

The layer was introduced in Milestone M2 and is governed by
ADR-047. It
has four parts:

1. [`kit.json`](#1-kitjson--the-index) — the aggregate index.
2. [Skill frontmatter](#2-skill-frontmatter) — per-skill structured fields.
3. [The `bin/*` JSON envelope and exit codes](#3-the-bin-json-envelope-and-exit-codes).
4. [Programmatic surfaces](#4-programmatic-surfaces) — read-only scripts.

A validator, [`bin/validate-kit-json`](#5-keeping-it-honest), keeps the
index and frontmatter in agreement.

---

## 1. `kit.json` — the index

[`kit.json`](../kit.json) at the repo root lists every skill and the
kit's conventions in one document. Read it first.

Top-level keys:

| Key | What it holds |
|---|---|
| `kit`, `kitVersion`, `schemaVersion` | identity and version |
| `contract` | pointers to this ADR, this doc, and the validator |
| `permissionCategories` | the `1` / `2` / `3` legend |
| `exitCodes` | the `0`–`4` exit-code legend |
| `workflowOrder` | the happy-path skill sequence |
| `skills` | one entry per skill (see below) |
| `bin` | the programmatic scripts and which envelope each speaks |

Each `skills[]` entry:

```json
{
  "name": "prepare-issue",
  "slashCommand": "/prepare-issue",
  "permissionCategory": 1,
  "description": "Auto-fill an implementation prompt from a GitHub issue and ADRs.",
  "inputs":  [ { "name": "issue-number", "required": true, "description": "..." } ],
  "outputs": [ { "artefact": "prompts/issue-NNN-short-title.md", "description": "..." } ],
  "next":    [ { "skill": "claude-issue-executor", "when": "the prompt file is written" } ]
}
```

Useful queries (the kit ships `jq`-friendly JSON):

```bash
# List every skill and its permission category
jq -r '.skills[] | "\(.permissionCategory)  \(.name)"' kit.json

# What runs after prepare-issue?
jq '.skills[] | select(.name=="prepare-issue") | .next' kit.json

# Which skills are non-substitutable (cat-3, always need approval)?
jq -r '.skills[] | select(.permissionCategory==3) | .name' kit.json

# The happy-path order
jq -r '.workflowOrder[]' kit.json
```

`kit.json` is **hand-maintained** and lives at the kit level. It is not
copied into target projects in this iteration (see ADR-047 deferrals);
target-project agents read the per-skill frontmatter instead, which
*is* copied with `skills/`.

### Permission categories

| Category | Meaning |
|---|---|
| `1` | substitutable — local and reversible; auto-mode runs without asking |
| `2` | operator-acknowledged-bypass — auto-mode proceeds but pauses to acknowledge a bypass in chat |
| `3` | non-substitutable — public or hard-to-reverse; always requires explicit approval |

The full classification rules are in
[`workflow-guide.md` §7](workflow-guide.md#7-auto-mode-permission-contract-adr-041).

### Operating modes and routing (for agents)

An agent driving the kit should read two more things from
[`workflow-control.md`](workflow-control.md) (ADR-048):

- **Operating mode** — the session runs in `interactive` (default),
  `assisted`, or `autonomous`. The mode relaxes friction on cat-1/cat-2
  operations but **never** on cat-3: regardless of mode, every cat-3
  skill requires explicit operator approval. Treat the permission
  `category` in this index as the hard constraint; the mode only changes
  how cat-1/cat-2 work is gated.
- **Next action** — `design/state.md` carries an optional, marker-fenced
  `next-action` zone (`skill`, `args`, `preconditions`, `blocked-by`).
  When present and unblocked it is the deterministic next step; the
  `start` skill is the router that turns it (or, in its absence, the
  project's artefacts) into a recommendation. Use exact skill names from
  this index, not the human verb layer, when acting on it.

---

## 2. Skill frontmatter

Every `skills/<name>/SKILL.md` carries the same structured fields in its
YAML frontmatter, so a single installed skill describes itself wherever
it runs:

```yaml
---
name: prepare-issue
description: ...
permission-category: 1  # ...
inputs:
  - name: "issue-number"
    required: true
    description: "GitHub issue number (no #)"
outputs:
  - artefact: "prompts/issue-NNN-short-title.md"
    description: "Filled implementation prompt"
next:
  - skill: claude-issue-executor
    when: "the prompt file is written"
---
```

- `inputs` — the skill's arguments and flags (positional names, or
  `--flag` names), each with `required` and a short `description`. A
  skill that takes no arguments has `inputs: []`.
- `outputs` — produced artefacts. `artefact` is a file path
  (`design/prd.md`) or an external artefact in parentheses
  (`(GitHub PR)`, `(git tag + GitHub Release)`).
- `next` — handoff skills with a one-phrase `when` condition. A
  terminal skill has `next: []`.

`kit.json` mirrors these fields for the aggregate view. The
[validator](#5-keeping-it-honest) checks that the two agree on the
skill set, names, permission categories, and `next` targets.

---

## 3. The `bin/*` JSON envelope and exit codes

New `bin/*` scripts that support `--format json` emit one **standard
envelope** and use a shared exit-code convention. The normative helper
is [`bin/lib/json-envelope.sh`](../bin/lib/json-envelope.sh); source it
to get the exit-code constants and the `print_envelope` function.

Standard envelope:

```json
{
  "skill":   "prepare-issue",
  "version": "1",
  "status":  "ok",
  "outputs": { },
  "next":    [ { "skill": "...", "args": "...", "when": "..." } ],
  "errors":  [ { "code": "...", "message": "..." } ]
}
```

- `skill` — the script name.
- `version` — the envelope schema version (currently `1`).
- `status` — a machine-readable status string (script-specific, e.g.
  `ok`, `stale`, `prompt-missing`, `degraded`, `in-sync`, `drift`).
- `outputs` — the script's result payload (object or array).
- `next` — recommended follow-up actions, each naming a `skill`.
- `errors` — structured errors; empty on success.

Exit-code convention (a script uses the subset it needs):

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | domain failure (the script ran; the answer is "no/failed") |
| `2` | invocation error (bad flags, unreadable input, missing file) |
| `3` | auth/service failure (gh/network/credential problem) |
| `4` | user cancellation (operator declined at a prompt) |

The exit code is the contract for control flow; `status` adds detail.
For example, `bin/prepare-issue` exits `0` whether the prompt is
present, missing, or stale — that distinction is in `status`, not the
exit code — and exits `3` only when GitHub itself is unreachable.

### Legacy surface: `bin/check-plan`

[`bin/check-plan`](../bin/check-plan) predates this convention
(ADR-043). It emits
`{criteria-set, result, criteria[]}` and uses `0/1/2`. It is the
documented **legacy** surface, kept stable for its in-tree callers
(`adr-writer`, `prepare-issue`); re-aligning it to the standard
envelope is a tracked follow-up, not a contradiction of this contract.

---

## 4. Programmatic surfaces

| Script | Envelope | Exit codes | Purpose |
|---|---|---|---|
| [`bin/prepare-issue`](../bin/prepare-issue) | standard | 0, 2, 3 | Read-only analysis of an issue's prepared prompt |
| [`bin/validate-kit-json`](../bin/validate-kit-json) | standard | 0, 1, 2 | Check `kit.json` agrees with frontmatter |
| [`bin/validate-carry-forward`](../bin/validate-carry-forward) | standard | 0, 1, 2 | Check carry-forward design-questions blocks against [`schemas/design-questions.v1.yaml`](../schemas/design-questions.v1.yaml) |
| [`bin/check-consistency`](../bin/check-consistency) | standard | 0, 1, 2 | Check docs and metadata agree (skills.md, verb layer, bin registry, schema refs) |
| [`bin/write-receipt`](../bin/write-receipt) | standard | 0, 1, 2 | Write or look up an idempotency receipt under `.claude/receipts/` ([`docs/receipts.md`](receipts.md)) |
| [`bin/self-test`](../bin/self-test) | standard | 0, 1, 2 | Run the non-mutating validation surface against the kit + a stub, timed ([`docs/self-test.md`](self-test.md)) |
| [`bin/review-pr`](../bin/review-pr) | standard | 0, 1, 2, 3 | Generate a dry-run AI review artifact from a PR diff; posts nothing ([`docs/ai-review.md`](ai-review.md)) |
| [`bin/publish-review`](../bin/publish-review) | standard | 0, 1, 2, 3, 4 | Preview, then post an AI review to a PR on the `--confirm` token; idempotent via receipts |
| [`bin/review-eval`](../bin/review-eval) | standard | 0, 1, 2 | Offline fixture harness for AI PR review quality |
| [`bin/fence`](../bin/fence) | standard | 0, 1, 2 | Read and splice marker-fenced zones (state/summary/workflow-docs/adr-index) without clobbering surrounding prose (ADR-054) |
| [`bin/fences-eval`](../bin/fences-eval) | standard | 0, 1, 2 | Offline golden-fixture harness proving byte-stable fence round-trip/splice |
| [`bin/validate-schema`](../bin/validate-schema) | standard | 0, 1, 2 | Validate a JSON data file against a kit schema (unit/required/type/const/enum/pattern) |
| [`bin/adr-alloc`](../bin/adr-alloc) | standard | 0, 1, 2 | Allocate the next ADR number/date/slug/index-row stub; writes nothing |
| [`bin/changelog-collect`](../bin/changelog-collect) | standard | 0, 2 | Parse git history into grouped, deduped, categorized commit data as JSON (offline via `--git-log-file`/`--gh-mock`) |
| [`bin/pr-context`](../bin/pr-context) | standard | 0, 2 | Extract PR-packaging context (issue/ADR/commit-groups/carry-forward) as JSON; writes nothing |
| [`bin/docs-render`](../bin/docs-render) | standard | 0, 1, 2 | Splice marker-fenced doc sections (omit/merge/collapse) for the workflow-docs skill |
| [`bin/collect-eval`](../bin/collect-eval) | standard | 0, 1, 2 | Offline fixture harness asserting changelog-collect / pr-context outputs are byte-stable |
| [`bin/check-plan`](../bin/check-plan) | legacy | 0, 1, 2 | Validate an ADR or prompt against criteria |
| [`bin/sync-adr-index`](../bin/sync-adr-index) | none (text) | 0, 1, 2 | Regenerate the ADR index table |

The standard-envelope validators (`validate-kit-json`,
`validate-carry-forward`, `check-consistency`) and `self-test` are the
checks wired into CI ([`.github/workflows/kit-checks.yml`](../.github/workflows/kit-checks.yml)).
The canonical schemas they read live under [`schemas/`](../schemas/).

### `bin/prepare-issue --issue N --format json`

A read-only companion to the interactive `/prepare-issue` skill. It
**writes nothing**; it reports the current state so an agent can decide
the next step without prose interpretation:

```bash
bin/prepare-issue --issue 17 --format json
```

`outputs` payload:

```json
{
  "issue": 17,
  "promptPath": "prompts/issue-017-...md",
  "promptExists": true,
  "stale": false,
  "adrReferences": ["ADR-013", "ADR-014"],
  "gaps": []
}
```

- `stale` is `true` when the live issue was updated after the prompt
  file was last written (mtime vs. `gh issue view ... --json updatedAt`),
  `false` when current, and `null` when undetermined (no prompt, or
  GitHub unreachable).
- `next` always recommends `claude-issue-executor`, with a `when`
  condition reflecting the current `status`.
- If `gh` is missing, unauthenticated, or the issue lookup fails, the
  script still emits a degraded (`status: "degraded"`) prompt-only
  analysis and exits `3`.

---

## 5. Keeping it honest

`kit.json` and the skill frontmatter are two representations of the
same facts, so they can drift. [`bin/validate-kit-json`](../bin/validate-kit-json)
is the synchronization check:

```bash
bin/validate-kit-json            # text; exit 0 in sync, 1 on drift
bin/validate-kit-json --format json
```

It verifies that `kit.json` is well-formed, that its skill set exactly
matches the `skills/` directories, that each skill's `name` and
`permissionCategory` match that skill's frontmatter, and that every
`next.skill` target names a skill in the index. Run it whenever you
add, rename, remove, or re-categorise a skill, or edit a skill's
frontmatter.

Deeper doc↔metadata consistency (the prose in `docs/skills.md` and the
verb layer vs. the structured index, the `bin` registry, and the schema
references) is the job of its sibling
[`bin/check-consistency`](../bin/check-consistency), added in M4. Run both:
`validate-kit-json` owns `skills/` ↔ `kit.json` agreement, and
`check-consistency` owns the surrounding docs and registry.

---

## Preflight for GitHub-integrated skills

Several skills and `bin/prepare-issue` call `gh`. Before an agent
starts a run that touches GitHub, confirm authentication and scopes per
the preflight checklist in
[`docs/github-setup.md`](github-setup.md#github-credentials-and-required-scopes).
