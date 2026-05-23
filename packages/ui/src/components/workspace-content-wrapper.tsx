'use client';

import { EditorSearchProvider } from '@notopia-uit/ui/components/editor-search-provider';

import { NoteSearchModal } from './note-search-modal';

interface WorkspaceContentWrapperProps {
  children: React.ReactNode;
  workspaceId: string;
}

export function WorkspaceContentWrapper({
  children,
  workspaceId,
}: WorkspaceContentWrapperProps) {
  return (
    <EditorSearchProvider
      workspaceId={workspaceId}
      host={
        process.env.NEXT_PUBLIC_MEILISEARCH_HOST || 'http://localhost:7700'
      }
    >
      {children}
      <NoteSearchModal workspaceId={workspaceId} />
    </EditorSearchProvider>
  );
}
