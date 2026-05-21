'use client';

import { getNoteGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { GraphData, GraphNode, GraphLink, D3Config } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';

import { Dialog, DialogContent, DialogHeader, DialogTitle } from './shadcn/dialog';
import { Button } from './shadcn/button';
import { GraphSettingsDialog } from './graph-settings-dialog';
import { Icons } from './icons';

const defaultLocalGraphSettings: Partial<D3Config> = {
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
};

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
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [graphSettings, setGraphSettings] = useState<Partial<D3Config>>(
    defaultLocalGraphSettings
  );

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

  const handleSaveSettings = (settings: Partial<D3Config>) => {
    setGraphSettings(settings);
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="flex h-[600px] max-w-4xl flex-col">
          <DialogHeader className="flex flex-row items-center justify-between">
            <DialogTitle>Note Graph</DialogTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsSettingsOpen(true)}
              aria-label="Graph settings"
            >
              <Icons.Settings className="size-4" />
            </Button>
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
                  localGraph: graphSettings,
                }}
                className="size-full"
              />
            )}
          </div>
        </DialogContent>
      </Dialog>
      <GraphSettingsDialog
        isOpen={isSettingsOpen}
        onOpenChange={setIsSettingsOpen}
        isLocalGraph={true}
        currentSettings={graphSettings}
        onSave={handleSaveSettings}
      />
    </>
  );
}
