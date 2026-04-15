'use client';

import {
  NoteUserWorkspace,
  getMyWorkspacesOptions,
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
import {
  RadioGroup,
  RadioGroupItem,
} from '@notopia-uit/ui/components/shadcn/radio-group';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@notopia-uit/ui/components/shadcn/select';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { useSuspenseQuery } from '@tanstack/react-query';
import {
  Briefcase,
  MoreVertical,
  Pencil,
  Plus,
  Save,
  Shield,
  Trash2,
  User,
  X,
} from 'lucide-react';
import { useState } from 'react';

interface UserWorkspace {
  id: string;
  slug: string;
  name: string;
  userRole: 'owner' | 'admin' | 'member';
}

const mapUserWorkspaceDtoToDomain = (
  dtos: NoteUserWorkspace[]
): UserWorkspace[] =>
  dtos.map((dto) => ({
    id: dto.workspace.id,
    slug: dto.workspace.slug,
    name: dto.workspace.name,
    userRole: dto.role as 'owner' | 'admin' | 'member',
  }));

const generateSlug = (name: string) => {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
};

// TODO: handle create new workspace with api and router.push to workspace

const WorkspaceSwitcher = () => {
  const { data: allWorkspaceData } = useSuspenseQuery({
    ...getMyWorkspacesOptions({}),
    select: mapUserWorkspaceDtoToDomain,
  });

  const [workspaces, setWorkspaces] =
    useState<UserWorkspace[]>(allWorkspaceData);
  const [selectedId, setSelectedId] = useState<string>(workspaces[0]?.id || '');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [isAddingNew, setIsAddingNew] = useState(false);
  const [editForm, setEditForm] = useState<Partial<UserWorkspace>>({});

  const startEditing = (workspace: UserWorkspace) => {
    setEditingId(workspace.id);
    setEditForm({ ...workspace });
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditForm({});
  };

  const saveEdit = () => {
    if (editingId && editForm.name && editForm.slug && editForm.userRole) {
      setWorkspaces((prev) =>
        prev.map((workspace) =>
          workspace.id === editingId
            ? ({
                ...workspace,
                ...editForm,
              } as UserWorkspace)
            : workspace
        )
      );
      setEditingId(null);
      setEditForm({});
    }
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

  const saveNewWorkspace = () => {
    if (editForm.name && editForm.slug && editForm.userRole) {
      const newWorkspace: UserWorkspace = {
        id: Date.now().toString(),
        name: editForm.name,
        slug: editForm.slug,
        userRole: editForm.userRole,
      };
      setWorkspaces((prev) => [...prev, newWorkspace]);
      setIsAddingNew(false);
      setEditForm({});
      if (workspaces.length === 0) {
        setSelectedId(newWorkspace.id);
      }
    }
  };

  const deleteWorkspace = (id: string) => {
    setWorkspaces((prev) => prev.filter((workspace) => workspace.id !== id));
    if (selectedId === id && workspaces.length > 1) {
      setSelectedId(workspaces.find((w) => w.id !== id)?.id || '');
    }
    if (editingId === id) {
      setEditingId(null);
      setEditForm({});
    }
  };

  return (
    <section className="py-16 md:py-24">
      <div className="container max-w-2xl">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">
              Workspaces
            </h1>
            <p className="mt-1 text-muted-foreground">
              Select or manage your active workspaces
            </p>
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
                  selectedId === workspace.id &&
                    editingId !== workspace.id &&
                    'border-primary',
                  editingId === workspace.id && 'border-primary'
                )}
                onClick={() =>
                  editingId !== workspace.id && setSelectedId(workspace.id)
                }
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
                              saveEdit();
                            }}
                          >
                            <Save className="mr-2 size-4" />
                            Save
                          </Button>
                        </div>
                      </div>
                      <div className="grid gap-4 sm:grid-cols-2">
                        <div className="space-y-2 sm:col-span-2">
                          <Label htmlFor={`name-${workspace.id}`}>
                            Workspace Name
                          </Label>
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
                          <Label htmlFor={`role-${workspace.id}`}>
                            Your Role
                          </Label>
                          <Select
                            value={editForm.userRole || 'member'}
                            onValueChange={(
                              value: 'owner' | 'admin' | 'member'
                            ) => setEditForm({ ...editForm, userRole: value })}
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
                      <div className="flex size-12 items-center justify-center rounded-lg bg-muted/50">
                        {workspace.userRole === 'owner' ||
                        workspace.userRole === 'admin' ? (
                          <Shield className="size-6 text-primary" />
                        ) : (
                          <User className="size-6 text-muted-foreground" />
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
                        <p className="text-sm text-muted-foreground">
                          /{workspace.slug}
                        </p>
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
              <Card className="gap-0 border-primary p-0">
                <CardContent className="p-4">
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold">Create New Workspace</h3>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={cancelAddingNew}
                        >
                          <X className="size-4" />
                        </Button>
                        <Button size="sm" onClick={saveNewWorkspace}>
                          <Save className="mr-2 size-4" />
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
                          onValueChange={(
                            value: 'owner' | 'admin' | 'member'
                          ) => setEditForm({ ...editForm, userRole: value })}
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
              <Briefcase className="mb-4 size-12 text-muted-foreground" />
              <h2 className="text-xl font-semibold">No workspaces found</h2>
              <p className="mt-2 text-muted-foreground">
                Create a new workspace to get started
              </p>
              <Button className="mt-4" onClick={startAddingNew}>
                <Plus className="mr-2 size-4" />
                Create Workspace
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </section>
  );
};

export { WorkspaceSwitcher };
