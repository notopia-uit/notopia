import { RevisionSidebar } from '@notopia-uit/ui/components/revision-sidebar';

export default async function RevisionLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ noteId: string }>;
}) {
  const { noteId } = await params;
  return (
    <div className="bg-background mx-auto max-w-6xl rounded-lg border shadow-sm">
      <div className="flex h-175 w-full">
        <RevisionSidebar noteId={noteId} />
        <div className="flex min-w-0 flex-1 flex-col">{children}</div>
      </div>
    </div>
  );
}
