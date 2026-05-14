import { redirect } from 'next/navigation';

export default async function RevisionLayout({
  params,
}: {
  params: Promise<{ workspaceId: string; noteId: string }>;
}) {
  const { workspaceId, noteId } = await params;
  redirect(`/workspace/${workspaceId}/note/${noteId}`);
}
