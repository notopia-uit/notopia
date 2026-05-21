'use client';
import {
  NoteUserWorkspace,
  NoteWorkspaceRole,
  getMyWorkspacesOptions,
  useChangeWorkspaceSlugMutation,
  useCreateWorkspaceMutation,
  useLeaveWorkspaceMutation,
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
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { Briefcase, MoreVertical, Pencil, Plus, Save, Shield, Trash2, User, X } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';

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

const WorkspaceSwitcher = () => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { retry } = useQueryErrorHandler();

  const _data = useQuery({
    ...getMyWorkspacesOptions(),
    select: mapUserWorkspaceDtoToDomain,
  });

  const {
    data: allWorkspaceData,
    isPending: isGetMyWorkspacesPending,
    isError: isGetMyWorkspacesError,
    error: getMyWorkspacesError,
  } = _data;

  const [workspaces, setWorkspaces] = useState<UserWorkspace[]>([]);
  const [selectedId, setSelectedId] = useState<string>();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [isAddingNew, setIsAddingNew] = useState(false);
  const [editForm, setEditForm] = useState<Partial<UserWorkspace>>({});

  const { alert, showAlert } = useAlert();

  const handleSelectWorkspace = (workspaceId: string) => {
    const workspace = workspaces.find((w) => w.id === workspaceId);
    if (workspace && editingId !== workspaceId) {
      setSelectedId(workspaceId);
      router.push(`/workspace/${workspaceId}`);
    }
  };

  if (isGetMyWorkspacesError) {
    return (
      <div className="p-4">
        <QueryErrorFallback
          error={getMyWorkspacesError}
          onRetry={retry}
          title="Failed to Load Workspaces"
          description="Unable to load your workspaces. Please try again."
          compact
        />
      </div>
    );
  }
  useEffect(() => {
    if (allWorkspaceData) {
      setWorkspaces(allWorkspaceData);
    }
  }, [allWorkspaceData]);
  const { mutate: createWorkspace, isPending: isCreating } = useCreateWorkspaceMutation({
    onSuccess: async (responses, variables) => {
      await queryClient.invalidateQueries({
        queryKey: getMyWorkspacesOptions({}).queryKey,
      });
      setWorkspaces((prev) => [
        ...prev,
        {
          id: responses.id,
          slug: variables.body.slug,
          name: variables.body.name,
          userRole: 'owner',
        },
      ]);
      setIsAddingNew(false);
      setEditForm({});
      showAlert(
        'success',
        'Workspace Created',
        `Your new workspace "${variables.body.name}" has been created successfully.`
      );
    },
    onError: (error) => {
      showAlert(
        'error',
        'Creation Failed',
        `There was an error creating your workspace. Please try again. Error details: ${error.message}`
      );
    },
  });

  const { mutate: mutateWorkspaceSlug, isPending: isChangingSlug } = useChangeWorkspaceSlugMutation(
    {
      onSuccess: async (_, variables) => {
        await queryClient.invalidateQueries({
          queryKey: getMyWorkspacesOptions({}).queryKey,
        });

        setWorkspaces((prev) =>
          prev.map((workspace) =>
            workspace.id === variables.path.workspaceId
              ? { ...workspace, slug: variables.body.slug }
              : workspace
          )
        );
        setEditingId(null);
        setEditForm({});
        showAlert(
          'success',
          'Slug Updated',
          `Workspace slug has been updated to "${variables.body.slug}".`
        );
      },
      onError: () => {
        showAlert(
          'error',
          'Update Failed',
          `There was an error updating the workspace slug. Please try again.`
        );
      },
    }
  );

  const { mutate: leaveWorkspace, isPending: isLeavingWorkspace } = useLeaveWorkspaceMutation({
    onSuccess: async (_, variables) => {
      await queryClient.invalidateQueries({
        queryKey: getMyWorkspacesOptions({}).queryKey,
      });
      setWorkspaces((prev) =>
        prev.filter((workspace) => workspace.id !== variables.path.workspaceId)
      );
      if (selectedId === variables.path.workspaceId) {
        setSelectedId(undefined);
      }
      showAlert('success', 'Left Workspace', `You have left the workspace successfully.`);
    },
    onError: (error) => {
      showAlert(
        'error',
        'Action Failed',
        `There was an error leaving the workspace. Please try again. Error details: ${error.message}`
      );
    },
  });
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

        <RadioGroup value={selectedId} onValueChange={handleSelectWorkspace}>
          <div className="space-y-3">
            {workspaces.map((workspace) => (
               <Card
                 key={workspace.id}
                 className={cn(
                   'cursor-pointer gap-0 p-0 transition-colors',
                   selectedId === workspace.id && editingId !== workspace.id && `border-primary`,
                   editingId === workspace.id && 'border-primary'
                 )}
                 onClick={() => editingId !== workspace.id && handleSelectWorkspace(workspace.id)}
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
                          {workspace.userRole !== 'owner' && (
                            <DropdownMenuItem
                              className="text-destructive"
                              disabled={isLeavingWorkspace}
                              onClick={(e) => {
                                e.stopPropagation();
                                leaveWorkspace({
                                  path: {
                                    workspaceId: workspace.id,
                                  },
                                });
                              }}
                            >
                              {isLeavingWorkspace ? (
                                <Spinner></Spinner>
                              ) : (
                                <Trash2 className="mr-2 size-4" />
                              )}
                              Leave Workspace
                            </DropdownMenuItem>
                          )}
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
        {alert?.type === 'success' && <SuccessAlert title={alert.title} message={alert.message} />}
        {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}
      </div>
    </section>
  );
};

export { WorkspaceSwitcher };
