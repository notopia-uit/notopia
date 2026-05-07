import { getMyWorkspacesOptions } from '@notopia-uit/api-gen/index';
import { dehydrate, HydrationBoundary } from '@tanstack/react-query';
import { GeneralSettings } from '@ui/components/general-settings';

import getQueryClient from '#/get-query-client';

//TODO: handle prefetch with client
export default async function WorkspaceGeneralSettingsPage({
  params,
}: {
  params: Promise<{ workspaceId: string }>;
}) {
  const { workspaceId } = await params;
  const queryClient = getQueryClient();
  const { queryKey: getMyWorkspacesQueryKey, queryFn: getMyworkspacesQueryFn } =
    getMyWorkspacesOptions();
  await queryClient.prefetchQuery({
    queryKey: getMyWorkspacesQueryKey,
    queryFn: getMyworkspacesQueryFn,
  });
  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <GeneralSettings workspaceId={workspaceId} />
    </HydrationBoundary>
  );
}
