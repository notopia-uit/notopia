import { RevisionDetailContainer } from '@notopia-uit/ui/components/revision-detail-container';

export default async function RevisionPage({
  params,
}: {
  params: Promise<{ noteId: string; revisionId: string }>;
}) {
  const { revisionId } = await params;
  return <RevisionDetailContainer selectedRevisionId={revisionId} />;
}
