# AI PR review

The kit can generate an AI-assisted review of a GitHub pull request and —
**only after you preview the exact comments and explicitly approve** —
publish them to the PR. It is safe by default: review generation is local
and posts nothing; publishing is a cat-3 GitHub write gated by an explicit,
deterministic approval token and made idempotent by a receipt.

This capability is governed by
ADR-051 and the
follow-on installer distribution decision in
ADR-052. The human
entry point is the `/review-pr` verb. In target projects it is installed
only when the operator passes `--with-ai-review` during install.

## What to expect

AI PR review is **optional** and **bring-your-own-provider**. Before you
turn it on:

- **You supply the credentials.** The kit ships no API key and no provider
  account. You create an account with an OpenAI-compatible provider
  (OpenRouter is the documented first choice), generate your own key, and
  the provider — not the kit — bills you for the model calls it makes.
- **Your key never enters the kit's files.** It lives only in the
  environment variable named by `apiKeyEnv`. Nothing the kit writes —
  config, artifacts, receipts, or logs — contains the key, and the tools
  never ask you to paste it into chat.
- **The model only ever sees the PR diff and metadata** you point it at,
  sent to the provider endpoint you configure. No other repo content is
  transmitted.
- **Nothing is posted without you.** Review generation is local and posts
  nothing; publishing a comment requires a separate, explicit approval
  token (see §3).
- **It is off until you opt in.** A default install never copies this
  runtime, so the criteria above only apply once you pass
  `--with-ai-review` and configure a provider yourself.

## Install mode

Default installs stay lightweight:

```bash
bin/install-workflow-kit --target /path/to/project
```

A default install does not copy the AI-review runtime and does not expose a
broken target-local `/review-pr` skill.

To install AI PR review into a target project:

```bash
bin/install-workflow-kit --target /path/to/project --with-ai-review
```

That copies the runtime under `.claude/`:

- `.claude/bin/review-pr`
- `.claude/bin/publish-review`
- `.claude/bin/review-eval`
- `.claude/bin/write-receipt`
- `.claude/bin/lib/*review*` and `.claude/bin/lib/json-envelope.sh`
- `.claude/ai-review/prompts/`
- `.claude/ai-review/eval/`
- `.claude/ai-review/config.example.json`
- `.claude/schemas/ai-review-*.yaml` as reference schemas

The canonical schema files are `schemas/ai-review-config.v1.yaml` and
`schemas/ai-review-artifact.v1.yaml` in the kit source repo; installed target
copies live under `.claude/schemas/`.

Generated artifacts and user-local config are ignored by the target
`.gitignore`:

```gitignore
.claude/ai-review/artifacts/
.claude/ai-review/config.json
```

`--force` refreshes runtime scripts, prompt packs, schemas, and example
config, but preserves any existing `.claude/ai-review/config.json`.

## The three surfaces

- `.claude/bin/review-pr`: generate a dry-run review artifact from the PR diff. Posts nothing.
- `.claude/bin/publish-review`: preview, then post only with explicit `--confirm`.
- `.claude/bin/review-eval`: offline fixture harness for review quality.

In the kit source repo, use the repo-local `bin/*` equivalents.

## 1. Configure a provider without committing secrets

The model layer is provider-agnostic: any OpenAI-compatible
chat/completions endpoint works, with **OpenRouter** as the first provider.
Configuration is a small, **secret-free** JSON file. In a target project:

```bash
cp .claude/ai-review/config.example.json .claude/ai-review/config.json
```

Example config:

```json
{
  "schema": "ai-review-config",
  "version": 1,
  "provider": "openrouter",
  "baseURL": "https://openrouter.ai/api/v1",
  "model": "openai/gpt-4o-mini",
  "apiKeyEnv": "OPENROUTER_API_KEY",
  "timeoutSeconds": 60,
  "maxDiffBytes": 200000,
  "profile": "balanced"
}
```

The config never contains the key. `apiKeyEnv` names the environment
variable that holds it. Export the key in your shell or secret manager;
never paste it into chat and never commit it:

```bash
export OPENROUTER_API_KEY=sk-or-...   # shell only; do not put this in config.json
```

Safer persistent options:

- shell profile export in `~/.zshrc` or `~/.bashrc` for local dev
- `direnv` / `.envrc` if the project uses direnv, with `.envrc` gitignored
- CI secret store if running from automation

Config resolves in this order, with later values winning: built-in defaults
→ `$KIT_ROOT/ai-review/config.json` → `--config FILE` → CLI flags
(`--provider`, `--model`, `--base-url`, `--profile`). In a target project,
`$KIT_ROOT` is `.claude`; in the kit repo it is the repo root.

If no key is set, `review-pr` exits `3` with a setup message rather than
failing obscurely — and never asks you for the key.

### Using a different provider

OpenRouter is only the documented default. Any OpenAI-compatible
`chat/completions` endpoint works — typically `baseURL`, `model`, `apiKeyEnv`,
and any provider-specific `headers` change. The `provider` label is free-form:
it only tags receipts and artifact provenance and does not switch code paths.
`review-pr` always sends a single `POST` to `<baseURL>/chat/completions` with
an `Authorization: Bearer <key>` header, so the transport is identical for
every provider.

**OpenAI directly** (the GPT / Codex models), keyed by `OPENAI_API_KEY`:

```json
{
  "schema": "ai-review-config",
  "version": 1,
  "provider": "openai",
  "baseURL": "https://api.openai.com/v1",
  "model": "gpt-4o-mini",
  "apiKeyEnv": "OPENAI_API_KEY",
  "timeoutSeconds": 60,
  "maxDiffBytes": 200000,
  "profile": "balanced"
}
```

```bash
export OPENAI_API_KEY=sk-...   # shell only; do not put this in config.json
```

**Any other OpenAI-compatible endpoint** — a self-hosted gateway or a local
server such as Ollama, vLLM, or llama.cpp — follows the same shape: point
`baseURL` at the endpoint's `…/v1` root (no trailing `/chat/completions`,
which `review-pr` appends), set `model` to whatever that endpoint expects, and
name the key's environment variable in `apiKeyEnv`. Drop the
OpenRouter-specific `headers` (`HTTP-Referer`, `X-Title`) unless your provider
asks for them. A local server that does not check keys still needs the named
variable exported to some non-empty placeholder, because `review-pr` treats an
unset key as a setup error (exit `3`).

Because the transport is fixed — a `POST` to `<baseURL>/chat/completions` with
bearer auth — endpoints that need a different request path, extra query
parameters (such as Azure OpenAI's `api-version`), or a non-bearer auth header
are not supported by the current script. The request also sets
`response_format: {"type": "json_object"}` (JSON mode) and reads the reply from
`choices[0].message.content`, so the endpoint must support that OpenAI
request/response shape.

## 2. Generate a dry-run review

```bash
.claude/bin/review-pr --pr 42 --profile balanced --format md
.claude/bin/review-pr --pr 42 --format json
```

It fetches the diff (`gh pr diff`) and metadata (`gh pr view`), calls the
model, validates/coerces the reply to the AI-review artifact shape, and
writes `.claude/ai-review/artifacts/pr-42-<hash>.json` and `.md`.
Findings are classified **blocking / non-blocking / question / praise**
with severity, category, file/line, suggestion, confidence, and a
`commentable` flag.

Each finding is `commentable` only when it is high-confidence, not praise,
and its file+line actually appear in the diff. Oversized diffs are truncated
with an explicit notice rather than failing silently.

### Review profiles

Set with `--profile` or the config:

- **strict** — everything, including style nitpicks.
- **balanced** (default) — real risk; no style nits.
- **lightweight** — blocking findings only; low noise.

The rubric and profiles are the prompt pack under `.claude/ai-review/prompts/`.

## 3. Publish — preview, approve, post

```bash
# preview the exact comments — still posts nothing
.claude/bin/publish-review --artifact .claude/ai-review/artifacts/pr-42-<hash>.json --pr 42

# post, only after you approve, with the deterministic token
.claude/bin/publish-review --artifact .claude/ai-review/artifacts/pr-42-<hash>.json \
  --pr 42 --confirm publish-pr-42
```

- **Dry-run is the default.** With no `--confirm`, it prints the top-level
  review body and every inline comment and exits — nothing is posted.
- **Explicit, deterministic approval.** Posting requires
  `--confirm publish-pr-<N>`; any other value is refused.
- **No broken inline comments.** Findings without a valid diff line are
  folded into the top-level body.
- **Idempotent.** A successful post writes a receipt under
  `.claude/receipts/`; posting the same artifact again is refused unless
  `--force`.

## 4. Evaluate review quality offline

```bash
.claude/bin/review-eval
.claude/bin/review-eval --format json
```

Runs the generator in mocked, offline mode over fixtures in
`.claude/ai-review/eval/fixtures/` and checks each artifact against its
expectations. It also verifies duplicate-comment prevention without touching
a real PR. Run it before changing the prompt pack or model.

## Safety summary

- No secret is ever committed or requested in chat. Keys live only in the
  env var named by the config.
- Dry-run before any GitHub write, always.
- Explicit approval before posting, in every operating mode.
- Receipts make publishing idempotent and auditable.

## Offline / CI use

```bash
.claude/bin/review-pr \
  --pr 1 \
  --diff .claude/ai-review/eval/fixtures/simple-bugfix/input.diff \
  --mock .claude/ai-review/eval/fixtures/simple-bugfix/mock-response.json \
  --format text
```

This path uses no network, no API key, and no GitHub write.
