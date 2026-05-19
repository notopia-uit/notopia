'use client';

import { useRenameNoteMutation } from '@notopia-uit/api-gen';
import { useQueryClient } from '@tanstack/react-query';
import { Input } from './shadcn/input';
import { useState, useEffect, useCallback } from 'react';

interface NoteTitleProps {
  noteId: string;
  initialTitle?: string;
  workspaceId?: string;
  isEditing?: boolean;
}

export function NoteTitle({
  noteId,
  initialTitle = 'Untitled',
  workspaceId,
  isEditing = true,
}: NoteTitleProps) {
  const [title, setTitle] = useState(initialTitle);
  const [isFocused, setIsFocused] = useState(false);
  const queryClient = useQueryClient();

  const { mutate: renameNote, isPending: isRenamingNote } = useRenameNoteMutation({
    onSuccess: () => {
      if (workspaceId) {
        queryClient.invalidateQueries({ queryKey: ['workspace', workspaceId, 'tree'] });
      }
    },
  });

  useEffect(() => {
    setTitle(initialTitle);
  }, [initialTitle, noteId]);

  const handleTitleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newTitle = e.target.value;
      setTitle(newTitle);
    },
    []
  );

  const handleTitleBlur = useCallback(() => {
    setIsFocused(false);
    if (title && title !== initialTitle) {
      renameNote({
        path: {
          noteId: noteId,
        },
        body: {
          name: title,
        },
      });
    }
  }, [title, initialTitle, noteId, renameNote]);

  if (!isEditing) {
    return (
      <div className="px-4 py-3 text-2xl font-bold text-foreground">
        {title}
      </div>
    );
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
        className={`border-0 bg-transparent text-2xl font-bold focus-visible:ring-0 focus-visible:ring-offset-0 placeholder-muted-foreground ${
          isFocused ? 'text-foreground' : 'text-foreground'
        }`}
      />
    </div>
  );
}
