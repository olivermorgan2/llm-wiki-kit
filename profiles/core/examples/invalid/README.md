# Invalid core-profile examples

Each fixture here is a fully valid core-profile page with **exactly one
mutation** that trips **exactly one** validation rule. Validated on its own (a
single-file bundle) by `internal/validate/examples_fixtures_test.go`
(`TestInvalidExamplesFailExactlyOneRule`), each yields a single finding whose
`Code` is listed below. This single-rule isolation is what makes the negative
half of the Phase 1 acceptance corpus (ADR-004) machine-checkable.

The table below mirrors `invalidFixtureCodes` in the test, which is the source of
truth; `TestInvalidExamplesDirMatchesTable` fails if the two drift apart.

| Fixture | Mutation | Expected `Code` | Ruleset / Severity | Criterion |
|---|---|---|---|---|
| `malformed-yaml.md` | `title` becomes an unterminated YAML flow mapping | `okf-yaml-parse` | okf / error | 7 |
| `missing-type.md` | omit `type` | `okf-type-present` | okf / error | 5, 7 |
| `missing-title.md` | omit `title` | `core-required-title` | profile / error | 5, 7 |
| `missing-description.md` | omit `description` | `core-required-description` | profile / error | 7 |
| `wrong-field-type.md` | `title` is a sequence, not a string | `core-field-type` | profile / error | 7 |
| `broken-link.md` | body links to an absent `missing.md` | `core-broken-link` | profile / warning | 8 |
| `Not_Kebab.md` | non-kebab filename (contents fully valid) | `core-kebab-filename` | profile / warning | — |

Notes:

- A parse failure short-circuits: `malformed-yaml.md` emits only
  `okf-yaml-parse` and runs no further rule.
- `broken-link.md` resolves as broken only because it is validated as a
  single-file bundle — its link target is deliberately absent.
- `Not_Kebab.md` exercises the filename rule, which has no acceptance criterion
  of its own but rounds out single-rule coverage.
