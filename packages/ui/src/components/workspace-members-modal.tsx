'use client';

import {
  getWorkspaceMembersOptions,
  NoteWorkspaceMember,
  NoteWorkspaceRole,
  useUpdateWorkspaceMembersMutation,
} from '@notopia-uit/api-gen';
import { Badge } from '@notopia-uit/ui/components/shadcn/badge';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Search, Trash2, Users } from 'lucide-react';
import { useState, useMemo } from 'react';

import { ErrorAlert } from './error-alert';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './shadcn/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './shadcn/select';
import { SuccessAlert } from './success-alert';

type UserRole = (typeof NoteWorkspaceRole)[keyof typeof NoteWorkspaceRole];

interface WorkspaceMember {
  id: string;
  name: string;
  role: UserRole;
}

function mapDtoToWorkspaceMembers(dto: NoteWorkspaceMember[]): WorkspaceMember[] {
  return dto.map((member) => ({
    id: member.id,
    name: member.name?.toString() ?? 'Unknown User',
    role: member.role,
  }));
}

//TODO: search functionality is currently client-side, we may want to move it server-side if we have a large number of members in a workspace. For now, it's fine since most workspaces will likely have a manageable number of members.
function WorkspaceMembersModal({ workspaceId }: { workspaceId: string }) {
  const queryClient = useQueryClient();
  const { alert, showAlert } = useAlert();

  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [editingMemberId, setEditingMemberId] = useState<string | null>(null);
  const [editingRole, setEditingRole] = useState<UserRole | null>(null);

  const { data: members = [], isPending: isFetchingMembers } = useQuery({
    ...getWorkspaceMembersOptions({
      path: {
        workspaceId,
      },
    }),
    select: mapDtoToWorkspaceMembers,
    enabled: isOpen,
  });

  const { mutate: updateMembers, isPending: isUpdating } = useUpdateWorkspaceMembersMutation({
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: getWorkspaceMembersOptions({
          path: { workspaceId },
        }).queryKey,
      });
      setEditingMemberId(null);
      setEditingRole(null);
      showAlert('success', 'Member Updated', 'The member role has been updated successfully.');
    },
    onError: (error) => {
      showAlert('error', 'Update Failed', `Failed to update member. Error: ${error.message}`);
    },
  });

  const filteredMembers = useMemo(() => {
    return members.filter((member) =>
      member.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [members, searchQuery]);

  const handleUpdateRole = (memberId: string, newRole: UserRole) => {
    const updatedMembers = members.map((m) => (m.id === memberId ? { ...m, role: newRole } : m));

    updateMembers({
      path: { workspaceId },
      body: updatedMembers.map((m) => ({
        id: m.id,
        role: m.role,
      })),
    });
  };

  const handleDeleteMember = (memberId: string) => {
    const remainingMembers = members.filter((m) => m.id !== memberId);

    updateMembers({
      path: { workspaceId },
      body: remainingMembers.map((m) => ({
        id: m.id,
        role: m.role,
      })),
    });
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-2">
          <Users className="size-4" />
          Members
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Workspace Members</DialogTitle>
          <DialogDescription>
            Manage member roles and permissions for this workspace
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="relative">
            <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
            <Input
              placeholder="Search members..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          <div className="max-h-96 space-y-2 overflow-y-auto rounded-lg border p-4">
            {isFetchingMembers ? (
              <div className="flex justify-center py-8">
                <Spinner className="size-6" />
              </div>
            ) : filteredMembers.length === 0 ? (
              <div className="text-muted-foreground py-8 text-center">
                {members.length === 0 ? 'No members found' : 'No matching members'}
              </div>
            ) : (
              filteredMembers.map((member) => (
                <div
                  key={member.id}
                  className="hover:bg-muted/50 flex items-center justify-between gap-3 overflow-hidden rounded-lg border p-3 transition-colors"
                >
                  <div className="min-w-0 flex-1">
                    <div className="truncate font-medium">{member.name}</div>
                  </div>

                  {editingMemberId === member.id ? (
                    <div className="flex flex-shrink-0 items-center gap-2">
                      <Select
                        value={editingRole || member.role}
                        onValueChange={(value: UserRole) => {
                          setEditingRole(value);
                          handleUpdateRole(member.id, value);
                        }}
                      >
                        <SelectTrigger className="w-28">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="owner">Owner</SelectItem>
                          <SelectItem value="admin">Admin</SelectItem>
                          <SelectItem value="viewer">Viewer</SelectItem>
                        </SelectContent>
                      </Select>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setEditingMemberId(null);
                          setEditingRole(null);
                        }}
                      >
                        Done
                      </Button>
                    </div>
                  ) : (
                    <div className="flex flex-shrink-0 items-center gap-2">
                      <Badge variant="secondary">
                        {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                      </Badge>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setEditingMemberId(member.id);
                          setEditingRole(member.role);
                        }}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive hover:bg-destructive/10"
                        onClick={() => {
                          handleDeleteMember(member.id);
                          showAlert(
                            'success',
                            'Member Removed',
                            'The member has been removed from the workspace.'
                          );
                        }}
                        disabled={isUpdating}
                      >
                        {isUpdating ? (
                          <Spinner className="size-4" />
                        ) : (
                          <Trash2 className="size-4" />
                        )}
                      </Button>
                    </div>
                  )}
                </div>
              ))
            )}
          </div>

          {alert?.type === 'success' && (
            <SuccessAlert title={alert.title} message={alert.message} />
          )}
          {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}
        </div>
      </DialogContent>
    </Dialog>
  );
}

export { WorkspaceMembersModal };
