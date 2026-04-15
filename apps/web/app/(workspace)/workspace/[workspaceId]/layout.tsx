import {
  getMyWorkspacesOptions,
  getWorkspaceTreeOptions,
} from '@notopia-uit/api-gen/index';
import { HydrationBoundary, dehydrate } from '@tanstack/react-query';
import { Separator } from '@ui/components/shadcn/separator';
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@ui/components/shadcn/sidebar';
import WorkspaceSideBar from '@ui/components/workspace-sidebar';

import getQueryClient from '../../../get-query-client';

interface WorkspaceLayoutProps {
  children: React.ReactNode;
  params: Promise<{ workspaceId: string }>;
}

async function WorkspaceSideBarServerComponents({
  workspaceId,
}: {
  workspaceId: string;
}) {
  const queryClient = getQueryClient();
  const { queryKey: getMyWorkspacesQueryKey, queryFn: getMyworkspacesQueryFn } =
    getMyWorkspacesOptions();
  const {
    queryKey: getWorkspaceTreeQueryKey,
    queryFn: getWorkspaceTreeQueryFn,
  } = getWorkspaceTreeOptions({
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
      <WorkspaceSideBar currentWorkspaceSlug={workspaceId} />
    </HydrationBoundary>
  );
}

export default async function WorkspaceLayout({
  children,
  params,
}: WorkspaceLayoutProps) {
  const { workspaceId } = await params;
  return (
    <SidebarProvider defaultOpen={true}>
      <WorkspaceSideBarServerComponents workspaceId={workspaceId} />
      <SidebarInset>
        <header
          className="
            flex h-16 shrink-0 items-center gap-2 transition-[width,height]
            ease-linear
            group-has-data-[collapsible=icon]/sidebar-wrapper:h-12
          "
        >
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
          </div>
        </header>
        <div>{children}</div>
      </SidebarInset>
    </SidebarProvider>
  );
}
