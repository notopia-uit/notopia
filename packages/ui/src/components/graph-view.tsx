'use client';

import { getWorkspaceGraphOptions, NoteGraph } from '@notopia-uit/api-gen';
import { GraphData, GraphNode, GraphLink } from '@notopia-uit/ui/graph-view/graph';
import Graph from '@notopia-uit/ui/graph-view/graph';
import { useSuspenseQuery } from '@tanstack/react-query';

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
  const { data: graphData } = useSuspenseQuery({
    ...getWorkspaceGraphOptions({
      path: { workspaceId: workspaceId },
    }),
    select: (dto: NoteGraph) => mapDtoNoteData(dto),
  });

  return <Graph data={graphData} />;
}
