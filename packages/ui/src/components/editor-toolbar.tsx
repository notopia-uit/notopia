'use client';

import { Network } from 'lucide-react';
import { useRouter, useParams } from 'next/navigation';

import { Button } from './shadcn/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './shadcn/tooltip';
import { RevisionModal } from './revision-modal';

interface EditorToolbarProps {
  noteId: string;
  currentEditor: any;
}

export function EditorToolbar({ noteId, currentEditor }: EditorToolbarProps) {
  const router = useRouter();
  const params = useParams();

  const handleGraphOpen = () => {
    const workspaceId = params.workspaceId as string;
    router.push(`/workspace/${workspaceId}/note/${noteId}/graph`);
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
      </div>
    </TooltipProvider>
  );
}
