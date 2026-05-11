'use client';
import {
  NoteUserWorkspace,
  NoteWorkspaceRole,
  getMyWorkspacesOptions,
  useChangeWorkspaceSlugMutation,
  useCreateWorkspaceMutation,
} from '@notopia-uit/api-gen';
import { Badge } from '@notopia-uit/ui/components/shadcn/badge';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Card, CardContent } from '@notopia-uit/ui/components/shadcn/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@notopia-uit/ui/components/shadcn/dropdown-menu';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Label } from '@notopia-uit/ui/components/shadcn/label';
import { RadioGroup, RadioGroupItem } from '@notopia-uit/ui/components/shadcn/radio-group';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@notopia-uit/ui/components/shadcn/select';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { Briefcase, MoreVertical, Pencil, Plus, Save, Shield, Trash2, User, X } from 'lucide-react';
import { useState, useRef } from 'react';

import { ErrorAlert } from './error-alert';
import { SuccessAlert } from './success-alert';

type UserRole = (typeof NoteWorkspaceRole)[keyof typeof NoteWorkspaceRole];

interface UserWorkspace {
  id: string;
  slug: string;
  name: string;
  userRole: UserRole;
}

const mapUserWorkspaceDtoToDomain = (dtos: NoteUserWorkspace[]): UserWorkspace[] =>
  dtos.map((dto) => ({
    id: dto.workspace.id,
    slug: dto.workspace.slug,
    name: dto.workspace.name,
    userRole: dto.role as UserRole,
  }));

const generateSlug = (name: string) => {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
};

// TODO: handle create new workspace with api and router.push to workspace
//TODO: handle on error, loading states, and edge cases like slug conflicts for both create and edit flows. Currently the UI optimistically updates local state without any backend integration, so it will drift from server state on any refetch (focus, reconnect, manual invalidation, etc.) and doesn't persist across sessions. Consider using React Query's useMutation with query invalidation for a more robust implementation.
const WorkspaceSwitcher = () => {
  const queryClient = useQueryClient();
  const _data = useQuery({
    ...getMyWorkspacesOptions({}),
    select: mapUserWorkspaceDtoToDomain,
  });

  const {
    data: allWorkspaceData,
    isPending: isGetMyWorkspacesPending,
    isError: isGetMyWorkspacesError,
    error: getMyWorkspacesError,
  } = _data;

  //TODO: Local state mirroring server data will desync on refetch.
  // workspaces and selectedId are initialized from allWorkspaceData once at mount. If the underlying getMyWorkspaces query refetches (on focus, reconnect, manual invalidation, etc.), useSuspenseQuery updates allWorkspaceData but the local useState snapshot is never re-synced, so the UI will silently drift from server state. Combined with the TODO at Line 69, this whole component currently operates on local-only edits.
  // Consider either deriving the list directly from the query data (with a mutation that invalidates the query) or using useEffect to sync — but the former is the idiomatic React Query pattern.

  if (isGetMyWorkspacesError) {
    throw getMyWorkspacesError;
  }
  const [workspaces, setWorkspaces] = useState<UserWorkspace[]>([]);
  const [selectedId, setSelectedId] = useState<string>();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [isAddingNew, setIsAddingNew] = useState(false);
  const [editForm, setEditForm] = useState<Partial<UserWorkspace>>({});

  const [showCreateWorkspaceSuccessAlert, setShowCreateWorkspaceSuccessAlert] = useState(false);
  const [showChangeSlugSuccessAlert, setShowChangeSlugSuccessAlert] = useState(false);
  const [showCreateWorkspaceErrorAlert, setShowCreateWorkspaceErrorAlert] = useState(false);
  const [showChangeSlugErrorAlert, setShowChangeSlugErrorAlert] = useState(false);

  const hideAlertTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  if (allWorkspaceData && workspaces.length === 0) {
    setWorkspaces(allWorkspaceData);
  }
  if (workspaces.length > 0 && !selectedId) {
    setSelectedId(workspaces[0]?.id || '');
  }
  const { mutate: createWorkspace, isPending: isCreating } = useCreateWorkspaceMutation({
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: getMyWorkspacesOptions({}).queryKey,
      });
      const newWorkspacesData = await queryClient.fetchQuery({
        ...getMyWorkspacesOptions({}),
      });
      const mappedWorkspaces = mapUserWorkspaceDtoToDomain(newWorkspacesData);
      setWorkspaces(mappedWorkspaces);
      setIsAddingNew(false);
      setEditForm({});
      setShowCreateWorkspaceSuccessAlert(true);
      if (hideAlertTimerRef.current) clearTimeout(hideAlertTimerRef.current);
      hideAlertTimerRef.current = setTimeout(() => {
        setShowCreateWorkspaceSuccessAlert(false);
      }, 3000);
    },
    onError: () => {
      setShowCreateWorkspaceErrorAlert(true);
      if (hideAlertTimerRef.current) clearTimeout(hideAlertTimerRef.current);
      hideAlertTimerRef.current = setTimeout(() => {
        setShowCreateWorkspaceErrorAlert(false);
      }, 3000);
    },
  });

  const { mutate: mutateWorkspaceSlug, isPending: isChangingSlug } = useChangeWorkspaceSlugMutation(
    {
      onSuccess: async () => {
        await queryClient.invalidateQueries({
          queryKey: getMyWorkspacesOptions({}).queryKey,
        });
        const newWorkspacesData = await queryClient.fetchQuery({
          ...getMyWorkspacesOptions({}),
        });
        const mappedWorkspaces = mapUserWorkspaceDtoToDomain(newWorkspacesData);
        setWorkspaces(mappedWorkspaces);
        setEditingId(null);
        setEditForm({});
        setShowChangeSlugSuccessAlert(true);
        if (hideAlertTimerRef.current) clearTimeout(hideAlertTimerRef.current);
        hideAlertTimerRef.current = setTimeout(() => {
          setShowChangeSlugSuccessAlert(false);
        }, 3000);
      },
      onError: () => {
        setShowChangeSlugErrorAlert(true);

        if (hideAlertTimerRef.current) clearTimeout(hideAlertTimerRef.current);
        hideAlertTimerRef.current = setTimeout(() => {
          setShowChangeSlugErrorAlert(false);
        }, 3000);
      },
    }
  );

  const startEditing = (workspace: UserWorkspace) => {
    setEditingId(workspace.id);
    setEditForm({ ...workspace });
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditForm({});
  };

  const startAddingNew = () => {
    setIsAddingNew(true);
    setEditForm({
      name: '',
      slug: '',
      userRole: 'owner',
    });
  };

  const cancelAddingNew = () => {
    setIsAddingNew(false);
    setEditForm({});
  };

  //   TODO:delete are local-only and will not persist.
  //
  // deleteWorkspace only mutate local state — no API calls — and saveNewWorkspace assigns Date.now().toString() as id, which will collide with real backend ids once the mutation is wired up. The TODO at Line 69 acknowledges this, but flagging so it isn't missed before release. Consider wiring useMutation with query invalidation for getMyWorkspacesQueryKey().
  //

  const deleteWorkspace = (id: string) => {
    setWorkspaces((prev) => {
      const nextWorkspaces = prev.filter((workspace) => workspace.id !== id);
      if (selectedId === id) {
        if (prev.length === 1) {
          setSelectedId('');
        } else {
          setSelectedId(nextWorkspaces[0]?.id || '');
        }
      }
      return nextWorkspaces;
    });
    if (editingId === id) {
      setEditingId(null);
      setEditForm({});
    }
  };

  return isGetMyWorkspacesPending ? (
    <Spinner className="size-8" />
  ) : (
    <section className="py-16 md:py-24">
      <div className="container max-w-2xl">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">Workspaces</h1>
            <p className="text-muted-foreground mt-1">Select or manage your active workspaces</p>
          </div>
          <Button onClick={startAddingNew}>
            <Plus className="mr-2 size-4" />
            New Workspace
          </Button>
        </div>

        <RadioGroup value={selectedId} onValueChange={setSelectedId}>
          <div className="space-y-3">
            {workspaces.map((workspace) => (
              <Card
                key={workspace.id}
                className={cn(
                  'cursor-pointer gap-0 p-0 transition-colors',
                  selectedId === workspace.id && editingId !== workspace.id && `border-primary`,
                  editingId === workspace.id && 'border-primary'
                )}
                onClick={() => editingId !== workspace.id && setSelectedId(workspace.id)}
              >
                <CardContent className="p-4">
                  {editingId === workspace.id ? (
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <h3 className="font-semibold">Edit Workspace</h3>
                        <div className="flex gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              cancelEditing();
                            }}
                          >
                            <X className="size-4" />
                          </Button>
                          <Button
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              mutateWorkspaceSlug({
                                path: {
                                  workspaceId: workspace.id,
                                },
                                body: {
                                  slug: editForm.slug as string,
                                },
                              });
                            }}
                            disabled={isChangingSlug}
                          >
                            {isChangingSlug ? <Spinner /> : <Save className="mr-2 size-4" />}
                            Save
                          </Button>
                        </div>
                      </div>
                      <div className="grid gap-4 sm:grid-cols-2">
                        <div className="space-y-2 sm:col-span-2">
                          <Label htmlFor={`name-${workspace.id}`}>Workspace Name</Label>
                          <Input
                            id={`name-${workspace.id}`}
                            value={editForm.name || ''}
                            onChange={(e) =>
                              setEditForm({
                                ...editForm,
                                name: e.target.value,
                              })
                            }
                            onClick={(e) => e.stopPropagation()}
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor={`slug-${workspace.id}`}>Slug</Label>
                          <Input
                            id={`slug-${workspace.id}`}
                            value={editForm.slug || ''}
                            onChange={(e) =>
                              setEditForm({
                                ...editForm,
                                slug: e.target.value,
                              })
                            }
                            onClick={(e) => e.stopPropagation()}
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor={`role-${workspace.id}`}>Your Role</Label>
                          <Select
                            value={editForm.userRole || 'member'}
                            onValueChange={(value: UserRole) =>
                              setEditForm({
                                ...editForm,
                                userRole: value,
                              })
                            }
                          >
                            <SelectTrigger
                              id={`role-${workspace.id}`}
                              onClick={(e) => e.stopPropagation()}
                            >
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="owner">Owner</SelectItem>
                              <SelectItem value="admin">Admin</SelectItem>
                              <SelectItem value="member">Member</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center gap-4">
                      <RadioGroupItem value={workspace.id} id={workspace.id} />

                      {/* Icon based on role */}
                      <div className="bg-muted/50 flex size-12 items-center justify-center rounded-lg">
                        {workspace.userRole === 'owner' || workspace.userRole === 'viewer' ? (
                          <Shield className="text-primary size-6" />
                        ) : (
                          <User className="text-muted-foreground size-6" />
                        )}
                      </div>

                      {/* Details */}
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{workspace.name}</span>
                          <Badge variant="secondary" className="text-xs">
                            {workspace.userRole.charAt(0).toUpperCase() +
                              workspace.userRole.slice(1)}
                          </Badge>
                        </div>
                        <p className="text-muted-foreground text-sm">/{workspace.slug}</p>
                      </div>

                      {/* Actions */}
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-8"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <MoreVertical className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              startEditing(workspace);
                            }}
                          >
                            <Pencil className="mr-2 size-4" />
                            Edit Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={(e) => {
                              e.stopPropagation();
                              deleteWorkspace(workspace.id);
                            }}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {workspace.userRole === 'owner'
                              ? 'Delete Workspace'
                              : 'Leave Workspace'}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}

            {isAddingNew && (
              <Card className="border-primary gap-0 p-0">
                <CardContent className="p-4">
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold">Create New Workspace</h3>
                      <div className="flex gap-2">
                        <Button variant="ghost" size="sm" onClick={cancelAddingNew}>
                          <X className="size-4" />
                        </Button>
                        <Button
                          size="sm"
                          onClick={() =>
                            createWorkspace({
                              body: {
                                slug: editForm.slug as string,
                                name: editForm.name as string,
                              },
                            })
                          }
                          disabled={isCreating}
                        >
                          {isCreating ? (
                            <Spinner className="mr-2 size-4" />
                          ) : (
                            <Save className="mr-2 size-4" />
                          )}
                          Create
                        </Button>
                      </div>
                    </div>
                    <div className="grid gap-4 sm:grid-cols-2">
                      <div className="space-y-2 sm:col-span-2">
                        <Label htmlFor="new-name">Workspace Name</Label>
                        <Input
                          id="new-name"
                          value={editForm.name || ''}
                          onChange={(e) => {
                            const newName = e.target.value;
                            setEditForm({
                              ...editForm,
                              name: newName,
                              slug: generateSlug(newName),
                            });
                          }}
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="new-slug">Slug URL</Label>
                        <Input
                          id="new-slug"
                          value={editForm.slug || ''}
                          onChange={(e) =>
                            setEditForm({
                              ...editForm,
                              slug: e.target.value,
                            })
                          }
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="new-role">Initial Role</Label>
                        <Select
                          value={editForm.userRole || 'owner'}
                          onValueChange={(value: UserRole) =>
                            setEditForm({ ...editForm, userRole: value })
                          }
                        >
                          <SelectTrigger id="new-role">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="owner">Owner</SelectItem>
                            <SelectItem value="admin">Admin</SelectItem>
                            <SelectItem value="member">Member</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </RadioGroup>

        {workspaces.length === 0 && (
          <Card className="p-0">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Briefcase className="text-muted-foreground mb-4 size-12" />
              <h2 className="text-xl font-semibold">No workspaces found</h2>
              <p className="text-muted-foreground mt-2">Create a new workspace to get started</p>
              <Button className="mt-4" onClick={startAddingNew}>
                <Plus className="mr-2 size-4" />
                Create Workspace
              </Button>
            </CardContent>
          </Card>
        )}
        {showCreateWorkspaceSuccessAlert && (
          <SuccessAlert
            title="Workspace Created"
            message="Your new workspace has been created successfully."
          />
        )}
        {showChangeSlugSuccessAlert && (
          <SuccessAlert
            title="Workspace Updated"
            message="Your workspace details have been updated successfully."
          />
        )}

        {showCreateWorkspaceErrorAlert && (
          <ErrorAlert
            title="Creation Failed"
            message="There was an error creating your workspace. Please try again."
          />
        )}
        {showChangeSlugErrorAlert && (
          <ErrorAlert
            title="Update Failed"
            message="There was an error updating your workspace. Please try again."
          />
        )}
      </div>
    </section>
  );
};

export { WorkspaceSwitcher };
