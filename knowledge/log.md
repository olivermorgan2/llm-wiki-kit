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
