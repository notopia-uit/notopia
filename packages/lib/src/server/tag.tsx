import { createReactInlineContentSpec } from '@blocknote/react';

import { BlockNoteTagConfig } from '../block-note';

export const createServerBlockNoteTagSpec = () =>
  createReactInlineContentSpec(BlockNoteTagConfig, {
    render: (props) => (
      <span
        className="notopia-tag"
        data-notopia-tag={props.inlineContent.props.tag}
      >
        #{props.inlineContent.props.tag}
      </span>
    ),

    toExternalHTML: (props) => {
      const tag = props.inlineContent.props.tag;
      return (
        <a href={`#${tag}`} data-notopia-tag={tag}>
          #{tag}
        </a>
      );
    },

    parse: (element) => {
      const tag = element.getAttribute('data-notopia-tag');
      if (tag) {
        return { tag };
      }
      return undefined;
    },
  });
