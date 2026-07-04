# Authoring skill: location, packaging, and adapter contract

This note records the design decisions behind the first *shipped* Claude Code
skill asset, `wiki-authoring` (issue #38), and the engine contract the skill
adapter depends on. It is the contract-doc home for the authoring flow, alongside
[`exit-codes.md`](exit-codes.md), [`validation.md`](validation.md),
[`install.md`](install.md), and [`platform-selection.md`](platform-selection.md).

## Shipped-skill location / packaging

**Decision: shipped plugin skills live at repo-root `skills/<skill-name>/SKILL.md`.**
Issue #38 adds `skills/wiki-authoring/SKILL.md`.

Rationale:

- The kit is distributed as a **Claude Code plugin package** whose root already
  holds the release-built `bin/<goos>_<goarch>/llm-wiki` tree and
  `bin/SHA256SUMS` (see [`install.md`](install.md),
  [`platform-selection.md`](platform-selection.md)). Claude Code's plugin layout
  expects `skills/<name>/SKILL.md` at the plugin root, so the repo root doubles
  as the plugin-package source root: `skills/` (source-controlled assets) sits
  sibling to `bin/` (release-built binaries).
- `.claude/skills/` is **workflow-kit process tooling** for developing this
  repository and is *not* shipped. The build-out-plan sketch of "`.claude/` —
  plugin manifest, skills" is superseded on this point.

### Explicitly deferred (not part of #38)

- **Plugin manifest** (`.claude-plugin/plugin.json`), **marketplace packaging**,
  and the **plugin name** (open questions Q1/QB1). Adding `skills/` requires none
  of them.
- **Wiring `skills/` into the release-bundle job.** Shipping the skill inside the
  built release archive is a follow-up for the packaging issue (Phase 6/7); #38
  only adds the source asset.
- **Install-time skill delivery.** `llm-wiki install` still writes exactly the
  four-file wiki bundle (ADR-009); skill delivery is Claude Code plugin install,
  outside the engine.

## Adapter flow contract

The skill is a thin adapter (MVP principle 3: skill scripts carry no independent
logic). All policy — validation, hashing, staging, the citation-loss gate — lives
in the engine. The adapter's job is to run the engine in the right order and
relay the results. The flow is **draft → validate → plan (diff) → apply →
confirm**.

| Step | Command | Engine role |
| ---- | ------- | ----------- |
| Draft | (none — scratch file or stdin, outside the bundle) | — |
| Validate | `llm-wiki page inspect <path> --content <file\|-> --json` | Validate proposed bytes in full bundle context; stage nothing. |
| Plan | `llm-wiki page plan <path> --content <file\|-> --json` | Normalize frontmatter, bind base hash, stage under `.llm-wiki/staging/<txn>/`, render the diff, raise the citation-loss gate (ADR-008 sub-decision 6). |
| Apply | `llm-wiki page apply <txn> [--approve] --json` | Re-verify base hashes, then commit the staged change set (ADR-006). |
| Confirm | `llm-wiki validate <root> --json` | Confirm the committed page is valid. |

### `page inspect --content` (the one engine addition in #38)

The flow puts *validate* before *plan diff*, but no shipped surface could
validate content that is not yet on disk: `validate` and `page inspect` read the
live tree, `page plan` defers content validation to apply, and post-apply
validation would transiently commit an invalid page. So #38 extends `page inspect`
with `--content <file|->` (mirroring `page plan`'s flag):

- The proposed bytes are validated over an **overlay of the live bundle** — the
  draft is served at its target path and injected into the bundle file set, so
  broken-link and citation membership see it (criterion 15: identical findings
  across surfaces). Intermediate directories that do not exist on disk are
  synthesized for the run only.
- The target **may be absent** (a new-page draft); a non-`.md` path or a boundary
  escape is still rejected exactly as live inspect rejects it.
- Nothing is staged or written. `contentHash` in the report is the SHA-256 of the
  proposed bytes as read (pre-normalization; normalization stays `page plan`'s
  job).

No contract change accompanies this: no new envelope fields, no new exit codes,
no new command — one new flag on an existing subcommand, carried by the existing
`page` payload and `findings` members.

### Envelope fields the adapter consumes

- `status` / exit code — control flow (see the exit-code table below).
- `findings[].code` — `core-citation-*` (a sourced claim's citation does not
  resolve) and error-severity findings gate iteration at the validate step.
- `plan.noOp` — already up to date; stop (no duplicate).
- `plan.diff` — shown verbatim before any write.
- `plan.transaction` — the id `page apply` consumes.
- `approval` — a raised citation-loss gate; surfaced for explicit human approval.
- `apply.committed` — the paths committed.

### `--approve` discipline

The adapter passes `--approve` to `page apply` **only** when the human explicitly
approved a citation loss reported by `page plan` (the `approval` block). It never
adds `--approve` on its own; an un-granted approval requirement is exit `3`, which
means stop and ask.

## Exit codes the adapter branches on

Frozen contract ([`exit-codes.md`](exit-codes.md); source of truth
`internal/contract`):

| Code | Meaning for the adapter |
| ---- | ----------------------- |
| 0 | ok |
| 1 | ok with warnings |
| 2 | findings — fix the draft and re-validate |
| 3 | approval required — a reported citation loss; stop and ask |
| 4 | re-plan or bad invocation — for a stale apply, re-plan and re-show the diff |
| 5 | system/boundary failure — stop and report |

## Acceptance-criteria mapping (prd-addenda/001)

| Criterion | Where it is enforced / tested |
| --------- | ----------------------------- |
| 9 — sourced claims carry resolvable citations | The evidence-context citation resolver (`page inspect --content`); `TestAuthoringSkillFlowHappyPath` asserts zero `core-citation-*` findings on the sourced-claim fixture. |
| 10 — repeated unchanged input creates no duplicate | `page plan` no-op detection; `TestAuthoringSkillFlowIdempotent`. |
| 11 — edits previewed before application | `page plan` diff shown before `page apply`; `TestAuthoringSkillFlowHappyPath` / `…EditPreview` assert the tree is untouched before apply. |
| 12 — skills mutate only through the engine's staged plan/apply | The adapter performs no direct write; `TestAuthoringSkillFlowEngineMediatedMutation` asserts only staging changes before apply and only the target page after. |

## Non-goals

No plugin manifest, marketplace packaging, or plugin naming; no release-bundle
change to ship `skills/`; no enrichment skill (Phase 5); no hooks (Phase 6); no
academic-research profile rules or profile-loaded evidence vocabulary (Phase 4 —
the interim `LLM_WIKI_EVIDENCE_SECTIONS` channel stays); no new envelope fields,
exit-code changes, or staging garbage collection.
