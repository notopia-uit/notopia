'use client';

import { getNoteGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { GraphSettingsDialog } from '@notopia-uit/ui/components/graph-settings-dialog';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { GraphData, GraphNode, GraphLink, D3Config } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

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

interface LocalNoteGraphViewProps {
  noteId: string;
  workspaceId?: string;
}

export default function LocalNoteGraphView({ noteId, workspaceId }: LocalNoteGraphViewProps) {
  const { retry } = useQueryErrorHandler();
  const router = useRouter();
  const [showSettings, setShowSettings] = useState(true);
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
  });

  const handleSettingsChange = (settings: Partial<D3Config>) => {
    setGraphSettings(settings);
  };

  const handleNodeClick = (clickedNodeId: string, nodeType: 'note' | 'tag') => {
    // Only navigate for notes
    if (nodeType === 'note' && workspaceId) {
      router.push(`/workspace/${workspaceId}/note/${clickedNodeId}/`);
    }
  };

  if (isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex h-full flex-col items-center justify-center p-4">
        <QueryErrorFallback
          error={error}
          onRetry={retry}
          title="Failed to Load Graph"
          description="Unable to load the note graph. Please try again."
        />
      </div>
    );
  }

  return (
    <div className="relative size-full">
      <Graph
        data={graphData}
        currentSlug={noteId}
        options={{
          localGraph: graphSettings,
        }}
        onNodeClick={handleNodeClick}
      />
      <GraphSettingsDialog
        isOpen={showSettings}
        onOpenChange={setShowSettings}
        isLocalGraph={true}
        currentSettings={graphSettings}
        onSettingsChange={handleSettingsChange}
      />
    </div>
  );
}
