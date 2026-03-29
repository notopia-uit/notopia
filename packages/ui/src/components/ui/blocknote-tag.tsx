import { createReactInlineContentSpec } from '@blocknote/react';

export const BlocknoteTagConfig = {
  type: 'tag',
  propSchema: {
    tag: { default: '' },
  },
  content: 'none',
} as const;

export const blocknoteTagSpec = createReactInlineContentSpec(
  BlocknoteTagConfig,
  {
    render: (props) => (
      <span
        // TODO: tailwind shadcn
        className="notopia-tag bg-gray-200 text-gray-800 rounded px-1 font-semibold"
        data-notopia-tag={props.inlineContent.props.tag}
      >
        #{props.inlineContent.props.tag}
      </span>
    ),

    toExternalHTML: (props) => (
      <span data-notopia-tag={props.inlineContent.props.tag}>
        #{props.inlineContent.props.tag}
      </span>
    ),

    parse: (element) => {
      if (element.hasAttribute('data-notopia-tag')) {
        return { tag: element.getAttribute('data-notopia-tag') || '' };
      }
      return undefined;
    },
  }
);
