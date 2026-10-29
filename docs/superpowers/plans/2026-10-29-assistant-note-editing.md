# Workspace Assistant Note Editing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let workspace chat find, open, and edit one existing note at a time through its live collaborative editor.

**Architecture:** One browser-owned Chat instance coordinates an active Hocuspocus/BlockNote session. A backend-owned open-note tool hands execution to the browser; Redis records authoritative task state and one-shot continuations so the model can resume with live document context without a second credit debit. Existing BlockNote operations and review controls handle edits.

**Tech Stack:** TypeScript, React, Next.js, NestJS, AI SDK v6, installed BlockNote 0.55.0, Hocuspocus/Yjs, ioredis/Lua, Vitest, Nx.

**Spec:** `docs/superpowers/specs/2026-10-29-assistant-note-editing-design.md` (approved).

## Global Constraints

- One active target note; other notes may provide references. Existing note content only.
- Reuse BlockNote's exported integration; do not copy its internal operation executor.
- Keep task/user/workspace/target binding authoritative on the server. Client history is not proof of a paid task.
- One credit per started user generation, including continuations. Bound total steps and task lifetime.
- Preserve live collaboration, permission checks, normal editor commands, and separate Commit behavior.
- No snapshot-based document rollback. Verify installed BlockNote review semantics before wiring the UI.
- Billing setup is deferred. Preserve unrelated changes already present in AI, tests, docs, and lint configuration.
- Use existing dependencies unless a necessary addition is explained and approved. Shared UI receives URLs as props, never process.env.
- Run formatting and affected lint/typecheck/tests. Never bypass a failed check or claim blocked live checks passed.

## Review Focus

1. Late chunks after Stop or switching: cannot mutate the new target (Task 4).
2. Disconnect/reconnect during opening: cannot read an empty unsynced document or hang indefinitely (Task 3).
3. Concurrent/replayed continuations: cannot execute twice or evade credit/concurrency limits (Task 1).
4. Collaborator edits during review: rejecting AI changes preserves unrelated edits (Tasks 3 and 5).
5. Refresh, workspace navigation, and failure between stream completion and continuation: no replay or orphaned reusable task (Tasks 2 and 4).

## Evidence path and budget

The executor owns verification; the final reviewer challenges the evidence rather than repeating all checks. Preferred evidence: mock-model backend stream tests for protocol behavior, real isolated Redis for atomic transitions, browser tests with controlled AI responses for actual BlockNote/Yjs effects, then a live-stack smoke test when prerequisites permit. Pure mocked editor tests are a fallback for lifecycle logic only, not proof that editing/rejection works. A live-model-only test is weaker for races and is not the primary evidence.

Use `MockLanguageModelV3` already used in `apps/ai/tests/stream.spec.ts`. Add a reusable scripted SSE fixture under UI test support only if required to drive real browser editing; never add a production bypass for auth or billing. Keep temporary runners under `tmp/ai/`. Do not install a browser-testing dependency without approval. Existing browser tools can provide manual browser evidence. Redis tests must use an isolated instance and namespace.

## File responsibilities

| Files | Responsibility |
| --- | --- |
| `apps/ai/src/ai/assistant-task.types.ts` | Server task states, continuation inputs, and stream data contract |
| `apps/ai/src/usage/assistant-task.scripts.ts`, `assistant-task.service.ts` | Atomic one-shot task transitions and persisted authoritative history |
| `apps/ai/src/usage/usage.service.ts`, `usage.scripts.ts`, `usage.module.ts` | Lease-only continuation admission using existing per-user limits |
| `apps/ai/src/ai/chat-request.dto.ts`, `ai.controller.ts`, `ai.service.ts`, `ai.module.ts` | Parse, route, and orchestrate start/continue/cancel requests |
| `apps/ai/src/ai/assistant-stream.ts`, `tools/open-note.ts` | Stream recording, target validation, handoff, total step budget |
| `packages/ui/src/lib/assistant-protocol.ts`, `assistant-session.ts` | Typed protocol and target/session lifecycle independent of rendering |
| `packages/ui/src/lib/assistant-ai.ts`, `assistant-transport.ts` | BlockNote shared-chat bridge and start/continue/cancel transport |
| `packages/ui/src/components/assistant-note-editor.tsx`, `editor-core.tsx` | Synced authenticated target editor and supported AI review integration |
| `packages/ui/src/hooks/use-assistant-chat.ts`, `components/assistant-view.tsx` | Chat lifecycle, tool dispatch, target panel, cancellation/errors |
| Web assistant `page.tsx`, `assistant-boundary.tsx` | Supply API/WebSocket configuration from the app |

Existing tests are under `apps/ai/tests/` and colocated UI `*.spec.ts`. Recheck current paths at execution because a separate test-colocation plan exists in the working tree.

### Task 1: Atomic continuation state and stream leases

**Files:** Create task types, task service/scripts and `apps/ai/tests/assistant-task.spec.ts`; modify usage service/scripts/module and `apps/ai/tests/usage.spec.ts`.

**Interfaces:**
- `TaskOwner = { userId: string; workspaceId: string }`.
- `AssistantTask` holds random task ID, owner, original prompt ID, server-recorded UI transcript, fixed deadline, remaining model steps, optional target note ID, expected client calls, state (`running | awaiting-client | completed | cancelled | failed`), and revision.
- `AssistantTaskService.create(owner, promptId, messages, limits): Promise<AssistantTask>`; `suspend(taskId, revision, transcript, pendingCalls, targetNoteId, stepsUsed): Promise<{ taskId: string; token: string; expiresAt: number }>`; `claim(owner, taskId, token): Promise<AssistantTask>`; `finish(owner, taskId, outcome): Promise<void>`.
- `UsageService.acquireContinuationLease(userId: string): Promise<{ release(): Promise<void> }>` uses the same active-stream/rate keys as initial admission without modifying monthly usage.
- Use `config.timeoutMs` as a fixed end-to-end task lifetime and `config.maxSteps` as the total model-step budget across requests. No sliding expiry. Store hashed random continuation tokens; never log tokens or document contents.

- [ ] Write failing Redis tests for `claimIsSingleUse`, `ownerCannotBeChanged`, `expiredTaskCannotResume`, `terminalTaskCannotResume`, `suspendCannotOverwriteCancelledTask`, and `continuationLeaseDoesNotDebit`. Assert two parallel claims yield exactly one success; rate/concurrency caps apply equally to both request types; task expiry does not extend after suspension.
- [ ] Run `nx test ai --skip-nx-cache` and the same suite with `AI_TEST_REDIS_URL` pointing to isolated Redis; confirm the new assertions fail for the missing behavior.
- [ ] Implement atomic compare-and-set transitions in Lua, bounded record size (512000 bytes), owner/revision checks, and fail-closed Redis errors. Failed lease acquisition after a claimed continuation terminates that task rather than reopening a consumed token. Initial reservation remains responsible for the only credit debit/refund policy.
- [ ] Run the focused tests against real Redis; assert original usage tests still pass. Format and run `nx lint ai` and `nx typecheck ai`.
- [ ] Commit only this task's changes: `feat(ai): track bounded assistant continuations`.

### Task 2: Authorized open-note handoff and resumable generation

**Files:** Create `assistant-stream.ts`, `tools/open-note.ts`, `apps/ai/tests/assistant-stream.spec.ts`; modify AI DTO/controller/service/module, request/access/stream tests, and task types from Task 1.

**Interfaces:**
- Preserve existing ordinary chat/BlockNote requests. New workspace assistant starts opt in with `assistantMode: true`.
- Continuation request to existing `POST /ai`: `{ action: 'continue', workspaceId, taskId, token, results, noteId?, documentState?, toolDefinitions? }`. Do not accept replacement messages or a new prompt on this branch. `results` must match the recorded pending tool calls by ID/name and allowed output schema.
- Cancellation request to the same authenticated route: `{ action: 'cancel', workspaceId, taskId }`, returning 204 after owner-scoped termination. Parse this as a distinct branch, without requiring messages.
- Fixed tool `openNoteForEditing({ noteId: UUID })`; result is `{ noteId: UUID, ready: true }` or a bounded tool error. Tool identity/schema are server-owned, not supplied by the browser.
- Stream `data-assistant-task` at initial admission (task ID/deadline); emit `data-assistant-continuation` with Task 1's token only after a valid handoff is durably stored. Continuation tokens are not inserted into model messages.

- [ ] Add failing tests: search→open→edit resumes with one reserve/debit; a forged token, extra tool result, changed target, user message injection, or wrong owner/workspace is rejected; a viewer cannot open for editing; trashed/missing targets fail; parallel/different targets in one task cannot open; step count is cumulative; oversized live context fails before model use.
- [ ] Run `nx test ai --skip-nx-cache` and confirm new behavior fails while existing ordinary requests remain covered.
- [ ] Implement the discriminated parser and authorized tool. Validate note/workspace/write access before forwarding a completed open-tool call to the browser. Buffer that tool's partial input until validation completes. Pause generation at a client handoff; do not run later model steps without its output. Once a task selects a target, it cannot select a different target within that task.
- [ ] Implement `streamAssistantTask(...)` in `assistant-stream.ts`, keeping `AiService` as admission/routing orchestration. Persist actual server-emitted transcript/calls, reconstruct continuation history from Redis, and attach fresh BlockNote document state to the recorded user-message metadata for `injectDocumentStateMessages`. Never substitute client history for recorded messages. Recheck access on continuation; retain already-paid admission without charging again or requiring unused quota for that same task.
- [ ] Handle finish versus abort explicitly: normal HTTP completion after a saved handoff must not cancel the waiting task; disconnect/provider error/timeout closes it. Cancel marks server state terminal even if the connection has already ended. Any in-flight response after cancellation cannot create a usable continuation. Clip each HTTP/model timeout to the original deadline.
- [ ] Test that a failure after starting retains one credit, before starting follows current refund rules, and cancelling between requests invalidates the token. Run AI tests, real-Redis suite, typecheck, and lint.
- [ ] Commit: `feat(ai): resume authorized assistant editing tasks`.

### Task 3: Live target editor and BlockNote bridge

**Files:** Create UI `assistant-protocol.ts`, `assistant-session.ts`, `assistant-ai.ts`, `assistant-note-editor.tsx`, and their focused tests; modify `editor-core.tsx` only for optional shared-AI integration and cleanup callbacks.

**Interfaces:**
- `AssistantTarget = { noteId: string; name: string; sessionId: string }`.
- `AssistantEditorSession` exposes `open(noteId: string, signal: AbortSignal): Promise<AssistantTarget>`, `getReadyEditor(sessionId: string): MyEditor`, `close(): Promise<void>`, and `cancel(): void`.
- `AssistantNoteEditor` receives workspace/API/WebSocket configuration and the target, emits synced editor readiness/disposal/error, and owns exactly one keyed Hocuspocus room. Use client-side access-token retrieval and existing navigation/schema/user providers.
- `AssistantAiBridge` exposes `prepare(signal): Promise<AIRequest>`, `send(chat, message, options, signal): Promise<void>`, `accept(): void`, `reject(): void`, and `abort(): Promise<void>`; `AIRequest` and `Chat` are library types. `send` checks both chat error state and the exported helper's `Result`.

- [ ] Verify installed 0.55.0 exports (`chatProvider`, `buildAIRequest`, `sendMessageWithAIRequest`, extension review methods) and inspect their behavior before integration. Record how streaming suggestions sync and how rejection preserves collaborator edits. If public helpers cannot provide this behavior, report the concrete conflict before copying internals or changing review semantics.
- [ ] Write failing session/bridge tests: no readiness before initial sync; abort during opening rejects and disposes; auth failure is surfaced; duplicate open calls share one session; a failed switch cannot keep a stale writable target; old-session callbacks cannot mutate the new editor. Use a real editor browser check for review semantics, not just mocked methods.
- [ ] Run `nx test packages/ui --skip-nx-cache` and confirm new lifecycle tests fail.
- [ ] Implement readiness gating and cleanup, bounded by the task signal/deadline. Build full-note requests with `useSelection: false` and `deleteEmptyCursorBlock: false`. Supply a shared chat provider instead of simultaneously setting both transport and chatProvider. Reuse supported extension review setup, not merely menu labels; verify accept/reject actually owns the streamed suggestions.
- [ ] Exercise the actual bridge with deterministic edit operations and a concurrent collaborator insertion: rejecting AI suggestions preserves that insertion. Verify normal editor AI remains unchanged when optional props are omitted. Run UI tests/typecheck/lint.
- [ ] Commit: `feat(ui): add live assistant note editing session`.

### Task 4: Connect assistant chat, continuation transport, and editor panel

**Files:** Create `assistant-transport.ts`, `use-assistant-chat.ts` and focused tests; modify `assistant-view.tsx`, web assistant `page.tsx`/`assistant-boundary.tsx`; consume Tasks 1–3 contracts.

**Interfaces:**
- `useAssistantChat({ workspaceId, apiUrl, documentWsUrl })` owns one `Chat<UIMessage>`, the target session, task/continuation state, submit/stop/review/close actions, and observable UI state.
- Transport sends ordinary user prompts as initial assistant requests and continuations using only validated results, fresh context, and the latest one-shot token. It delegates auth/error handling to existing `fetchAi` and access-token helpers.
- Execute open calls once by task ID/toolCallId. Wait for the previous HTTP stream to finish and the note to sync before invoking the explicit continuation. Disable generic automatic tool-result resubmission to avoid duplicate requests and bypassing BlockNote setup.

- [ ] Write failing tests for a scripted search/open/edit response, duplicate tool notification, Stop during open and edit, late input chunks, workspace navigation, expired token, and failure restoring prompt text with a new message ID. Assert no second debit request and no simultaneous target editors. Cancel must invoke both local AbortController/chat stop and authenticated task cancellation; cancellation network failures are surfaced, while local mutations still stop immediately.
- [ ] Run UI tests and confirm the new orchestration assertions fail.
- [ ] Implement transport and shared-chat ownership. For an already-open target, a new user prompt builds fresh live context; for an unopened target, continue only after open succeeds. On follow-up prompts while review is pending, require accepting/discarding current suggestions first. Read-only chat remains usable for viewers without advertising edit tools.
- [ ] Add the responsive editor panel with note identity, loading/connection errors, existing review actions, Close, and Commit via `useEditorState`. Guard target switches during generation/review, preserve one mounted session on responsive layout changes, and ensure Commit waits for applied changes and provider sync. Keep reference attachments distinct from the active edit target.
- [ ] Pass API and WebSocket configuration from the web page/boundary. Derive ws/wss from the configured API origin, keeping secrets out of props. Verify normal error boundaries/navigation remain in place.
- [ ] Run UI and AI tests, `nx typecheck packages/ui`, `nx typecheck web`, `nx lint packages/ui`, and `nx lint web`; format changed JS/TS with oxfmt.
- [ ] Commit: `feat(web): enable assistant-driven single-note editing`.

### Task 5: End-to-end evidence and explanatory documentation

**Files:** Update AI/UI AGENTS, `apps/ai/HOW_TO_ADD_TOOLS.md`, `deploy/compose/AI_TESTING.md`; create `docs/ai/ASSISTANT_NOTE_EDITING.md`. Add narrow regression tests in preceding tasks' test files if integration reveals a missing case.

- [ ] Run the deterministic full flow: find an unopened note → open synced editor → edit → review/accept → commit. Inspect document content through the editor/provider and resulting commit state, not only assistant prose. Repeat reject with concurrent collaborator input, Stop during streamed mutation, failed open, viewer access, and navigation cleanup.
- [ ] Run final affected suites once after final changes: `nx test ai --skip-nx-cache`, isolated real-Redis AI tests, `nx test packages/ui --skip-nx-cache`, AI/UI/web typechecks and lint, and `nx build ai` / `nx build web`. Repeat only checks invalidated by subsequent changes or concrete failures.
- [ ] Attempt live-stack browser smoke only if billing/auth/model prerequisites are ready. Otherwise document the exact blocker and which deterministic checks ran; do not use fabricated credentials or report a live success.
- [ ] Write the walkthrough with a request example, browser executor versus server tool explanation, continuation/credit diagram, relevant code excerpts, collaboration/review semantics verified in Task 3, Stop/switch behavior, commands, and troubleshooting. Remove outdated “client continuations unsupported” claims only where behavior now supports them.
- [ ] Review the diff against the approved spec, verify unrelated changes are preserved, and obtain the chosen execution workflow's final review. Fix supported findings and rerun only affected evidence.
- [ ] Commit: `docs(ai): explain and verify assistant note editing`.

## Self-review

Spec requirements map to Tasks 1–2 (authorization/accounting/protocol), 3 (live editor/review), 4 (one-target UX/lifecycle), and 5 (evidence/docs). Review-focus cases each have an owning test. Task types are established before consumers; browser wire types mirror the server DTO and get contract fixtures in Task 4. No production code has been changed by this plan.
