import { createInlineContentSpec } from '@blocknote/core';
import { TagConfig, TagInlineContentSpec } from '@notopia-uit/lib/block-note';

export const createTagSpec = (): TagInlineContentSpec =>
  createInlineContentSpec(TagConfig, {
    render: (inlineContent) => {
      const tag = inlineContent.props.tag;
      const currentDocument = typeof globalThis !== 'undefined' ? globalThis.document : undefined;
      if (!currentDocument) {
        throw new Error(
          "DOM 'document' is not available. Ensure this is wrapped inside BlockNote Server Editor's context."
        );
      }
      const a = document.createElement('a');
      a.setAttribute('href', `#${tag}`);
      a.setAttribute('data-notopia-tag', tag);
      a.textContent = `#${tag}`;
      return {
        dom: a,
      };
    },
  });
