# Project brief — llm-wiki-kit

Distilled from the PRD ([`design/prd.md`](../design/prd.md) §1–§5) and the
normalized PRD ([`design/prd-normalized.md`](../design/prd-normalized.md)).
Curated summary — see those files for the authoritative detail.

## What it is

A versioned **Claude Code plugin** for creating and maintaining portable,
repository-native knowledge bundles based on the **Open Knowledge Format
(OKF v0.1)**. It combines:

- Claude Code **skills** as the user-facing workflow and semantic
  reasoning layer.
- A bundled, self-contained **Go engine** (`llm-wiki` CLI) that is the
  deterministic authority for scaffolding, validation, indexing, profile
  management, diff planning, and safe filesystem writes.
- Optional Claude Code + Git **hooks** for local guardrails.
- Ready-to-use **CI** as the authoritative team enforcement mechanism.
- Versioned **profiles** defining domain page types, templates, and
  validation policy.

Core content stays plain Markdown + YAML and remains readable, editable,
and portable without Claude Code or the kit.

## Problem

Agent-system teams scatter knowledge across docs, papers, code, schemas,
notes, and memory; it's hard for agents to consume because it lacks stable
structure, explicit relationships, and predictable provenance. OKF is a
minimal interchange format only — no opinionated workflow, domain
templates, deterministic validation beyond base conformance, or repo
maintenance automation. Teams otherwise rebuild these per wiki.

## Vision

Create a useful LLM wiki quickly, keep its structure trustworthy as it
grows, and adapt it to a domain without sacrificing OKF portability.
Claude orchestrates; the Go engine stays deterministic; the repository is
the source of truth.

## Users

- **Primary:** technically proficient researcher / team already on Claude
  Code + Git wanting durable, inspectable, agent-readable knowledge.
- **Initial domain:** academic / scientific literature researcher
  (sources, concepts, claims, methods, entities, syntheses w/ provenance).
- **Secondary:** AI engineers maintaining agent context; software/data/
  analytics/ops teams; technical writers & OSS maintainers; teams defining
  custom profiles.

## Jobs to be done

1. Starting a wiki → valid structure, default profile, templates,
   instructions, and validation without runtime setup.
2. Authoring → create/enrich correctly structured pages while preserving
   existing human work.
3. Using sources → keep a traceable distinction between evidence,
   inference, and unsupported statements.
4. Editing directly → the same validation contract as model-assisted flows.
5. Stronger domain conventions → custom profile without recompiling the CLI.
6. Kit evolution → upgrade without silently overwriting customizations.

## Architectural spine (load-bearing principles)

- **Format, not platform** — no proprietary runtime to read content.
- **OKF baseline + explicit profiles** — base vs profile conformance are
  reported separately; never present a profile rule as a universal OKF rule.
- **Deterministic enforcement** — parsing/validation/indexing/FS safety
  live in Go; model judgment is never described as deterministic validation.
- **One deterministic implementation** — skills, hooks, CI, direct CLI all
  call the same engine; skill scripts are thin adapters only.
- **Review proportional to risk** — new pages fast; existing-page edits
  previewed; renames/deletes/migrations require explicit approval.
- **Provenance without citation theatre** — sourced model claims cite
  resolvable sources; human statements need not unless the profile says so.

## Status

Bootstrapped 2026-06-30 from `claude-workflow-kit` v5.0.1. Phase: PRD
normalization complete; next gate is adversarial Codex review of
`design/prd-normalized.md` (see [`decisions.md`](decisions.md)).
