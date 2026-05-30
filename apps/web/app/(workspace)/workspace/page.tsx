import { getMyWorkspacesOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, QueryClient, dehydrate } from '@tanstack/react-query';
import { WorkspaceSwitcher } from '@ui/components/workspace-switcher';
import { fetchAccessTokenServerSide } from '@lib/get-access-token';

// import getQueryClient from '#/get-query-client';

export const dynamic = 'force-dynamic';

export default async function WorkspacePage() {
  const queryClient = new QueryClient();
  const { queryKey, queryFn } = getMyWorkspacesOptions({
    auth: await fetchAccessTokenServerSide(),
  });
  await queryClient.prefetchQuery({
    queryKey: queryKey,
    queryFn: queryFn,
    staleTime: 1000 * 60 * 60, // 1 hour
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <main className="bg-muted/20 flex min-h-screen flex-col items-center justify-center p-4 md:p-8">
        <div className="w-full max-w-3xl">
          <WorkspaceSwitcher />
        </div>
      </main>
    </HydrationBoundary>
  );
}
