<!--
  Template: Decisions log
  Filled by: the clarify skill, or a human
  Output in a target project: design/decisions.md
  The decisions log captures informal-but-settled context that sits
  below ADR weight: gray areas surfaced before ADR drafting, deep-dive
  resolutions, and the rationale that hardens them. The skill is
  OPT-IN — small projects can skip it and go straight from
  prd-to-mvp to adr-writer.

  Append-only by design. Re-runs of the clarify skill never delete or
  rewrite earlier entries; they add new sections or skip questions
  whose topics already appear here OR are locked by accepted ADRs.
  Marker fences (<!-- decisions:NN:start --> / <!-- decisions:NN:end -->)
  wrap each entry so editorial commentary outside the fences is
  preserved across re-runs.

  Graduate-to-ADR criterion. If a decision recorded here later turns
  out to be architectural — i.e. superseding it would require a new
  ADR — graduate it via adr-writer. The decisions log is for choices
  whose reversal would be a routine code change, not an architectural
  shift.
-->

# {{PRODUCT_NAME}} — Decisions

**Last updated:** {{YYYY-MM-DD}}
**Source PRD:** [`design/prd-normalized.md`](prd-normalized.md)
**Source MVP:** [`design/mvp.md`](mvp.md)

## Purpose

{{One paragraph: this log captures informal-but-settled context for
the project — implementation choices that downstream agents (planner,
adr-writer, executor) can rely on without re-asking. Decisions here
are below ADR weight by design. If a decision later proves
architectural, graduate it via `adr-writer` — see the criterion in
the header comment.}}

## Decisions

<!-- Each decision is its own self-contained section with marker
     fences. Append new entries below; never delete or rewrite earlier
     ones. The clarify skill skips questions whose topic already
     appears here, so re-runs are idempotent. -->

<!-- decisions:01:start -->

### D1 — {{Question or topic}}

- **Question:** {{The gray area surfaced before this was settled. One sentence.}}
- **Options weighed:**
  - {{Option A — short description.}}
  - {{Option B — short description.}}
  - {{Option C, if applicable.}}
- **Decision:** {{Which option was chosen, in one short paragraph.}}
- **Rationale:** {{Why — the constraint, principle, or trade-off that drove the choice. One short paragraph.}}

<!-- decisions:01:end -->

<!-- decisions:02:start -->

### D2 — {{Question or topic}}

- **Question:** {{...}}
- **Options weighed:**
  - {{...}}
  - {{...}}
- **Decision:** {{...}}
- **Rationale:** {{...}}

<!-- decisions:02:end -->

<!-- Append D3, D4, ... below as new gray areas are resolved. -->

## Acceptance criteria for this document

This decisions log is acceptable when it:

- captures every gray area the clarify skill surfaced and the user
  chose to resolve,
- never duplicates content already locked by an accepted ADR,
- carries enough rationale that a future reader can reconstruct the
  *why* without re-running the deep-dive,
- preserves earlier entries verbatim across re-runs.
