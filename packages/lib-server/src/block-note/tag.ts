import { createInlineContentSpec } from '@blocknote/core';
import { TagConfig, TagInlineContentSpec } from '@notopia-uit/lib/block-note';
import { JSDOM } from 'jsdom';

export const createTagSpec = (): TagInlineContentSpec =>
  createInlineContentSpec(TagConfig, {
    render: (inlineContent) => {
      const dom = new JSDOM().window.document;
      const tag = inlineContent.props.tag;
      const a = dom.createElement('a');
      a.setAttribute('href', `#${tag}`);
      a.setAttribute('data-notopia-tag', tag);
      a.textContent = `#${tag}`;
      return {
        dom: a,
      };
    },
  });
