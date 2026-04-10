import {
  CustomInlineContentConfig,
  type InlineContentSpec,
} from '@blocknote/core';
import { createReactInlineContentSpec } from '@blocknote/react';
import { useState } from 'react';

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

const ReferenceLink = ({ noteId }: { noteId: string }) => {
  const [noteName, _] = useState('Loading...');

  // TODO: Inject whatever context provider, to get the client to render here

  // useEffect(() => {
  //   let isMounted = true;
  //
  //   const fetchNote = async () => {
  //     try {
  //       const name = await getNoteName(noteId);
  //       if (isMounted) setNoteName(name || 'Untitled Note');
  //     } catch {
  //       if (isMounted) setNoteName('Unknown Note');
  //     }
  //   };
  //
  //   void fetchNote();
  //   return () => {
  //     isMounted = false;
  //   };
  // }, [noteId, getNoteName]);

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

export const createBlockNoteReferenceSpec = (
  type: 'client' | 'server'
): BlockNoteReferenceInlineContentSpec =>
  createReactInlineContentSpec(BlockNoteReferenceConfig, {
    render:
      type === 'client'
        ? (props) => {
            return <ReferenceLink noteId={props.inlineContent.props.noteId} />;
          }
        : (props) => {
            // Server side rendering, just render the noteId for now
            return (
              <a
                href={`@${props.inlineContent.props.noteId}`}
                data-notopia-ref={props.inlineContent.props.noteId}
              >
                @{props.inlineContent.props.noteId}
              </a>
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
