'use client';

import { getWorkspaceSearchTokenOptions } from '@notopia-uit/api-gen';
import { useQuery } from '@tanstack/react-query';
import { MeilisearchProvider } from '@notopia-uit/ui/contexts/meilisearch-context';

interface EditorSearchProviderProps {
  children: React.ReactNode;
  workspaceId: string;
  host: string;
}

export function EditorSearchProvider({ children, workspaceId, host }: EditorSearchProviderProps) {
  const { data: tokenData } = useQuery({
    ...getWorkspaceSearchTokenOptions({
      path: {
        workspaceId,
      },
    }),
  });

  return (
    <MeilisearchProvider host={host} apiKey={tokenData?.token}>
      {children}
    </MeilisearchProvider>
  );
}
