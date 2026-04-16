import { showTrashOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, dehydrate } from '@tanstack/react-query';
import TrashedItemManagement from '@ui/components/trashed-file-managment';

import getQueryClient from '#/get-query-client';

async function TrasedItemsManagementServerComponents({
  workspaceId,
}: {
  workspaceId: string;
}) {
  const queryClient = getQueryClient();
  const { queryKey: showTrashQueryKey, queryFn: showTrashQueryFn } =
    showTrashOptions({ path: { workspaceId: workspaceId } });
  await queryClient.prefetchQuery({
    queryKey: showTrashQueryKey,
    queryFn: showTrashQueryFn,
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TrashedItemManagement workspaceId={workspaceId}></TrashedItemManagement>
    </HydrationBoundary>
  );
}

export default async function TrashPage({
  params,
}: {
  params: Promise<{ workspaceId: string }>;
}) {
  const { workspaceId } = await params;
  return <TrasedItemsManagementServerComponents workspaceId={workspaceId} />;
}
