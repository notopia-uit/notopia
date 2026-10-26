# Local AI testing

The application services run on the host through Nx. Compose runs their infrastructure.
The AI backend needs note and authorization gRPC, billing gRPC, Meilisearch, a model server,
and its own Redis. Browser testing also needs web, Traefik and Authentik; document and
search-worker provide live editor content and committed search results.

From the repository root, start the normal infrastructure, including the new AI Redis:

```sh
docker compose -f deploy/compose/compose.yaml up -d
```

To start only the new Redis when the rest is already running:

```sh
docker compose -f deploy/compose/compose.yaml up -d --wait ai_redis
```

`apps/ai/.env` already uses `redis://127.0.0.1:6376`. Redis stores monthly credit counters
and throttling state with AOF persistence and `noeviction`. A named volume preserves data
across container recreation. `AI_REDIS_PORT` can override the published port; if changed,
update `NOTOPIA_AI_REDIS_URL` in your local AI environment too.

## Optional local model server

If Ollama is already running on the host at port 11434, skip this step. Otherwise:

```sh
docker compose -f deploy/compose/compose.yaml --profile ai-local-llm up -d --wait ai_ollama
docker compose -f deploy/compose/compose.yaml --profile ai-local-llm exec ai_ollama ollama pull qwen3:8b
```

This matches the AI app's existing model and `http://localhost:11434/v1` base URL.
The container uses CPU by default and stores downloaded models in a named volume. The
model pull happens only when you run the second command. `AI_OLLAMA_IMAGE` can pin an image
version; `AI_OLLAMA_PORT` can change the host port, with a matching AI base URL update.
A different OpenAI-compatible model endpoint can be configured through the AI app's
`NOTOPIA_AI_LLM_*` environment variables instead.

## Application and billing prerequisites

Start note, authorization, billing, document, search-worker, web and AI using their normal
Nx development targets. AI listens on 8085; its billing gRPC endpoint is localhost:18084.
The health endpoint is `http://api.notopia.localhost/ai/health`.

Use a workspace with AI credits: Free defaults to zero, while Plus/Pro provide credits.
There is no need to make a real payment to test: the existing `just fixtures-billing`
workflow creates a development Pro subscription with fake provider IDs for workspace
`00000000-0000-4000-8000-000000000110`. Follow the billing fixture prerequisites and use a
user with membership in that workspace. Fake provider IDs support entitlement checks, not
real Stripe checkout/refund actions.

Open the workspace assistant, ask it to find/read notes, then test an editor AI command.
Commit edited notes to make their latest content available to search and the assistant.
Automatic commits may also do this when enabled by the workspace plan.

AI currently charges one credit per started generation, including server tool steps, and
resets usage on UTC calendar months. Provider errors and cancellation after start retain
the credit. Missing Redis or billing causes a 503; a workspace without credits gets a 402.

## Automated Redis tests

Use a separate test Redis rather than the Redis holding your manual-test credit counters.
The integration tests take `AI_TEST_REDIS_URL` and use a random namespace:

```sh
AI_TEST_REDIS_URL=redis://127.0.0.1:16385 nx test ai --skip-nx-cache
```

Without that variable, the real-Redis cases are skipped; the remaining tests still run.
