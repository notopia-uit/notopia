'use client';
import { NoteTrashedFolder, NoteTrashedNote, showTrashOptions, useRestoreTrashedWorkspaceItemsMutation} from '@notopia-uit/api-gen';
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
import { useQueryClient, useSuspenseQuery } from '@tanstack/react-query';
import { FileText, Folder, MoreVertical, RotateCcw, Trash2 } from 'lucide-react';
import React, { useMemo, useState } from 'react';

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

export default function TrashedFileManager({ workspaceId }: { workspaceId: string }) {
  const queryClient =useQueryClient();
  const { data: trashedData } = useSuspenseQuery({
    ...showTrashOptions({ path: { workspaceId: workspaceId } }),
    select: (data) => mapDtoTrashedData(data),
  });

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

const { mutate: restoreItems, isPending: isRestoring } = useRestoreTrashedWorkspaceItemsMutation({
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: showTrashOptions({ path: { workspaceId: workspaceId } }).queryKey });
      setSelectedItems(new Set());
    }
  }

  )
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

  return (
    <div className="w-full font-sans text-zinc-300">
      <div className="w-full">
        {/* Header Actions */}
        <div className="flex items-center justify-end pr-6 pb-6">
          <div className="flex gap-3 space-x-0.5">
            <Button
              variant="destructive"
              className="gap-0.5 space-x-0.5 border-white/10 bg-transparent text-zinc-300 hover:bg-white/5 hover:text-white"
            >
              Empty Trash
            </Button>
            {selectedItems.size > 0 && (
              <Button className="bg-blue-600 text-white hover:bg-blue-700"
                onClick= {() =>
              restoreItems({path:{
                workspaceId: workspaceId,
              },body:{
                noteIds: Array.from(selectedItems).filter(id => displayData.find(item => item.id === id)?.displayType === 'Note'),
                folderIds: Array.from(selectedItems).filter(id => displayData.find(item => item.id === id)?.displayType === 'Folder'),
              }})} disabled={isRestoring}>{isRestoring? <Spinner></Spinner> : 'Restore Selected'}</Button>
            )}
          </div>
        </div>

        {/* Data Table */}
        <Table>
          <TableHeader className="bg-transparent hover:bg-transparent">
            <TableRow className="border-white/10 hover:bg-transparent">
              <TableHead className="w-12 pl-6 text-center">
                <Checkbox
                  checked={selectedItems.size === displayData.length && displayData.length > 0}
                  onCheckedChange={toggleAll}
                  className="border-white/20 data-[state=checked]:border-blue-600 data-[state=checked]:bg-blue-600"
                />
              </TableHead>
              <TableHead className="font-medium text-zinc-400">Name</TableHead>
              <TableHead className="w-24 text-right font-medium text-zinc-400">Type</TableHead>
              <TableHead className="w-48 pr-6 text-right font-medium text-zinc-400">
                Deleted Date
              </TableHead>
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
                  className={`border-white/5 transition-colors hover:bg-white/5 ${isSelected ? `bg-white/5` : ''}`}
                >
                  <TableCell className="pl-6">
                    <Checkbox
                      checked={isSelected}
                      onCheckedChange={() => toggleSelection(item.id)}
                      className="border-white/20 data-[state=checked]:border-blue-600 data-[state=checked]:bg-blue-600"
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-3">
                      <Icon
                        className={`size-5 ${
                          item.displayType === 'Folder' ? `text-blue-400` : `text-zinc-400`
                        }`}
                      />
                      <span className="font-medium text-zinc-200">{item.name}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right text-zinc-400">{item.displayType}</TableCell>
                  <TableCell className="pr-6 text-right text-zinc-400">
                    {formatDate(item.trashed.at)}
                  </TableCell>
                  <TableCell className="pr-6 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          className="size-8 p-0 text-zinc-400 hover:bg-white/10 hover:text-zinc-100"
                        >
                          <MoreVertical className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent
                        align="end"
                        className="w-48 border-white/10 bg-[#2c2c2e] text-zinc-200"
                      >
                        <DropdownMenuItem className="cursor-pointer hover:bg-white/10 focus:bg-white/10 focus:text-white">
                          <RotateCcw className="mr-2 size-4" />
                          <span>Restore</span>
                        </DropdownMenuItem>
                        <DropdownMenuItem className="cursor-pointer text-red-400 hover:bg-red-500/10 focus:bg-red-500/10 focus:text-red-400">
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
