'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import GraphView from '@ui/components/graph-view';

export function GraphViewBoundary({
  workspaceId,
}: {
  workspaceId: string;
}) {
  return (
    <ErrorBoundary fallbackTitle="Graph Error">
      <GraphView workspaceId={workspaceId} />
    </ErrorBoundary>
  );
}
