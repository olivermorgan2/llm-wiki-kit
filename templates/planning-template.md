<!--
  Template: Planning artefact
  Filled by: the planning skill, or a human
  Output in a target project: design/planning.md
  The planning doc sits between design/mvp.md and the first ADR. It
  captures requirements decomposition, risks, assumptions, sequencing
  rationale, and open research questions. The doc is OPT-IN — small
  projects can skip it and go straight from MVP to adr-writer.
  Larger projects use it to harden ambiguity into decisions before
  ADRs are drafted.

  Marker fences (`<!-- planning:start -->` / `<!-- planning:end -->`)
  wrap the auto-generated body of each section. Editorial text outside
  the fences is preserved across re-runs of the planning skill.
-->

# {{PRODUCT_NAME}} — Planning

**Last updated:** {{YYYY-MM-DD}}
**Source PRD:** [`design/prd-normalized.md`](prd-normalized.md)
**Source MVP:** [`design/mvp.md`](mvp.md)

## Objective

{{One paragraph: what this planning doc covers and how it relates to
the MVP. The planning doc is for "things we need to settle before
drafting ADRs" — decomposition, risk, sequencing rationale, open
questions. Anything that hardens into an architectural decision
graduates to an ADR via `adr-writer`.}}

## Requirements decomposition

<!-- planning:requirements:start -->

{{Break the in-scope MVP capabilities into concrete requirements. Each
requirement should be small enough to drive one or two GitHub issues.
Group by capability. Use stable IDs (e.g. R1, R2, R3) so downstream
ADRs and issues can reference them.}}

### {{Capability A}}

- **R1 — {{requirement}}.** {{One-line description.}}
- **R2 — {{requirement}}.** {{One-line description.}}

### {{Capability B}}

- **R3 — {{requirement}}.** {{One-line description.}}

<!-- planning:requirements:end -->

## Risks

<!-- planning:risks:start -->

{{Risks that could derail the MVP. Each risk has a one-line
description, an impact estimate, and a mitigation. Keep this honest;
risks that turn out not to matter are easy to drop later.}}

### Risk 1 — {{NAME}}

- **Impact:** {{low | medium | high}}
- **Likelihood:** {{low | medium | high}}
- **Mitigation:** {{one line}}

### Risk 2 — {{NAME}}

- **Impact:** {{...}}
- **Likelihood:** {{...}}
- **Mitigation:** {{...}}

<!-- planning:risks:end -->

## Assumptions

<!-- planning:assumptions:start -->

{{Things this plan depends on. Tools, access, prior decisions,
external services, team availability. If an assumption is wrong, the
plan needs revising — record assumptions explicitly so it is obvious
when one breaks.}}

- **A1 — {{assumption}}.** {{If this is wrong: ...}}
- **A2 — {{assumption}}.** {{If this is wrong: ...}}
- **A3 — {{assumption}}.** {{If this is wrong: ...}}

<!-- planning:assumptions:end -->

## Sequencing rationale

<!-- planning:sequencing:start -->

{{Why the build-out plan's phases are in the order they are. This is
the section `issue-planner` reads to decide phase ordering of the
issue backlog. Keep it tight: one short paragraph per phase
explaining why it has to come before the next.}}

### Phase 1 first because {{...}}

{{One paragraph.}}

### Phase 2 follows because {{...}}

{{One paragraph.}}

### Phase 3 follows because {{...}}

{{One paragraph.}}

<!-- planning:sequencing:end -->

## Open research questions

<!-- planning:research:start -->

{{Questions that need answering before specific ADRs can be drafted.
Each question has an owner and a target answer date. Once answered,
the resolution either graduates into an ADR (architectural) or into
`design/decisions.md` (informal — see ADR-033 once shipped).}}

- **Q1 — {{question}}.** Owner: {{NAME}}. Target: {{YYYY-MM-DD}}.
- **Q2 — {{question}}.** Owner: {{NAME}}. Target: {{YYYY-MM-DD}}.
- **Q3 — {{question}}.** Owner: {{NAME}}. Target: {{YYYY-MM-DD}}.

<!-- planning:research:end -->

## Decisions needing ADRs

{{List the architectural questions that this planning round has
surfaced and that need formal ADRs. Hand this list to `adr-writer` as
its input batch. Questions still open (Q1–Qn above) are NOT in this
list — only questions whose answers will become decisions.}}

- {{Decision topic 1}}
- {{Decision topic 2}}
- {{Decision topic 3}}

## Acceptance criteria for this document

This planning doc is acceptable when it:

- decomposes every in-scope MVP capability into requirements,
- names risks with explicit mitigations,
- records assumptions whose breakage would invalidate the plan,
- justifies the build-out plan's phase ordering, and
- lists the decision topics that should next be drafted as ADRs.
