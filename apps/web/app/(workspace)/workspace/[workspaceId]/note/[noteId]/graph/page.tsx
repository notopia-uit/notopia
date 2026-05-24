import LocalNoteGraphView from '@notopia-uit/ui/components/local-note-graph-view';

export default async function GraphPage({
  params,
}: {
  params: Promise<{ noteId: string; workspaceId: string }>;
}) {
  const { noteId, workspaceId } = await params;
  return (
    <div className="h-screen w-full">
      <LocalNoteGraphView noteId={noteId} workspaceId={workspaceId} />
    </div>
  );
}
