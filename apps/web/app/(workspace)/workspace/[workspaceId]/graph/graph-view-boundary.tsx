'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import GraphView from '@ui/components/graph-view';

export function GraphViewBoundary({
  workspaceId,
  initialData,
}: {
  workspaceId: string;
  initialData?: unknown;
}) {
  return (
    <ErrorBoundary fallbackTitle="Graph Error">
      <GraphView workspaceId={workspaceId} initialData={initialData} />
    </ErrorBoundary>
  );
}
