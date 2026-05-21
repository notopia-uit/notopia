'use client';

import { getWorkspaceGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { GraphSettingsDialog } from '@notopia-uit/ui/components/graph-settings-dialog';
import { Icons } from '@notopia-uit/ui/components/icons';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { GraphData, GraphNode, GraphLink, D3Config } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';

const defaultGlobalGraphSettings: Partial<D3Config> = {
  drag: true,
  zoom: true,
  depth: -1,
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

export default function GraphView({ workspaceId }: { workspaceId: string }) {
  const { retry } = useQueryErrorHandler();
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [graphSettings, setGraphSettings] = useState<Partial<D3Config>>(defaultGlobalGraphSettings);

  const {
    data: graphData = { nodes: [], links: [] },
    isError,
    error,
    isPending,
  } = useQuery({
    ...getWorkspaceGraphOptions({
      path: { workspaceId: workspaceId },
    }),
    select: (dto: NoteGraph) => mapDtoNoteData(dto),
  });

  const handleSaveSettings = (settings: Partial<D3Config>) => {
    setGraphSettings(settings);
  };

  if (isPending) {
    return <Spinner />;
  }

  if (isError) {
    return (
      <div className="flex h-[400px] items-center justify-center">
        <QueryErrorFallback
          error={error}
          onRetry={retry}
          title="Failed to Load Graph"
          description="Unable to load the workspace graph. Please try again."
        />
      </div>
    );
  }

  return (
    <div className="relative h-full w-full">
      <Graph
        data={graphData}
        options={{
          localGraph: graphSettings,
        }}
      />
      <div className="absolute top-4 right-4 z-10">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setIsSettingsOpen(true)}
          aria-label="Graph settings"
        >
          <Icons.Settings className="mr-2 size-4" />
          Settings
        </Button>
      </div>
      <GraphSettingsDialog
        isOpen={isSettingsOpen}
        onOpenChange={setIsSettingsOpen}
        isLocalGraph={false}
        currentSettings={graphSettings}
        onSave={handleSaveSettings}
      />
    </div>
  );
}
