import { createInlineContentSpec } from '@blocknote/core';
import { ReferenceConfig, ReferenceInlineContentSpec } from '@notopia-uit/lib/block-note';
import { JSDOM } from 'jsdom';

export const createReferenceSpec = (): ReferenceInlineContentSpec =>
  createInlineContentSpec(ReferenceConfig, {
    render: (inlineContent) => {
      const noteId = inlineContent.props.noteId;
      const dom = new JSDOM().window.document;
      const a = dom.createElement('a');
      a.setAttribute('href', `/note/${noteId}`);
      a.setAttribute('data-notopia-ref', noteId);
      a.textContent = `@${noteId}`;
      return {
        dom: a,
      };
    },
  });
