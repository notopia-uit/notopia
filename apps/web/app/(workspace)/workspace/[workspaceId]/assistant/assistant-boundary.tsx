'use client';

import { AssistantView } from '@notopia-uit/ui/components/assistant-view';
import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';

export function AssistantBoundary({
  workspaceId,
  aiApiUrl,
}: {
  workspaceId: string;
  aiApiUrl: string;
}) {
  return (
    <ErrorBoundary fallbackTitle="Assistant Error">
      <AssistantView workspaceId={workspaceId} aiApiUrl={aiApiUrl} />
    </ErrorBoundary>
  );
}
