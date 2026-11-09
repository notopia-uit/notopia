# Workspace assistant note editing

The workspace assistant can search for an existing note, open its live editor beside chat,
and suggest changes. One note is the edit target; attached notes are references. Accept or
discard suggestions before sending the next prompt. **Commit** remains a separate action:
it publishes the accepted document version for indexing/search.

Example: “Find my meeting notes and add a short action-items section.” The assistant first
uses authorized search/read tools, then requests `openNoteForEditing({ noteId })`. Opening
is a browser action: the browser authenticates to Hocuspocus and waits for initial sync.
The model then receives the live BlockNote document, rather than the indexed snapshot.

## Request flow and credits

```text
User prompt (assistantMode: true)
  -> authorization + billing allowance + one Redis credit reservation
  -> model search/read -> authorized openNoteForEditing
  -> Redis stores transcript, expected calls, target and hashed one-shot token
  -> browser opens and syncs the target editor
  -> continuation(token, matching results, fresh document state)
  -> owner/token/access checks + rate/concurrency lease (no extra credit)
  -> streamed applyDocumentOperations -> browser BlockNote executor
  -> continuation with operation acknowledgement -> model response
  -> accept/discard -> optional Commit
```

All continuations share the original fixed timeout and total model-step budget. A started
generation costs one credit even if stopped or the provider fails. Only a reservation known
not to have started is refunded. An expired, consumed, cancelled, or owner-mismatched token
cannot resume a task. A continuation cannot inject a new prompt or replace server history.

The existing authenticated `POST /ai` accepts three branches:

```ts
// Start: ordinary UI messages ending in a new user prompt.
{ assistantMode: true, workspaceId, messages }

// Continue: token comes from transient data-assistant-continuation SSE metadata.
{
  action: 'continue', workspaceId, taskId, token,
  results: [{ toolCallId, toolName: 'openNoteForEditing', output: { noteId, ready: true } }],
  noteId, documentState, toolDefinitions,
}

// Cancel, including between streams; returns 204.
{ action: 'cancel', workspaceId, taskId }
```

`applyDocumentOperations` acknowledges successful execution with `{ status: 'ok' }`.
Results must match the recorded call IDs and names exactly. The server validates the live
note's workspace and write access before exposing a completed open call, buffers its partial
input, and rechecks access on continuation. Viewers can use read-only chat.

## Browser ownership and review

- `AssistantRunner` owns one Chat, cancellation, handoffs and review state.
- `AssistantEditorSession` resolves opening only after the target editor is ready. Stale
  session callbacks are ignored; disconnected editors become unavailable for execution.
- `AssistantNoteEditor` owns the authenticated Hocuspocus room and existing Commit action.
- `AssistantAiBridge` uses exported `buildAIRequest` and `sendMessageWithAIRequest` with
  full-document context. The AI extension receives the shared Chat through `chatProvider`.

BlockNote forks the collaborative Y.Doc for suggestions. Reject rebinds the original live
document; accept merges a CRDT update into it. This preserves unrelated collaborator edits
instead of restoring a stale snapshot. The bridge guards exported tool execution with both
the request signal and a review revision, and disables executor animation delays, so queued
operations cannot mutate the document after Stop or review. Stop leaves applied suggestions
available to accept/discard. Navigation disposes the runner and discards its pending fork.

The installed executor observes the latest assistant message. A temporary, uniquely identified
empty assistant message isolates continuation execution from old completed tools; it is removed
in `finally` before future requests, including failed requests.

## Running and verifying

Use the normal local stack and AI configuration in `deploy/compose/AI_TESTING.md`. A model
supporting tool calls, authenticated note access, billing allowance, and AI Redis are required.
OpenRouter's free model allowance does not replace workspace billing credits. API and document
WebSocket URLs are supplied by the web app, with `ws`/`wss` derived from the API origin.

```sh
nx test ai --skip-nx-cache
nx test packages/ui --skip-nx-cache
nx run-many -t typecheck,lint -p ai,packages/ui,web
nx build ai
nx build web
# Use only a disposable isolated Redis instance:
AI_TEST_REDIS_URL=redis://localhost:16385 nx test ai --skip-nx-cache
```

The real-editor deterministic browser fixture uses existing Vite and system Firefox; it
adds no production authentication/billing bypass. In three terminals:

```sh
mkdir -p tmp/ai/firefox-profile
firefox --headless --remote-debugging-port 9224 --profile "$PWD/tmp/ai/firefox-profile"
pnpm exec vite --config packages/ui/tests/browser/vite.config.mts
node packages/ui/tests/browser/run.mjs
```

It mounts real BlockNote editors and collaborating Y.Docs, drives scripted Chat streams
through open/edit/continue, checks actual content, accepts/rejects suggestions, preserves
concurrent collaborator insertions, and rejects a queued delete after cancellation. Unit
tests cover token/result validation, single-debit continuation, target readiness, Stop,
navigation during bridge loading, and Strict Mode runner reactivation.

### Evidence limits

The deterministic browser fixture exercises the actual editor/executor and Yjs review path;
it does not exercise a real Hocuspocus server, billing provider, authenticated user, or Commit
RPC. The full live-stack open → edit → accept → Commit smoke remains unverified because local
billing/auth/model prerequisites were not established. The latest environment has no Docker
socket or isolated Redis endpoint, so Redis integration cases are skipped in ordinary runs.
Task 1's atomic Lua transitions were previously tested against disposable Redis; rerun the
command above when an isolated instance is available. Do not interpret skips as passes.

Final affected checks: 66 AI tests passed, 18 Redis cases skipped, and 19 UI tests passed;
AI/UI/web typechecks and lint exited successfully, and AI/web builds passed. Lint still
reports existing repository warnings. An independent code review's three important findings
(queued mutation after cancellation, temporary-message leakage, Strict Mode reactivation)
were reproduced and fixed with regression evidence. A minor inconsistency remains: assistant
quota failures enforce 402 but do not attach the ordinary-chat `requiredPlan` metadata.

## Troubleshooting

- **Opening fails or times out:** check token retrieval, document WebSocket connectivity,
  note workspace/write permissions, and initial Hocuspocus sync. An empty unsynced document
  is never valid live context.
- **409 on continuation:** task/token expired, was consumed or cancelled, or results/target
  differ from the recorded handoff. Submit a new prompt with a new message ID.
- **402:** provision workspace AI credits through billing; a free LLM endpoint alone is insufficient.
- **429:** initial and continuation streams share rate and active-stream limits.
- **Suggestions remain after Stop:** accept or discard them before another prompt or Commit.
- **Search misses edits:** accept and Commit first, then allow search indexing to catch up.
