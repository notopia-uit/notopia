'use client';

import { useRouter } from 'next/navigation';
import { WorkspaceContentWrapper } from '@notopia-uit/ui/components/workspace-content-wrapper';

export function WorkspaceContentWrapperWithNav({
  workspaceId,
  meilisearchHost,
  children,
}: {
  workspaceId: string;
  meilisearchHost?: string;
  children: React.ReactNode;
}) {
  const router = useRouter();
  return (
    <WorkspaceContentWrapper
      workspaceId={workspaceId}
      meilisearchHost={meilisearchHost}
      onNavigate={(href) => router.push(href)}
    >
      {children}
    </WorkspaceContentWrapper>
  );
}
