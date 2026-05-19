import { createInlineContentSpec } from '@blocknote/core';
import { ReferenceConfig, ReferenceInlineContentSpec } from '@notopia-uit/lib/block-note';

export const createReferenceSpec = (): ReferenceInlineContentSpec =>
  createInlineContentSpec(ReferenceConfig, {
    render: (inlineContent) => {
      const noteId = inlineContent.props.noteId;
      const currentDocument = typeof globalThis !== 'undefined' ? globalThis.document : undefined;
      if (!currentDocument) {
        throw new Error(
          "DOM 'document' is not available. Ensure this is wrapped inside BlockNote Server Editor's context."
        );
      }
      const a = document.createElement('a');
      a.setAttribute('href', `/note/${noteId}`);
      a.setAttribute('data-notopia-ref', noteId);
      a.textContent = `@${noteId}`;
      return {
        dom: a,
      };
    },
  });
