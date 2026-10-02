import { getMyWorkspacesOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, QueryClient, dehydrate } from '@tanstack/react-query';
import { WorkspaceSwitcherWrapper } from './workspace-switcher-wrapper';
import { fetchAccessTokenServerSide } from '@lib/get-access-token';
import { notFound } from 'next/navigation';

export const dynamic = 'force-dynamic';

export default async function WorkspacePage() {
  const queryClient = new QueryClient();
  const { queryKey, queryFn } = getMyWorkspacesOptions({
    auth: await fetchAccessTokenServerSide(),
  });

  try {
    await queryClient.prefetchQuery({
      queryKey: queryKey,
      queryFn: queryFn,
      staleTime: 1000 * 60 * 60,
    });
  } catch {
    notFound();
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <main className="bg-muted/20 flex min-h-screen flex-col items-center justify-center p-4 md:p-8">
        <div className="w-full max-w-3xl">
          <WorkspaceSwitcherWrapper />
        </div>
      </main>
    </HydrationBoundary>
  );
}
