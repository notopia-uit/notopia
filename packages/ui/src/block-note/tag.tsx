'use client';

import { createReactInlineContentSpec } from '@blocknote/react';
import { TagConfig } from '@notopia-uit/lib/block-note';
import { searchNotesByTag, type SearchResult } from '@notopia-uit/ui/block-note';
import { Dialog, DialogContent } from '@notopia-uit/ui/components/shadcn/dialog';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useMeilisearch } from '@notopia-uit/ui/contexts/meilisearch-context';
import { FileText } from 'lucide-react';
import { useRouter, useParams } from 'next/navigation';
import { useCallback, useEffect, useState } from 'react';

function TagPreview({
  tag,
  open,
  onOpenChange,
}: {
  tag: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [notes, setNotes] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const meilisearchClient = useMeilisearch();
  const router = useRouter();
  const params = useParams();
  const workspaceId = (params?.workspaceId as string) || '';

  useEffect(() => {
    if (!open || !meilisearchClient) return;
    setIsLoading(true);
    searchNotesByTag(meilisearchClient, tag)
      .then(setNotes)
      .catch(console.error)
      .finally(() => setIsLoading(false));
  }, [open, tag, meilisearchClient]);

  const handleSelect = (noteId: string) => {
    onOpenChange(false);
    router.push(`/workspace/${workspaceId}/note/${noteId}`);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg p-0" showCloseButton={true}>
        <div className="px-4 py-3 text-sm font-medium">Notes tagged #{tag}</div>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Spinner />
          </div>
        ) : notes.length === 0 ? (
          <div className="text-muted-foreground px-4 py-8 text-center text-sm">
            No notes found with this tag
          </div>
        ) : (
          <div className="max-h-72 overflow-y-auto px-2 pb-2">
            {notes.map((note) => (
              <button
                key={note.id}
                onClick={() => handleSelect(note.id)}
                className="hover:bg-muted flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-sm"
              >
                <FileText className="size-4 shrink-0" />
                <span className="truncate">{note.name}</span>
              </button>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

export const createBlockNoteTagSpec = () =>
  createReactInlineContentSpec(TagConfig, {
    render: (props) => {
      const tag = props.inlineContent.props.tag;
      const [showPreview, setShowPreview] = useState(false);

      return (
        <>
          <span
            className="notopia-tag cursor-pointer rounded-sm px-1 font-semibold"
            data-notopia-tag={tag}
            onMouseEnter={() => setShowPreview(true)}
          >
            #{tag}
          </span>
          <TagPreview tag={tag} open={showPreview} onOpenChange={setShowPreview} />
        </>
      );
    },

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
