'use client';

import { AssistantView } from '@notopia-uit/ui/components/assistant-view';
import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';

export function AssistantBoundary({
  workspaceId,
  apiUrl,
  documentWsUrl,
}: {
  workspaceId: string;
  apiUrl: string;
  documentWsUrl: string;
}) {
  return (
    <ErrorBoundary fallbackTitle="Assistant Error">
      <div className="h-[calc(100dvh-4rem)] min-h-0">
        <AssistantView
          key={workspaceId}
          workspaceId={workspaceId}
          apiUrl={apiUrl}
          documentWsUrl={documentWsUrl}
        />
      </div>
    </ErrorBoundary>
  );
}
