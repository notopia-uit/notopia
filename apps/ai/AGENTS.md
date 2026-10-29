# AI Service (nx: `ai`)

NestJS HTTP backend for BlockNote AI and workspace assistant chat. No database or Kafka;
uses note, authorization and billing gRPC, Meilisearch, and Redis for shared accounting.

## Modules

| Module | Role |
| --- | --- |
| `ai/` | `POST /ai`, runtime validation, tool selection, AI SDK v6 UI-message stream |
| `note/` | User-aware note metadata and workspace lookups over gRPC |
| `authorization/` | Workspace discovery, read/write permission checks |
| `billing/` | Uncached workspace entitlements and live upgrade plan lookup over gRPC |
| `search/` | Workspace-filtered committed content search and bounded indexed content reads |
| `usage/` | Redis Lua: per-user throttle/stream leases, workspace monthly credits, duplicate requests |
| `common/` | Gateway user guard, normalized HTTP errors |

## Request flow and context

Browser → Traefik JWT middleware → `POST /ai` → LLM → AI SDK SSE response.
There is no web proxy. `GET /ai/health` (alias of `/live`) and
`/ai/health/{live,startup,ready}` remain unauthenticated.

Requests contain `messages`, optional BlockNote `toolDefinitions`, and context supplied as
`workspaceId`, `noteId`, or `X-Workspace-Id`. Conflicting contexts are rejected. `noteId`
is resolved using a user-aware note RPC. Without explicit context, exactly one accessible
workspace is required; zero or multiple workspaces return 400 `workspaceRequired`.
The web assistant sends workspaceId and editors (including previews) send noteId through the shared UI transport; other clients need to send context; do not infer billing from prompt text,
attached note labels, a model tool argument, or an untrusted Referer.

HTTP trusts gateway-injected `x-forwarded-id`; the backend must not be exposed to untrusted
traffic bypassing Traefik. User ID is the throttle key. Forwarded client IP is not used:
a trusted proxy chain has not been configured for this service.

## Billing and accounting

Billing owns subscriptions and resolves `ai_credits_monthly`. AI fetches it on every generation
(no entitlement cache or subscription-event consumer). Missing quota means zero, `-1` means
unlimited. Permission is checked before billing. Billing failure → 503
`billingServiceUnavailable`; exhausted quota → 402 `planLimitExceeded` with the same
`details` shape as note/document (`kind`, `key`, `current`, `max`, `requiredPlan`) and `resetAt`.
Upgrade options come from billing's `ListPlans`, not hardcoded quota values. `requiredPlan()`
returns the cheapest tier whose live `ai_credits_monthly` exceeds the exhausted limit; it never
invents a tier. A missing/invalid plan limit, or no tier above the current limit, fails closed
with 503 `billingServiceUnavailable`.

Policy: one credit per accepted generation, including server tool steps. Usage resets at
00:00 UTC on the first of each calendar month, not the payment-provider renewal date (the
current gRPC contract has no billing-period fields). A reserved credit is refunded only on
confirmed failure/cancellation before generation starts. Once started, provider errors,
timeouts and client disconnects retain the charge, because provider work may have occurred.
A crash or uncertain Redis write is conservatively not refunded automatically.

Last user message IDs deduplicate requests per user/workspace/month. A duplicate returns 409
without another debit or replaying a response. Retry/regenerate with a new user message ID.
Requests must end with a user prompt; client tool-result-only continuation is not supported.
BlockNote's streamed editor tools and server tool loops are supported within a generation.

`NOTOPIA_AI_REDIS_URL` is required: the service refuses to boot without it. It must point to a
dedicated, persistent, **noeviction** Redis 6.2+. There is no in-memory fallback for quota
accounting; a down Redis fails `/ai/health/ready` with 503 and returns 503 `aiUsageUnavailable`
to requests. Liveness never checks Redis, so an outage drains traffic without restarting pods.
The Lua scripts use multiple keys and require standalone Redis (not Redis Cluster).
Counters/deduplication survive until seven days after the month resets; active stream leases
expire after the bounded request timeout plus a grace period.
Redis loss can lose usage history: persistence/backups are an operational requirement.

## Tools and errors

Read `HOW_TO_ADD_TOOLS.md` before adding tools. All server tools recheck workspace read access
at execution. Search is filtered by validated workspace ID and results pass live note
permission/trash/workspace checks. Indexed content may be stale and is never advertised as
live document content. No tool calls another service's public REST API.

Only BlockNote `applyDocumentOperations` client definitions are accepted, and require write
permission at admission. Execution is client-side, with document authorization enforced by
Hocuspocus. General chat includes workspace discovery; editor requests omit that tool.
Do not set `toolChoice: 'none'` or force a tool call.

Before response headers, errors are JSON `{ code, message, ... }` with 400/401/402/403/409/
413/429/503/504 as appropriate. 429 includes `Retry-After`. After headers, provider errors use
AI SDK stream error parts with a safe message; server tool failures are sanitized before
AI SDK serializes them. The shared UI transport normalizes HTTP JSON into readable errors for workspace chat and BlockNote. Remote file parts are rejected to avoid server-side URL fetches.

## Configuration and checks

`NOTOPIA_AI_SERVICES_BILLING_GRPC_URL` joins existing note/authorization endpoints.
Defaults: `NOTOPIA_AI_REQUESTS_PER_MINUTE=20`, `NOTOPIA_AI_CONCURRENT_STREAMS=2`,
`NOTOPIA_AI_TIMEOUT_MS=120000`, `NOTOPIA_AI_MAX_OUTPUT_TOKENS=4096` (per model step),
`NOTOPIA_AI_MAX_STEPS=10`. The timeout covers setup and generation. Model credentials remain
`NOTOPIA_AI_LLM_{BASE_URL,API_KEY,MODEL}`.

- `nx typecheck ai`, `nx lint ai --fix`, `nx test ai`
- Real Redis integration tests: `AI_TEST_REDIS_URL=redis://localhost:16385 nx test ai --skip-nx-cache`.
  Tests use a random key namespace and remove only their own keys. Without the URL, Redis
  integration cases are skipped. Do not point tests at production.

Keep `@ai-sdk/openai-compatible` on v2 while BlockNote pins AI SDK v6. Config tokens are
Symbols: `ConfigService.get(SERVICES_CONFIG)` versus `@Inject(servicesConfig.KEY)`.
