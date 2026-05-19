'use client';

import { getNoteGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { GraphData, GraphNode, GraphLink } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { useQuery } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { Button } from './shadcn/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from './shadcn/dialog';

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

  if (isError) {
    return (
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-4xl h-[600px]">
          <DialogHeader>
            <DialogTitle>Note Graph</DialogTitle>
          </DialogHeader>
          <div className="flex items-center justify-center h-full">
            <p className="text-red-500">
              Error loading graph: {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[600px] flex flex-col">
        <DialogHeader>
          <DialogTitle>Note Graph</DialogTitle>
        </DialogHeader>
        <div className="flex-1 overflow-hidden">
          {isPending ? (
            <div className="flex items-center justify-center h-full">
              <Spinner />
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
              className="h-full w-full"
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
