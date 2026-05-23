'use client';

import { useGetNoteLinksQuery } from '@notopia-uit/api-gen';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { ArrowUpRight, Link2 } from 'lucide-react';
import { useRouter, useParams } from 'next/navigation';
import { useState } from 'react';

import { Button } from './shadcn/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './shadcn/dialog';
import { ScrollArea } from './shadcn/scroll-area';
import { Spinner } from './shadcn/spinner';

interface NoteLinksModalProps {
  noteId: string;
}

export function NoteLinksModal({ noteId }: NoteLinksModalProps) {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const params = useParams();
  const { retry } = useQueryErrorHandler();

  const {
    data,
    isPending,
    isError,
    error,
  } = useGetNoteLinksQuery({
    path: { noteId },
    query: { outgoingLinks: true, backlinks: true },
  });

  const outgoingLinks = data?.outgoingLinks ?? [];
  const backlinks = data?.backlinks ?? [];
  const workspaceId = params.workspaceId as string;

  const handleLinkClick = (linkId: string) => {
    setOpen(false);
    router.push(`/workspace/${workspaceId}/note/${linkId}`);
  };

  const content = () => {
    if (isPending) {
      return (
        <div className="flex items-center justify-center py-12">
          <Spinner />
        </div>
      );
    }

    if (isError) {
      return (
        <div className="p-6">
          <QueryErrorFallback
            error={error}
            onRetry={retry}
            title="Failed to Load Links"
            description="Unable to load note links. Please try again."
          />
        </div>
      );
    }

    return (
      <div className="flex flex-col gap-6 p-6">
        <section>
          <h3 className="mb-3 text-sm font-medium text-muted-foreground">
            Outgoing Links ({outgoingLinks.length})
          </h3>
          {outgoingLinks.length === 0 ? (
            <p className="text-sm text-muted-foreground">No outgoing links</p>
          ) : (
            <ul className="space-y-1">
              {outgoingLinks.map((link) => (
                <li key={link.id}>
                  <button
                    onClick={() => handleLinkClick(link.id)}
                    className="hover:bg-accent flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors"
                  >
                    <ArrowUpRight className="size-4 shrink-0" />
                    <span className="truncate">{link.name}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section>
          <h3 className="mb-3 text-sm font-medium text-muted-foreground">
            Backlinks ({backlinks.length})
          </h3>
          {backlinks.length === 0 ? (
            <p className="text-sm text-muted-foreground">No backlinks</p>
          ) : (
            <ul className="space-y-1">
              {backlinks.map((link) => (
                <li key={link.id}>
                  <button
                    onClick={() => handleLinkClick(link.id)}
                    className="hover:bg-accent flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors"
                  >
                    <ArrowUpRight className="size-4 shrink-0" />
                    <span className="truncate">{link.name}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    );
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          aria-label="View note links"
        >
          <Link2 className="size-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Note Links</DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[60vh]">
          {content()}
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
