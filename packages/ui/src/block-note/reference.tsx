'use client';

import { createReactInlineContentSpec } from '@blocknote/react';
import { ReferenceConfig, ReferenceInlineContentSpec } from '@notopia-uit/lib/block-note';
import { useGetNoteQuery } from '@notopia-uit/api-gen';

const ReferenceLink = ({ noteId }: { noteId: string }) => {
  const { data: note, isPending, isError } = useGetNoteQuery({
    path: { noteId },
  });

  const displayName = isPending ? '' : isError ? 'Unknown Note' : note?.name || 'Untitled Note';

  return (
    <a
      href={`/note/${noteId}`}
      className="notopia-reference cursor-pointer rounded-sm bg-primary/10 px-1 text-primary hover:bg-primary/20"
      data-notopia-ref={noteId}
    >
      @{displayName}
    </a>
  );
};

export const createBlockNoteReferenceSpec = (): ReferenceInlineContentSpec =>
  createReactInlineContentSpec(ReferenceConfig, {
    render: (props) => {
      return <ReferenceLink noteId={props.inlineContent.props.noteId} />;
    },
    toExternalHTML: (props) => {
      const id = props.inlineContent.props.noteId;
      return (
        <a href={`@${id}`} data-notopia-ref={id}>
          @{props.inlineContent.props.noteId}
        </a>
      );
    },

    parse: (element) => {
      const noteId = element.getAttribute('data-notopia-ref');
      if (!noteId) {
        return undefined;
      } else {
        return {
          noteId,
        };
      }
    },
  });
