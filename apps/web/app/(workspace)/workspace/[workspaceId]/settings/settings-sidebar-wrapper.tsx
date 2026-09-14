'use client';

import { usePathname } from 'next/navigation';
import { SettingsSidebar } from '@notopia-uit/ui/components/settings-sidebar';

export function SettingsSidebarWrapper({
  workspaceId,
}: {
  workspaceId: string;
}) {
  const pathname = usePathname();
  return <SettingsSidebar workspaceId={workspaceId} pathname={pathname} />;
}
