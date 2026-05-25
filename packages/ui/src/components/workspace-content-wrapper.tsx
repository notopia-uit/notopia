'use client';

import { getWorkspaceSearchTokenOptions } from '@notopia-uit/api-gen';
import { useQuery } from '@tanstack/react-query';
import { MeilisearchProvider } from '@notopia-uit/ui/contexts/meilisearch-context';
import { WorkspaceEventsProvider } from '@notopia-uit/ui/contexts/workspace-events-context';

import { NoteSearchModal } from './note-search-modal';

interface WorkspaceContentWrapperProps {
  children: React.ReactNode;
  workspaceId: string;
}

export function WorkspaceContentWrapper({
  children,
  workspaceId,
}: WorkspaceContentWrapperProps) {
  const { data: tokenData } = useQuery({
    ...getWorkspaceSearchTokenOptions({
      path: {
        workspaceId,
      },
    }),
  });

  return (
    <MeilisearchProvider
      host={process.env.NEXT_PUBLIC_MEILISEARCH_HOST || 'http://localhost:7700'}
      apiKey={tokenData?.token}
    >
      <WorkspaceEventsProvider workspaceId={workspaceId}>
        {children}
      </WorkspaceEventsProvider>
      <NoteSearchModal workspaceId={workspaceId} />
    </MeilisearchProvider>
  );
}
