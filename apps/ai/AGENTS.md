# AI Service (nx: `ai`)

Stateless BlockNote AI backend. NestJS HTTP in (no Kafka, no DB); it is a
gRPC *client* of the note service and an HTTP client of Meilisearch for custom
tools. Receives `{ messages, toolDefinitions }` from the web proxy, streams
back an AI SDK v6 UI-message stream via an OpenAI-compatible LLM.

## Modules (`src/`)

| Module | Note |
| --- | --- |
| `ai/` | `AiController` (`POST /ai`), `AiService` (streamText + BlockNote + custom tools), `tools/` (see `HOW_TO_ADD_TOOLS.md`) |
| `note/` | gRPC client for the note service (same pattern as `apps/document/src/note/`) |
| `authorization/` | gRPC client for the authorization service |
| `search/` | `SearchService` — Meilisearch HTTP client (`notes` index) |
| `common/` | `HttpUserGuard` + `@ReqUser()` — reads Traefik-forwarded `x-forwarded-*` headers |
| `config.ts`, `config.factory.ts`, `env.validation.ts` | Zod config (`NOTOPIA_AI_*`) |

## Request flow

Browser (`DefaultChatTransport`, Bearer from better-auth) → `apps/web` `POST /api/ai`
(proxy, forwards `Authorization`) → Traefik `jwt` middleware (validates, injects
`x-forwarded-*`) → `POST /ai` (guarded) → OpenAI-compatible LLM → SSE stream back.

`GET /ai/health` is unauthenticated (Traefik bypass, like other services).

## Things worth knowing

- `toolDefinitionsToToolSet(body.toolDefinitions ?? {})` — the client may omit
  tool definitions; the helper throws on `undefined`.
- `AiService` returns `result.toUIMessageStream()`; the controller pipes it with
  the official `pipeUIMessageStreamToResponse()` helper ([Nest cookbook](https://raw.githubusercontent.com/vercel/ai/refs/heads/main/content/cookbook/15-api-servers/50-nest.mdx)).
  `await` only resolves headers — chunks stream through as they arrive.
- LLM is configured via `NOTOPIA_AI_LLM_{BASE_URL,API_KEY,MODEL}` — any
  OpenAI-compatible endpoint (Ollama, vLLM, OpenAI itself).
- `@ai-sdk/openai-compatible` stays on **v2**: v3 targets the v4 model spec
  (needs `ai` v7) but `@blocknote/xl-ai` pins `ai` v6.
- Config tokens are Symbols: `ConfigService.get()` needs the raw Symbol
  (`get(SERVICES_CONFIG)`), `@Inject()` needs the factory KEY
  (`@Inject(servicesConfig.KEY)`). Mixing them up fails silently (`undefined`).
- No OpenAPI spec — the SSE stream doesn't fit `openapi-generator`.
