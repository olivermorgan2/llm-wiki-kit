<!--
  Template: Product Requirements Document (PRD)
  Filled by:  the user (by hand) or by an external LLM given this
              file as the shape to follow. Tool-agnostic — any
              external LLM, or no LLM at all, works equally well.
  Output in a target project: design/prd.md
  Consumer:   the `prd-normalizer` skill.

  The PRD is the starting artifact of the workflow. Once filled, hand
  it to `prd-normalizer`, which maps it onto the canonical 11-field
  Normalized PRD at design/prd-normalized.md. If this template is
  filled faithfully, the normalizer has almost nothing to do — it
  just confirms the two identity fields and writes the file through.

  Hard-required fields must not be left blank or [TBD]:
    Product name, Problem, Primary user, at least one core capability,
    at least one user story.

  Soft fields may be left as [TBD] if you genuinely do not know yet;
  downstream skills expect and handle them.

  See ADR-027 for the decision record behind this template.
-->

# {{PRODUCT_NAME}} — PRD

**Last updated:** {{YYYY-MM-DD}}

## Product name

{{PRODUCT_NAME}}

## One-line description

{{One sentence a stranger can understand. "A thing that does X for Y so
they can Z."}}

## Problem

{{One or two sentences. What problem does this solve? Name the current
alternative (manual work, another tool, doing without) and why it falls
short.}}

## Target users

### Primary user

{{Who this is for first. Be specific — "a solo maintainer of an
open-source CLI" beats "developers".}}

### Secondary user

{{Who else benefits. Delete this subsection if there is no meaningful
secondary user.}}

## Goal

{{Two or three sentences on the intended outcome. What does success look
like at the product level? Not a feature list — the outcome the features
enable.}}

## User stories / scenarios

{{2–5 concrete "user does X so that Y" stories. Each story should be
something a user actually does end-to-end, not a feature fragment.}}

1. {{As a {{user}}, I {{action}} so that {{outcome}}.}}
2. {{As a {{user}}, I {{action}} so that {{outcome}}.}}
3. {{As a {{user}}, I {{action}} so that {{outcome}}.}}

## Core capabilities

{{The capabilities the product must have. Un-prioritized — scoping
the first release is the MVP's job, not the PRD's. Be concrete enough
that a reader can picture each one.}}

- {{Capability 1.}}
- {{Capability 2.}}
- {{Capability 3.}}

## Non-goals

{{Product-level things this product will not do — ever, or at least not
in its current framing. Keep MVP-scope decisions (things you might ship
later) out of here; those belong in the MVP statement.}}

- {{Thing this product is not.}}
- {{Another explicit non-goal.}}

## Constraints and preferences

{{Technical constraints, platform preferences, known stack decisions,
timing constraints, deployment preferences, or compliance requirements.
This is the primary hint source for the `adr-writer` skill — be as
specific as you can.}}

- {{Constraint or preference 1.}}
- {{Constraint or preference 2.}}

## Success signals

{{How will you know the product is working? Qualitative signals are
fine ("I stop dreading release days") alongside quantitative ones
("release notes take under 10 minutes instead of 30+").}}

- {{Signal 1.}}
- {{Signal 2.}}

## Open questions

{{What you do not yet know. Leaving this non-empty is expected and
healthy — an empty list usually means the questions have not been
surfaced yet, not that they have all been answered.}}

- {{Question 1.}}
- {{Question 2.}}
