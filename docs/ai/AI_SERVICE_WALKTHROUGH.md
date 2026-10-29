# Notopia AI implementation walkthrough

Written: 27 October 2026. This describes the code currently in the repository.

This guide starts from the basics. You do not need to understand NestJS, Redis, or
BlockNote before reading it. Read the examples first, then open the linked source files.

Code blocks marked **Actual code** are excerpts from the implementation. Imports and
surrounding code may be omitted. Blocks marked **Illustration** or **Future proposal**
explain an idea; they are not additional implemented features.

Suggested reading order:

- For the overall idea: sections 1–2, then the two request examples in sections 4 and 14.
- For the editing-tool question: sections 14–16 and the future plan in section 24.
- For implementation work: use section 25 as a file map, then read the relevant task.

## 1. What works, and what does not

The completed work connects the existing AI backend to billing, shared usage accounting,
authorized note retrieval, and the frontend. It does not create a new subscription service.

| Task | Current result |
| --- | --- |
| Validate incoming AI requests | Implemented |
| Identify and authorize the workspace | Implemented |
| Read AI credit allowances from billing | Implemented |
| Count monthly credits using Redis | Implemented |
| Limit request frequency and simultaneous generations | Implemented |
| Reject duplicate prompt submissions | Implemented |
| Secure the existing note-search tool | Implemented |
| Read committed note content through an AI tool | Implemented |
| Combine server tools with BlockNote editor tools | Implemented |
| Stream answers and handle errors/cancellation | Implemented |
| Send workspace/note context from the UI | Implemented |
| Share the AI URL through React context | Implemented |
| Show tool activity, errors, Stop, and prompt recovery | Implemented |
| Make the commit button clearer | Implemented |
| Provide AI Redis and optional Ollama through Compose | Implemented |
| Use existing BlockNote AI inside an open note | Integrated; the underlying editor tools come from BlockNote |
| Edit notes from the standalone workspace assistant | **Not implemented** |
| Assistant sidebar connected to the current editor | **Not implemented** |
| Open arbitrary notes in managed editor sessions for the assistant | **Not implemented** |
| Limit requests by client IP | **Not implemented**; limits use authenticated user ID |
| Full manual browser verification with every service running | **Still needs to be performed** |

The five read tools are not all new. `searchNotes`, `getNote`, and `listWorkspaces`
already existed; I revised their integration and access handling. `getNoteWorkspace`
and `readIndexedNoteContent` extend what the assistant can retrieve.

## 2. Understand the pieces first

Imagine a user asking: “Find my meeting notes and summarize them.”

- **Web/UI** displays the chat and sends the question.
- **AI service** coordinates permissions, credits, the model, and tools.
- **Model**, such as Qwen through Ollama, decides what text or tool calls to produce.
- **Tools** are functions the model can request. Application code executes them.
- **Authorization service** answers whether this user may access the workspace.
- **Billing service** supplies the workspace's plan allowance.
- **Redis** tracks how much AI has been used and how many requests are active.
- **Note service** supplies note metadata and authoritative access checks.
- **Meilisearch** supplies searchable content from committed documents.
- **Document service** owns the live collaborative document and revisions.

A model is not directly connected to your database. When it requests a tool, our
application decides whether to execute that function and what data to return.

```text
Browser
  │ question + workspaceId or noteId + access token
  ▼
Traefik gateway
  │ validates token and forwards user identity
  ▼
AI controller
  │ validates request and starts cancellation/timeout handling
  ▼
AI service
  ├── Redis: request throttle
  ├── Authorization: workspace permissions
  ├── Billing: monthly AI allowance
  ├── Redis: reserve one credit and one active-stream slot
  └── Model
        ├── asks for a tool → our code checks access and retrieves data
        └── produces text → streamed back to the browser
```

The browser sees a response arrive in pieces. That is what “streaming” means.

### How the backend files fit together

These names describe responsibilities, not different computers:

| File/class kind | Plain meaning | Example |
| --- | --- | --- |
| Controller | Receives an HTTP request and sends the response | `AiController` |
| Service | Coordinates application behavior | `AiService` |
| Adapter/client | Calls another service or storage system | AI-local `BillingService` |
| Module | Tells NestJS which pieces to create and connect | `AiModule` |
| Schema | Rules for the shape and allowed values of data | `requestSchema` |

The AI-local `BillingService` class is a client adapter. It is not a second billing
application. It calls the existing Go billing service.

**Actual module registration:**

```ts
@Module({
  imports: [NoteModule, SearchModule, AuthorizationModule, BillingModule],
  controllers: [AiController],
  providers: [AiService, UsageService],
  exports: [AiService],
})
export class AiModule {}
```

NestJS creates these objects and passes dependencies to their constructors. This is called
**dependency injection**. For example, `AiService` receives its billing client rather than
constructing a new billing client for every question. `@Inject(...)` identifies a particular
configuration/provider to supply.

A gRPC call is an internal network request between services using the contract in `proto/`.
The generated TypeScript types describe that contract; they do not replace the running
service that actually answers the call.

## 3. Task: validate requests before using them

**Problem:** a TypeScript type does not protect an HTTP endpoint at runtime. Someone
can send an empty message list, a fake workspace ID, or unsupported tool definitions.

**Files:** [chat-request.dto.ts](../../apps/ai/src/ai/chat-request.dto.ts),
[main.ts](../../apps/ai/src/main.ts).

**Actual code, shortened to the message fields:**

```ts
messages: z
  .array(
    z.object({
      id: z.string().min(1).max(200),
      role: z.enum(['user', 'assistant']),
      parts: z.array(z.unknown()).min(1).max(200),
    }).passthrough()
  )
  .min(1)
  .max(100)
```

Read that as: “messages must be a list containing 1–100 messages. Each has a bounded
ID, an allowed role, and 1–200 parts.” The AI SDK also validates the message structure.
The final message must be a user prompt containing non-empty text.

Other checks include:

- Workspace and note IDs must be UUIDs.
- The request body is limited to 512,000 bytes, approximately 500 KiB.
- Client editing definitions may only use `applyDocumentOperations`.
- User-supplied system-role messages are rejected.
- File message parts are rejected, preventing this endpoint from accepting arbitrary
  remote-file requests that could trigger server-side downloads.

**Illustration of a valid workspace request:**

```json
{
  "workspaceId": "00000000-0000-4000-8000-000000000001",
  "messages": [
    {
      "id": "prompt-001",
      "role": "user",
      "parts": [{ "type": "text", "text": "Find notes about meetings" }]
    }
  ]
}
```

The UUID above is an example, not a promise that the workspace exists. Authentication
comes through the access token and gateway; a user ID in the JSON would not grant access.

**When developing later:** add new request fields to the runtime schema, not just to
a TypeScript interface. A compile-time type and runtime validation solve different problems.

## 4. Task: decide which workspace the request belongs to

**Problem:** a user can belong to several workspaces. We must know which workspace's
notes to search and which workspace pays for the generation.

**File:** [ai.service.ts](../../apps/ai/src/ai/ai.service.ts), `resolveWorkspace()`.

The backend accepts three ways to provide context:

| Input | Meaning |
| --- | --- |
| Body `workspaceId` | “This request belongs to this workspace.” |
| Body `noteId` | “Resolve the workspace containing this note.” |
| Header `X-Workspace-Id` | An alternative way for a client to identify the workspace |

If several inputs are supplied, they must agree. The backend never asks the model to
choose the workspace that pays.

**Actual code:**

```ts
let workspaceId = body.workspaceId ?? headerWorkspace;
if (body.noteId) {
  const workspace = await this.noteClient.getNoteWorkspace(userId, body.noteId);
  if (!workspace || (workspaceId && workspaceId !== workspace.id)) {
    throw new AiError(403, 'forbidden', 'Note is outside this workspace.');
  }
  workspaceId = workspace.id;
}
```

`??` means “use the other value if this one is missing.” The note lookup includes the
authenticated `userId`, so knowing a note ID is not enough to access it.

For older clients that supply no context, the backend checks the user's workspaces.
Exactly one accessible workspace can be used automatically. Zero or several produces
`400 workspaceRequired`; choosing the first workspace would risk incorrect billing.

**Example:** you are viewing a linked note from another workspace. Its editor sends
that linked note's ID. The backend resolves the linked note's workspace rather than
charging the workspace of the page behind the preview.

## 5. Task: enforce permissions at admission and during tools

**Problem:** being signed in does not mean a user may read every workspace. Access
can also change after a conversation starts.

**Files:** [authorization.service.ts](../../apps/ai/src/authorization/authorization.service.ts),
[custom-tools.ts](../../apps/ai/src/ai/tools/custom-tools.ts),
[user.guard.ts](../../apps/ai/src/common/user.guard.ts).

At admission, the AI service checks:

```ts
const permissions = await this.authorization.permissions(userId, workspaceId);
if (!permissions.canRead) {
  throw new AiError(403, 'forbidden', 'Workspace access denied.');
}
if (Object.keys(body.toolDefinitions ?? {}).length && !permissions.canWrite) {
  throw new AiError(403, 'forbidden', 'Editing tools require workspace write permission.');
}
```

Each server tool also rechecks read access immediately before its work:

```ts
const checkAccess = async () => {
  deps.abortSignal.throwIfAborted();
  if (!(await deps.authorization.canRead(deps.userId, deps.workspaceId))) {
    throw new AiError(403, 'forbidden', 'Workspace access denied.');
  }
  deps.abortSignal.throwIfAborted();
};
```

The `deps` object contains trusted dependencies passed in by the service: user,
workspace, adapters, and cancellation signal. The model does not choose the acting user.

This is not continuous reauthorization for every streamed word. It checks the request
and each server-tool execution. Browser editing is additionally subject to the document
service's existing Hocuspocus authorization.

The HTTP guard trusts identity headers inserted by Traefik. Deployment must keep that
trust boundary intact; directly exposing the backend to arbitrary forwarded headers
would bypass the intended gateway authentication.

## 6. Task: connect to billing without changing subscription ownership

**Problem:** AI must respect paid allowances, but it should not duplicate checkout,
subscription status rules, or pricing logic.

**Files:** [billing.service.ts](../../apps/ai/src/billing/billing.service.ts),
[billing.module.ts](../../apps/ai/src/billing/billing.module.ts).

The billing service already exposes a gRPC method returning workspace entitlements.
An **entitlement** is simply something a workspace's plan allows.

**Actual code:**

```ts
const response = await firstValueFrom(
  this.client.getWorkspaceEntitlements({ workspaceId }).pipe(timeout(5000))
);
const raw = response.entitlements.limits?.ai_credits_monthly;
const credits = raw === undefined ? 0 : Number(raw);
```

The real method also checks that the response belongs to the requested workspace and
that the quota is a safe integer of at least `-1`.

- Missing quota means zero.
- Zero means no AI credits.
- A positive value is the monthly allowance.
- `-1` means unlimited credits, but request/concurrency limits still apply.

`firstValueFrom()` waits for the gRPC response. `timeout(5000)` bounds that wait to five
seconds. Billing is queried for each generation; the AI service does not cache plan state.

The existing default plans are Free: 0, Plus: 200, Pro: 2,000 credits. Those are defaults,
not hardcoded enforcement values in AI. Billing's response is what AI uses.

If the allowance is exhausted, the AI service also reads `ListPlans` to suggest a larger
plan. If billing is unavailable, it returns 503 rather than assuming the request is free.

**When developing later:** payment-provider status remains billing's responsibility.
AI should consume the resolved allowance, not reinterpret Stripe statuses itself.

## 7. Task: give Redis responsibility for shared counters

**Problem:** memory inside one Node process is private to that process.

```text
AI instance A counts 10 requests.
AI instance B counts 10 requests.
Neither knows about the other's count.
```

Restarting either process also loses its in-memory data. Redis gives both instances a
shared place to read and update counters.

**Files:** [usage.service.ts](../../apps/ai/src/usage/usage.service.ts),
[usage.scripts.ts](../../apps/ai/src/usage/usage.scripts.ts).

These are the key families used by the code:

| Redis key | Purpose |
| --- | --- |
| `ai:attempts:<hashed-user>` | Count parsed requests attempting to use AI |
| `ai:rate:<hashed-user>` | Count successfully admitted reservations in its rate window |
| `ai:streams:<hashed-user>` | Track active requests and their lease expiry times |
| `ai:usage:<workspaceId>:<YYYY-MM>` | Count workspace credits consumed/reserved this month |
| `ai:request:<workspaceId>:<YYYY-MM>:<hash>` | Remember submitted message IDs and their reservation state |

The hashes keep raw user IDs out of these key names. Hashing here is for key construction;
it is not a replacement for securing access to Redis.

No Redis URL, unavailable Redis, or a failed accounting command causes a 503. There is
no silent in-memory replacement for credit accounting.

## 8. Task: prevent concurrent requests from overspending credits

**Problem:** imagine the workspace has one credit remaining and two users submit together.
A naive implementation lets both read “one remaining” before either updates the counter.
Both could start generating.

The implementation uses a Redis Lua script. The relevant read/check/write sequence runs
atomically on the Redis server: another request cannot run its own sequence between those
operations. This does not mean Redis rolls back arbitrary script errors like a SQL transaction.

**Actual excerpt from the reservation script:**

```lua
local used = tonumber(redis.call('GET', KEYS[1]) or '0')
if quota ~= -1 and used >= quota then return 'quota' end
redis.call('INCR', KEYS[1])
redis.call('EXPIREAT', KEYS[1], expiry)
redis.call('SET', KEYS[2], 'reserved:' .. ARGV[7], 'EXAT', expiry)
```

In this excerpt:

- `KEYS[1]` is the monthly workspace counter.
- `KEYS[2]` is the duplicate-request/reservation key.
- `quota` comes from billing.
- `ARGV[7]` is this reservation's unique ownership token.
- `expiry` is when those keys may be removed.

The full script also checks duplicate requests, simultaneous-stream limits, and the
admitted-request rate limit before spending a credit.

**Example:** allowance 200, used 199. Request A increments used to 200. Request B then
sees 200 and receives a quota rejection. Only A receives the last credit.

## 9. Task: define credit reservation and refund behavior

The policy implemented for this version is **one credit per generation**, including
its server-side tool steps. It is not one credit per token or one credit per tool call.

```text
Reserve credit
  │
  ├── confirmed failure before start → refund reservation
  │
  └── mark generation started
         ├── successful answer → keep credit
         ├── model/provider failure → keep credit
         ├── timeout → keep credit
         └── user presses Stop → keep credit
```

**Actual orchestration:**

```ts
const reservation = await this.usage.reserve(
  workspaceId,
  userId,
  body.messages.at(-1)!.id,
  credits
);

await reservation.start();
started = true;
// streamText() is started after this point.
```

The cleanup path calls:

```ts
await reservation.release(!started);
```

`!started` means “refund only if generation has not started.” `release()` also removes
the active-stream lease. It does not erase a started request's duplicate marker.

The start marker is written **before** calling the model. It is a conservative accounting
boundary, not proof that the provider produced output. An uncertain write or process crash
can retain a reservation without a successful answer; automatic crash reconciliation has
not been implemented.

Usage resets by **UTC calendar month**, not the Stripe renewal date. The existing billing
contract does not expose a billing-period start/end to AI. Old counter keys remain for
seven days after the reset, but the new month uses a different counter immediately.

Example: `ai:usage:<workspace>:2026-10` and `ai:usage:<workspace>:2026-11` are different
keys. Changing to the November key is how the allowance resets.

Redis accounting is not a complete financial ledger. Normal container restarts preserve
AOF data; loss of Redis data can lose usage history. Persistence and recovery strategy matter
if this later becomes production billing infrastructure.

## 10. Task: reject duplicates and limit request bursts

A duplicate is identified by the combination of user, workspace, month, and the last user
message's ID. Replaying that message ID returns 409 rather than charging and generating again.

The implementation does **not** replay a stored answer. It rejects the duplicate.
A deliberate new generation needs a new user message ID.

Request frequency and concurrency are separate:

| Setting | Question it answers | Default |
| --- | --- | --- |
| Requests per minute | “How often can this user try?” | 20 |
| Concurrent streams | “How many can this user run at once?” | 2 |
| Monthly workspace credits | “How much can the workspace use this month?” | From billing |

The attempt throttle runs before workspace resolution and billing. Repeated validly shaped
requests therefore cannot freely hammer those downstream services.

**Actual throttle script:**

```lua
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('EXPIRE', KEYS[1], 60) end
if count > tonumber(ARGV[1]) then return 0 end
return 1
```

This is a fixed 60-second window beginning with its first increment, not a sliding-window
algorithm. The reservation script independently limits admitted requests.

Active streams have expiry times, so a crashed process does not block a user's stream slots
forever. Those expiries are tied to the bounded request timeout plus a grace period.

Client IP is not used. The trusted proxy chain has not been configured for reliable IP
attribution; accepting arbitrary `X-Forwarded-For` values would make an IP limit misleading.

## 11. Task: define useful server tools

**File:** [custom-tools.ts](../../apps/ai/src/ai/tools/custom-tools.ts).

A tool has three important parts:

```ts
getNote: tool({
  description:
    'Read note metadata (title, tags, folder), not its content. Returns null for a missing or trashed note.',
  inputSchema: z.object({ noteId: z.uuid() }),
  execute: ({ noteId }) => execute(() => getNote(noteId))(),
}),
```

Read this as:

1. `description`: explain the tool to the model.
2. `inputSchema`: explain and validate the arguments it accepts.
3. `execute`: application code that runs if the model requests the tool.

The outer `execute(...)` is our local permission/error wrapper. It is different from the
`execute:` property expected by the SDK. Both happen to use the same name in this file.

| Tool | What it returns | Important boundary |
| --- | --- | --- |
| `searchNotes` | Note IDs, metadata, and short committed-content excerpts | Current authorized workspace only |
| `getNote` | Title, tags, folder, and ID | Metadata, not document content |
| `getNoteWorkspace` | Workspace details for an accessible current-workspace note | Cannot switch the request's billing workspace |
| `readIndexedNoteContent` | Up to 20,000 characters of indexed committed content | May be stale and may be truncated |
| `listWorkspaces` | Accessible workspace IDs and roles | Does not grant access to their content in the current request |

Tools act with `deps.userId`, supplied by the controller/service. Never add a model argument
such as `actingUserId` that can override the authenticated user.

## 12. Task: make search respect real permissions

The original search adapter accepted an optional workspace filter. Without one it could
search the whole index. The revised adapter requires a validated workspace ID:

```ts
const result = await this.meili.index('notes').search(query, {
  filter: `workspaceId = "${workspaceId}"`,
  limit: Math.min(20, Math.max(1, limit)),
  // The implementation also selects the returned attributes.
});
```

**Files:** [search.service.ts](../../apps/ai/src/search/search.service.ts),
[note.service.ts](../../apps/ai/src/note/note.service.ts).

Search-index data can lag behind the live services. Therefore, before returning a hit,
the tool also uses the note service to verify that the note is accessible, not trashed,
and still belongs to the current workspace.

```ts
const note = await deps.noteClient.getNote({ userId: deps.userId, noteId });
if (!note) return null;
const workspace = await deps.noteClient.getNoteWorkspace(deps.userId, noteId);
if (workspace?.id !== deps.workspaceId) {
  throw new AiError(403, 'forbidden', 'Note is outside this workspace.');
}
return note;
```

Search is limited to 20 results; excerpts are limited to 1,500 characters per result.
A full indexed-content read is limited to 20,000 characters and includes `truncated`
and `mayBeStale` fields.

**Why “committed” matters:**

```text
Edit text in browser
  → live collaborative document updates
  → document is persisted
  → manual or plan-enabled automatic commit creates/updates committed state
  → commit event reaches search-worker
  → search index reflects the new content
```

Collaborative persistence, committing a revision, and updating the search index are
separate stages. This pipeline already existed; I connected the assistant to its indexed
output. I did not add a live-document read API.

## 13. Task: supply tools to the model and stream its answer

**File:** [ai.service.ts](../../apps/ai/src/ai/ai.service.ts).

**Actual code:**

```ts
const clientTools = toolDefinitionsToToolSet(body.toolDefinitions ?? {});
const tools: ToolSet = {
  ...clientTools,
  ...createCustomTools({
    userId,
    workspaceId,
    noteClient: this.noteClient,
    search: this.search,
    authorization: this.authorization,
    abortSignal: signal,
  }),
};
if (Object.keys(clientTools).length) delete tools.listWorkspaces;
```

The source includes a TypeScript type assertion on `body.toolDefinitions`; it is omitted
above to make the flow easier to read. `...` merges the named tools into one object.

This is the current extent of dynamic tool selection: general chat gets the read tools,
and editor requests also supply the allowlisted BlockNote tool while omitting workspace
discovery. It is not yet a general plugin registry or a sophisticated per-task tool planner.

The generation then uses:

```ts
const result = streamText({
  model,
  system: systemPrompt,
  messages,
  tools,
  abortSignal: signal,
  maxOutputTokens: this.config.maxOutputTokens,
  stopWhen: stepCountIs(this.config.maxSteps),
  maxRetries: 0,
});
```

This excerpt substitutes `systemPrompt` for the longer actual expression and omits the
logging callback. The real prompt includes BlockNote HTML guidance, workspace context,
and a reminder to treat note contents as data rather than instructions.

The model may answer directly or request tools. There is no `toolChoice: 'none'` and no
forced-tool setting. Server tool calls can lead to further model steps within the same
generation. The default maximum is 10 steps and 4,096 output tokens **per model step**.
There is not a separate aggregate token budget across all steps.

`maxRetries: 0` disables automatic model-call retries at this layer. The service logs total
reported token usage when the generation finishes; credits are still generation-based.

The result becomes an AI SDK UI-message stream, then SSE over HTTP. A stream can carry
text, tool-call information, tool results, and errors, not just plain strings.

## 14. Task: reuse BlockNote editing inside the existing editor

This is the distinction that caused confusion earlier:

> I connected the existing note editor's BlockNote AI integration to our backend.
> I did not give the standalone workspace assistant the ability to edit notes.

**File:** [editor-core.tsx](../../packages/ui/src/components/editor-core.tsx).

**Actual code, with unrelated editor options omitted:**

```ts
const transport = useAiTransport({ noteId });
const provider = useHocuspocusProvider();

const editor = useCreateBlockNote(
  withCollaboration({
    schema: mySchema,
    collaboration: {
      provider: {
        awareness: provider.awareness ? provider.awareness : undefined,
      },
      fragment: provider.document.getXmlFragment('prosemirror'),
      user: sessionUser,
    },
    extensions: [
      AIExtension({ transport }),
    ],
  })
);
```

The real code also supplies a fallback collaboration user when `sessionUser` is absent.

Think of this as creating a worktable for one particular note. On that table are the
current blocks, your selection, and a connection to the shared document. `AIExtension`
attaches BlockNote's AI editing behavior to that editor instance.

### What an “editor session” means here

“Editor session” is an explanatory term, not an existing class or database table named
`EditorSession` in this implementation. It means the running objects needed to edit a
particular note. It is different from your login session:

```text
Editor for Meeting notes
  ├── noteId: the identity of Meeting notes
  ├── BlockNote editor instance
  ├── current document blocks and block IDs
  ├── cursor and selected text
  ├── Yjs document + Hocuspocus connection
  └── AI editing/review state
```

Yjs represents the shared document. Hocuspocus carries synchronization and document access.
BlockNote supplies the editing UI and operations. These objects work together; a note ID
alone does not contain them.

An editor need not always be visible, but some execution environment must hold or access
the live document state. The current workspace assistant does not create or manage such
an environment for its attachments.

### Where `applyDocumentOperations` actually comes from

BlockNote creates the definition. Inside the installed library, its `streamToolsToToolSet`
helper returns an object shaped like this:

```ts
return {
  applyDocumentOperations: {
    inputSchema: jsonSchema(createStreamToolsArraySchema(streamTools)),
    outputSchema: jsonSchema({ type: 'object' }),
  },
};
```

The library then builds a request from the editor's current state, sends those definitions,
and sets up its browser-side tool-call processing.

The AI backend receives the serialized definition in `body.toolDefinitions`, converts it
with `toolDefinitionsToToolSet()`, and supplies it to the model. That is the backend reuse
point described in Task 13.

Notice that this definition has no server-side `execute` function. Actual document edits
are applied by BlockNote's browser integration. A tool definition is a description and
schema; it is not the running editor or executable document connection.

The existing flow is:

```text
Select a paragraph inside Meeting notes
  → choose a BlockNote AI command
  → BlockNote sends current document/selection + tool definition
  → AI service supplies definition to model
  → model streams applyDocumentOperations arguments
  → BlockNote applies suggested operations to that editor
  → user reviews/accepts/rejects through the existing BlockNote flow
```

The underlying editing machinery already existed in BlockNote and was already used by
the editor integration. I did not copy its implementation into `custom-tools.ts`.
My changes preserve that path, add request context and error handling, and make the backend
validate permissions and the accepted client-tool name.

Do not import BlockNote's private source paths just because they are useful for reading
how it works. Future integrations should prefer its public supported exports.

## 15. Why the standalone assistant cannot edit yet

**File:** [assistant-view.tsx](../../packages/ui/src/components/assistant-view.tsx).

Its chat starts like this:

```ts
const transport = useAiTransport({ workspaceId });
const { messages, sendMessage, status, error, stop, clearError } = useChat({
  id: `workspace-assistant-${workspaceId}`,
  transport,
});
```

There is no BlockNote editor instance here. There is also no registration of BlockNote's
streamed operation executor for this chat.

Attaching a note currently adds a data part and text identifying the note. It helps the
model choose a read tool, but it does not open a live editor or authorize an editing action.

**Example:**

```text
Request: “Shorten the introduction in Meeting notes.”

Inside Meeting notes' editor:
  We already have its live blocks and an operation executor.

Inside workspace assistant:
  We know the workspace and possibly a note ID.
  We do not yet have an editor for that note or a route for applying its edits.
```

Adding a function named `applyDocumentOperations` to the backend alone would not fix this.
We still need to connect the requested operations to the right live document and its
review/synchronization behavior.

This is integration work, not a fundamental limitation. A sidebar beside an existing
editor is the smallest useful next step because the target editor already exists.

## 16. Task: share the AI URL without passing it through every component

Previously, `aiApiUrl` travelled through pages, boundary wrappers, editors, schemas,
reference links, and preview components. Most of those components did not use it directly.

**Files:** [ai-context.tsx](../../packages/ui/src/contexts/ai-context.tsx),
[use-ai-transport.ts](../../packages/ui/src/hooks/use-ai-transport.ts),
[workspace layout](../../apps/web/app/(workspace)/workspace/[workspaceId]/layout.tsx).

The app now supplies the URL once:

```tsx
<AiProvider aiApiUrl={`http://${apiUrl}/ai`}>
  {/* Workspace UI and route content */}
</AiProvider>
```

The provider stores a URL, not a chat or editor session:

```tsx
const AiContext = createContext<string | null>(null);

export function AiProvider({ aiApiUrl, children }: {
  aiApiUrl: string;
  children: ReactNode;
}) {
  return <AiContext.Provider value={aiApiUrl}>{children}</AiContext.Provider>;
}
```

A component below it can create a transport with the context it actually knows:

```ts
// Workspace assistant:
const transport = useAiTransport({ workspaceId });

// Note editor or linked-note preview:
const transport = useAiTransport({ noteId });
```

The hook memoizes the transport so ordinary re-renders do not recreate it unnecessarily.
The provider module is separate from the transport module to avoid pulling the chat SDK
into pages that only need to wrap their content with the provider.

The same URL context does **not** make the assistant and editor share one conversation.
Sharing configuration is different from sharing running editor/chat objects.

The editor sandbox was also updated to supply the provider, because it consumes the same
shared UI components.

## 17. Task: share authentication and normalize frontend errors

**File:** [ai-transport.ts](../../packages/ui/src/lib/ai-transport.ts).

Both interfaces now use the same transport factory:

```ts
export function createAiTransport(api: string, context: {
  workspaceId?: string;
  noteId?: string;
}) {
  return new DefaultChatTransport({
    api,
    body: context,
    headers: async () => ({
      Authorization: `Bearer ${await fetchAccessTokenClientSide()}`,
    }),
    fetch: fetchAi,
  });
}
```

The body supplies context. BlockNote can still add its tool definitions to the request.
The async header function obtains the current access token for each request.

`fetchAi` passes successful responses through unchanged. That matters because reading the
whole body as JSON would destroy a successful streaming response.

For an unsuccessful response, it reads the error JSON and throws `AiRequestError`, which
extends the application's existing `AppError`. It retains the HTTP status and error code
while choosing a readable message.

**Example of what the user sees:**

```text
HTTP 402 → This workspace has no AI credits available. Credits reset on …
HTTP 429 → Too many AI requests. Wait a minute, then try again.
HTTP 503 → AI is temporarily unavailable. Please try again later.
```

An HTML gateway error is not rendered as raw HTML/text. If no useful JSON is available,
the UI uses a generic request-failed message.

## 18. Task: distinguish HTTP failures from streaming failures

**Files:** [ai-error.ts](../../apps/ai/src/common/ai-error.ts),
[ai.controller.ts](../../apps/ai/src/ai/ai.controller.ts).

Before streaming starts, the server can choose an HTTP status and JSON body.
After it has sent the response headers, it cannot change a 200 response into a 503.
It must send an error inside the already-open stream instead.

| Situation | Representation |
| --- | --- |
| Invalid input before generation | HTTP 400 + JSON |
| Missing authentication | HTTP 401 + JSON |
| No credits | HTTP 402 + JSON |
| No permission | HTTP 403 + JSON |
| Duplicate message ID | HTTP 409 + JSON |
| Oversized request | HTTP 413 + JSON |
| Rate/concurrency limit | HTTP 429 + JSON, with `Retry-After` |
| Redis/billing unavailable | HTTP 503 + JSON |
| Timeout detected before response streaming | HTTP 504 + JSON |
| Model error after headers | AI SDK error part inside the stream |
| Server tool fails | Sanitized tool failure through the SDK's tool-result handling |

**Illustration of a plan error:**

```json
{
  "code": "planLimitExceeded",
  "message": "Monthly AI credits are exhausted.",
  "details": {
    "kind": "limit",
    "key": "ai_credits_monthly",
    "current": 200,
    "max": 200,
    "requiredPlan": "pro"
  },
  "resetAt": "2026-11-01T00:00:00.000Z"
}
```

**Actual model-stream error customization:**

```ts
result.toUIMessageStream({
  onError: (error) => {
    this.logger.error({ err: error, workspaceId }, 'AI stream failed');
    return 'The AI generation failed. Please try again with a new message.';
  },
});
```

Private error details go to server logs; the stream gets a safe message. Our tool wrapper
similarly logs the original failure and throws a sanitized message for SDK serialization.

## 19. Task: cancel work and release resources

**Problem:** a request may end because it finished, the user pressed Stop, the browser
closed, the model failed, or the timeout expired. Every exit path needs cleanup.

The controller creates a cancellation signal:

```ts
const abort = new AbortController();
const timer = setTimeout(
  () => abort.abort(new AiError(504, 'aiRequestTimeout', 'The AI request timed out.')),
  this.config.timeoutMs
);
```

It passes `abort.signal` to the AI service, which passes the signal to the model request.
The controller also listens for response close/finish events and clears listeners/timers
when it is done.

The service wraps the outgoing stream so that completion, stream failure, or cancellation
releases its Redis active-stream lease. Cleanup errors are logged. A cancelled stream also
cancels its reader.

The timeout bounds the generation and is passed through setup, but this is not a guarantee
that every dependency stops at the exact same millisecond. gRPC/search calls also have their
own bounded timeouts. Background accounting is not a distributed transaction with the model.

**Actual UI button configuration:**

```tsx
<PromptInputSubmit
  status={status}
  onStop={() => void stop()}
  aria-label={isTyping ? 'Stop generating' : 'Send message'}
  disabled={!isTyping && !input.trim() && attachedNotes.length === 0}
/>
```

The button stays enabled while generating so the user can actually stop it.

## 20. Task: make frontend state and feedback understandable

### Tool activity

The assistant previously rendered text and attached-note chips. It now recognizes tool
parts too, displaying labels such as “Search notes · Working…” and “Read committed note ·
Complete.” Tool failures are also visible.

This is feedback about server read tools. Showing tool activity does not mean the assistant
has gained document-editing powers.

### Prompt recovery

Before submission, the assistant saves the current prompt text and selected attachments in
component state. An error exposes an **Edit prompt** button that restores those values.
Submitting again creates a new message ID, matching the backend's duplicate policy.

This is local recovery within the mounted component, not persisted conversation history.
Reloading the page does not restore a server-side saved chat.

### Editor error feedback

[editor-ai-error.tsx](../../packages/ui/src/components/editor-ai-error.tsx) subscribes to
BlockNote's AI state. When the state is `error`, it displays the actual readable error
message beside the editor, rather than relying only on BlockNote's generic status label.

```ts
const ai = editor.getExtension(AIExtension);
const getSnapshot = useCallback(
  () => ai?.store.state.aiMenuState ?? 'closed',
  [ai]
);
```

The component uses `useSyncExternalStore` with a stable subscription and a server snapshot
of `closed`. This is React's mechanism for subscribing to a store outside React state.

### Switching notes/workspaces

The workspace assistant is keyed by workspace ID, and the editor is keyed by note ID.
A key change remounts that component instead of keeping the previous note's editor/chat
objects. Linked-note previews also key their editor by the target note ID.

```tsx
<AssistantView key={workspaceId} workspaceId={workspaceId} />

<Editor
  key={noteId}
  noteId={noteId}
  workspaceId={workspaceId}
  apiUrl={apiUrl}
  onNavigate={(href) => router.push(href)}
/>
```

The dynamic editor wrapper was given an explicit props type so its generated declaration
tracks the editor's current props when building the shared UI package.

### Commit changes

The existing commit action is now a labeled **Commit changes** button with a pending state.
It is hidden for viewers. The action still calls the existing document commit API; this
work did not rewrite that service's commit or indexing pipeline.

```tsx
<Button variant="outline" onClick={handleSave} disabled={isCommitingDocument}>
  {isCommitingDocument ? <Spinner /> : <Icons.Save />}
  {isCommitingDocument ? 'Committing…' : 'Commit changes'}
</Button>
```

The spelling `isCommitingDocument` is the current existing identifier. The important idea
is that collaborative saving and committing searchable content are separate actions.

## 21. Task: make local infrastructure available

**Files:** [AI Redis](../../deploy/compose/services/ai.yaml),
[optional Ollama](../../deploy/compose/infrastructure/ollama.yaml),
[Compose entrypoint](../../deploy/compose/compose.yaml).

AI Redis has its own named volume and localhost port 6376:

```yaml
ai_redis:
  image: redis:8.4.7-alpine3.22
  command: ["redis-server", "--appendonly", "yes", "--maxmemory-policy", "noeviction"]
  ports:
    - "127.0.0.1:${AI_REDIS_PORT:-6376}:6379"
  volumes:
    - ai_redis_data:/data
```

`appendonly yes` enables Redis's append-only persistence. `noeviction` prevents Redis
from choosing usage keys to discard under memory pressure. The named volume survives normal
container recreation. This is not a promise of zero data loss under every crash scenario.

The Redis file lives in `services/` because the existing repository puts per-service
Redis dependencies there. The optional shared model server lives in `infrastructure/`.

Ollama is optional and uses the `ai-local-llm` profile. It is CPU-only by default, persists
models in its own volume, and listens on localhost:11434. It is not started by an ordinary
Compose `up` unless the profile is enabled or that service is explicitly targeted.

The default Ollama image is `latest` for this development helper. `AI_OLLAMA_IMAGE` allows
pinning a chosen version. The model download is a separate explicit command.

AI Redis is the required new dependency. A different OpenAI-compatible endpoint may replace
Ollama. The AI service itself still runs on the host through Nx in this local setup.

## 22. Configuration you will likely change later

| Variable | Default/current local value | Meaning |
| --- | --- | --- |
| `NOTOPIA_AI_PORT` | `8085` | AI HTTP server port |
| `NOTOPIA_AI_SERVICES_BILLING_GRPC_URL` | `127.0.0.1:18084` | Existing billing gRPC server |
| `NOTOPIA_AI_REDIS_URL` | `redis://127.0.0.1:6376` | Shared usage and throttle storage |
| `NOTOPIA_AI_LLM_BASE_URL` | `http://localhost:11434/v1` | Model's OpenAI-compatible endpoint |
| `NOTOPIA_AI_LLM_MODEL` | `qwen3:8b` | Model name sent to that endpoint |
| `NOTOPIA_AI_REQUESTS_PER_MINUTE` | `20` | Per-user request limit |
| `NOTOPIA_AI_CONCURRENT_STREAMS` | `2` | Per-user simultaneous generation limit |
| `NOTOPIA_AI_TIMEOUT_MS` | `120000` | Request timeout in milliseconds |
| `NOTOPIA_AI_MAX_OUTPUT_TOKENS` | `4096` | Maximum model output tokens per step |
| `NOTOPIA_AI_MAX_STEPS` | `10` | Maximum automatic model/tool steps |

Schemas live in [config.ts](../../apps/ai/src/config.ts), environment validation in
[env.validation.ts](../../apps/ai/src/env.validation.ts), and environment-to-config mapping in
[config.factory.ts](../../apps/ai/src/config.factory.ts).

One second is 1,000 milliseconds. A token is a unit the model uses to process text; it is
not necessarily a word. A step is one model-generation round, which can include tool calls.

Local secret overrides belong in ignored `.env.local` files. No model key should be
embedded into a UI component or sent as a browser-side model-provider credential.

## 23. What was tested, and what those tests do not prove

The implementation session passed **38 AI tests** and **10 UI transport tests**. It also
passed the AI build/typecheck/lint, UI and sandbox typechecks, a direct web TypeScript
check, and UI/web lint with existing warnings. Compose configuration validation passed.
These are previous validation results; writing this guide did not rerun the full stack.

| Test file | What it checks |
| --- | --- |
| [request.spec.ts](../../apps/ai/tests/request.spec.ts) | Valid/invalid message shapes, forbidden file parts, oversized requests |
| [access.spec.ts](../../apps/ai/tests/access.spec.ts) | Workspace selection, conflicting context, viewer restrictions, revoked access, stale hits |
| [billing.spec.ts](../../apps/ai/tests/billing.spec.ts) | Uncached quota reads, invalid responses, missing/unlimited quotas, billing failure, upgrade lookup |
| [usage.spec.ts](../../apps/ai/tests/usage.spec.ts) | Calendar periods, missing Redis, real Lua concurrency, duplicates, refunds, stream leases, throttling |
| [stream.spec.ts](../../apps/ai/tests/stream.spec.ts) | Actual SDK behavior with a mock model: multi-step tool calls, safe errors, cancellation, timeout cleanup |
| [ai-transport.spec.ts](../../packages/ui/src/lib/ai-transport.spec.ts) | Request context, auth header, BlockNote request-body merging, SSE passthrough, readable HTTP errors |

The model in streaming tests is mocked. These tests do not prove that a particular real
model reliably follows your editing instructions. UI tests exercise the transport, not
complete rendered browser workflows. Real Redis cases require an explicitly supplied test
Redis URL and are skipped without it.

Full manual testing with Authentik, Traefik, all applications, live collaborative editing,
and a real model remains necessary. See [AI_TESTING.md](../../deploy/compose/AI_TESTING.md).

### Repeat the automated checks

From the repository root:

```sh
nx test packages/ui
nx test ai
```

To include real Redis tests, create an isolated temporary Redis:

```sh
docker run --rm -d --name notopia-ai-test-redis \
  -p 127.0.0.1:16385:6379 redis:8.4.7-alpine3.22

AI_TEST_REDIS_URL=redis://127.0.0.1:16385 nx test ai --skip-nx-cache

docker stop notopia-ai-test-redis
```

The test namespace is random and the tests remove their own keys. Keep that test instance
separate from Redis holding your manual-test usage counters.

### Small manual acceptance checklist

1. Start the Compose infrastructure and an available model endpoint.
2. Start note, authorization, billing, document, search-worker, AI, and web.
3. Sign in with membership in a populated workspace with AI credits.
4. Ask the assistant to find and read a known committed note.
5. Confirm its Network request includes `workspaceId`.
6. Open a note and run a BlockNote AI command; confirm the request includes `noteId`.
7. Review the suggested editor changes. Commit them and wait for search indexing.
8. Ask the assistant about the new content.
9. Stop a generation and confirm another request can start.
10. Use a Free workspace to verify a readable 402 error.
11. Replay a request with the same message ID to verify 409 and no second debit.
12. Try an inaccessible workspace to verify 403.

The current generation-based credit policy means manual test generations consume credits,
including requests stopped after the start marker has been recorded.

## 24. A practical development path for assistant editing

Everything in this section is a **future proposal**, not completed functionality.

### Phase A: put the assistant beside an already-open editor

Start with this user story:

> “While viewing Meeting notes, I can ask the sidebar to shorten the introduction.”

The page already owns the correct live editor. Expose that instance to a sidebar through
a React context or an explicit prop. Keep the selected-note identity visible to the user.

```text
Note page
  ├── EditorCore → creates editor for Meeting notes
  └── Assistant sidebar → receives that same editor instance
```

The missing integration must then:

1. Build the AI request from that editor's current document/selection.
2. Include the editor tool definitions alongside the assistant conversation.
3. Register the operation-processing bridge for that conversation.
4. Route streamed operations back to that editor instance.
5. Expose review, accept/reject, cancellation, and errors in the sidebar.

BlockNote documents `chatProvider`, `buildAIRequest`, and `sendMessageWithAIRequest`
for sharing/customizing its chat and request flow. Check their installed public exports
and version-specific behavior before implementing the bridge.

A small prototype could invoke the existing editor AI directly:

```ts
// Future proposal: not currently called by AssistantView.
const ai = editor.getExtension(AIExtension);
if (!ai) throw new Error('This editor does not have AI enabled.');

await ai.invokeAI({
  userPrompt: 'Shorten the introduction without removing the action items.',
  useSelection: false,
});
```

That reuses the editor's AI flow, but it is not automatically the same conversation as the
workspace assistant. Invoking it from another model tool could create a second model request
and an additional credit charge. Also, completion of `invokeAI()` is not proof that the user
has accepted the changes. A proper shared-chat design needs to coordinate those states.

First acceptance test: the sidebar modifies the displayed note, the user can reject the
suggestion, and an unrelated open note remains unchanged.

### Phase B: support notes that are not already open

Only after Phase A works should the assistant manage arbitrary target notes.

**Illustration of a possible future structure:**

```ts
// Design sketch only. These types/functions do not exist yet.
const sessions = new Map<NoteId, ManagedEditorSession>();

// A managed session would own:
// - note identity and current permissions
// - live document connection and editor instance
// - AI editing/review state
// - cleanup when closed, cancelled, or no longer needed
```

Load a session only when needed. Do not create an editor for every note in a workspace at
startup. Before editing, confirm the target note/workspace, write permission, and the
workspace that pays for the request. Do not let model arguments silently change that context.

You must also decide what happens when the note is already open in another component,
when the user changes the target mid-generation, and when edits are still awaiting review.
Concurrent human edits mean that stale search-index block positions are not a safe editing
basis. Use the live document's current state.

First acceptance test: the assistant opens an unopened note, proposes an edit, shows the
correct target for review, and releases its session when finished.

### Phase C: extend backend tools through proper contracts

Possible future tools include workspace trees, note links, live document reads, and explicit
commit actions. The current note gRPC contract does not expose every existing HTTP use case,
and the AI service does not have a live document-content RPC.

Add suitable internal contracts and authorization checks before adding those tools. Keep
the repository rule: internal service calls use gRPC rather than another service's public
REST endpoint. Do not make AI write directly into another service's database.

### Phase D: improve the product policies separately

Useful later decisions include:

- Whether one credit should remain one generation or reflect token/model cost.
- Whether allowances reset by calendar month or the actual billing renewal period.
- Whether failed provider calls should be refunded, with reliable reconciliation.
- Whether conversations and responses should be persisted for safe retry/replay.
- Whether a durable usage ledger is needed in addition to Redis.
- Whether tools and system prompts should differ more substantially between chat and editor.
- Whether IP throttling can be added after the proxy trust configuration is explicit.
- Whether upgrade hints should explicitly represent “no higher plan exists.” The current
  `requiredPlan()` falls back to `pro`, which is not a useful upgrade suggestion when Pro
  itself has exhausted its credits. The current UI does not display that hint as an upgrade link.

These are decisions to make, not hidden completed features.

## 25. Where to make future changes

| You want to change… | Start here |
| --- | --- |
| Accepted request shape | `apps/ai/src/ai/chat-request.dto.ts` |
| Request lifecycle, timeout, response piping | `apps/ai/src/ai/ai.controller.ts` |
| Workspace resolution, tool selection, model call | `apps/ai/src/ai/ai.service.ts` |
| Tool descriptions, input schemas, and execution | `apps/ai/src/ai/tools/custom-tools.ts` |
| Credit allowance/upgrade lookup | `apps/ai/src/billing/billing.service.ts` |
| Shared credit/rate/concurrency state | `apps/ai/src/usage/usage.service.ts` and `usage.scripts.ts` |
| Search filtering and content bounds | `apps/ai/src/search/search.service.ts` |
| Permission RPC adapters | `apps/ai/src/authorization/authorization.service.ts` |
| Note metadata/workspace RPC adapters | `apps/ai/src/note/note.service.ts` |
| Backend public errors | `apps/ai/src/common/ai-error.ts` |
| AI URL provider | `packages/ui/src/contexts/ai-context.tsx` |
| Memoized context-aware transport hook | `packages/ui/src/hooks/use-ai-transport.ts` |
| Browser authentication/error transport | `packages/ui/src/lib/ai-transport.ts` |
| Workspace conversation and attachments | `packages/ui/src/components/assistant-view.tsx` |
| Live editor and AI extension | `packages/ui/src/components/editor-core.tsx` |
| Editor AI error display | `packages/ui/src/components/editor-ai-error.tsx` |
| Commit button | `packages/ui/src/components/editor.tsx` |
| URL supplied by the web application | `apps/web/app/(workspace)/workspace/[workspaceId]/layout.tsx` |
| Linked-note editor previews | `packages/ui/src/block-note/reference.tsx` |
| Local Redis/model server | `deploy/compose/services/ai.yaml`, `infrastructure/ollama.yaml` |

Read the scoped `AGENTS.md` before editing an area. Billing, note, document, and authorization
service implementation changes were not part of this AI implementation. Their existing APIs
and behavior were consumed through AI-local adapters.

## 26. Quick troubleshooting reference

| What you observe | What to inspect first |
| --- | --- |
| `400 workspaceRequired` | Browser request is missing workspace/note context and the user has multiple workspaces |
| `402 planLimitExceeded` | Workspace's billing allowance and current month's Redis counter |
| `403` | Membership, read/write permissions, and whether note/workspace context agrees |
| `409 duplicateAiRequest` | The last user message ID was already submitted this month |
| `429` | Per-user attempt/admission windows or active-stream slots |
| `503 aiUsageUnavailable` | AI Redis URL, connection, and accounting commands |
| `503 billingServiceUnavailable` | Billing gRPC endpoint and entitlement response |
| Answer omits recently edited text | Whether the document was committed and search-worker indexed it |
| Assistant says something but the note does not change | Standalone assistant editing has not been implemented |
| Editor AI returns an operation but the wrong editor seems active | Editor lifetime, note ID, and shared-chat/executor binding |
| All automated tests pass but a real model edits poorly | Model/tool-following behavior and actual document context need manual evaluation |

## Further reading

The repository code and installed package source are the primary references for this guide.
These upstream references explain the library concepts used here:

- [AI SDK tool calling](https://ai-sdk.dev/docs/ai-sdk-core/tools-and-tool-calling)
- [AI SDK UI error handling](https://ai-sdk.dev/docs/ai-sdk-ui/error-handling)
- [BlockNote AI reference](https://github.com/typecellos/blocknote/blob/main/docs/content/docs/features/ai/reference.mdx)
- [React useSyncExternalStore](https://react.dev/reference/react/useSyncExternalStore)
- [ioredis documentation](https://github.com/redis/ioredis)

Use the documentation matching the installed versions before changing API calls. This
repository currently integrates AI SDK v6 with BlockNote `0.55.0`; upgrading those packages
may require revisiting the request, streaming, and editor bridges together.
