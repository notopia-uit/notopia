'use client';

import { getRevisionsOptions } from '@notopia-uit/api-gen';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useQuery } from '@tanstack/react-query';
import { formatDistanceToNow } from 'date-fns';
import { History, User } from 'lucide-react';
import Link from 'next/link';
import { useParams } from 'next/navigation';

import { ScrollArea } from './shadcn/scroll-area';

interface Revision {
  id: string;
  name: string;
  createdAt: string;
}

//TODO: move it later
type ExtractOptionsData<T> = T extends (...args: any[]) => any
  ? ReturnType<T> extends import('@tanstack/react-query').UseQueryOptions<
      infer TData,
      any,
      any,
      any
    >
    ? TData
    : never
  : never;

const mapDtoToRevisionData = ({
  data,
  pagination,
}: ExtractOptionsData<typeof getRevisionsOptions>) => {
  return data.map(
    (revision) =>
      ({
        id: revision.id,
        name: revision.name,
        createdAt: revision.createdAt.toString(),
      }) as Revision
  );
};

export function RevisionSidebar({ noteId }: { noteId: string }) {
  const {
    data: RevisionData,
    isPending,
    isError,
    error,
  } = useQuery({
    ...getRevisionsOptions({
      query: {
        documentId: noteId,
      },
    }),
    select: mapDtoToRevisionData,
  });
  if (isError) {
    throw error;
  }

  const params = useParams();
  const activeRevisionId = params.revisionId as string;

  return isPending ? (
    <Spinner />
  ) : (
    <div className="bg-muted/10 flex w-80 flex-shrink-0 flex-col border-r">
      <div className="flex items-center gap-2 border-b p-4">
        <History className="text-muted-foreground h-5 w-5" />
        <h2 className="font-semibold">Version History</h2>
      </div>

      <ScrollArea className="flex-1">
        <div className="flex flex-col gap-1 p-2">
          {RevisionData.map((revision) => {
            const isActive = activeRevisionId === revision.id;

            return (
              <Link
                key={revision.id}
                href={`/note/${noteId}/revision/${revision.id}`}
                className={`hover:bg-accent flex flex-col items-start rounded-md p-3 text-left text-sm transition-colors ${
                  isActive
                    ? 'bg-accent border-l-primary border-l-4'
                    : 'border-l-4 border-l-transparent'
                }`}
              >
                <div className="mb-1 flex w-full items-center justify-between">
                  <span className="font-medium">
                    {formatDistanceToNow(revision.createdAt, { addSuffix: true })}
                  </span>
                </div>
                <span className="text-muted-foreground mb-2 line-clamp-1">{revision.name}</span>
              </Link>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}
