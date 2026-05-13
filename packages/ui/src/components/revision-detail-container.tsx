'use client';
import { Block } from '@blocknote/core';
import { useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import { DocumentRevisionWithContent, getRevisionWithContentOptions } from '@notopia-uit/api-gen';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner.jsx';
import { useQuery } from '@tanstack/react-query';

import { ScrollArea } from './shadcn/scroll-area';

function ReadOnlyEditor({ initialContent }: { initialContent: Block[] }) {
  const editor = useCreateBlockNote({
    initialContent: initialContent,
  });

  return (
    <div className="p-4">
      <BlockNoteView editor={editor} editable={false} />
    </div>
  );
}

type RevisionDetail = {
  id: string;
  name: string;
  content: Block[];
  createdAt: string;
};

function mapDtoToRevisionDetail(dto: DocumentRevisionWithContent): RevisionDetail {
  return {
    id: dto.id,
    name: dto.name ?? '',
    content: dto.content as Block[],
    createdAt: dto.createdAt.toString(),
  };
}

export function RevisionDetailContainer({ selectedRevisionId }: { selectedRevisionId: string }) {
  const { data, isError, isPending, error } = useQuery({
    ...getRevisionWithContentOptions({
      path: {
        revisionId: selectedRevisionId,
      },
    }),
    select: mapDtoToRevisionDetail,
  });
  if (isError) {
    throw error;
  }

  return isPending ? (
    <Spinner />
  ) : (
    <ScrollArea className="flex-1">
      <div className="mx-auto max-w-3xl py-8">
        <ReadOnlyEditor key={selectedRevisionId} initialContent={data?.content ?? []} />
      </div>
    </ScrollArea>
  );
}
