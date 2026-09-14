import { getWorkspaceGraphOptions } from '@notopia-uit/api-gen/index';
import { dehydrate, HydrationBoundary } from '@tanstack/react-query';
import { fetchAccessTokenServerSide } from '@lib/get-access-token';
import { notFound } from 'next/navigation';

import getQueryClient from '#/get-query-client';

import { GraphViewBoundary } from './graph-view-boundary';

export default async function GraphPage({ params }: { params: Promise<{ workspaceId: string }> }) {
  const { workspaceId } = await params;
  const queryClient = getQueryClient();
  const { queryKey: getWorkspaceGraphQueryKey, queryFn: getWorkspaceGraphQueryFn } =
    getWorkspaceGraphOptions({
      path: { workspaceId: workspaceId },
      auth: fetchAccessTokenServerSide,
    });

  try {
    await queryClient.prefetchQuery({
      queryKey: getWorkspaceGraphQueryKey,
      queryFn: getWorkspaceGraphQueryFn,
    });
  } catch {
    notFound();
  }

  return (
    <div className="h-screen w-full overflow-hidden">
      <HydrationBoundary state={dehydrate(queryClient)}>
        <GraphViewBoundary workspaceId={workspaceId} />
      </HydrationBoundary>
    </div>
  );
}
