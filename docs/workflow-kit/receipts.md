# Idempotency receipts for mutating skills

A **receipt** is a small JSON file a mutating skill writes when it
finishes (or partially finishes) a unit of work, so a later run can
inspect what was already done and **avoid duplicating a hard-to-reverse
action** — a GitHub issue, PR, milestone, release, or a branch's commits.
Receipts also give a human a concise audit trail of agent activity.

This is the canonical convention; it is governed by
ADR-050. The
machine-readable field list is
[`schemas/receipt.v1.yaml`](../schemas/receipt.v1.yaml).

## Where receipts live

Receipts are written under **`.claude/receipts/`** in the project being
worked on, one JSON file per *(skill, key)*:

```
.claude/receipts/<skill>__<key>.json
```

`<key>` is the **idempotency key** — a stable identifier for the unit of
work (an issue number, a PR number, a milestone slug, a release tag). The
file is **overwritten as the status advances** (`started` →
`completed`/`partial`/`failed`), so it always holds the latest state for
that unit. Receipts are **local, per-run state, not committed artefacts**:
the target-project `.gitignore` excludes `.claude/receipts/`.

## The receipt shape

```json
{
  "schema": "receipt",
  "version": 1,
  "skill": "claude-issue-executor",
  "key": "42",
  "status": "completed",
  "timestamp": "2026-06-05T14:03:00Z",
  "inputs":  { "issue": "42" },
  "outputs": { "branch": "issue-42-foo", "pr": "#118" },
  "next-action": "none"
}
```

`schema`, `version`, `skill`, `key`, `status`, and `timestamp` are
required; `inputs`, `outputs`, and `next-action` are recommended. `status`
is one of `started`, `completed`, `partial`, `failed`. See the schema file
for the full field semantics.

## Writing a receipt

Two equivalent ways — a skill uses whichever is available where it runs:

1. **The helper** (kit repo, or any project that has it on `PATH`):

   ```bash
   # before the mutation
   bin/write-receipt --skill claude-issue-executor --key 42 --status started
   # after it succeeds
   bin/write-receipt --skill claude-issue-executor --key 42 --status completed \
     --outputs '{"pr":"#118","branch":"issue-42-foo"}' --next none
   ```

2. **Direct file write.** A target project that does not ship the helper
   writes the JSON above to `.claude/receipts/<skill>__<key>.json`
   directly, following the schema. The helper is a convenience, not a
   dependency — the convention is the file, not the script.

## Reading receipts to avoid duplicate work

Before performing a mutation, a skill checks for an existing terminal
receipt for its key and **skips the already-done work** if one is found:

```bash
# exit 0 if a completed receipt exists for this key, else exit 1
bin/write-receipt --find --skill release --key v4.2.0 --status completed
```

or, without the helper, read `.claude/receipts/release__v4.2.0.json` and
check its `status`. A `completed` receipt means the unit is done; a
`partial`/`failed` receipt names what remains in `next-action`.

## Which skills write receipts

Receipts target **mutating skills whose actions are public or
hard-to-reverse** — the cat-2 and cat-3 skills in
[`kit.json`](../kit.json):

| Skill | Category | Key | Guards against |
|---|---|---|---|
| `claude-issue-executor` | 2 | issue number | re-running commits/branch work on a resumed issue |
| `pr-review-packager` | 3 | issue or PR number | opening a duplicate PR |
| `issue-planner` | 3 | milestone/plan id | filing the same issues twice |
| `complete-milestone` | 3 | milestone number | double-closing a milestone |
| `release` | 3 | release tag | cutting a duplicate tag/Release |

**Boundary (documented, not a gap).** Cat-1 file-writing skills
(`adr-writer`, `idea-to-prd`, `prepare-issue`, …) do **not** write
receipts: their outputs are local files, already idempotent (re-running
overwrites the same path reversibly), so a receipt would add audit noise
without preventing harm. Receipts are for the actions you cannot simply
re-run. Extending receipts to cat-1 skills, and shipping `write-receipt`
into target projects via the installer, are deferred follow-ups recorded
in ADR-050.

## Pointers

- Schema: [`schemas/receipt.v1.yaml`](../schemas/receipt.v1.yaml)
- Writer/reader: [`bin/write-receipt`](../bin/write-receipt)
- ADR: `adr-050-reliability-validation-self-test.md`
- Related — the structured next step a resuming run reads:
  [`docs/workflow-control.md` §4](workflow-control.md#4-finding-the-next-step)
