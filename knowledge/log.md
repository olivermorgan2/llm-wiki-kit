# Log — decisions & review outcomes — llm-wiki-kit

Chronological log of decisions and review outcomes (knowledge-layer `log.md`).
Lightweight; an item
that is genuinely architectural should be promoted to an ADR in
`design/adr/` via `/adr-writer`, with a back-reference here.

Format: date — decision — rationale — links.

---

### 2026-06-30 — Bootstrap from `claude-workflow-kit` v5.0.1

Bootstrapped `llm-wiki-kit` into `/Users/hermes/llm-wiki-kit` (mirrors the
existing public repo `olivermorgan2/llm-wiki-kit`) using the public
release's `bootstrap-workflow-kit` installer at pinned tag **v5.0.1**.

- **Visibility:** public (repo already existed, public, and empty; cloned
  and initialized in place — not deleted/recreated).
- **Flags used:** `--with-docs`, `--license=mit`,
  `--license-holder="Oliver Morgan"`. **Omitted `--with-ai-review`** — our
  review mechanism is Codex on the PRD, not the kit's OpenRouter PR-review
  runtime.
- **Day-one CLAUDE.md fields set:** `PROJECT_NAME=llm-wiki-kit`,
  `GITHUB_OWNER=olivermorgan2`, `GITHUB_REPO=llm-wiki-kit`,
  `DEFAULT_BRANCH=main`.

Rationale: the PRD is a finished, standard-shaped artifact, so the flow
skips `idea-to-prd` and starts at PRD normalization.

### 2026-06-30 — Start at PRD normalization; PRD placed at `design/prd.md`

The source PRD ("Claude Code Kit for OKF Wikis", updated 2026-06-20) was
copied verbatim to `design/prd.md` (the kit's expected input location) and
archived verbatim at `knowledge/sources/prd-original.md`.

### 2026-06-30 — Normalized PRD generated

Ran the `prd-normalizer` procedure against `design/prd.md`, producing the
canonical 11-field `design/prd-normalized.md`. The source PRD is rich, so
all hard-required fields resolved without invention; the only `[TBD]`
carries the pending Codex review. `design/prd.md` was left untouched per
the skill contract.

### 2026-06-30 — License = MIT, holder "Oliver Morgan"

Used the installer's `--license=mit` option (ADR-025/ADR-030 in the kit).
Note: PRD §19 still lists the *plugin/marketplace* license as an open
product decision (see [`open-questions.md`](open-questions.md) Q1) — this
decision covers the repository scaffold's LICENSE file.

### 2026-06-30 — Revision-target convention = `design/prd-addenda/`

Codex review fixes/addenda go in `design/prd-addenda/NNN-*.md` (additive,
non-destructive) rather than in-place edits to `design/prd-normalized.md`,
unless a kit command itself must regenerate normalized output. See
[`../design/prd-addenda/README.md`](../design/prd-addenda/README.md).

### 2026-06-30 — Codex adversarial PRD review — verdict `NEEDS_REVISION`

The adversarial Codex review of `design/prd-normalized.md` (with
`design/prd.md` for context) ran and returned **`NEEDS_REVISION`**. The
verbatim review artifact is archived at
[`reviews/2026-06-30-codex-prd-review.md`](reviews/2026-06-30-codex-prd-review.md).

**Three blocking (High) findings, all accepted:**

1. The normalized PRD declares itself the only downstream input but the
   11-field form dropped the source PRD's §17 MVP acceptance criteria.
2. Open questions Q3/Q6/Q7/Q8 (CI scope, Claude Code version floor, binary
   selection, JSON-contract compatibility) shape implementation and
   acceptance tests — must resolve or assumption-lock before MVP scoping.
3. The MVP-scope academic-research profile was prose-only with open
   templates (Q4) — unplannable/untestable without a concrete contract.

**Non-blocking (Medium) findings, accepted:** wide MVP surface needs a
slice order; success signals need gate-vs-measurement labels and
thresholds; custom-profile vs third-party-registry boundary should be
drawn.

**Resolution — required addenda created before `/prd-to-mvp`:**

- [`design/prd-addenda/001-mvp-acceptance-criteria.md`](../design/prd-addenda/001-mvp-acceptance-criteria.md) — carries §17 forward as a 12th normalized field; labels success signals.
- [`design/prd-addenda/002-mvp-planning-assumptions.md`](../design/prd-addenda/002-mvp-planning-assumptions.md) — assumption-locks Q3/Q6/Q7/Q8 (GitHub Actions only; single CC version floor, no shim; one binary-selection mechanism; contract v1 with no pre-release backward-compat).
- [`design/prd-addenda/003-academic-research-profile-contract.md`](../design/prd-addenda/003-academic-research-profile-contract.md) — minimum research-profile contract + per-type fixtures (Q4).
- [`design/prd-addenda/004-mvp-slice-order-and-fixture-plan.md`](../design/prd-addenda/004-mvp-slice-order-and-fixture-plan.md) — must-pass spine (Slices 0–3) vs in-MVP hardening (4–6) + fixture plan.
- [`design/prd-addenda/005-custom-profile-boundary.md`](../design/prd-addenda/005-custom-profile-boundary.md) — MVP custom profiles local-file only; registry/trust → Phase 3 (Q5).

**Knowledge-layer updates:** QB2 closed; Q3/Q4/Q6/Q7/Q8 marked
assumption-locked / must-resolve, Q5 scoped to Phase 3
([`open-questions.md`](open-questions.md)); three review risks + an
assumption-creep risk added ([`risks.md`](risks.md)).

**Gate status:** the normalized PRD plus addenda 001–005 now carry the
acceptance criteria, planning assumptions, domain contract, slice order,
and scope boundary that the review required. The PRD gate is **READY for
`/prd-to-mvp`** under the recorded assumptions; the addenda named
assumptions Oliver can still override (which would revise the affected MVP
issues, not silently persist). `/prd-to-mvp` was **not** started in this
session, per scope.

### 2026-06-30 — MVP scoped via `/prd-to-mvp`; working name + Go/YAML assumption-locked

Ran the `prd-to-mvp` skill against `design/prd-normalized.md` **plus** addenda
001–005, producing [`design/mvp.md`](../design/mvp.md) and
[`design/build-out-plan.md`](../design/build-out-plan.md) from the kit
templates. No user elicitation was needed — every blocker was already
assumption-locked by the addenda; the two remaining items were locked per the
phase mandate.

- **Granularity:** `standard`; **7 phases** mapping 1:1 to the addendum-004
  slices (Slice 0→Phase 1 … Slice 6→Phase 7) plus a continuous cross-platform
  gate. Must-pass spine = Phases 1–4; in-MVP hardening = Phases 5–7.
- **Acceptance gate:** all 21 addendum-001 criteria mapped to an owning
  phase/milestone (explicit table in the build-out plan). Only hard
  quantitative gate = zero out-of-boundary writes.
- **Assumption lock — working name (QB1/Q1):** `llm-wiki-kit` repo/plugin +
  `llm-wiki` CLI for MVP; final public name/namespace/license/marketplace
  stays open and Oliver-overridable before packaging (bounded rename, not a
  scope change).
- **Assumption lock — Go/YAML (Q2):** Go 1.24.x (conservative current-stable
  line) + `github.com/goccy/go-yaml` (node-aware, round-trip-preserving;
  `yaml.v3` archived). Recorded as the **ADR-001** candidate, not a settled
  decision — reversible, ratified/revised by the first engine ADR. A YAML
  round-trip risk was added to [`risks.md`](risks.md).
- **ADR handoff:** the build-out plan surfaces **ADR-001–012** candidates;
  ADR-001–005 are Phase-1 prerequisites. Next step is `/adr-writer`.

Q2 marked assumption-locked, QB1 marked working-name-locked in
[`open-questions.md`](open-questions.md). No addendum or PRD content was
rewritten; the addenda remain source-of-truth refinements. No product/source
code was created (out of phase scope). Validation: no leftover template
placeholders, all 21 criteria present in the map, all relative md links in the
two new files resolve.

### 2026-07-01 — ADR-001–005 drafted (`proposed`) — Phase-1 engine-foundation gate

Drafted the five Phase-1-prerequisite ADRs from the build-out plan's "Decisions
needing ADRs" list (topics 1–5), in one batch so `adr-alloc` numbered them
001–005. ADR bodies carry **`Date: 2026-06-30`** (inherited from the plan-mode
drafts, consistent with the surrounding MVP/PRD design batch); the drafting and
gate ran 2026-07-01. All five ship as **`proposed`** — acceptance is a human act
(see [`SCHEMA.md`](SCHEMA.md); no open question flipped to `closed` here).

Topic → ADR map (and the question/criteria each advances):

- [ADR-001](../design/adr/adr-001-go-toolchain-and-yaml.md) — Go 1.24.x +
  `goccy/go-yaml`. **Ratifies the Q2 assumption-lock** (round-trip criterion 6
  needs a node-aware library; `yaml.v3` archived).
- [ADR-002](../design/adr/adr-002-platform-binary-selection.md) — ship + select
  + checksum-verify one per-platform binary (criteria 2, 21). **Advances Q7.**
  Scope: ship/select/verify only; install/upgrade/uninstall **asset ownership**
  (criteria 1, 3, 20) stays **deferred to ADR-009**.
- [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md) — one
  versioned JSON envelope + fixed six-value exit-code set (criteria 14, 15).
  **Advances Q8.** Hash-bound stale-plan rejection (criterion 13) referenced but
  owned by ADR-006.
- [ADR-004](../design/adr/adr-004-validation-and-severity-model.md) — single
  validation engine, ruleset-tagged (OKF vs profile) three-severity findings
  (criteria 5, 7, 8). Provenance (9) → ADR-008; index consistency → ADR-010.
- [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) — one mandatory
  filesystem-safety gate (criterion 17, the only hard quantitative release
  gate). inspect/plan/apply mechanics → ADR-006.

**Conventions/validation:** files match `templates/adr-template.md`; `**Phase:**`
line omitted on all five (cross-cutting engine foundations, not single-phase) so
`sync-adr-index` renders no Phase column. `check-plan --criteria-set adr` passes
deterministic ADR-C1–C4 on all five (exit 0); ADR-C5 emits expected forward-ref
warnings (ADR-006/008/009/010 not yet drafted); ADR-C6 is deferred in the tool.
`sync-adr-index` populated the 5-row table in `design/adr/README.md`. No
template placeholders remain.

**Knowledge-layer updates (curate, don't accrete):** Q2/Q7/Q8 annotated
`assumption-locked → ADR-00N proposed` (still **not** `closed`) in
[`open-questions.md`](open-questions.md); ADR back-references added to the
relevant rows in [`risks.md`](risks.md) (statuses left `open` — the ADR bounds
the risk, it does not retire it before the engineering exists);
[`index.md`](index.md) phase/next-action advanced to the Codex ADR review gate.

**Gate status:** ADR-001–005 are drafted and self-validated but **`proposed`**.
Next step is the Codex ADR/milestone review gate before any human acceptance;
only on acceptance do Q2/Q7/Q8 flip to `closed` (with a back-reference here) and
do accepted ADRs become inputs to `/prepare-issue` for Phase 1. No product/source
code was created; no remote side effects.

### 2026-07-01 — Codex ADR-001–005 review — verdict `NEEDS_REVISION`

Ran the adversarial Codex milestone/ADR review gate over proposed ADR-001–005.
The verbatim review artifact is archived at
[`reviews/2026-07-01-codex-adr-001-005-review.md`](reviews/2026-07-01-codex-adr-001-005-review.md).

**Four blocking findings, accepted as revision targets:**

1. ADR-001 overstates acceptance criterion 6 by treating YAML comment
   preservation as binding; criterion 6 only says unknown frontmatter fields
   survive. Distinguish binding requirement vs best-effort design quality, or
   promote comment preservation to an authoritative addendum/fixture.
2. ADR-002 specifies checksum verification but not the checksum trust root or
   residual supply-chain risk. Define manifest/source/signing boundary or
   explicitly defer provenance/signing.
3. ADR-005 claims all-or-nothing multi-file interruption safety without a
   transaction model. Define staging/manifest/recovery behavior or narrow the
   claim to per-file atomicity and defer cross-file transaction semantics.
4. ADR-004 baseline suppression conflicts ambiguously with hard validation
   failures. Clarify whether malformed YAML / missing required fields can ever
   be baseline-suppressed and how exit codes behave.

**Gate status:** ADR-001–005 remain **`proposed`** and **not accepted**.
Next step is a bounded Claude Code revision pass, then a second Codex ADR review
before human acceptance / closing Q2/Q7/Q8.

### 2026-07-01 — ADR-001–005 revised per Codex review (still `proposed`)

Applied the bounded revision pass addressing all four blocking findings plus the
cheap non-blocking refinements. No ADR status flipped — all five stay
**`proposed`**; Q2/Q7/Q8 stay **open** (assumption-locked); ADR boundaries to
ADR-006/008/009/010 preserved.

- **ADR-001 (blocking 1):** distinguished the **binding** criterion 6 (unknown
  frontmatter fields survive) from **best-effort, non-gated** comment
  preservation across Context, Options, Decision, and Consequences. Repeated the
  same distinction in [`build-out-plan.md`](../design/build-out-plan.md) §Phase-1
  toolchain and §Risk-2, and in the [`risks.md`](risks.md) round-trip row.
- **ADR-002 (blocking 2):** added a **trust-boundary** decision — CI-generated
  checksum manifest bundled in the payload catches wrong-platform/mismatched/
  corrupt binaries but **not** a maliciously rebuilt payload; **signing /
  provenance attestation deferred** (ADR-009 / dedicated supply-chain ADR).
  Residual supply-chain risk made explicit in [`risks.md`](risks.md).
- **ADR-005 (blocking 3):** **narrowed** the atomicity claim to **per-file**
  atomicity; the **cross-file transaction model** (staging manifest, commit
  ordering, recovery/rollback, partial-commit detection) is explicitly deferred
  to **ADR-006**. Softened the criteria 3/20 "follows" claim to "necessary
  primitive, completed by ADR-006" and the untrusted-input claim to
  "contributes to" — mirrored in [`risks.md`](risks.md). ADR-006 remains the
  owner of staged inspect/plan/apply mechanics.
- **ADR-004 (blocking 4):** clarified baseline is a **differential filter** that
  **cannot** suppress malformed YAML (parse must precede comparison) or override
  criterion 7 in release-gate/CI runs; specified exit-code behavior; **moved
  severity precedence into the Decision section**.
- **ADR-003 (non-blocking):** documented JSON opt-in via **`--json`** on every
  command and **deferred exact numeric exit-code values** to the implementation
  issue.

**Validation:** re-ran `check-plan --criteria-set adr` on all five (deterministic
ADR-C1–C4 pass; ADR-C5 expected forward-ref warnings for ADR-006/008/009/010).
No `{{`/`}}`/`_TBD_` placeholders introduced; all five statuses remain
`proposed`. **Gate status:** ready to **rerun the Codex ADR review gate**.

### 2026-07-01 — Codex ADR-001–005 re-review — verdict `READY`

Reran the adversarial Codex milestone/ADR gate after revision commit `dda3687`.
The verbatim re-review artifact is archived at
[`reviews/2026-07-01-codex-adr-001-005-rereview.md`](reviews/2026-07-01-codex-adr-001-005-rereview.md).

Codex marked all four prior blockers **resolved**:

1. ADR-001 now separates binding unknown-frontmatter preservation from
   best-effort, non-gated comment preservation.
2. ADR-002 now states checksum verification is an integrity/corruption check,
   not an authenticity check, and records residual signing/provenance risk.
3. ADR-005 now guarantees per-file atomicity only and defers cross-file
   transaction semantics to ADR-006.
4. ADR-004 now bounds baseline suppression against malformed YAML, CI/release
   gates, required-field errors, and exit-code behavior.

**Gate status:** Codex says **`READY`** for human acceptance. ADR-001–005 still
remain **`proposed`** until Oliver explicitly accepts them. Q2/Q7/Q8 remain
**open** until that acceptance step; after acceptance, flip ADR statuses to
`accepted`, close Q2/Q7/Q8 with this log back-reference, and prepare dependent
Phase 1 implementation issues.

### 2026-07-01 — ADR-001–005 accepted by Oliver; Q2/Q7/Q8 closed

Oliver reviewed the Codex `READY` re-review and **accepted ADR-001 through
ADR-005**. Applied the acceptance bookkeeping (documentation/decision only — no
product/source code):

- **ADR status flips:** all five ADR bodies moved `proposed` → **`accepted`**;
  `sync-adr-index` regenerated the 5-row table in
  [`../design/adr/README.md`](../design/adr/README.md) to match. ADR `**Date:**`
  lines left at their authoring date (2026-06-30); acceptance date is recorded
  here (2026-07-01).
- **Open questions closed** ([`open-questions.md`](open-questions.md)):
  - **Q2** (Go version + YAML library) → **closed** by accepted ADR-001
    (Go 1.24.x + `goccy/go-yaml`).
  - **Q7** (platform-binary selection) → **closed for the ship/select/verify
    scope only** by accepted ADR-002. Full install/upgrade/uninstall **asset
    ownership** (criteria 1, 3, 20) and release **signing/provenance** remain
    **deferred to ADR-009**; residual supply-chain risk stays `open` in
    [`risks.md`](risks.md).
  - **Q8** (JSON contract versioning/compat) → **closed** by accepted ADR-003.
    Carry-forward (not a reopened question): the **exact numeric exit-code
    values** are still a Phase 1 implementation/spec item and must be published
    as a stable code→meaning table **before Phase 1 closes**.
- **Risk register** ([`risks.md`](risks.md)): ADR annotations updated
  `(proposed)` → `(accepted)`; all bounded risks **stay `open`** per
  [`SCHEMA.md`](SCHEMA.md) — an accepted ADR bounds a risk but does not retire it
  before the engineering exists.
- **Phase 1 planning artifact:** drafted
  [`../notes/phase-1-first-issue-spec.md`](../notes/phase-1-first-issue-spec.md)
  — a pre-issue spec for the first Phase 1 implementation issue (deterministic
  CLI skeleton + JSON-contract spine) enabled by accepted ADR-001/003/004/005.
  It carries forward: the ADR-003 numeric exit-code table (before Phase 1 close),
  the ADR-006 draft prerequisite (before any cross-file mutation), and the
  ADR-002≠ADR-009 installer boundary.

**Preserved deferred boundaries:** ADR-006 (staged mutation / inspect-plan-apply
transaction semantics), ADR-008 (provenance/audit), ADR-009 (install/upgrade/
uninstall asset ownership + signing/provenance handoff), ADR-010 (index
consistency) — none decided by this acceptance.

**Validation:** `check-plan --criteria-set adr --non-interactive` re-run on all
five (deterministic ADR-C1–C4 pass; ADR-C5 expected forward-ref warnings for
ADR-006/008/009/010, still undrafted); `sync-adr-index` clean; no
`{{`/`}}`/`_TBD_` placeholders in the ADRs or the planning artifact.

**Next step:** file the Phase 1 backlog as GitHub issues (`/issue-planner`) and
`/prepare-issue` the first one from the pre-issue spec — plus draft **ADR-006**
before any cross-file mutation/transaction work begins. No GitHub issues were
created in this session; no remote side effects.

### 2026-07-01 — Phase 1 backlog filed as GitHub issues #1–#6; issue #1 prompt prepared

Ran the `issue-planner` flow over `design/build-out-plan.md` §"Foundation —
Phase 1" and filed the six Phase-1 (Slice 0, "Engine + contract spine") issues
in `olivermorgan2/llm-wiki-kit`, all under a new **`Foundation`** milestone
(CLAUDE.md canonical; the build-out plan groups Phases 1–2 there). Only the
Phase 1 slice was filed this session; Phases 2–7 remain in the plan backlog.

- **#1** CLI skeleton + versioned JSON-contract spine (ADR-001, ADR-003) —
  `feature`. Merges build-out bullets "scaffold `cmd`/`internal`" + "JSON
  contract envelope + exit codes" per the documented first-issue boundary in
  [`../notes/phase-1-first-issue-spec.md`](../notes/phase-1-first-issue-spec.md).
- **#2** OS/arch detection + release-artifact checksum verification (ADR-002) —
  `infra`. Ship/select/verify **only**; installer/asset-ownership + signing
  stay **ADR-009** (named out-of-scope in the body).
- **#3** Core-profile `validate`, OKF-vs-profile, three severities (ADR-004) —
  `feature`.
- **#4** Broken-link detection at configured severity (ADR-004) — `feature`.
- **#5** Safe filesystem layer — per-file atomic write, symlink/path-traversal
  rejection (ADR-005) — `security`. Per-file atomicity **only**; cross-file
  transaction/staged mutation deferred to **ADR-006** (named out-of-scope).
- **#6** Core-profile fixtures + traversal/symlink testdata (ADR-004, ADR-005)
  — `infra`.

**Labels:** created the CLAUDE.md primary labels the drafts needed
(`feature`, `infra`, `security`) — the freshly-bootstrapped repo carried only
GitHub defaults. One primary label per issue, per CLAUDE.md.

**Project board: skipped.** The authenticated `gh` token
(`gist, read:org, repo`) lacks the `project` OAuth scope, so ADR-012's Project
board step was skipped rather than silently downgraded. Re-run with
`gh auth refresh -s project,read:project` if a board is wanted later.

**Prompt prepared:** ran the `prepare-issue` fill for #1 →
[`../prompts/issue-001-cli-contract-spine.md`](../prompts/issue-001-cli-contract-spine.md).
No `bin/check-plan` ships in this scaffold, so the pre-write gate was run
**manually** against `check-plan` `criteria.md` PROMPT-C1–C6: deterministic
C1/C2/C3/C6 pass, warning C4/C5 clean; all ADR path refs resolve. `design/state.md`
does not exist (this project uses the `knowledge/` layer), so that step was skipped.

**Scope discipline:** no product/source code written; no ADR or plan file
edited in place. Carry-forwards preserved in issue bodies — ADR-003 numeric
exit-code table (before Phase 1 closes), ADR-006 before any cross-file mutation,
ADR-002 ≠ ADR-009 installer boundary.

**Next step:** implement **#1** via `/claude-issue-executor` (plan-first, one
issue per session); draft **ADR-006** before any cross-file mutation work.

### 2026-07-03 — ADR-006/007/009 drafted (`proposed`) — Phase-2 install/init prerequisites

Opened Phase 2 (**`Phase 2 — Install/init`** milestone, issue **#14**, label
`design`, branch `docs/adr-006-007-009-install-init-foundations`) by drafting the
three ADR prerequisites the build-out plan §Phase 2 names as dependencies
(seeds #6, #7, #9 in `design/build-out-plan.md`). All three content-derive from
accepted ADRs + addenda — no un-inferable product decision — so they were
authored under the 2026-07-03 autonomous-phase mandate (Fable plans, Opus 4.8
auto builds, Hermes drives).

- **[ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md)** —
  staged mutation lifecycle + **cross-file transaction model** that
  [ADR-005](../design/adr/adr-005-safe-filesystem-layer.md) explicitly deferred:
  stage the full change set under `.llm-wiki/staging/<txn-id>/`, staging
  manifest with source/target base hashes, validate-before-commit, journaled
  ordered per-file atomic-rename commit, partial-commit detection +
  roll-forward/back, temp cleanup. Phase 2 consumes the **transaction** half;
  Phase 3 consumes the `inspect/plan/apply` UX + hash-bound stale-plan
  rejection. Criteria 3, 11, 12, 13, 20.
- **[ADR-007](../design/adr/adr-007-profile-system-boundary.md)** — declarative
  data profiles, **one** inheritance level (`academic-research` extends `core`),
  id-based resolution with a reserved Phase-7 local-path seam; ratifies the
  existing `internal/profile` `Loader` shape; `init` materializes a profile
  **reference** (not a rule-data copy). Phase 2 exercises only `core`. Criteria
  4, 5; addenda 003/005; Q5 registry/trust stays Phase 3.
- **[ADR-009](../design/adr/adr-009-install-upgrade-uninstall-ownership.md)** —
  explicit **plugin-owned vs repo-owned** asset classification + version-record
  file (plugin/CLI/OKF/profile versions); install via the ADR-006 transaction,
  `--dry-run` zero-mutation preview, silent-overwrite refusal, non-empty-repo
  no-file-loss (install half, Phase 2); upgrade/uninstall preservation
  **decided** here, **implemented** Phase 7. Takes the install-ownership
  remainder ADR-002 carved out (criteria 1, 3, 20). **Signing / provenance**
  attestation is explicitly deferred to a dedicated supply-chain ADR (mirrors
  ADR-002); residual supply-chain risk stays `open`.

**Status = `proposed` (not yet accepted).** Per the plan's Phase-1-precedent
gate, acceptance follows an **independent Codex adversarial ADR review** (run by
Hermes after the PR opens) reaching `READY`; only then are statuses flipped to
`accepted` under the mandate and **flagged for Oliver's async ratification**,
with verbatim review artifacts archived under `knowledge/reviews/`. This entry
records the drafting; no acceptance is claimed here and **Q7 is not closed** —
`open-questions.md` carries a Phase-2 annotation, not a resolution.

**Validation:** `check-plan --criteria-set adr --non-interactive` deterministic
**ADR-C1–C5 pass** on all three (C6 is the standing best-effort warning);
`sync-adr-index` regenerated `design/adr/README.md` (008 gap preserved);
`GOTOOLCHAIN=local go test ./...` green; no `{{`/`}}`/`_TBD_`/`TODO`
placeholders. No product/source code written; no accepted ADR edited in place.

**Next step:** Hermes runs the independent Codex ADR gate on the PR; on `READY`,
flip statuses to `accepted`, re-sync the index, close Q7's install-ownership
remainder, and archive the review. Then Phase 2 train items 1–7 proceed.

### 2026-07-03 — ADR-006/007/009 Codex re-review `READY`; accepted under mandate

The independent Codex adversarial ADR gate re-ran on PR #22 after the loop-1
revision commit and returned **`READY` (score 5/5), no blockers**. Both loop-1
blockers are resolved — ADR-006's rollback is now **preimage-backed, not
hash-inferred** (durable preimage records + absent sentinel + reverse-order
journaled restore); ADR-009 fully specifies the ownership/version manifest
(`.llm-wiki/manifest.json`, plugin-owned, written only through the ADR-006
transaction, minimum schema, first-install/missing/corrupt/user-modified policy,
default skip-and-report with explicit `--force`). All three loop-1 non-blocking
findings (ADR-007 Phase 3/7 split, ADR-006 base-hash sentinel/non-regular
handling, ADR-009 managed-asset vs repo-inventory wording) are also fixed. Two
residual non-blocking items (ADR-006 roll-forward/rollback determinism, ADR-009
`lastInstalledHash` for repo-owned assets) are schema/implementation detail, not
ADR-acceptance blockers. Verbatim artifact archived at
[`reviews/2026-07-03-codex-adr-006-007-009-rereview.md`](reviews/2026-07-03-codex-adr-006-007-009-rereview.md)
(loop-1 `NEEDS_REVISION` record preserved alongside it).

**Acceptance:** per the Phase-1-precedent gate, on `READY` the three statuses were
flipped `proposed` → **`accepted`** under the **2026-07-03 autonomous-phase
mandate** and are **flagged for Oliver's async ratification** (not a substitute
for it). `sync-adr-index` regenerated `design/adr/README.md` — ADR-001–009 now
show `accepted` (008 gap preserved).

- **[ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md)** accepted — cross-file transaction model (criteria 3, 11, 12, 13, 20).
- **[ADR-007](../design/adr/adr-007-profile-system-boundary.md)** accepted — data-driven profile boundary, one inheritance level (criteria 4, 5).
- **[ADR-009](../design/adr/adr-009-install-upgrade-uninstall-ownership.md)** accepted — install/upgrade/uninstall asset-ownership + manifest (criteria 1, 3, 20).

**Open-question movement:** with ADR-009 accepted, **Q7's install-ownership
remainder is now closed** (see [`open-questions.md`](open-questions.md)); Q7 was
already `closed (binary-selection scope only)` via ADR-002. **Signing / provenance
is *not* closed** — ADR-009 re-defers it to a dedicated supply-chain ADR, so that
residual supply-chain risk stays `open` in [`risks.md`](risks.md).

**Validation:** `check-plan --criteria-set adr --non-interactive` deterministic
**ADR-C1–C5 pass** on all three (C6 standing best-effort WARN); `sync-adr-index`
clean; `GOTOOLCHAIN=local go test ./...` PASS across 8 packages; no
`{{`/`}}`/`_TBD_`/`TODO` placeholders in the ADRs or review artifacts. No product
code written; no accepted ADR edited in place; PR not merged.

**Next step:** Oliver async-ratifies the three acceptances; the Phase 2 train
proceeds (CI test matrix, ADR-006 transaction layer, release builds + selection,
`init` core profile, `install` lifecycle, acceptance fixtures, closeout).

### 2026-07-03 — Phase 2 (Install/init) implementation train merged; phase closed out

The full Phase 2 train (issues #14–#20) merged via PRs #22–#28; the closeout
issue **#21** (this entry) refreshes [`../design/state.md`](../design/state.md)
and the knowledge layer and records validation evidence. It merged as closeout
PR **#31** (`main` = `30cfbac`), closing issue #21 and dropping the milestone
(`Phase 2 — Install/init`, #2) to zero open issues; Hermes then closed the
milestone manually (state=closed, 9 closed / 0 open, 2026-07-03). **Docs-only —
no product behavior changed.** [`../design/state.md`](../design/state.md) is the
authoritative evidence artifact; this entry is the log pointer.

**Merged train** (merge order):

| Issue | Title (abbrev.) | ADR | PR | Merge |
|---|---|---|---|---|
| #14 | Draft + accept ADR-006/007/009 | 006/007/009 | #22 | `4136e70` |
| #16 | Cross-file transaction commit on fsafe staging | 006 | #23 | `9077a6a` |
| #18 | `init` core profile + wiki bundle scaffold | 007 | #24 | `c8c40d2` |
| #19 | Install new + non-empty, `--dry-run`, refusal, version record | 009 | #25 | `d5bc4cd` |
| #15 | Cross-platform CI test matrix (5 platforms) | — | #26 | `fb65639` |
| #20 | Install/init acceptance corpus (gate evidence) | 002/003/005/006/009 | #27 | `a078007` |
| #17 | Multi-platform release bundle + selfcheck smoke | 002 | #28 | `33dd78a` |

**Gate evidence** (from [`../notes/eval-issue-020.md`](../notes/eval-issue-020.md)
+ PR #27/#28 CI; refreshed at closeout): acceptance corpus **6/6 PASS on 4 of 5
platforms** (linux-amd64, linux-arm64, macos-arm64, windows-amd64);
`cross-compile-smoke` + per-platform selfcheck smoke green; local
`go build`/`vet`/`test ./...` green on the Unix dev host (11 packages) and the
`TestAcceptance` corpus 6/6.

**Two documented caveats — deferred, not hidden, not fixed here:**

1. **windows-amd64 full `go test ./...` is RED** — pre-existing Unix
   permission-mode assertions in `internal/fsafe` + `internal/txn`
   (`mode = -rw-rw-rw-, want -rw-r-----`); predate #20, fail identically on
   `main` at `fb65639`. Follow-up issue **#29** (`bug`).
2. **macos-amd64 (`macos-13`) produced no evidence** — runner never dispatched
   (queue/availability), on branch and `main` alike; `main` tip run at `33dd78a`
   still pending with no jobs at closeout. Closed on inference (identical code
   path to green macos-arm64 + clean `darwin/amd64` cross-compile), not observed
   CI. Follow-up issue **#30** (`infra`).

**Exit-criteria verdict:** the Phase 2 gate says criteria 2/3/4(core) pass on
**all five** platforms; **4/5 were observed**, macos-amd64 by inference only.
The closeout records the gap plainly and **does not claim 5/5** — the gate call
was Hermes/Oliver's. Closeout PR #31 merged (`30cfbac`) and closed issue #21,
dropping milestone #2 to zero open issues; Hermes then closed the milestone
manually (state=closed, 9 closed / 0 open, 2026-07-03).

**Next:** Oliver async-ratifies ADR-006/007/009 (still flagged, not blocking);
**ADR-008 (provenance/citation) must be drafted + accepted before Phase 3**
authoring work; then Phase 3 issues.

### 2026-07-03 — ADR-008 (provenance & citation model) drafted, Codex-gated, accepted

Drafted [ADR-008](../design/adr/adr-008-provenance-and-citation-model.md) — the
Phase 3 provenance/citation prerequisite (issue **#32**) — on branch
`docs/adr-008-provenance-citation-model`, filling the deliberately-preserved 008
numbering gap. **Decision (Option A):** citations are **ordinary inline Markdown
links** `[label](target)` (the grammar `internal/validate/links.go` already
parses); evidentiary status comes from a **profile-designated evidence context**
(no new syntax, no frontmatter registry, no micro-syntax parser); "resolvable"
is a **total, deterministic, offline** three-class predicate (http(s) syntactic
validity / in-bundle membership via the shipped `resolveTarget` / repo-path
read-only `stat`) sharing **one resolver** with `core-broken-link`;
sourced-ness is asserted **structurally** (profile data, per addendum 003) plus
**procedurally** (adapter contract, criterion-9 fixtures) — the engine never
infers authorship; severities live inside [ADR-004](../design/adr/adr-004-validation-and-severity-model.md)'s
precedence; preservation is [ADR-001](../design/adr/adr-001-go-toolchain-and-yaml.md)/[ADR-006](../design/adr/adr-006-staged-mutation-transaction-model.md)
byte mechanics **plus a plan-time citation-loss gate ADR-008 owns**, routed
through the **existing [ADR-003](../design/adr/adr-003-json-contract-and-exit-codes.md)
`approval` member** (no envelope change). Rejected a frontmatter `citations:`
registry (B: can't express section-scoped rules, severs claim/evidence
adjacency, duplicates `resource`) and a citation micro-syntax (C: needs a
deferred parser, "citation theatre").

**Codex gate (Hermes `openai-codex` provider fallback; standalone codex CLI
unavailable):**

- **Loop 1 — `NEEDS_REVISION` (3/5)** ([archive](reviews/2026-07-03-codex-adr-008-review.md)):
  four decision-level blockers — (B1) preservation falsely attributed an
  approval-on-deletion discipline to ADR-006 (which gates only on stale hashes);
  (B2) the three resolution classes didn't partition the grammar (fragment-only,
  protocol-relative, invalid-host http(s) fell between codes); (B3)
  `profile-citation-unresolved` double-counted a condition ADR-004 models as a
  severity promotion; (B4) the ruleset tag was undecided and contradicted by the
  shipped enum (`okf`/`profile` only) and `core-broken-link`'s `RulesetProfile`
  tag. All four addressed while keeping `proposed`.
- **Loop 2 — `READY` (5/5)** ([archive](reviews/2026-07-03-codex-adr-008-rereview.md)):
  all four blockers **fixed** at the mechanism level (verified against the
  accepted ADRs + shipped `internal/contract`/`links.go`), all four non-blocking
  items folded, **no new blockers**. Two carry-ins recorded for the Phase 3/4
  implementation issues (gate class 3 on `isIntraWiki`; define whether
  present-but-unresolved satisfies a require-citation obligation).

Flipped `proposed` → **`accepted`** under the **2026-07-03 autonomous-phase
mandate**, **flagged for Oliver's async ratification** (matching ADR-006/007/009);
re-ran `sync-adr-index` (008 now `accepted` between 007 and 009). **Phase 3
authoring is unblocked.** Scope note: ADR-008 covers **content-provenance**
(citations); **supply-chain provenance** (release signing/attestation) is a
**separate deferred ADR** — not conflated — so that residual risk and the
"sourced claims lose provenance" risk stay `open` in [`risks.md`](risks.md), the
latter now with its mechanism decided. Validation: `check-plan` ADR-C1–C5 PASS
(C6 WARN); `go build`/`vet`/`test ./...` green (11 packages); placeholder scan
clean. Next: **file Phase 3 issues** (`page inspect`/`plan`/`apply` + authoring
adapter).

### 2026-06-30 — Knowledge layer reconciled to canonical file set

Aligned `knowledge/` with the project's explicit layer spec: added
[`SCHEMA.md`](SCHEMA.md) (canonical conventions) and [`index.md`](index.md)
(live front door), renamed `decisions.md` → `log.md` (this file), and
slimmed `README.md` to defer to the two. The layer now provides
`SCHEMA.md`, `index.md`, `log.md`, `project-brief.md`, `risks.md`,
`open-questions.md`, and `reviews/`. Intra-layer and addenda links were
repointed to `log.md`; the verbatim review archive under `reviews/` was
left unmodified. No `design/` artifact, PRD, or addendum content changed —
this was a curation/structure pass only. Also verified PRD provenance:
`design/prd.md`, `knowledge/sources/prd-original.md`, and the supplied
source share an identical SHA-256.
