import { createReactInlineContentSpec } from '@blocknote/react';

export const BlocknoteReference = createReactInlineContentSpec(
  {
    type: 'reference',
    propSchema: {
      note: {
        default: 'Unknown',
      },
    },
    content: 'none',
  }
  // {
  //   render: (props) => {},
  // }
);
