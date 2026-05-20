'use client';

import { Block, BlockNoteEditor } from '@blocknote/core';
import { useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import {
  DocumentRevisionWithContent,
  getRevisionWithContentOptions,
  getRevisionsOptions,
} from '@notopia-uit/api-gen';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQuery } from '@tanstack/react-query';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { formatDistanceToNow } from 'date-fns';
import { History, RotateCcw, Search } from 'lucide-react';
import { useState, useMemo } from 'react';

import { ErrorAlert } from './error-alert';
import { Button } from './shadcn/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './shadcn/dialog';
import { Input } from './shadcn/input';
import { ScrollArea } from './shadcn/scroll-area';
import { Spinner } from './shadcn/spinner';
import { SuccessAlert } from './success-alert';

interface Revision {
  id: string;
  name: string;
  createdAt: string;
}

function ReadOnlyRevisionEditor({ initialContent }: { initialContent: Block[] }) {
  const editor = useCreateBlockNote({
    initialContent: initialContent,
  });

  return <BlockNoteView editor={editor} editable={false} />;
}

interface RevisionModalProps {
  noteId: string;
  currentEditor: BlockNoteEditor | null;
}

export function RevisionModal({ noteId, currentEditor }: RevisionModalProps) {
  const [open, setOpen] = useState(false);
  const [selectedRevisionId, setSelectedRevisionId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const { alert, showAlert } = useAlert();
  const { retry } = useQueryErrorHandler();

  const {
    data: revisions,
    isPending: isLoadingRevisions,
    isError: isRevisionsError,
    error: revisionsError,
  } = useQuery({
    ...getRevisionsOptions({
      query: {
        documentId: noteId,
      },
    }),
    select: (response) =>
      response.data.map((revision) => {
        const createdAt = revision.createdAt.toString();
        const name = revision.name ?? createdAt;
        return {
          id: revision.id,
          name: name,
          createdAt: createdAt,
        } satisfies Revision;
      }),
  });

  const {
    data: selectedRevisionData,
    isPending: isLoadingRevision,
    isError: isRevisionError,
    error: revisionError,
  } = useQuery({
    ...getRevisionWithContentOptions({
      path: {
        revisionId: selectedRevisionId ?? '',
      },
    }),
    enabled: !!selectedRevisionId,
    select: (dto: DocumentRevisionWithContent) => ({
      id: dto.id,
      name: dto.name ?? '',
      content: dto.content as Block[],
      createdAt: dto.createdAt.toString(),
    }),
  });

  const filteredRevisions = useMemo(() => {
    if (!revisions) return [];
    return revisions.filter(
      (rev) =>
        rev.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        rev.id.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [revisions, searchQuery]);

  const handleApplyRevision = () => {
    if (!currentEditor || !selectedRevisionData) return;

    try {
      const blocks = selectedRevisionData.content;
      currentEditor.replaceBlocks(
        currentEditor.document.map((block) => block.id),
        blocks
      );

      showAlert(
        'success',
        'Revision Applied',
        `Successfully applied revision "${selectedRevisionData.name}" to your current note.`
      );
      setOpen(false);
      setSelectedRevisionId(null);
      setSearchQuery('');
    } catch (error) {
      showAlert(
        'error',
        'Failed to apply revision',
        `An error occurred while applying the revision. ${
          error instanceof Error ? error.message : 'Please try again.'
        }`
      );
    }
  };

  if (isRevisionsError) {
    return (
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button variant="outline" size="sm" className="gap-2" title="View version history">
            <History className="size-4" />
            History
          </Button>
        </DialogTrigger>
        <DialogContent className="flex h-[90vh] max-w-6xl flex-col p-0">
          <DialogHeader className="border-b px-6 py-4">
            <DialogTitle>Version History</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-auto p-6">
            <QueryErrorFallback
              error={revisionsError}
              onRetry={retry}
              title="Failed to Load Revisions"
              description="Unable to load version history. Please try again."
            />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (isRevisionError) {
    return (
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button variant="outline" size="sm" className="gap-2" title="View version history">
            <History className="size-4" />
            History
          </Button>
        </DialogTrigger>
        <DialogContent className="flex h-[90vh] max-w-6xl flex-col p-0">
          <DialogHeader className="border-b px-6 py-4">
            <DialogTitle>Version History</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-auto p-6">
            <QueryErrorFallback
              error={revisionError}
              onRetry={retry}
              title="Failed to Load Revision"
              description="Unable to load the selected revision. Please try again."
            />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2" title="View version history">
          <History className="size-4" />
          History
        </Button>
      </DialogTrigger>

      <DialogContent className="flex h-[90vh] max-w-6xl flex-col p-0">
        <DialogHeader className="border-b px-6 py-4">
          <DialogTitle>Version History</DialogTitle>
        </DialogHeader>

        <div className="flex flex-1 overflow-hidden">
          <div className="flex w-72 flex-none flex-col overflow-y-auto border-r">
            <div className="border-b p-4">
              <div className="relative">
                <Search className="text-muted-foreground absolute top-1/2 left-2 size-4 -translate-y-1/2" />
                <Input
                  placeholder="Search versions..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>

            <ScrollArea className="flex-1">
              {isLoadingRevisions ? (
                <div className="flex items-center justify-center py-8">
                  <Spinner />
                </div>
              ) : (
                <div className="flex flex-col gap-1 p-2">
                  {filteredRevisions.map((revision) => {
                    const isActive = selectedRevisionId === revision.id;
                    return (
                      <button
                        key={revision.id}
                        onClick={() => setSelectedRevisionId(revision.id)}
                        className={`hover:bg-accent flex flex-col items-start rounded-md p-3 text-left text-sm transition-colors ${
                          isActive
                            ? 'bg-accent border-l-primary border-l-4'
                            : 'border-l-4 border-l-transparent'
                        }`}
                      >
                        <div className="mb-1 flex w-full items-center justify-between">
                          <span className="font-medium">
                            {formatDistanceToNow(new Date(revision.createdAt), { addSuffix: true })}
                          </span>
                        </div>
                        <span className="text-muted-foreground line-clamp-1">{revision.name}</span>
                      </button>
                    );
                  })}
                </div>
              )}
            </ScrollArea>
          </div>

          <div className="flex flex-1 flex-col">
            {!selectedRevisionId ? (
              <div className="flex flex-1 items-center justify-center text-center">
                <div>
                  <p className="text-muted-foreground">Select a version to view</p>
                </div>
              </div>
            ) : isLoadingRevision ? (
              <div className="flex flex-1 items-center justify-center">
                <Spinner />
              </div>
            ) : (
              <>
                <div className="border-b px-6 py-4">
                  <h3 className="text-lg font-semibold">{selectedRevisionData?.name}</h3>
                </div>
                <ScrollArea className="flex-1">
                  <div className="p-4">
                    {selectedRevisionData && (
                      <ReadOnlyRevisionEditor initialContent={selectedRevisionData.content} />
                    )}
                  </div>
                </ScrollArea>
                <div className="bg-muted/50 border-t px-6 py-4">
                  <Button
                    onClick={handleApplyRevision}
                    variant="default"
                    size="sm"
                    className="gap-2"
                    disabled={!currentEditor}
                  >
                    <RotateCcw className="size-4" />
                    Apply this version
                  </Button>
                </div>
              </>
            )}
          </div>
        </div>

        {alert?.type === 'success' && <SuccessAlert title={alert.title} message={alert.message} />}
        {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}
      </DialogContent>
    </Dialog>
  );
}
