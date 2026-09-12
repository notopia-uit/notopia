'use client';

import { useRouter } from 'next/navigation';
import { WorkspaceSwitcher } from '@notopia-uit/ui/components/workspace-switcher';

export function WorkspaceSwitcherWrapper() {
  const router = useRouter();
  return <WorkspaceSwitcher onNavigate={(href) => router.push(href)} />;
}
