import { LocalNoteGraphViewBoundary } from './note-graph-view-boundary';

export default async function GraphPage({
  params,
}: {
  params: Promise<{ noteId: string; workspaceId: string }>;
}) {
  const { noteId, workspaceId } = await params;
  return (
    <div className="h-screen w-full">
      <LocalNoteGraphViewBoundary noteId={noteId} workspaceId={workspaceId} />
    </div>
  );
}
