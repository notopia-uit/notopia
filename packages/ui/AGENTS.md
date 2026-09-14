# UI Package (`packages/ui`)

Shared React components and utilities for `@notopia-uit/ui`. Uses shadcn/ui + Tailwind CSS.

## Error Handling

### Components

| File | Export | Purpose |
|---|---|---|
| `components/error-boundary.tsx` | `ErrorBoundary` | Class-based React error boundary with retry. Props: `fallbackTitle`, `fallback` (render prop), `retry` (callback). Default UI shows error badge + retry button. |
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
