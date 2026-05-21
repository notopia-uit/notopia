import LocalNoteGraphView from '@notopia-uit/ui/components/local-note-graph-view';

export default async function GraphPage({ params }: { params: Promise<{ noteId: string }> }) {
  const { noteId } = await params;
  return (
    <div className="h-screen w-full">
      <LocalNoteGraphView noteId={noteId} />
    </div>
  );
}
