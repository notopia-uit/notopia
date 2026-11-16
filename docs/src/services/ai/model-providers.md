---
authorship:
  - ai
---

# Model providers: OpenRouter and OpenCode

Notopia already uses `@ai-sdk/openai-compatible` to call a model over HTTP. Switching a
compatible provider requires configuration, not a new chat implementation.

Three values select the model, in `apps/ai/.env.local`:

```dotenv
NOTOPIA_AI_LLM_BASE_URL=https://provider.example/v1
NOTOPIA_AI_LLM_API_KEY=replace-with-your-provider-key
NOTOPIA_AI_LLM_MODEL=replace-with-a-provider-model-id
```

The base URL is the API prefix, without `/chat/completions`. The service calls the provider's
Chat Completions API and streams tool calls. Choose a model supporting that API, streaming
and tools. A Responses-only or Anthropic-only endpoint is not interchangeable with this adapter.
Keep keys server-side in the ignored local file, and restart `nx run ai:run` after changes.

## What the names mean

| Name                            | Role in this setup                                               |
| ------------------------------- | ---------------------------------------------------------------- |
| Notopia AI service              | Runs Notopia's chat/tool loop, permissions and credit accounting |
| OpenRouter                      | Hosted gateway offering many models through one API              |
| OpenCode Console (formerly Zen) | Hosted model API that can also supply models to Notopia          |
| OpenCode coding agent           | Separate development tool for working on this repository         |

Using OpenCode's model API does not run the coding agent inside Notopia. Conversely, using
OpenCode to develop the repository does not configure the application's model credentials.

## Option 1: OpenRouter

1. Create an [OpenRouter API key](https://openrouter.ai/settings/keys).
2. Choose a model from the [catalog](https://openrouter.ai/models) that supports tool calling.
3. Create or update `apps/ai/.env.local`:

```dotenv
NOTOPIA_AI_LLM_BASE_URL=https://openrouter.ai/api/v1
NOTOPIA_AI_LLM_API_KEY=replace-with-your-openrouter-key
NOTOPIA_AI_LLM_MODEL=openrouter/free
```

`openrouter/free` routes among available free models, accounting for the request's supported
features. It is useful for trying the integration, but model availability, rate limits and
behavior can vary. For repeatable editing tests, choose a specific tool-capable model ID
from the catalog instead. Free usage is still subject to OpenRouter's account/provider limits.

No OpenRouter SDK or special provider headers are required by the current service. The
existing adapter sends the API key as Bearer authentication.

Official references:

- [OpenRouter quickstart](https://openrouter.ai/docs/quickstart)
- [Free model router](https://openrouter.ai/docs/guides/routing/routers/free-router)

## Option 2: OpenCode's hosted model API

OpenCode's hosted offering was called **Zen**; current documentation describes it as
**OpenCode Console**. Use its inference API directly, without starting an OpenCode server.

1. Sign in to the [OpenCode Console](https://opencode.ai/console), configure account access
   and create an API key.
2. Choose an available model that supports Chat Completions and tools from the
   [model documentation](https://opencode.ai/docs/console/models/).
3. Configure `apps/ai/.env.local`, for example:

```dotenv
NOTOPIA_AI_LLM_BASE_URL=https://opencode.ai/inference/openai/v1
NOTOPIA_AI_LLM_API_KEY=replace-with-your-opencode-console-key
NOTOPIA_AI_LLM_MODEL=glm-5.3
```

This uses a paid-model example from the current catalog; availability and account access
may change. Use the provider's bare model ID (`glm-5.3`), not the coding agent's
`opencode/glm-5.3` selection string.

Current OpenCode documentation restricts free-tier chat models to the OpenCode application;
other clients can receive `403` even with a key. Notopia is another client, so use a paid
model or eligible Go subscription access rather than assuming an OpenCode free model will work.

Older examples use `https://opencode.ai/zen/v1`; use the currently documented inference base
URL for new configurations. This service does not consume the OpenCode session/server API,
and its own `/ai` endpoint is not an OpenAI-compatible provider endpoint for OpenCode.

## Optional: use the OpenCode coding agent with OpenRouter

This configures your development tool, independently of the Notopia runtime:

1. Start OpenCode in the repository.
2. Run `/connect`, select OpenRouter and enter your OpenRouter key.
3. Run `/models` and choose a model available to that account.

OpenCode model selection adds a provider prefix to the full provider model ID. For example,
the following optional `opencode.json` setting selects the OpenRouter free router:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "openrouter/openrouter/free"
}
```

The double `openrouter` is intentional: the first identifies OpenCode's provider; the
remaining `openrouter/free` is OpenRouter's model ID. In Notopia's `.env.local`, use only
`openrouter/free`. OpenCode can also read `OPENROUTER_API_KEY` from its process environment;
Notopia instead reads `NOTOPIA_AI_LLM_API_KEY`. Neither automatically reads the other's login.

See [OpenCode provider setup](https://opencode.ai/docs/providers/#openrouter).

## Option 3: local Ollama

The checked-in `apps/ai/.env` defaults to Ollama:

```dotenv
NOTOPIA_AI_LLM_BASE_URL=http://localhost:11434/v1
NOTOPIA_AI_LLM_API_KEY=ollama
NOTOPIA_AI_LLM_MODEL=qwen3:8b
```

If Ollama is not already running on the host, use the optional Compose profile:

```sh
docker compose -f deploy/compose/compose.yaml --profile ai-local-llm up -d --wait ai_ollama
docker compose -f deploy/compose/compose.yaml --profile ai-local-llm exec ai_ollama ollama pull qwen3:8b
```

The container runs on CPU by default and stores models in a named volume. Do not start it
on port `11434` if a host Ollama process already uses that port. Remove or update conflicting
provider overrides in `.env.local` when switching back to local inference.

## Verify the complete integration

Provider configuration only supplies model access. Follow [local setup](./index#local-setup)
to start Redis, the application services and web, and use a workspace with AI credits.

Verify all three behaviors:

1. A simple prompt streams a text response.
2. A note question invokes search/read tools and returns authorized workspace content.
3. An editing request opens a note, applies suggestions in the live editor, and supports
   Accept/Discard followed by a separate Commit.

A successful text response does not establish tool compatibility. Likewise, provider
credits and Notopia workspace AI credits are independent checks. These examples document
the integration configuration; they do not assert a live provider smoke test has passed.
