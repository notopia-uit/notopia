# AI tools

Tools live in `src/ai/tools/custom-tools.ts`. `createCustomTools` receives authenticated
`userId`, the resolved billing `workspaceId`, service adapters, and the request abort signal.
Never accept the acting user or billing workspace from model arguments.

Existing tools:

| Tool | Data source |
| --- | --- |
| `listWorkspaces` | Authorization gRPC; general chat only |
| `searchNotes` | Workspace-filtered Meilisearch, followed by live note checks |
| `getNote` | User-aware note metadata gRPC; not document content |
| `getNoteWorkspace` | User-aware note gRPC; restricted to the active workspace |
| `readIndexedNoteContent` | Last committed search content, at most 20,000 characters, marked potentially stale |

Use `tool()` from `ai`, bounded Zod schemas and the existing execution wrapper, which rechecks
read permission and sanitizes failures. Return `null` for missing/trashed records. Check the
note's live workspace before returning content; search data alone is not authoritative.
New data sources belong in top-level NestJS modules. Internal service calls use existing
gRPC contracts; adding a tool must not silently add dependencies on public REST endpoints.

The AI service merges server tools with the allowlisted BlockNote client editing tool.
Only `applyDocumentOperations` definitions are accepted and only for users with write
permission. These operations execute in the browser. Reuse BlockNote's exported integration
rather than copying its internal implementation. Additional editor tools, live document reads,
workspace trees and graphs require separate integration work when suitable contracts exist.

Tool choice remains automatic. `NOTOPIA_AI_MAX_STEPS` bounds the server-side loop; all steps
within one accepted generation consume one credit. A tool failure must not expose raw gRPC,
Redis, search-provider or model-provider errors to the client/model. Log details server-side.
