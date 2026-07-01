# Valid core-profile examples

Complete, conformant core-profile pages. Validated together as one bundle by
`internal/validate/examples_fixtures_test.go`
(`TestValidExamplesHaveNoErrorFindings`), they yield **no findings** at all —
this is the positive half of the Phase 1 acceptance corpus (ADR-004).

Each page carries every required field (`type`, `title`, `description`) and every
recommended field (`timestamp`, `tags`, `aliases`, `resource`), uses a
kebab-case filename, and links only to targets that exist in the bundle.

| Page | Proves | Criterion |
|---|---|---|
| `alpha-page.md` | A complete page with a resolving intra-wiki link raises nothing. | 5, 7 |
| `beta-page.md` | Two pages cross-link inside the bundle without a broken-link warning. | 5, 7, 8 |

The bundle is loaded whole (both pages together) precisely so the cross-links
resolve; the test excludes this `README.md`, which is documentation rather than
an OKF page.
