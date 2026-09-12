'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import LocalNoteGraphView from '@notopia-uit/ui/components/local-note-graph-view';
import { useRouter } from 'next/navigation';

export function LocalNoteGraphViewBoundary({
  noteId,
  workspaceId,
}: {
  noteId: string;
  workspaceId: string;
}) {
  const router = useRouter();
  return (
    <ErrorBoundary fallbackTitle="Graph Error">
      <LocalNoteGraphView noteId={noteId} workspaceId={workspaceId} onNavigate={(href) => router.push(href)} />
    </ErrorBoundary>
  );
}
