import {
  CustomInlineContentConfig,
  type InlineContentSpec,
} from '@blocknote/core';
import { createReactInlineContentSpec } from '@blocknote/react';
import { useEffect, useState } from 'react';

export const BlockNoteReferenceConfig = {
  type: 'reference',
  propSchema: {
    noteId: { default: 'unknown' },
  },
  content: 'none',
} as const satisfies CustomInlineContentConfig;

export type BlockNoteReferenceInlineContentSpec = InlineContentSpec<
  typeof BlockNoteReferenceConfig
>;

export type getNoteNameFn = (noteId: string) => Promise<string>;

const ReferenceLink = ({
  noteId,
  getNoteName,
}: {
  noteId: string;
  getNoteName: getNoteNameFn;
}) => {
  const [noteName, setNoteName] = useState('Loading...');

  useEffect(() => {
    let isMounted = true;

    const fetchNote = async () => {
      try {
        const name = await getNoteName(noteId);
        if (isMounted) setNoteName(name || 'Untitled Note');
      } catch {
        if (isMounted) setNoteName('Unknown Note');
      }
    };

    fetchNote();
    return () => {
      isMounted = false;
    };
  }, [noteId, getNoteName]);

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

export const createBlockNoteReferenceSpec = ({
  getNoteName,
}: {
  getNoteName: getNoteNameFn;
}): BlockNoteReferenceInlineContentSpec =>
  createReactInlineContentSpec(BlockNoteReferenceConfig, {
    render: (props) => {
      return (
        <ReferenceLink
          noteId={props.inlineContent.props.noteId}
          getNoteName={getNoteName}
        />
      );
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
      if (element.hasAttribute('data-notopia-ref')) {
        return {
          noteId: element.getAttribute('data-notopia-ref')!,
        };
      }
      return undefined;
    },
  });
