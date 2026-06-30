<!--
  Template: MVP statement
  Filled by: the prd-to-mvp skill, or a human after PRD intake
  Output in a target project: design/mvp.md (recommended) or similar
  The MVP statement is a scoping document. It turns a PRD into an
  explicit yes/no list: what the first release must do, what it must
  not do, and how you'll know you've succeeded. Keep it short — one
  page is healthy.
-->

# {{PRODUCT_NAME}} — MVP

**Last updated:** {{YYYY-MM-DD}}

## Product name

{{PRODUCT_NAME}}

## One-line description

{{One sentence a stranger can understand. "A thing that does X for Y so
they can Z."}}

## Product goal

{{Two or three sentences: what success looks like for the first
release. Resist the urge to list every future feature — this is the
goal of the MVP only.}}

## Target users

### Primary user

{{Who this is for first. Be specific — "a solo technical founder" beats
"developers".}}

### Secondary user

{{Who else benefits. If there is no meaningful secondary user, delete
this section.}}

## Core problem

{{What problem does this solve that isn't solved well enough already?
Name the alternatives and why they fall short.}}

## Product principles

1. {{A principle that will resolve scope arguments later.}}
2. {{Another principle — terse and decisive.}}
3. {{3–6 principles is healthy; more than 8 is noise.}}

## MVP scope

### In scope

- {{Capability 1 the MVP must ship.}}
- {{Capability 2.}}
- {{Capability 3.}}

### Out of scope

- {{Thing that is explicitly excluded from this release.}}
- {{Another non-goal — tempting but deferred.}}
- {{Include anything users commonly assume is in scope but is not.}}

## Primary outputs

{{The concrete artifacts the MVP produces or ships — a CLI, a PR pack,
a deployed service, a set of files in a repo, etc.}}

## Success criteria

The MVP succeeds if a user can:

1. {{End-to-end outcome 1.}}
2. {{End-to-end outcome 2.}}
3. {{End-to-end outcome 3.}}

## Deferred to later

- {{Feature or capability intentionally pushed to a later phase.}}
- {{Include why it's deferred if the reason is not obvious.}}

## Acceptance criteria for this document

This MVP statement is acceptable when it:

- names a clear product and user,
- lists what is in and out of scope without ambiguity,
- and can drive the build-out plan, ADRs, and issue backlog without
  further interpretation.
