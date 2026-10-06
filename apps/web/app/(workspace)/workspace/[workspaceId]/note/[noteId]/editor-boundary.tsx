'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import { Editor } from '@ui/components/dynamic-editor';
import { useRouter } from 'next/navigation';

export function EditorBoundary({
  noteId,
  workspaceId,
  aiApiUrl,
  apiUrl,
}: {
  noteId: string;
  workspaceId: string;
  aiApiUrl: string;
  apiUrl: string;
}) {
  const router = useRouter();
  return (
    <ErrorBoundary fallbackTitle="Editor Error">
      <Editor
        noteId={noteId}
        workspaceId={workspaceId}
        aiApiUrl={aiApiUrl}
        apiUrl={apiUrl}
        onNavigate={(href) => router.push(href)}
      />
    </ErrorBoundary>
  );
}
