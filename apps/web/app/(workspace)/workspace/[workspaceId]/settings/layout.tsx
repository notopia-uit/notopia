import { SettingsSidebar } from '@ui/components/settings-sidebar';
import { Separator } from '@ui/components/shadcn/separator';

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
    <div className="min-h-screen p-10 font-sans">
      <div className="mx-auto max-w-5xl space-y-6">
        <div className="space-y-1">
          <h2 className="text-2xl font-bold tracking-tight">Workspace Settings</h2>
          <p className="text-sm">Manage your workspace details, members, and preferences.</p>
        </div>

        <Separator className="my-6" />

        <div className="flex flex-col gap-8 lg:flex-row lg:gap-12">
          <aside className="w-full lg:w-1/4">
            <SettingsSidebar workspaceId={workspaceId} />
          </aside>

          <div className="max-w-2xl flex-1">{children}</div>
        </div>
      </div>
    </div>
  );
}
