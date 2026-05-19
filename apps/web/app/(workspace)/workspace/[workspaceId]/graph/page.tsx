import { getWorkspaceGraphOptions } from '@notopia-uit/api-gen/index';
import { dehydrate, HydrationBoundary } from '@tanstack/react-query';
import GraphView from '@ui/components/graph-view';
import { fetchAccessTokenServerSide } from '@ui/lib/get-access-token';

import getQueryClient from '#/get-query-client';

export default async function GraphPage({ params }: { params: Promise<{ workspaceId: string }> }) {
  const { workspaceId } = await params;
  const queryClient = getQueryClient();
  const { queryKey: getWorkspaceGraphQueryKey, queryFn: getWorkspaceGraphQueryFn } =
    getWorkspaceGraphOptions({
      path: { workspaceId: workspaceId },
      auth: fetchAccessTokenServerSide,
    });
  await queryClient.prefetchQuery({
    queryKey: getWorkspaceGraphQueryKey,
    queryFn: getWorkspaceGraphQueryFn,
  });

  return (
    <div className="h-full w-full">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <GraphView workspaceId={workspaceId} />
      </HydrationBoundary>
    </div>
  );
}
