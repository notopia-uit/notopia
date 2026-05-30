'use client';
import {
  NoteTrashedFolder,
  NoteTrashedNote,
  showTrashOptions,
  getWorkspaceTreeOptions,
  useRestoreTrashedWorkspaceItemsMutation,
  usePermanentlyDeleteWorkspaceItemsMutation,
} from '@notopia-uit/api-gen';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Checkbox } from '@notopia-uit/ui/components/shadcn/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@notopia-uit/ui/components/shadcn/dropdown-menu';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@notopia-uit/ui/components/shadcn/table';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { FileText, Folder, MoreVertical, RotateCcw, Trash2 } from 'lucide-react';
import React, { useMemo, useState, useRef } from 'react';

interface TrashedBy {
  by: 'purpose' | 'parent';
  at: string;
}

interface TrashedItem {
  id: string;
  name: string;
  trashed: TrashedBy;
}

interface TrashedData {
  notes: TrashedItem[];
  folders: TrashedItem[];
}
type TrashedDataDto = {
  notes: NoteTrashedNote[];
  folders: NoteTrashedFolder[];
};

export function mapDtoTrashedData(dto: TrashedDataDto): TrashedData {
  const mapItem = (item: NoteTrashedNote | NoteTrashedFolder): TrashedItem => {
    const dateString =
      item.trashed.at instanceof Date
        ? item.trashed.at.toISOString()
        : new Date(item.trashed.at).toISOString();

    return {
      id: item.id,
      name: item.name,
      trashed: {
        by: item.trashed.by,
        at: dateString,
      },
    };
  };

  return {
    notes: dto.notes ? dto.notes.map(mapItem) : [],
    folders: dto.folders ? dto.folders.map(mapItem) : [],
  };
}

const formatDate = (isoString: string) => {
  const date = new Date(isoString);
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
  }).format(date);
};

const EMPTY_TRASH_DATA: TrashedData = { notes: [], folders: [] };
export default function TrashedFileManager({ workspaceId }: { workspaceId: string }) {
  const queryClient = useQueryClient();
  const { retry } = useQueryErrorHandler();

  const { data, isError, error, isPending } = useQuery({
    ...showTrashOptions({ path: { workspaceId: workspaceId } }),
    select: mapDtoTrashedData,
  });

  if (isError) {
    return (
      <div className="p-4">
        <QueryErrorFallback
          error={error}
          onRetry={retry}
          title="Failed to Load Trash"
          description="Unable to load deleted items. Please try again."
        />
      </div>
    );
  }

  const trashedData = data || EMPTY_TRASH_DATA;
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());

  const displayData = useMemo(() => {
    const foldersMapped = trashedData.folders.map((folder) => ({
      ...folder,
      displayType: 'Folder',
      icon: Folder,
    }));

    const notesMapped = trashedData.notes.map((note) => ({
      ...note,
      displayType: 'Note',
      icon: FileText,
    }));

    return [...foldersMapped, ...notesMapped].sort(
      (a, b) => new Date(b.trashed.at).getTime() - new Date(a.trashed.at).getTime()
    );
  }, [trashedData]);

  const { showAlert } = useAlert();
  const { mutate: deleteItems, isPending: isDeleting } = usePermanentlyDeleteWorkspaceItemsMutation(
    {
      onSuccess: async (_, variables) => {
        const { noteIds, folderIds } = variables.body;
        queryClient.setQueryData<TrashedDataDto>(
          showTrashOptions({ path: { workspaceId } }).queryKey,
          (oldData) => {
            if (!oldData) return oldData;
            return {
              ...oldData,
              notes:
                oldData.notes?.filter((note) => !(noteIds ? noteIds : []).includes(note.id)) || [],
              folders:
                oldData.folders?.filter(
                  (folder) => !(folderIds ? folderIds : []).includes(folder.id)
                ) || [],
            };
          }
        );

        setSelectedItems(new Set());
        await queryClient.invalidateQueries({
          queryKey: showTrashOptions({ path: { workspaceId: workspaceId } }).queryKey,
        });
        showAlert({
          type: 'success',
          title: 'Successfully Deleted',
          message: `Selected items have been permanently deleted.`,
        });
      },
      onError: (error) => {
        showAlert({
          type: 'error',
          title: 'Failed to Delete',
          message: `An error occurred while trying to permanently delete the selected items. Please try again.
          ${error instanceof Error ? error.message : ''}`,
        });
      },
    }
  );
  const { mutate: restoreItems, isPending: isRestoring } = useRestoreTrashedWorkspaceItemsMutation({
    onSuccess: async (_, variables) => {
      const { noteIds, folderIds } = variables.body;
      queryClient.setQueryData<TrashedDataDto>(
        showTrashOptions({ path: { workspaceId } }).queryKey,
        (oldData) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            notes:
              oldData.notes?.filter((note) => !(noteIds ? noteIds : []).includes(note.id)) || [],
            folders:
              oldData.folders?.filter(
                (folder) => !(folderIds ? folderIds : []).includes(folder.id)
              ) || [],
          };
        }
      );
      await queryClient.invalidateQueries({
        queryKey: getWorkspaceTreeOptions({ path: { workspaceId } }).queryKey,
      });
      await queryClient.invalidateQueries({
        queryKey: showTrashOptions({ path: { workspaceId: workspaceId } }).queryKey,
      });
      setSelectedItems(new Set());
      showAlert({
        type: 'success',
        title: 'Successfully Restored',
        message: `Selected items have been restored.`,
      });
    },
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Failed to Restore',
        message: `An error occurred while trying to restore the selected items. Please try again.
        ${error instanceof Error ? error.message : ''}`,
      });
    },
  });
  const toggleSelection = (id: string) => {
    const newSelection = new Set(selectedItems);
    if (newSelection.has(id)) {
      newSelection.delete(id);
    } else {
      newSelection.add(id);
    }
    setSelectedItems(newSelection);
  };
  const toggleAll = () => {
    if (selectedItems.size === displayData.length) {
      setSelectedItems(new Set());
    } else {
      setSelectedItems(new Set(displayData.map((item) => item.id)));
    }
  };
  const handleRestoreSelected = () => {
    const noteIds: string[] = [];
    const folderIds: string[] = [];
    for (const item of displayData) {
      if (!selectedItems.has(item.id)) continue;
      if (item.displayType === 'Note') noteIds.push(item.id);
      else folderIds.push(item.id);
    }
    restoreItems({
      path: { workspaceId },
      body: { noteIds, folderIds },
    });
  };
  const handleDeleteSelected = () => {
    const noteIds: string[] = [];
    const folderIds: string[] = [];
    if (selectedItems.size > 0) {
      for (const item of displayData) {
        if (!selectedItems.has(item.id)) continue;
        if (item.displayType === 'Note') noteIds.push(item.id);
        else folderIds.push(item.id);
      }
    } else {
      noteIds.push(...trashedData.notes.map((n) => n.id));
      folderIds.push(...trashedData.folders.map((f) => f.id));
    }
    deleteItems({
      path: { workspaceId },
      body: { noteIds, folderIds },
    });
  };
  return isPending ? (
    <Spinner />
  ) : (
    <div className="w-full font-sans">
      <div className="w-full">
        {/* Header Actions */}
        <div className="flex items-center justify-end pr-6 pb-6">
          <div className="flex gap-3 space-x-0.5">
            <Button
              variant="destructive"
              className="gap-0.5 space-x-0.5 bg-transparent"
              onClick={handleDeleteSelected}
              disabled={isDeleting}
            >
              {isDeleting ? (
                <Spinner />
              ) : selectedItems.size > 0 ? (
                `Delete Permanently (${selectedItems.size})`
              ) : (
                'Empty Trash'
              )}{' '}
            </Button>
            {selectedItems.size > 0 && (
              <Button onClick={handleRestoreSelected} disabled={isRestoring}>
                {isRestoring ? <Spinner></Spinner> : 'Restore Selected'}
              </Button>
            )}
          </div>
        </div>

        {/* Data Table */}
        <Table>
          <TableHeader className="bg-transparent hover:bg-transparent">
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-12 pl-6 text-center">
                <Checkbox
                  checked={selectedItems.size === displayData.length && displayData.length > 0}
                  onCheckedChange={toggleAll}
                />
              </TableHead>
              <TableHead className="font-medium">Name</TableHead>
              <TableHead className="w-24 text-right font-medium">Type</TableHead>
              <TableHead className="w-48 pr-6 text-right font-medium">Deleted Date</TableHead>
              <TableHead className="w-12"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {displayData.map((item) => {
              const Icon = item.icon;
              const isSelected = selectedItems.has(item.id);

              return (
                <TableRow
                  key={item.id}
                  className={`transition-colors ${isSelected ? `bg-accent-foreground` : ''}`}
                >
                  <TableCell className="pl-6">
                    <Checkbox
                      checked={isSelected}
                      onCheckedChange={() => toggleSelection(item.id)}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-3">
                      <Icon
                        className={`size-5 ${
                          item.displayType === 'Folder' ? `text-foreground` : `text-background`
                        }`}
                      />
                      <span className="font-medium">{item.name}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">{item.displayType}</TableCell>
                  <TableCell className="pr-6 text-right">{formatDate(item.trashed.at)}</TableCell>
                  <TableCell className="pr-6 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="size-8 p-0">
                          <MoreVertical className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="w-48">
                        <DropdownMenuItem className="cursor-pointer">
                          <RotateCcw className="mr-2 size-4" />
                          <span>Restore</span>
                        </DropdownMenuItem>
                        <DropdownMenuItem className="cursor-pointer">
                          <Trash2 className="mr-2 size-4" />
                          <span>Delete Permanently</span>
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
