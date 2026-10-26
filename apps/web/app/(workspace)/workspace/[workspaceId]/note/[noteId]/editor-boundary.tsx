'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import { Editor } from '@ui/components/dynamic-editor';
import { useRouter } from 'next/navigation';

export function EditorBoundary({
  noteId,
  workspaceId,
  apiUrl,
}: {
  noteId: string;
  workspaceId: string;
  apiUrl: string;
}) {
  const router = useRouter();
  return (
    <ErrorBoundary fallbackTitle="Editor Error">
      <Editor
        key={noteId}
        noteId={noteId}
        workspaceId={workspaceId}
        apiUrl={apiUrl}
        onNavigate={(href) => router.push(href)}
      />
    </ErrorBoundary>
  );
}
