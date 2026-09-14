'use client';

import { useRouter } from 'next/navigation';
import WorkspaceSideBar from '@notopia-uit/ui/components/workspace-sidebar';

export function WorkspaceSideBarWrapper({
  currentWorkspaceId,
}: {
  currentWorkspaceId: string;
}) {
  const router = useRouter();
  return <WorkspaceSideBar currentWorkspaceId={currentWorkspaceId} onNavigate={(href) => router.push(href)} />;
}
