import { GeneralSettings } from '@ui/components/general-settings';

//TODO: handle prefetch with client
export default async function WorkspaceGeneralSettingsPage({
  params,
}: {
  params: Promise<{ workspaceId: string }>;
}) {
  const { workspaceId } = await params;
  return <GeneralSettings workspaceId={workspaceId} />;
}
