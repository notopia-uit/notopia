'use client';

import { useHocuspocusProvider } from '@hocuspocus/provider-react';
import { useCommitDocumentMutation } from '@notopia-uit/api-gen';
import { useIsDocModified } from '@notopia-uit/ui/hooks/use-is-doc-modified';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';

export const useEditorState = (noteId: string) => {
  const provider = useHocuspocusProvider();
  const { showAlert } = useAlert();

  const { isModified, setModified } = useIsDocModified(
    provider.document,
    provider.awareness?.clientID.toString() ?? 'anonymous'
  );

  const { mutate: commitDocument, isPending: isCommitingDocument } = useCommitDocumentMutation({
    onSuccess: (responses) => {
      setModified(false);
      showAlert({
        type: 'success',
        title: 'Document committed successfully!',
        message: `Your note ${responses.id} changes were committed.`,
      });
    },
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Failed to commit document',
        message: `An error occurred while committing your note changes. ${
          error instanceof Error ? error.message : 'Please try again.'
        }`,
      });
    },
  });

  const handleSave = () => {
    commitDocument({
      path: {
        documentId: noteId,
      },
    });
  };

  return {
    isModified,
    isCommitingDocument,
    handleSave,
  };
};
