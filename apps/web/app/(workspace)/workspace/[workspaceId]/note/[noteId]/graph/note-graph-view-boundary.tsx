'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import LocalNoteGraphView from '@notopia-uit/ui/components/local-note-graph-view';

export function LocalNoteGraphViewBoundary({
  noteId,
  workspaceId,
}: {
  noteId: string;
  workspaceId: string;
}) {
  return (
    <ErrorBoundary fallbackTitle="Graph Error">
      <LocalNoteGraphView noteId={noteId} workspaceId={workspaceId} />
    </ErrorBoundary>
  );
}
