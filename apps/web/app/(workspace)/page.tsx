import { getMyWorkspacesOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, dehydrate } from '@tanstack/react-query';
import { WorkspaceSwitcher } from '@ui/components/workspace-switcher';

import getQueryClient from '#/get-query-client';

async function WorkspaceSwitcherServerComponent() {
  const queryClient = getQueryClient();
  const { queryKey, queryFn } = getMyWorkspacesOptions();
  await queryClient.prefetchQuery({
    queryKey: queryKey,
    queryFn: queryFn,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <WorkspaceSwitcher />
    </HydrationBoundary>
  );
}

export default function WorkspacePage() {
  return (
    <main
      className="
        flex min-h-screen flex-col items-center justify-center bg-muted/20 p-4
        md:p-8
      "
    >
      <div className="w-full max-w-3xl">
        <WorkspaceSwitcherServerComponent />
      </div>
    </main>
  );
}
