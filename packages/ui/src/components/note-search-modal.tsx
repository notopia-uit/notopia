'use client';

import { searchNotesFromMeilisearch, type SearchResult } from '@notopia-uit/ui/block-note';
import { useMeilisearch } from '@notopia-uit/ui/contexts/meilisearch-context';
import { useSearchCache } from '@notopia-uit/ui/hooks/use-search-cache';
import { FileText, SearchX } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useCallback, useEffect, useState } from 'react';

import {
  Command,
  CommandDialog,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from './shadcn/command';
import { Spinner } from './shadcn/spinner';

interface NoteSearchModalProps {
  workspaceId: string;
}

export function NoteSearchModal({ workspaceId }: NoteSearchModalProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const meilisearchClient = useMeilisearch();
  const router = useRouter();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const noteSearchFn = useCallback(
    (q: string) => {
      if (!meilisearchClient) return Promise.resolve([] as SearchResult[]);
      return searchNotesFromMeilisearch(meilisearchClient, q);
    },
    [meilisearchClient]
  );

  const { search: searchNotesWithCache, isLoading, error } = useSearchCache(noteSearchFn, 300);

  const handleQueryChange = useCallback(
    async (value: string) => {
      setQuery(value);
      if (!value) {
        setResults([]);
        return;
      }
      const result = await searchNotesWithCache(value);
      setResults(result.data || []);
    },
    [searchNotesWithCache]
  );

  const handleSelect = (noteId: string) => {
    setOpen(false);
    setQuery('');
    setResults([]);
    router.push(`/workspace/${workspaceId}/note/${noteId}`);
  };

  const handleOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      setQuery('');
      setResults([]);
    }
  };

  return (
    <CommandDialog
      open={open}
      onOpenChange={handleOpenChange}
      title="Search Notes"
      description="Search for notes in your workspace"
    >
      <Command shouldFilter={false}>
        <CommandInput placeholder="Search notes..." value={query} onValueChange={handleQueryChange} />
        <CommandList>
          {!meilisearchClient && !isLoading && (
            <div className="text-muted-foreground flex items-center justify-center gap-2 py-6 text-sm">
              <Spinner className="size-4" />
              Initializing search...
            </div>
          )}
          {error && (
            <div className="flex items-center justify-center gap-2 py-6 text-sm text-red-500">
              <SearchX className="size-4" />
              Search failed. Please try again.
            </div>
          )}
          {isLoading && (
            <div className="flex items-center justify-center py-6">
              <Spinner />
            </div>
          )}
          {!isLoading && !error && meilisearchClient && query && results.length === 0 && (
            <div className="py-6 text-center text-sm">No notes found for "{query}"</div>
          )}
          {!isLoading && !error && results.length > 0 && (
            <CommandGroup heading="Notes">
              {results.map((note) => (
                <CommandItem key={note.id} onSelect={() => handleSelect(note.id)}>
                  <FileText className="size-4" />
                  <span>{note.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          )}
        </CommandList>
      </Command>
    </CommandDialog>
  );
}
