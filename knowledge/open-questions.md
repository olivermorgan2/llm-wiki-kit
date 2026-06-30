# Open questions — llm-wiki-kit

Unresolved product/technical decisions. Seeded from PRD §19 "Remaining
product decisions" ([`design/prd.md`](../design/prd.md)). When one is
resolved, record the resolution in [`decisions.md`](decisions.md) (and an
ADR in `design/adr/` if architectural), then mark it `closed` here.

Owner is **Oliver** unless noted.

| # | Question | Owner | Status |
|---|---|---|---|
| Q1 | Plugin name, command namespace, license, and marketplace location. (Repo/working name is `llm-wiki-kit`; CLI is `llm-wiki`. License chosen for the repo scaffold = MIT; "license" in PRD §19 also covers the distributed *plugin*'s license/marketplace listing — confirm they're the same.) | Oliver | open |
| Q2 | Exact Go YAML library and supported Go version for development. | Oliver | open |
| Q3 | Is GitHub Actions the only MVP CI template, or one of several? | Oliver | open |
| Q4 | Exact research-profile templates and conditional-section syntax. | Oliver | open |
| Q5 | Profile registry and trust model for third-party profiles. | Oliver | open |
| Q6 | Minimum supported Claude Code version. | Oliver | open |
| Q7 | Exact packaging mechanism for selecting the correct platform binary inside the plugin. | Oliver | open |
| Q8 | JSON contract versioning and compatibility policy. | Oliver | open |

## Questions raised during bootstrap

| # | Question | Owner | Status |
|---|---|---|---|
| QB1 | Final product/plugin name vs working repo name `llm-wiki-kit` — lock before issue/PR phases to avoid a later rename. | Oliver | open |
| QB2 | Codex adversarial PRD review findings (to be appended once review runs). | Oliver | open |
