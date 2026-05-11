'use client';

import { getWorkspaceGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { GraphData, GraphNode, GraphLink } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { fetchAccessTokenClientSide } from '@notopia-uit/ui/lib/get-access-token-client-side';
import { useQuery } from '@tanstack/react-query';

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
//TODO: handle loading and error states
export default function GraphView({ workspaceId }: { workspaceId: string }) {
  const {
    data: graphData = { nodes: [], links: [] },
    isError,
    error,
  } = useQuery({
    ...getWorkspaceGraphOptions({
      path: { workspaceId: workspaceId },
    }),
    select: (dto: NoteGraph) => mapDtoNoteData(dto),
  });

  if (isError) {
    throw error;
  }
  return <Graph data={graphData} />;
}
