import { createInlineContentSpec } from '@blocknote/core';
import { TagConfig, TagInlineContentSpec } from '@notopia-uit/lib/block-note';

export const createTagSpec = (): TagInlineContentSpec =>
  createInlineContentSpec(TagConfig, {
    render: (inlineContent) => {
      const tag = inlineContent.props.tag;
      const a = document.createElement('a');
      a.setAttribute('href', `#${tag}`);
      a.setAttribute('data-notopia-tag', tag);
      a.textContent = `#${tag}`;
      return {
        dom: a,
      };
    },
  });
