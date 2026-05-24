'use client';

import {
  getWorkspaceMembersOptions,
  NoteUser,
  NoteWorkspaceMember,
  NoteWorkspaceRole,
  searchUsersOptions,
  searchUsersQueryKey,
  useUpdateWorkspaceMembersMutation,
} from '@notopia-uit/api-gen';
import { Badge } from '@notopia-uit/ui/components/shadcn/badge';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useDebouncedValue } from '@notopia-uit/ui/hooks/use-debounced-value';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Search, Trash2, UserPlus, Users } from 'lucide-react';
import { useState, useMemo, useRef, useEffect } from 'react';

import { RoleSelectItems } from './role-select-items';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './shadcn/dialog';
import { Select, SelectContent, SelectTrigger, SelectValue } from './shadcn/select';

type UserRole = (typeof NoteWorkspaceRole)[keyof typeof NoteWorkspaceRole];

interface WorkspaceMember {
  id: string;
  name: string;
  role: UserRole;
}

interface SearchUserMember {
  id: string;
  username: string;
  name: string;
  email: string;
}

function mapDtoToWorkspaceMembers(dto: NoteWorkspaceMember[]): WorkspaceMember[] {
  return dto.map((member) => ({
    id: member.id,
    name: member.name?.toString() ?? 'Unknown User',
    role: member.role,
  }));
}

function mapDtoToSearchUserMembers(dto: NoteUser[]): SearchUserMember[] {
  return dto.map((user) => ({
    id: user.id,
    username: user.username,
    name: user.name ?? user.username ?? 'Unknown User',
    email: user.email ?? '',
  }));
}

//TODO: search functionality is currently client-side, we may want to move it server-side if we have a large number of members in a workspace. For now, it's fine since most workspaces will likely have a manageable number of members.
function WorkspaceMembersModal({ workspaceId }: { workspaceId: string }) {
  const queryClient = useQueryClient();
  const { showAlert } = useAlert();

  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [addUserSearch, setAddUserSearch] = useState('');
  const debouncedAddUserSearch = useDebouncedValue(addUserSearch, 300);
  const [showSearchResults, setShowSearchResults] = useState(false);
  const searchResultsRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [editingMemberId, setEditingMemberId] = useState<string | null>(null);
  const [editingRole, setEditingRole] = useState<UserRole | null>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        searchResultsRef.current &&
        !searchResultsRef.current.contains(event.target as Node) &&
        searchInputRef.current &&
        !searchInputRef.current.contains(event.target as Node)
      ) {
        setShowSearchResults(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const { data: members = [], isPending: isFetchingMembers } = useQuery({
    ...getWorkspaceMembersOptions({
      path: {
        workspaceId,
      },
    }),
    select: mapDtoToWorkspaceMembers,
    enabled: isOpen,
  });

  const { data: searchUserData = [], isFetching: isSearchingUsers } = useQuery({
    ...searchUsersOptions({
      query: {
        keyword: debouncedAddUserSearch,
        isActive: true,
        limit: 10,
        excludeMemberInWorkspaceId: workspaceId,
      },
    }),
    select: mapDtoToSearchUserMembers,
    enabled: isOpen && debouncedAddUserSearch.trim().length > 0,
  });

  const { mutate: updateMembers, isPending: isUpdating } = useUpdateWorkspaceMembersMutation({
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: getWorkspaceMembersOptions({
            path: { workspaceId },
          }).queryKey,
        }),
        queryClient.invalidateQueries({
          queryKey: searchUsersQueryKey({
            query: {
              keyword: debouncedAddUserSearch,
              isActive: true,
              limit: 10,
              excludeMemberInWorkspaceId: workspaceId,
            },
          }),
        }),
      ]);
      setEditingMemberId(null);
      setEditingRole(null);
    },
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Update Failed',
        message: `Failed to update member. Error: ${error.message}`,
      });
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

  const handleAddMember = (userId: string) => {
    const allMembers = [
      ...members.map((m) => ({ id: m.id, role: m.role })),
      { id: userId, role: NoteWorkspaceRole.VIEWER },
    ];

    updateMembers(
      {
        path: { workspaceId },
        body: allMembers,
      },
      {
        onSuccess: () => {
          setAddUserSearch('');
          setShowSearchResults(false);
          showAlert({
            type: 'success',
            title: 'Member Added',
            message: 'The user has been added to the workspace.',
          });
        },
      },
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-2">
          <Users className="size-4" />
          Members
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Workspace Members</DialogTitle>
          <DialogDescription>
            Manage member roles and permissions for this workspace
          </DialogDescription>
        </DialogHeader>

        <div className="min-w-0 space-y-4">
          <div className="relative">
            <UserPlus className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
            <Input
              ref={searchInputRef}
              placeholder="Search users to add..."
              value={addUserSearch}
              onChange={(e) => {
                setAddUserSearch(e.target.value);
                setShowSearchResults(true);
              }}
              onFocus={() => setShowSearchResults(true)}
              className="pl-10"
            />
            {showSearchResults && addUserSearch.trim().length > 0 && (
              <div
                ref={searchResultsRef}
                className="bg-popover absolute top-full z-50 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border shadow-md"
              >
                {isSearchingUsers ? (
                  <div className="flex justify-center py-4">
                    <Spinner className="size-4" />
                  </div>
                ) : searchUserData.length === 0 ? (
                  <div className="text-muted-foreground py-4 text-center text-sm">
                    No users found
                  </div>
                ) : (
                  searchUserData.map((user) => (
                    <button
                      key={user.id}
                      type="button"
                      className="hover:bg-muted/50 flex w-full items-center gap-3 px-3 py-2 text-left transition-colors disabled:opacity-50"
                      onClick={() => handleAddMember(user.id)}
                      disabled={isUpdating}
                    >
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-sm font-medium">{user.name}</div>
                        <div className="text-muted-foreground truncate text-xs">{user.email}</div>
                      </div>
                      <UserPlus className="text-muted-foreground size-4 shrink-0" />
                    </button>
                  ))
                )}
              </div>
            )}
          </div>

          <div className="relative">
            <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
            <Input
              placeholder="Filter members..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          <div className="max-h-96 max-w-full space-y-2 overflow-y-auto rounded-lg border p-4">
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
                  <div className="min-w-0 flex-1 overflow-hidden">
                    <div className="truncate font-medium">{member.name}</div>
                  </div>

                  {editingMemberId === member.id ? (
                    <div className="flex shrink-0 items-center gap-2">
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
                          <RoleSelectItems />
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
                    <div className="flex shrink-0 items-center gap-2">
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
                          showAlert({
                            type: 'success',
                            title: 'Member Removed',
                            message: 'The member has been removed from the workspace.',
                          });
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
        </div>
      </DialogContent>
    </Dialog>
  );
}

export { WorkspaceMembersModal };
