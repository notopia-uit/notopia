# Web App (`apps/web`)

Next.js 16 App Router. The main user-facing application.

## Error Handling

### Route Segments

| Segment | `error.tsx` | `not-found.tsx` | Notes |
|---|---|---|---|
| `app/` (root) | `error.tsx` | `not-found.tsx` | Root fallback |
| `app/` (global) | `global-error.tsx` | — | Catches errors outside root layout |
| `(workspace)/` | — | `not-found.tsx` | Workspace group |
| `(workspace)/workspace/` | pre-existing `error.tsx` | — | Workspace list |
| `(workspace)/workspace/[workspaceId]/` | `error.tsx` | `not-found.tsx` | Workspace detail |
| `(auth)/signin/` | — | — | Uses auth provider error handling |
| `(marketing)/` | — | — | Static pages, low risk |

### Component-Level Boundaries

Complex client components are wrapped with `ErrorBoundary` via thin `'use client'` wrapper files:

| Component | Page | Wrapper |
|---|---|---|
| `GraphView` | `workspace/[workspaceId]/graph/page.tsx` | `graph-view-boundary.tsx` |
| `Editor` | `workspace/[workspaceId]/note/[noteId]/page.tsx` | `editor-boundary.tsx` |
| `LocalNoteGraphView` | `workspace/[workspaceId]/note/[noteId]/graph/page.tsx` | `note-graph-view-boundary.tsx` |

**Pattern**: Server component pages import a client boundary wrapper instead of the component directly. The wrapper renders `<ErrorBoundary fallbackTitle="..."><Component /></ErrorBoundary>`.

### Server-Side Prefetching

All `queryClient.prefetchQuery()` calls are wrapped with `try/catch → notFound()`. This includes:
- `workspace/[workspaceId]/layout.tsx` (workspace + tree prefetch)
- `workspace/[workspaceId]/page.tsx`
- `workspace/page.tsx`
- `workspace/[workspaceId]/graph/page.tsx`

**Never leave `Promise.all` prefetch calls unprotected** — a rejected promise will crash the layout with no error UI.

### Auth Token

`lib/get-access-token.ts` calls `notFound()` instead of throwing when the token is missing. This triggers the nearest `not-found.tsx` boundary rather than crashing.

### Adding New Routes

When adding a new route segment:
1. Add `not-found.tsx` if the route has dynamic params
2. Add `error.tsx` if the segment does heavy data fetching or renders complex client components
3. Wrap complex client components with the `ErrorBoundary` pattern (create a `*-boundary.tsx` wrapper)
4. Wrap `prefetchQuery` calls with `try/catch → notFound()`

### Imports

- `ErrorBoundary` → `@notopia-uit/ui/components/error-boundary`
- `AppError`, `toAppError`, `is404` → `@notopia-uit/ui` (re-exported from `lib/errors`)
- `NotFound` → `@notopia-uit/ui/components/not-found`
