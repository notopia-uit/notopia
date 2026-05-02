import { createReactInlineContentSpec } from '@blocknote/react';
import { BlockNoteTagConfig } from '@notopia-uit/lib/block-note';

export const createBlockNoteTagSpec = () =>
  createReactInlineContentSpec(BlockNoteTagConfig, {
    render: (props) => (
      <span
        // TODO: tailwind shadcn
        className="notopia-tag rounded-sm bg-gray-200 px-1 font-semibold text-gray-800"
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
