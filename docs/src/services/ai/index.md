---
authorship:
  - ai
---

# AI service

The AI service powers Notopia's workspace assistant and the AI commands inside the note
editor. It runs as the NestJS application `apps/ai` (Nx project `ai`, local port `8085`).
It connects a language model to Notopia's authorized note tools. The model provider supplies
text generation; Notopia supplies workspace permissions, note access, editing and usage limits.

For provider credentials and examples, see [OpenRouter and OpenCode](./model-providers).

## Architecture

```text
Browser assistant / BlockNote editor
    -> Traefik (Authentik JWT validation)
    -> AI service (validation, permissions, quota, tool loop)
        -> OpenAI-compatible model provider
        -> note + authorization + billing (gRPC)
        -> Meilisearch (committed note content)
        -> Redis (usage, stream leases, continuation state)
    <- AI SDK UI-message stream (SSE)

Editing tool -> browser BlockNote editor -> Hocuspocus document session
             -> Accept / Discard -> Commit
```

The service has no application database or Kafka consumer. Redis is shared state, not an
optional cache. It must be persistent, standalone Redis 6.2+ with `noeviction`; the multi-key
Lua operations do not support Redis Cluster.

| Dependency          | Responsibility                                                                               |
| ------------------- | -------------------------------------------------------------------------------------------- |
| Note gRPC           | Resolve note metadata and its workspace using the requesting user                            |
| Authorization gRPC  | Discover workspaces and check current read/write access                                      |
| Billing gRPC        | Resolve the workspace's `ai_credits_monthly` allowance                                       |
| Meilisearch         | Search and read bounded, indexed content from committed notes                                |
| AI Redis            | Atomically reserve credits, deduplicate prompts, throttle requests and store assistant tasks |
| Document/Hocuspocus | Authorize and synchronize the browser's live editing session                                 |

### Search, editing and persistence

Server tools include `searchNotes`, `getNote`, `getNoteWorkspace`,
`readIndexedNoteContent` and, for general chat, `listWorkspaces`. They recheck access and
exclude unavailable, trashed or out-of-workspace notes. Search results describe committed
content and can lag behind unsaved editor changes.

The workspace assistant can use `openNoteForEditing` to open one existing note beside the
chat. After server authorization, the browser waits for an authenticated, synchronized
document session before supplying live document state and the supported BlockNote
`applyDocumentOperations` tool. Other notes can be references; one task cannot switch its
editing target or edit multiple notes.

Suggested changes use a fork of the collaborative Yjs document. **Accept** merges the AI
changes; **Discard** drops them while preserving collaborator changes. **Commit** is a
separate action that creates a revision and makes content available to indexing. Accepting
suggestions alone is not a commit; automatic commits depend on the workspace's plan/settings.

## HTTP contract

`POST /ai` accepts AI SDK UI messages, not OpenAI Chat Completions messages. The browser
uses the shared UI transport directly through Traefik; there is no Next.js AI proxy.

| Request                                                    | Purpose                                   |
| ---------------------------------------------------------- | ----------------------------------------- |
| `messages` with `workspaceId` or `noteId`                  | Start normal chat or an editor generation |
| Same fields plus `assistantMode: true`                     | Start a bounded workspace assistant task  |
| `action: 'continue'`, task/token and expected tool results | Resume that task after browser work       |
| `action: 'cancel'`, workspace/task IDs                     | Invalidate the task; returns `204`        |

Initial requests end with a nonempty user prompt and a unique message ID. Workspace context
may also come from `X-Workspace-Id`; conflicting contexts are rejected. If no context is
supplied, the user must have exactly one accessible workspace.

Assistant continuations use one-shot Redis tokens bound to the task owner, workspace and
expected calls. The server retains the authoritative transcript; clients cannot replace
history on continuation. Continuations retain the original timeout and cumulative step budget.

Traefik validates the user's JWT and injects `x-forwarded-id`. The service trusts that
gateway header, so public requests must go through Traefik. A model-provider API key is
outbound authentication to the model, not a Notopia user credential.

## Credits and limits

Billing owns the subscription. AI reads its allowance when admitting an initial generation;
continuations of that assistant task do not fetch billing again or debit another monthly credit.

- One started generation/task costs one Notopia credit, including its tool steps and continuations.
- Usage resets at the start of each UTC calendar month, independently of subscription renewal.
- Cancellation or failure before generation starts can refund the reservation. After start,
  cancellation, provider failure and timeout retain the charge.
- Zero allowance prevents generation; `-1` means unlimited. Default plans currently give
  Free no AI credits and Plus/Pro an allowance, but runtime enforcement uses billing's values.
- OpenRouter/OpenCode charges and Notopia credits are separate. A free provider model does
  not remove Notopia's workspace allowance check.

| Environment variable             | Default  | Meaning                                     |
| -------------------------------- | -------- | ------------------------------------------- |
| `NOTOPIA_AI_REQUESTS_PER_MINUTE` | `20`     | Per-user request rate limit                 |
| `NOTOPIA_AI_CONCURRENT_STREAMS`  | `2`      | Per-user active stream limit                |
| `NOTOPIA_AI_TIMEOUT_MS`          | `120000` | Bounded request/task lifetime               |
| `NOTOPIA_AI_MAX_OUTPUT_TOKENS`   | `4096`   | Output budget per model step                |
| `NOTOPIA_AI_MAX_STEPS`           | `10`     | Model-step budget, cumulative across a task |

## Local setup

Run commands from the repository root. Follow the project's normal local infrastructure,
Authentik and database setup first. Compose runs infrastructure; Nx runs application services.

```sh
docker compose -f deploy/compose/compose.yaml up -d
```

This includes `ai_redis` on `127.0.0.1:6376`, matching `apps/ai/.env`. Its AOF and named volume
preserve usage counters across container recreation. To start only AI Redis:

```sh
docker compose -f deploy/compose/compose.yaml up -d --wait ai_redis
```

Choose a [model provider](./model-providers) and put overrides in `apps/ai/.env.local`.
Nx loads the project environment files. Restart AI after changing them. The checked-in
`apps/ai/.env` also defines these local dependencies:

| Variable                                     | Local value                            |
| -------------------------------------------- | -------------------------------------- |
| `NOTOPIA_AI_SERVICES_NOTE_GRPC_URL`          | `127.0.0.1:18081`                      |
| `NOTOPIA_AI_SERVICES_AUTHORIZATION_GRPC_URL` | `127.0.0.1:18089`                      |
| `NOTOPIA_AI_SERVICES_BILLING_GRPC_URL`       | `127.0.0.1:18084`                      |
| `NOTOPIA_AI_MEILI_HOST`                      | `http://meilisearch.notopia.localhost` |
| `NOTOPIA_AI_MEILI_API_KEY`                   | Local Meilisearch development key      |
| `NOTOPIA_AI_REDIS_URL`                       | `redis://127.0.0.1:6376`               |

Use a workspace with a billing allowance. On a prepared development database,
`just fixtures-billing` creates a fake Pro lifetime subscription for workspace
`00000000-0000-4000-8000-000000000110`. It requires billing migrations and seeded prices
(`just seed-billing`), which in turn require the billing service's Stripe sandbox configuration.
The note and authorization fixtures must also give your user access to that workspace.
Fake billing provider IDs support entitlement testing, not real checkout or refunds.

```sh
nx run-many -t run -p note,authorization,billing,document,search-worker,ai --parallel=6
```

In another terminal:

```sh
nx dev web
```

Open `http://web.notopia.localhost`, sign in and open the workspace assistant. Ask it to find
and summarize a note, then ask it to edit that note. Check the opened target, review the
suggestions, Accept or Discard, and Commit accepted changes. Test the editor's inline AI too.

### Health and troubleshooting

```sh
curl -f http://api.notopia.localhost/ai/health/ready
```

`/ai/health/live` (also `/ai/health`) and `/ai/health/startup` report that the HTTP service
is running; `/ai/health/ready` pings Redis. Readiness does not prove that model credentials, other gRPC
services or browser document editing work.

| Symptom                             | Check                                                               |
| ----------------------------------- | ------------------------------------------------------------------- |
| `402 planLimitExceeded`             | Workspace billing allowance and remaining Notopia credits           |
| `429`                               | Per-user rate/concurrency limits; honor `Retry-After`               |
| `503 aiUsageUnavailable`            | Redis address, connectivity and persistence configuration           |
| `503 billingServiceUnavailable`     | Billing gRPC and valid plan allowance data                          |
| Provider authentication/model error | Provider base URL, key, model ID and account access                 |
| Text works but note tools fail      | Model support for streamed tool calls and structured tool input     |
| Note opens but editing fails        | Document logs, WebSocket authentication, schema and synchronization |
| Search shows old content            | Commit the note and check search-worker indexing                    |

Before streaming, failures are JSON errors. After streaming starts, failures appear as safe
AI SDK stream error parts; HTTP `200` alone does not prove a generation succeeded.

## Development checks

```sh
nx test ai
nx typecheck ai
nx lint ai --fix
nx build ai
```

Real Redis tests require a separate disposable Redis, not the instance holding manual-test
credits. Without the variable below, those integration cases are skipped:

```sh
AI_TEST_REDIS_URL=redis://127.0.0.1:16385 nx test ai --skip-nx-cache
```

Implementation guidance lives in `apps/ai/AGENTS.md` and `apps/ai/HOW_TO_ADD_TOOLS.md`.
The internal walkthrough `docs/ai/ASSISTANT_NOTE_EDITING.md` covers the continuation protocol,
shared browser runner and deterministic collaborative-editor fixture in more detail.
