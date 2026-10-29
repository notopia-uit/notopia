# UI Package (`packages/ui`)

Shared React components and utilities for `@notopia-uit/ui`. Uses shadcn/ui + Tailwind CSS.

## Error Handling

### Components

| File | Export | Purpose |
|---|---|---|
| `components/error-boundary.tsx` | `ErrorBoundary` | Class-based React error boundary with retry. Props: `fallbackTitle`, `fallback` (render prop) Default UI shows error badge + retry button. |
| `components/not-found.tsx` | `NotFound` | Reusable 404 page. Props: `resource` (optional, e.g. "Note", "Workspace"). Shows 404 badge, title, message, and "Go Home" link. |
| `components/error-alert.tsx` | `ErrorAlert` | Static destructive alert box for inline error display. |
| `components/error-button.tsx` | `ErrorButton` | 404 page with "Go Back Home" link button. |
| `hooks/query-error-fallback.tsx` | `QueryErrorFallback` | React Query error fallback with retry. Used in 6+ client components. |

### Utilities

| File | Export | Purpose |
|---|---|---|
| `lib/errors.ts` | `AppError` | Base error class extending `Error`. Properties: `code`, `moreInfo`, `statusCode`. |
| `lib/errors.ts` | `NoteServiceError` | Note service API error subclass. |
| `lib/errors.ts` | `DocumentServiceError` | Document service API error subclass. |
| `lib/errors.ts` | `UnauthorizedError` | 401 unauthorized error subclass. |
| `lib/errors.ts` | `NotFoundError` | 404 not found error subclass. |
| `lib/errors.ts` | `toAppError(err)` | Normalizes any caught value into typed `AppError`. |
| `lib/errors.ts` | `is404(err)` | Guard to detect 404 errors. |
| `lib/errors.ts` | `GeneratedError` | Interface matching codegen error shapes (`{ code, message, more_info? }`). |

### Architecture

```
Caught value (unknown)
    │
    ▼
toAppError(err)  ──→  AppError (typed)
    │                      │
    │              ┌───────┴───────┐
    │              │               │
    ▼              ▼               ▼
is404()       statusCode       code
check         switch           match
```

### Adding New Error Subclasses

1. Add the class to `lib/errors.ts` extending `AppError`
2. Update `toAppError()` to detect and instantiate it (match on status code or error shape)
3. Export from `lib/errors.ts` (already re-exported via `src/index.ts`)

### Pattern: Client Component Boundary Wrappers

Since Next.js App Router pages are server components, use this pattern to add error boundaries around client components:

```tsx
// my-component-boundary.tsx
'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import MyComponent from '@notopia-uit/ui/components/my-component';

export function MyComponentBoundary(props: Props) {
  return (
    <ErrorBoundary fallbackTitle="Component Error">
      <MyComponent {...props} />
    </ErrorBoundary>
  );
}
```

Then import the boundary wrapper in the server component page instead of the component directly.

### Avoid access process.env here

We will pass the from the consumer packages like `apps/web` as props. This is guarded by oxlint

## AI integration

`AiProvider` (`contexts/ai-context.tsx`) receives the AI URL from the consuming app. Use
`useAiTransport({ workspaceId })` for workspace chat or `useAiTransport({ noteId })` for an
editor, including linked-note previews. Do not pass the AI URL through editor/schema/reference
props or infer a billing workspace from prompt text. The web workspace layout owns the provider.
Other consumers, such as the editor sandbox, must wrap AI-enabled editor previews too.

`lib/ai-transport.ts` shares authentication and HTTP error normalization between useChat and
BlockNote. Successful SSE responses pass through untouched. `EditorAiError` subscribes to the
BlockNote AI store to display specific errors beside the editor. Workspace chat can restore a
failed prompt and stop generation; resubmitting creates a new message ID, because the backend
rejects duplicate generation IDs. Keep chat/editor state scoped to its workspace/note.

`nx test packages/ui` verifies transport context, BlockNote tool-body merging, SSE preservation,
and readable HTTP errors. Commit actions remain separate from collaborative persistence:
committing updates the content available to search and the assistant.

Workspace note editing uses `useAssistantChat`, `AssistantRunner`, and one shared Chat.
`AssistantEditorSession` gates execution on an authenticated, initially synced Hocuspocus
room. The assistant transport sends server-issued continuation tokens and matching results,
never replacement history. `AssistantAiBridge` uses BlockNote's exported full-document
request/executor helpers and fork-based accept/reject; do not implement snapshot rollback.
Its guarded, delay-free executors prevent queued operations from running after Stop/review.
Resolve suggestions before another prompt or Commit. API/WebSocket URLs come from web props.
The deterministic real-editor browser fixture lives in `tests/browser/`; setup and evidence
limits are documented in `docs/ai/ASSISTANT_NOTE_EDITING.md`.
