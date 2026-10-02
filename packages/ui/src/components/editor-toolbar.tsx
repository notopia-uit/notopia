'use client';

import { Network } from 'lucide-react';

import { Button } from './shadcn/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './shadcn/tooltip';
import { NoteLinksModal } from './note-links-modal';
import { RevisionModal } from './revision-modal';

interface EditorToolbarProps {
  noteId: string;
  workspaceId: string;
  currentEditor: any;
  onNavigate: (href: string) => void;
}

export function EditorToolbar({ noteId, workspaceId, currentEditor, onNavigate }: EditorToolbarProps) {
  const handleGraphOpen = () => {
    onNavigate(`/workspace/${workspaceId}/note/${noteId}/graph`);
  };

  return (
    <TooltipProvider>
      <div className="bg-background/95 supports-backdrop-filter:bg-background/60 sticky top-14 z-40 flex gap-2 border-b px-4 py-2 backdrop-blur-sm">
        <RevisionModal noteId={noteId} currentEditor={currentEditor} />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              onClick={handleGraphOpen}
              aria-label="Open note graph"
            >
              <Network className="size-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>View note graph</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <NoteLinksModal noteId={noteId} workspaceId={workspaceId} onNavigate={onNavigate} />
          </TooltipTrigger>
          <TooltipContent>View note links</TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  );
}
