'use client';

import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';
import { Editor } from '@ui/components/dynamic-editor';
import { useRouter } from 'next/navigation';

export function EditorBoundary({
  noteId,
  workspaceId,
}: {
  noteId: string;
  workspaceId: string;
}) {
  const router = useRouter();
  return (
    <ErrorBoundary fallbackTitle="Editor Error">
      <Editor noteId={noteId} workspaceId={workspaceId} onNavigate={(href) => router.push(href)} />
    </ErrorBoundary>
  );
}
