'use client';

import {
  NoteUserWorkspace,
  getMyWorkspacesOptions,
  useCreateWorkspaceMutation,
} from '@notopia-uit/api-gen';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { getAuthClient } from '@notopia-uit/ui/lib/auth-client';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  BadgeCheck,
  Bell,
  ChevronsUpDown,
  CreditCard,
  Folder,
  Forward,
  GalleryVerticalEnd,
  LogOut,
  MoreHorizontal,
  Plus,
  Save,
  Settings2,
  Sparkles,
  Trash2,
} from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

import { Avatar, AvatarFallback, AvatarImage } from './shadcn/avatar';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './shadcn/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from './shadcn/dropdown-menu';
import { Label } from './shadcn/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './shadcn/select';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from './shadcn/sidebar';
import TreeView from './tree-view';
import { WorkspaceMembersModal } from './workspace-members-modal';

const generateSlug = (name: string) => {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/(^-|-$)+/g, '');
};

export function CreateWorkspaceDialog() {
  const queryClient = useQueryClient();

  const [isOpen, setIsOpen] = useState(false);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [userRole, setUserRole] = useState('owner');

  const { mutate: createWorkspace, isPending: isCreating } = useCreateWorkspaceMutation({
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: getMyWorkspacesOptions({}).queryKey,
      });

      setName('');
      setSlug('');
      setUserRole('owner');

      setIsOpen(false);
    },
  });

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!name || !slug) return;

    createWorkspace({
      body: {
        name: name,
        slug: slug,
      },
    });
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <DropdownMenuItem className="gap-2 p-2" onSelect={(e) => e.preventDefault()}>
          <div className="bg-background flex size-6 items-center justify-center rounded-md border">
            <Plus className="size-4" />
          </div>
          <span className="text-muted-foreground font-medium">Add workspace</span>
        </DropdownMenuItem>
      </DialogTrigger>
      <DialogContent className="sm:max-w-106.25">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Workspace</DialogTitle>
            <DialogDescription>
              Set up a new workspace to organize your notes and graphs.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Workspace Name</Label>
              <Input
                id="name"
                placeholder="e.g. My Awesome Project"
                value={name}
                onChange={(e) => {
                  const newName = e.target.value;
                  setName(newName);
                  setSlug(generateSlug(newName));
                }}
                disabled={isCreating}
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="slug">Slug URL</Label>
              <Input
                id="slug"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                disabled={isCreating}
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="role">Initial Role</Label>
              <Select value={userRole} onValueChange={setUserRole} disabled={isCreating}>
                <SelectTrigger id="role">
                  <SelectValue placeholder="Select a role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="owner">Owner</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="member">Member</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsOpen(false)}
              disabled={isCreating}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isCreating}>
              {isCreating ? (
                <Spinner className="mr-2 size-4 animate-spin" />
              ) : (
                <Save className="mr-2 size-4" />
              )}
              Create
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

const data = [
  {
    name: 'Settings',
    url: (workspaceId: string) => `/workspace/${workspaceId}/settings/general`,
    icon: Settings2,
  },
  {
    name: 'Graph',
    url: (workspaceId: string) => `/workspace/${workspaceId}/graph`,
    icon: Sparkles,
  },
];

export default function WorkspaceSideBar({ currentWorkspaceId }: { currentWorkspaceId: string }) {
  const { data: sessionData } = getAuthClient().useSession();
  const { retry } = useQueryErrorHandler();

  const [activeWorkspacenow, setActiveWorkspace] = useState<NoteUserWorkspace>();

  const router = useRouter();

  const {
    data: allWorkspaceData,
    isError,
    error,
    isPending,
  } = useQuery({
    ...getMyWorkspacesOptions({}),
  });
  const currentWorkspace = allWorkspaceData?.find((ws) => ws.workspace.id === currentWorkspaceId);

  useEffect(() => {
    if (currentWorkspace) {
      setActiveWorkspace(currentWorkspace);
    }
  }, [currentWorkspaceId, currentWorkspace]);

  if (!sessionData) {
    return;
  }
  if (isPending) {
    return <Spinner />;
  }
  if (isError) {
    return (
      <div className="p-4">
        <QueryErrorFallback
          error={error}
          onRetry={retry}
          title="Failed to Load Workspaces"
          description="Unable to load your workspaces. Please try again."
          compact
        />
      </div>
    );
  }
  if (!allWorkspaceData || !currentWorkspace) {
    return;
  }

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                    <GalleryVerticalEnd className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-sm/tight">
                    <span className="truncate font-semibold">
                      {activeWorkspacenow?.workspace.name}
                    </span>
                    {/* <span className="truncate text-xs"> */}
                    {/*   {currentWorkspace.plan} */}
                    {/* </span> */}
                  </div>
                  <ChevronsUpDown className="ml-auto" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                align="start"
                side="bottom"
                sideOffset={4}
              >
                <DropdownMenuLabel className="text-muted-foreground text-xs">
                  Workspace
                </DropdownMenuLabel>
                {isPending ? (
                  <Spinner />
                ) : (
                  allWorkspaceData.map((ws, index) => (
                    <DropdownMenuItem
                      key={ws.workspace.name}
                      onClick={() => {
                        setActiveWorkspace(ws);
                        router.push(`/workspace/${index}`);
                      }}
                      className="gap-2 p-2"
                    >
                      <div className="flex size-6 items-center justify-center rounded-sm border">
                        <GalleryVerticalEnd className="size-4 shrink-0" />
                      </div>
                      {ws.workspace.name}
                      <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
                    </DropdownMenuItem>
                  ))
                )}
                <DropdownMenuSeparator />
                <DropdownMenuItem className="gap-2 p-2">
                  <div className="bg-background flex size-6 items-center justify-center rounded-md border">
                    <Plus className="size-4" />
                  </div>
                  <div className="text-muted-foreground font-medium">Add workspace</div>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent className="flex flex-col overflow-y-auto">
        <SidebarGroup className="flex flex-col group-data-[collapsible=icon]:hidden">
          <SidebarGroupLabel>Platform</SidebarGroupLabel>
          <SidebarMenu className="flex flex-col">
            <TreeView currentWorkspaceId={currentWorkspaceId} />
          </SidebarMenu>
        </SidebarGroup>
        <SidebarGroup className="shrink-0 group-data-[collapsible=icon]:hidden">
          <SidebarGroupLabel>Projects</SidebarGroupLabel>
          <SidebarMenu>
            {data.map((item) => (
              <SidebarMenuItem key={item.name}>
                <SidebarMenuButton onClick={() => router.push(item.url(currentWorkspaceId))}>
                  <item.icon />
                  <span>{item.name}</span>
                </SidebarMenuButton>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <SidebarMenuAction showOnHover>
                      <MoreHorizontal />
                      <span className="sr-only">More</span>
                    </SidebarMenuAction>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent className="w-48 rounded-lg" side="bottom" align="end">
                    <DropdownMenuItem>
                      <Folder className="text-muted-foreground" />
                      <span>View Project</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Forward className="text-muted-foreground" />
                      <span>Share Project</span>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem>
                      <Trash2 className="text-muted-foreground" />
                      <span>Delete Project</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </SidebarMenuItem>
            ))}
            <SidebarMenuItem>
              <SidebarMenuButton className="text-sidebar-foreground/70">
                <MoreHorizontal className="text-sidebar-foreground/70" />
                <span>More</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>
        <SidebarGroup className="shrink-0 group-data-[collapsible=icon]:hidden">
          <WorkspaceMembersModal workspaceId={currentWorkspaceId} />
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  {/* TODO: get user data from betterauth */}
                  <Avatar className="size-8 rounded-lg">
                    <AvatarImage src={sessionData.user?.image || ''} alt={'avatar here'} />
                    <AvatarFallback className="rounded-lg">
                      {sessionData.user?.name?.slice(0, 2)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm/tight">
                    <span className="truncate font-semibold">{sessionData.user.name}</span>
                    <span className="truncate text-xs">{sessionData.user.email}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                side="bottom"
                align="end"
                sideOffset={4}
              >
                <DropdownMenuLabel className="p-0 font-normal">
                  <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                    <Avatar className="size-8 rounded-lg">
                      <AvatarImage src={sessionData.user.image || ''} alt={sessionData.user.name} />
                      <AvatarFallback className="rounded-lg">
                        {sessionData.user?.name?.slice(0, 2)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm/tight">
                      <span className="truncate font-semibold">{sessionData.user.name}</span>
                      <span className="truncate text-xs">{sessionData.user.email}</span>
                    </div>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem>
                    <Sparkles />
                    Upgrade to Pro
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem>
                    <BadgeCheck />
                    Account
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <CreditCard />
                    Billing
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Bell />
                    Notifications
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={async () => {
                    await getAuthClient().signOut();
                    window.location.href = '/api/auth/logout';
                  }}
                >
                  <LogOut />
                  Log out
                </DropdownMenuItem>
                <CreateWorkspaceDialog />
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
