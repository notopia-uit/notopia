'use client';
import { Block } from '@blocknote/core';
import { useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import {
  commitDocumentMutation,
  DocumentRevisionWithContent,
  getRevisionWithContentOptions,
  useCommitDocumentMutation,
} from '@notopia-uit/api-gen';
import { ErrorAlert } from '@notopia-uit/ui/components/error-alert';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { SuccessAlert } from '@notopia-uit/ui/components/success-alert';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQuery } from '@tanstack/react-query';
import { RotateCcw } from 'lucide-react';

import { Button } from './shadcn/button';
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

export function RevisionDetailContainer({
  selectedRevisionId,
  documentId,
}: {
  selectedRevisionId: string;
  documentId: string;
}) {
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
  const { alert, showAlert } = useAlert();
  const {
    mutate: restoreVersion,
    isError: isRestoreError,
    error: restoreError,
    isPending: isRestoring,
  } = useCommitDocumentMutation({
    onSuccess: (responses, _) => {
      showAlert(
        'success',
        'Version restored successfully!',
        `The version ${data?.name} of ${responses.id} was restored.`
      );
    },
    onError: (error) => {
      showAlert(
        'error',
        'Failed to restore version',
        `An error occurred while restoring the version. ${
          error instanceof Error ? error.message : 'Please try again.'
        }`
      );
    },
  });
  if (isRestoreError) {
    throw restoreError;
  }
  return isPending ? (
    <Spinner />
  ) : (
    <ScrollArea className="flex-1">
      <div className="mx-auto max-w-3xl">
        <div className="flex items-center justify-between border-b px-8 py-4">
          <h3 className="text-lg font-semibold">{data?.name}</h3>
          <Button
            onClick={() =>
              restoreVersion({
                path: {
                  documentId: documentId,
                },
              })
            }
            variant="outline"
            size="sm"
            className="gap-2"
          >
            {isRestoring ? <Spinner /> : <RotateCcw className="size-4" />}
            Restore this version
          </Button>
        </div>
        <div className="py-8">
          <ReadOnlyEditor key={selectedRevisionId} initialContent={data?.content ?? []} />
        </div>

        {alert?.type === 'success' && <SuccessAlert title={alert.title} message={alert.message} />}
        {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}
      </div>
    </ScrollArea>
  );
}
