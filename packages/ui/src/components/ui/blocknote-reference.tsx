import { createReactInlineContentSpec } from '@blocknote/react';
import { getNote } from '@notopia-uit/api-gen';
import type { Client } from '@notopia-uit/api-gen/client';
import { useEffect, useState } from 'react';

export const BlocknoteReferenceConfig = {
  type: 'reference',
  propSchema: {
    noteId: { default: 'unknown' },
  },
  content: 'none',
} as const;

const ReferenceLink = ({
  noteId,
  apiClient,
}: {
  noteId: string;
  apiClient: Client;
}) => {
  const [noteName, setNoteName] = useState('Loading...');

  useEffect(() => {
    let isMounted = true;

    const fetchNote = async () => {
      try {
        const res = await getNote({
          client: apiClient,
          path: { noteId },
        });
        if (isMounted) setNoteName(res.data?.name || 'Untitled Note');
      } catch {
        if (isMounted) setNoteName('Unknown Note');
      }
    };

    fetchNote();
    return () => {
      isMounted = false;
    };
  }, [noteId, apiClient]);

  return (
    <a
      // TODO: tailwind shadcn
      href={`/note/${noteId}`}
      className="notopia-reference bg-blue-100 text-blue-700 rounded px-1 cursor-pointer"
      data-notopia-ref={noteId}
    >
      @{noteName}
    </a>
  );
};

export const CreateBlocknoteReferenceSpec = (apiClient?: Client) =>
  createReactInlineContentSpec(BlocknoteReferenceConfig, {
    render: (props) => {
      if (!apiClient) {
        return;
      }
      return (
        <ReferenceLink
          noteId={props.inlineContent.props.noteId}
          apiClient={apiClient}
        />
      );
    },

    toExternalHTML: (props) => (
      <a
        // TODO: endpoint from backend here
        href={`/note/${props.inlineContent.props.noteId}`}
        data-notopia-ref={props.inlineContent.props.noteId}
      >
        @{props.inlineContent.props.noteId}
      </a>
    ),

    parse: (element) => {
      if (element.hasAttribute('data-notopia-ref')) {
        return {
          noteId: element.getAttribute('data-notopia-ref') || 'unknown',
        };
      }
      return undefined;
    },
  });
