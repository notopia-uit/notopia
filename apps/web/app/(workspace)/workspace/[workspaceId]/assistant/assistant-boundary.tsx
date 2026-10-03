'use client';

import { AssistantView } from '@notopia-uit/ui/components/assistant-view';
import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';

export function AssistantBoundary({ workspaceId }: { workspaceId: string }) {
  return (
    <ErrorBoundary fallbackTitle="Assistant Error">
      <AssistantView workspaceId={workspaceId} />
    </ErrorBoundary>
  );
}
