# Addendum 001 — MVP acceptance criteria (carried into canonical planning input)

- **Source finding:** Codex PRD review (2026-06-30), Blocking finding #1
  (High). `design/prd-normalized.md` declares itself the only file
  downstream skills read, but the 11-field normalized form has no field
  that carries the source PRD's MVP acceptance criteria (`design/prd.md`
  §17). The criteria were dropped from the canonical downstream input.
- **Decision:** **Accept.** Carry the full §17 acceptance criteria forward
  here so `/prd-to-mvp` has them as a first-class planning input without
  re-reading `design/prd.md`. This addendum is authoritative alongside
  `prd-normalized.md`; the normalized file is not regenerated.
- **Knowledge links:** `knowledge/log.md` (verdict entry);
  `knowledge/risks.md` ("Canonical normalized PRD omits acceptance
  criteria"); dogfood note in `notes/workflow-kit-notes.md`.

## Precise change to the normalized PRD's fields

Treat the list below as an additional, twelfth normalized field —
**"MVP acceptance criteria"** — appended to `prd-normalized.md` §10
(Success signals). It is copied verbatim in intent from `design/prd.md`
§17; no criterion is invented or relaxed.

The MVP is ready when:

1. The plugin installs and updates through a documented Claude Code plugin
   distribution flow.
2. The correct checksum-verified Go CLI runs on every supported platform
   without runtime setup.
3. Installation succeeds in new and non-empty test repositories without
   losing existing files.
4. Users can initialize with core, academic-research, or a valid custom
   profile.
5. Generated documents report separate OKF and profile conformance.
6. Unknown frontmatter fields survive authoring and enrichment.
7. Malformed YAML and missing profile-required fields fail validation.
8. Broken links are reported at the configured severity.
9. Model-generated sourced claims contain resolvable citations in
   acceptance fixtures.
10. Repeating unchanged authoring inputs does not create a duplicate page.
11. Existing-page edits are previewed before application.
12. Skills perform managed existing-page writes through the engine's
    staged plan/apply workflow.
13. A changed target causes a previously generated plan to be rejected
    without mutation.
14. Structured engine output conforms to its documented contract and exit
    codes.
15. Skills, hooks, CI, and direct CLI validation produce the same findings
    for the same repository state.
16. Renames, deletions, and migrations require explicit approval.
17. Pre-write guards reject tested path traversal and symlink escapes.
18. Post-write validation is described as feedback, not prevention.
19. The CI workflow fails on configured errors.
20. Upgrade and uninstall preserve wiki content and local profile
    extensions.
21. Automated tests pass on all supported operating systems and
    architectures.

## Success-signal gate labeling

(Resolves Codex non-blocking finding: success signals listed in
`prd-normalized.md` §10 mostly have no threshold and no gate/measurement
distinction.)

Each success signal is labeled **release gate** (must hold to ship MVP) or
**measurement only** (tracked, not a ship blocker). Thresholds are stated
only where a true gate exists; where a number is not yet decided it is an
explicit planning assumption, not a fabricated target.

| Signal | Label | Threshold |
|---|---|---|
| Confirmed writes outside configured boundaries | release gate | **zero** (PRD §15: target zero; ties to acceptance criterion 17) |
| Generated pages passing validation without manual structural repair | measurement only | no numeric gate for MVP; acceptance fixtures (criterion 9) provide the pass/fail gate instead |
| Sourced model-generated claims with resolvable citations | release gate (fixtures) | 100% **within acceptance fixtures** (criterion 9); fleet-wide rate is measurement only |
| Validator false-positive rate | measurement only | no MVP threshold (instrument first, set a gate post-MVP) |
| Broken-link findings per 100 pages | measurement only | no MVP threshold |
| Authoring diffs accepted without substantive correction | measurement only | no MVP threshold |
| Successful upgrade rate | measurement only, backed by gate | no rate threshold; acceptance criterion 20 (upgrade preserves content) is the hard gate |
| Interrupted operations recovered without corruption | release gate | acceptance criteria 3 & 20 + §14 reliability: no partial/corrupt writes |
| Median time install → first conformant page | measurement only | PRD §14 usability **target** < 15 min for a proficient user; treated as a target, not a ship blocker |

**Assumption lock:** for MVP, the only hard quantitative release gate is
"zero out-of-boundary writes"; all other quantitative signals are
measurement-only until a baseline exists. The functional acceptance
criteria above (1–21) are the binding ship gate.
