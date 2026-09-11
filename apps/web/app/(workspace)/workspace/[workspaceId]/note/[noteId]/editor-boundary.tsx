'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import { Editor } from '@ui/components/dynamic-editor';

export function EditorBoundary({
  noteId,
  workspaceId,
}: {
  noteId: string;
  workspaceId: string;
}) {
  return (
    <ErrorBoundary fallbackTitle="Editor Error">
      <Editor noteId={noteId} workspaceId={workspaceId} />
    </ErrorBoundary>
  );
}
