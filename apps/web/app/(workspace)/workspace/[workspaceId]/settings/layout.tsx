import { Separator } from '@ui/components/shadcn/separator';

import { SettingsSidebarWrapper } from './settings-sidebar-wrapper';

interface WorkspaceSettingsLayoutProps {
  children: React.ReactNode;
  params: Promise<{ workspaceId: string }>;
}

export default async function WorkspaceSettingsLayout({
  children,
  params,
}: WorkspaceSettingsLayoutProps) {
  const { workspaceId } = await params;

  return (
    <div className="min-h-screen p-4 font-sans sm:p-6 lg:p-10">
      <div className="mx-auto max-w-7xl space-y-6">
        <div className="space-y-1">
          <h2 className="text-2xl font-bold tracking-tight">Workspace Settings</h2>
          <p className="text-sm">Manage your workspace details, members, and preferences.</p>
        </div>

        <Separator className="my-6" />

        <div className="flex flex-col gap-8 lg:flex-row lg:gap-12">
          <aside className="w-full overflow-x-auto lg:w-48 lg:shrink-0">
            <SettingsSidebarWrapper workspaceId={workspaceId} />
          </aside>

          <div className="min-w-0 flex-1">{children}</div>
        </div>
      </div>
    </div>
  );
}
