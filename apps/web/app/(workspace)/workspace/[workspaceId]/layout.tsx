import { getMyWorkspacesOptions, getWorkspaceTreeOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, dehydrate } from '@tanstack/react-query';
import { Separator } from '@ui/components/shadcn/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@ui/components/shadcn/sidebar';
import WorkspaceSideBar from '@ui/components/workspace-sidebar';

import getQueryClient from '../../../get-query-client';

interface WorkspaceLayoutProps {
  children: React.ReactNode;
  params: Promise<{ workspaceId: string }>;
}

//TODO: Prefetch errors will crash the layout / leak via Next.js error boundary.
// If either getMyWorkspaces or getWorkspaceTree rejects (e.g., invalid workspaceId, unauthenticated user, backend down), Promise.all rejects and the whole workspace layout — including the sidebar chrome — fails to render. Prefer prefetchQuery inside a try/catch or using queryClient.fetchQuery only where hydration is strictly needed, and let the client useSuspenseQuery boundary handle errors with a proper error UI. At minimum, consider an error.tsx boundary for this route segment so users don't see an unstyled error page.
export default async function WorkspaceLayout({ children, params }: WorkspaceLayoutProps) {
  const { workspaceId } = await params;
  const queryClient = getQueryClient();
  const { queryKey: getMyWorkspacesQueryKey, queryFn: getMyworkspacesQueryFn } =
    getMyWorkspacesOptions();
  const { queryKey: getWorkspaceTreeQueryKey, queryFn: getWorkspaceTreeQueryFn } =
    getWorkspaceTreeOptions({
      path: {
        workspaceId: workspaceId,
      },
    });

  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: getMyWorkspacesQueryKey,
      queryFn: getMyworkspacesQueryFn,
    }),
    queryClient.prefetchQuery({
      queryKey: getWorkspaceTreeQueryKey,
      queryFn: getWorkspaceTreeQueryFn,
    }),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <SidebarProvider defaultOpen={true}>
        <WorkspaceSideBar currentWorkspaceId={workspaceId} />
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
            <div className="flex items-center gap-2 px-4">
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
            </div>
          </header>
          <div>{children}</div>
        </SidebarInset>
      </SidebarProvider>
    </HydrationBoundary>
  );
}
