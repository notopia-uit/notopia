'use client';

import {
  useRenameNoteMutation,
  useGetNoteQuery,
  getNoteOptions,
  getWorkspaceTreeOptions,
} from '@notopia-uit/api-gen';
import { useQueryClient } from '@tanstack/react-query';
import { useState, useEffect, useCallback } from 'react';

import { Input } from './shadcn/input';

interface NoteTitleProps {
  noteId: string;
  workspaceId?: string;
  isEditing?: boolean;
}

export function NoteTitle({ noteId, workspaceId, isEditing = true }: NoteTitleProps) {
  const [title, setTitle] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const queryClient = useQueryClient();

  const { data: noteData } = useGetNoteQuery({
    path: { noteId },
  });

  const { mutate: renameNote, isPending: isRenamingNote } = useRenameNoteMutation({
    onSuccess: () => {
      if (workspaceId) {
        queryClient.invalidateQueries({
          queryKey: getNoteOptions({ path: { noteId } }).queryKey,
        });
        queryClient.invalidateQueries({
          queryKey: getWorkspaceTreeOptions({ path: { workspaceId } }).queryKey,
        });
      }
    },
  });

  useEffect(() => {
    if (noteData?.name) {
      setTitle(noteData.name);
    }
  }, [noteData?.name, noteId]);

  const handleTitleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const newTitle = e.target.value;
    setTitle(newTitle);
  }, []);

  const handleTitleBlur = useCallback(() => {
    setIsFocused(false);
    if (title && title !== noteData?.name) {
      renameNote({
        path: {
          noteId: noteId,
        },
        body: {
          name: title,
        },
      });
    }
  }, [title, noteData?.name, noteId, renameNote]);

  if (!isEditing) {
    return <div className="text-foreground px-4 py-3 text-2xl font-bold">{title}</div>;
  }

  return (
    <div className="bg-background/95 supports-backdrop-filter:bg-background/60 sticky top-0 z-40 border-b px-4 py-3 backdrop-blur-sm">
      <Input
        value={title}
        onChange={handleTitleChange}
        onFocus={() => setIsFocused(true)}
        onBlur={handleTitleBlur}
        placeholder="Untitled"
        disabled={isRenamingNote}
        className={`placeholder-muted-foreground border-0 bg-transparent text-2xl font-bold focus-visible:ring-0 focus-visible:ring-offset-0 ${
          isFocused ? 'text-foreground' : 'text-foreground'
        }`}
      />
    </div>
  );
}
