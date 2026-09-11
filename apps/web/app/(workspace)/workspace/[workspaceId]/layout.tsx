import { getMyWorkspacesOptions, getWorkspaceTreeOptions } from '@notopia-uit/api-gen/index';
import { HydrationBoundary, dehydrate } from '@tanstack/react-query';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@ui/components/shadcn/sidebar';
import { ModeToggle } from '@ui/components/theme-mode-toggle';
import { WorkspaceContentWrapper } from '@ui/components/workspace-content-wrapper';
import WorkspaceSideBar from '@ui/components/workspace-sidebar';
import { fetchAccessTokenServerSide } from '@lib/get-access-token';
import { notFound } from 'next/navigation';

import getQueryClient from '#/get-query-client';

interface WorkspaceLayoutProps {
  children: React.ReactNode;
  params: Promise<{ workspaceId: string }>;
}

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
      auth: fetchAccessTokenServerSide,
    });

  try {
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
  } catch {
    notFound();
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <WorkspaceContentWrapper
        workspaceId={workspaceId}
        meilisearchHost={process.env.MEILISEARCH_HOST}
      >
        <SidebarProvider defaultOpen={true}>
          <WorkspaceSideBar currentWorkspaceId={workspaceId} />
          <SidebarInset className="min-w-0">
            <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
              <div className="flex items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1" />
              </div>
              <div className="ml-auto mr-4">
                <ModeToggle />
              </div>
            </header>
            <div>{children}</div>
          </SidebarInset>
        </SidebarProvider>
      </WorkspaceContentWrapper>
    </HydrationBoundary>
  );
}
