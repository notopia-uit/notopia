'use client';

import { getWorkspaceGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { GraphSettingsDialog } from '@notopia-uit/ui/components/graph-settings-dialog';
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
  const [showSettings, setShowSettings] = useState(true);
  const [graphSettings, setGraphSettings] = useState<Partial<D3Config>>(defaultGlobalGraphSettings);
  const [showOrphansOnly, setShowOrphansOnly] = useState(false);

  const {
    data: graphData = { nodes: [], links: [] },
    isError,
    error,
    isPending,
  } = useQuery({
    ...getWorkspaceGraphOptions({
      path: { workspaceId: workspaceId },
      query: { includeOrphans: showOrphansOnly },
    }),
    select: (dto: NoteGraph) => mapDtoNoteData(dto),
  });

  const handleSettingsChange = (settings: Partial<D3Config>) => {
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
    <div className="relative size-full">
      <Graph
        data={graphData}
        options={{
          localGraph: graphSettings,
        }}
      />
      <GraphSettingsDialog
        isOpen={showSettings}
        onOpenChange={setShowSettings}
        isLocalGraph={false}
        currentSettings={graphSettings}
        onSettingsChange={handleSettingsChange}
        showOrphansOnly={showOrphansOnly}
        onOrphansOnlyChange={setShowOrphansOnly}
      />
    </div>
  );
}
