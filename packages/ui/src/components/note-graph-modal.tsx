'use client';

import { getNoteGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { GraphData, GraphNode, GraphLink } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { useQuery } from '@tanstack/react-query';

import { Dialog, DialogContent, DialogHeader, DialogTitle } from './shadcn/dialog';

export function mapDtoNoteData(dto: NoteGraph): GraphData {
  return {
    nodes: dto.nodes.map(
      (node): GraphNode => ({
        id: node.id,
        name: node.name,
        type: node.type,
        ...(node.weight !== undefined ? { weight: node.weight } : {}),
      })
    ),
    links: dto.links.map(
      (link): GraphLink => ({
        source: link.source,
        target: link.target,
      })
    ),
  };
}

interface NoteGraphModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  noteId: string;
}

export function NoteGraphModal({ isOpen, onOpenChange, noteId }: NoteGraphModalProps) {
  const { retry } = useQueryErrorHandler();

  const {
    data: graphData = { nodes: [], links: [] },
    isError,
    error,
    isPending,
  } = useQuery({
    ...getNoteGraphOptions({
      path: { noteId },
    }),
    select: (dto: NoteGraph) => mapDtoNoteData(dto),
    enabled: isOpen,
  });

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="flex h-[600px] max-w-4xl flex-col">
        <DialogHeader>
          <DialogTitle>Note Graph</DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-hidden">
          {isPending ? (
            <div className="flex h-full items-center justify-center">
              <Spinner />
            </div>
          ) : isError ? (
            <div className="flex h-full flex-col items-center justify-center p-4">
              <QueryErrorFallback
                error={error}
                onRetry={retry}
                title="Failed to Load Graph"
                description="Unable to load the note graph. Please try again."
              />
            </div>
          ) : (
            <Graph
              data={graphData}
              currentSlug={noteId}
              options={{
                localGraph: {
                  drag: true,
                  zoom: true,
                  depth: 2,
                  scale: 0.9,
                  repelForce: 0.5,
                  centerForce: 0.2,
                  linkDistance: 30,
                  fontSize: 0.6,
                  opacityScale: 1,
                  showTags: true,
                  removeTags: [],
                  focusOnHover: true,
                  enableRadial: true,
                },
              }}
              className="size-full"
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
