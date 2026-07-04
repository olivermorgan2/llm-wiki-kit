# academic-research invalid fixtures

Each fixture is an otherwise-complete page with exactly one deliberate defect, so
validated in isolation under the academic-research profile it trips exactly one
finding (per the addendum-003 fixtures table). Asserted by
`TestAcademicInvalidExamplesFailExactlyOneRule`.

The first six rows are the addendum-003 fixture table (≥1 invalid per
added/tightened type); the enum and list-min rows additionally fixture-verify
those rule kinds against the shipped profile data.

| Fixture | Trips |
|---|---|
| `source-missing-authors.md` | `profile-required-field` (authors) |
| `source-no-doi-or-url.md` | `profile-recommended-pair` (doi \| canonical_url) — suggestion |
| `claim-supported-no-citation.md` | `profile-citation-required` |
| `method-missing-limitations.md` | `profile-required-section` (## Limitations) |
| `claim-cites-question.md` | `profile-citation-target-type` (cites a question; validated with a resolvable `question.md` companion) |
| `synthesis-missing-disagreement.md` | `profile-required-section` (## Disagreement) |
| `source-bad-source-type.md` | `profile-field-enum` (source_type) |
| `source-empty-authors.md` | `profile-list-min` (authors present but empty) |
| `claim-bad-confidence.md` | `profile-field-enum` (confidence) |
