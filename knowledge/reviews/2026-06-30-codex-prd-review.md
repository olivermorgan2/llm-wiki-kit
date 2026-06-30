1. Verdict: NEEDS_REVISION

2. Summary:
- The product problem, target user, non-goals, and deterministic-engine spine are clear enough for MVP planning.
- The normalized PRD is not gate-ready as the canonical downstream input because it drops the source PRD’s explicit MVP acceptance criteria.
- Several open questions are not harmless implementation details; they affect packaging, test matrix, compatibility, CI scope, and contract design.
- The academic-research profile is in MVP scope but under-specified enough that planning would produce ambiguous issues.
- Scope is large but not impossible if the revision locks a thin MVP slice and test fixtures.

3. Blocking findings:

- Severity: High
  Evidence/path: [design/prd-normalized.md](/Users/hermes/llm-wiki-kit/design/prd-normalized.md:4) says downstream skills read the normalized PRD only; [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:596) contains the actual MVP acceptance criteria.
  Why it matters: MVP planning will miss or dilute acceptance requirements if it reads only the normalized file.
  Required fix: Add a PRD addendum carrying the full MVP acceptance criteria into the canonical planning input, or regenerate/update the normalized artifact workflow so the criteria are explicitly available to `prd-to-mvp`.

- Severity: High
  Evidence/path: [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:600), [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:613), [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:620), and open questions Q3/Q6/Q7/Q8 in [knowledge/open-questions.md](/Users/hermes/llm-wiki-kit/knowledge/open-questions.md:14).
  Why it matters: Platform packaging, minimum Claude Code version, CI template scope, and JSON contract compatibility directly determine implementation shape and acceptance tests.
  Required fix: Resolve these before MVP issue generation, or record explicit MVP planning assumptions such as “GitHub Actions only,” “minimum Claude Code version TBD but compatibility shim not in MVP,” “one binary-selection mechanism,” and “contract version starts at v1 with no backward compatibility until first release.”

- Severity: High
  Evidence/path: Academic profile is in MVP scope at [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:127), described only at [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:254), while exact templates remain open in [knowledge/open-questions.md](/Users/hermes/llm-wiki-kit/knowledge/open-questions.md:15).
  Why it matters: The initial domain value proposition depends on this profile. Without concrete page templates, required fields, citation expectations, and valid/invalid fixtures, “academic-research profile” cannot be planned or tested.
  Required fix: Define the minimum MVP research profile contract, including page types, required vs recommended fields, section rules, and at least one acceptance fixture per page type.

4. Important non-blocking findings:

- Severity: Medium
  Evidence/path: MVP includes plugin packaging, five platform binaries, install/upgrade/uninstall, custom profiles, hooks, CI, authoring, enrichment, staged mutation, and indexes at [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:624).
  Why it matters: This is a wide MVP surface. It risks producing a project plan that is technically coherent but too broad for a first shippable slice.
  Required fix: Add a “MVP slice order” section separating must-pass first release paths from follow-on hardening inside MVP.

- Severity: Medium
  Evidence/path: Success signals are listed in [design/prd-normalized.md](/Users/hermes/llm-wiki-kit/design/prd-normalized.md:138), but most have no threshold.
  Why it matters: Metrics like generated-page pass rate, false-positive rate, and diff acceptance rate are directionally useful but not decision-grade.
  Required fix: Add “measurement only” vs “release gate” labels, with thresholds only where they are true gates.

- Severity: Medium
  Evidence/path: Custom profiles are in MVP scope at [design/prd.md](/Users/hermes/llm-wiki-kit/design/prd.md:128), while third-party profile trust is open at [knowledge/open-questions.md](/Users/hermes/llm-wiki-kit/knowledge/open-questions.md:16).
  Why it matters: A local custom-profile template can ship without solving third-party registry trust. Leaving both adjacent creates unnecessary scope ambiguity.
  Required fix: Explicitly state that MVP custom profiles are local-file only, and third-party registry/trust is Phase 3 unless decided otherwise.

5. Suggested PRD addenda:

- `001-mvp-acceptance-criteria.md`
- `002-mvp-planning-assumptions.md`
- `003-academic-research-profile-contract.md`
- `004-mvp-slice-order-and-fixture-plan.md`
- `005-custom-profile-boundary.md`

6. Knowledge-layer updates:

- `knowledge/open-questions.md`: add/replace QB2 with this review outcome; mark Q3, Q4, Q6, Q7, Q8 as “must resolve or assumption-lock before MVP planning.”
- `knowledge/risks.md`: add “Canonical normalized PRD omits acceptance criteria,” “Academic profile underspecified,” and “MVP surface too wide for first release.”
- `knowledge/decisions.md`: record verdict `NEEDS_REVISION` and the required addenda before `/prd-to-mvp`.

7. What not to change:

- The core problem statement is strong.
- Target users are specific enough for MVP.
- Non-goals are clear and useful.
- The deterministic Go engine as shared authority is a good architectural constraint.
- The OKF baseline vs profile-conformance distinction is well framed.
- The safety model around staged plans, stale rejection, and bounded writes is good enough for MVP planning once acceptance criteria are carried forward.