import { createReactInlineContentSpec } from '@blocknote/react';
import { ReferenceConfig, ReferenceInlineContentSpec } from '@notopia-uit/lib/block-note';
import { useState } from 'react';
import { useEffect } from 'react';

const ReferenceLink = ({ noteId }: { noteId: string }) => {
  const [noteName, setNoteName] = useState('Loading...');

  // TODO: Inject whatever context provider, to get the client to render here

  useEffect(() => {
    let isMounted = true;

    const fetchNote = async () => {
      try {
        const name = await Promise.resolve(''); // TODO: inject client heyapi fetch here
        if (isMounted) setNoteName(name || 'Untitled Note');
      } catch {
        if (isMounted) setNoteName('Unknown Note');
      }
    };

    void fetchNote();
    return () => {
      isMounted = false;
    };
  }, [noteId]);

  return (
    <a
      // TODO: tailwind shadcn
      href={`/note/${noteId}`}
      className="notopia-reference cursor-pointer rounded-sm bg-blue-100 px-1 text-blue-700"
      data-notopia-ref={noteId}
    >
      @{noteName}
    </a>
  );
};

export const createBlockNoteReferenceSpec = (): ReferenceInlineContentSpec =>
  createReactInlineContentSpec(ReferenceConfig, {
    render: (props) => {
      return <ReferenceLink noteId={props.inlineContent.props.noteId} />;
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
