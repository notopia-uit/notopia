import LocalNoteGraphView from '@notopia-uit/ui/components/local-note-graph-view';

interface GraphPageProps {
  params: {
    noteId: string;
  };
}

export default function GraphPage({ params }: GraphPageProps) {
  return (
    <div className="h-screen w-full">
      <LocalNoteGraphView noteId={params.noteId} />
    </div>
  );
}
