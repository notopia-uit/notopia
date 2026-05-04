import { createReactInlineContentSpec } from '@blocknote/react';

import { BlockNoteReferenceConfig, BlockNoteReferenceInlineContentSpec } from '../block-note';

export const createServerBlockNoteReferenceSpec = (): BlockNoteReferenceInlineContentSpec =>
  createReactInlineContentSpec(BlockNoteReferenceConfig, {
    render: (props) => {
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
