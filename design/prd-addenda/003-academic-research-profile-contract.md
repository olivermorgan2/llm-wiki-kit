# Addendum 003 — Academic-research profile contract (minimum MVP)

- **Source finding:** Codex PRD review (2026-06-30), Blocking finding #3
  (High). The academic-research profile is in MVP scope (`design/prd.md`
  §8, §11) but described only in prose (§11.3), with exact templates left
  open (Q4). Without concrete page types, required/recommended fields,
  section rules, and valid/invalid fixtures, the profile cannot be planned
  or tested — and it is the initial domain value proposition.
- **Decision:** **Accept — define the minimum MVP contract here.** This is
  the *minimum* shippable contract, not the final schema. Exact YAML
  library, conditional-section syntax, and richer fields stay open (Q4
  remains open, now assumption-locked to this minimum). The contract below
  is data-driven (per PRD §11) and adds **one inheritance level** over
  `core`.
- **Knowledge links:** `knowledge/open-questions.md` Q4 (assumption-locked
  to this minimum); `knowledge/risks.md` ("Academic profile
  underspecified").

## Inheritance

```yaml
profile:
  id: academic-research
  version: "1.0"
  extends: core
```

Inherits all `core` rules: required `type`/`title`/`description`;
recommended `timestamp`/`tags`/`aliases`/`resource`; kebab-case filenames;
bundle-root-relative links; unknown-field preservation; core page types
`concept`, `entity`, `source`, `synthesis`. This addendum **adds** types
and **tightens** `source`/`synthesis`; it never relaxes a core rule.

## Page types (MVP)

`concept`, `entity` — inherited from `core` unchanged.
Added / tightened: `source`, `claim`, `method`, `question`, `synthesis`.

The profile must keep evidence-bearing pages (`source`, `claim`,
`synthesis`) distinct from workflow state (`question`) and derived output:
a `question` or a generated output must never silently become evidence for
another `claim` (PRD §11.3).

### `source` (tightened over core)
- **Required:** `type`, `title`, `description`, `authors` (list, ≥1),
  `source_type` (one of `paper | preprint | book | dataset | webpage |
  report | other`).
- **Recommended:** `publication_date` (ISO 8601), `doi` **or**
  `canonical_url` (at least one recommended; a `source` with neither emits
  a **suggestion**), `review_status` (one of `unreviewed | in-review |
  reviewed`), `tags`, `aliases`.
- **Sections:** none required beyond a title/description; a summary section
  is recommended.

### `claim` (new)
- **Required:** `type`, `title`, `description`, `confidence` (one of
  `low | medium | high`), `assessment` (one of `supported | contested |
  refuted | open`).
- **Recommended:** `tags`, `timestamp`.
- **Required sections:** `## Evidence`, `## Counterevidence`,
  `## Assessment`. (`Counterevidence` may be explicitly empty — e.g. "None
  found" — but the heading must be present so absence is deliberate, not
  forgotten.)
- **Provenance rule:** any evidence statement derived from an identifiable
  source must cite a resolvable URL, repo path, or OKF `source` document
  (inherits FR5). A `claim` whose `assessment` is `supported` with zero
  citations in `## Evidence` is an **error**.

### `method` (new)
- **Required:** `type`, `title`, `description`.
- **Recommended:** `tags`, `timestamp`.
- **Required sections:** `## Assumptions`, `## Strengths`,
  `## Limitations`.

### `question` (new — workflow state, not evidence)
- **Required:** `type`, `title`, `description`, `status` (one of
  `open | answered | abandoned`).
- **Recommended:** `tags`, `timestamp`.
- **Required sections:** `## Evidence gap`.
- **Rule:** a `question` page is never a valid citation target for a
  `claim`'s evidence (enforced as an error if referenced as evidence).

### `synthesis` (tightened over core)
- **Required:** `type`, `title`, `description`.
- **Required sections:** `## Scope`, `## Findings`, `## Agreement`,
  `## Disagreement`, `## Evidence gaps`.
- **Recommended:** links to the `claim`/`source` pages it synthesizes.

## Severity defaults (this profile)

Inherits core defaults (invalid YAML / missing required field / wrong field
type = **error**; broken links / stale index = **warning**). Profile
additions: missing required section on a profiled type = **error**;
`source` with neither `doi` nor `canonical_url` = **suggestion**;
`claim` evidence without citation when `assessment: supported` = **error**.

## Acceptance fixtures (≥1 valid + ≥1 invalid per added/tightened type)

Ship under `profiles/academic-research/examples/{valid,invalid}/`. Each
invalid fixture targets exactly one rule so the failing finding is
unambiguous. These fixtures satisfy acceptance criteria 5, 7, 8, and 9 for
the academic-research profile.

| Type | Valid fixture proves | Invalid fixture proves (expected finding) |
|---|---|---|
| `source` | full required set + DOI resolves | missing `authors` → error; (separate) no DOI/URL → suggestion |
| `claim` | required fields + all 3 sections + cited evidence | `assessment: supported` with no citation in `## Evidence` → error |
| `method` | required fields + all 3 sections | missing `## Limitations` → error |
| `question` | required fields + `## Evidence gap`, `status: open` | a `claim` citing a `question` as evidence → error |
| `synthesis` | all 5 sections present | missing `## Disagreement` → error |

**Assumption lock (Q4):** the field lists, section names, and enum values
above are the **MVP minimum**. Conditional-section syntax and any richer
field set are deferred; if Oliver later specifies different templates, this
contract is superseded by a new addendum and the fixtures are regenerated —
the page-type *set* (`source`/`claim`/`method`/`question`/`synthesis`) is
the stable part planning can rely on.
