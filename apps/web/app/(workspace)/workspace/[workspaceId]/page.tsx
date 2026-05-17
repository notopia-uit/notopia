import { getMyWorkspacesOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, QueryClient, dehydrate } from '@tanstack/react-query';
import { WorkspaceWelcome } from '@ui/components/workspace-welcome';
import { fetchAccessTokenServerSide } from '@ui/lib/get-access-token';
import { redirect } from 'next/navigation';

interface WorkspacePageProps {
  params: Promise<{ workspaceId: string }>;
}

export const dynamic = 'force-dynamic';

export default async function WorkspaceIndexPage({ params }: WorkspacePageProps) {
  const { workspaceId } = await params;
  const queryClient = new QueryClient();

  const { queryKey, queryFn } = getMyWorkspacesOptions({
    auth: await fetchAccessTokenServerSide(),
  });

  await queryClient.prefetchQuery({
    queryKey: queryKey,
    queryFn: queryFn,
    staleTime: 1000 * 60 * 60,
  });

  const workspacesData = queryClient.getQueryData<any>(queryKey);
  const currentWorkspace = workspacesData?.find((w: any) => w.workspace.id === workspaceId);

  if (!currentWorkspace) {
    redirect('/workspace');
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <WorkspaceWelcome
        workspaceId={currentWorkspace.workspace.id}
        workspaceName={currentWorkspace.workspace.name}
        workspaceSlug={currentWorkspace.workspace.slug}
      />
    </HydrationBoundary>
  );
}
