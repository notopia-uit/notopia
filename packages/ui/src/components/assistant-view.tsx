'use client';

import { useChat } from '@ai-sdk/react';
import { getWorkspaceTreeOptions } from '@notopia-uit/api-gen';
import type { NoteWorkspaceTreeFolder } from '@notopia-uit/api-gen';
import {
  Conversation,
  ConversationContent,
  ConversationEmptyState,
  ConversationScrollButton,
} from '@notopia-uit/ui/components/ai-elements/conversation';
import {
  Message,
  MessageContent,
  MessageResponse,
} from '@notopia-uit/ui/components/ai-elements/message';
import {
  PromptInput,
  PromptInputBody,
  PromptInputButton,
  type PromptInputMessage,
  PromptInputSubmit,
  PromptInputTextarea,
  PromptInputFooter,
  PromptInputHeader,
  PromptInputTools,
} from '@notopia-uit/ui/components/ai-elements/prompt-input';
import { Badge } from '@notopia-uit/ui/components/shadcn/badge';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@notopia-uit/ui/components/shadcn/command';
import { Popover, PopoverContent, PopoverTrigger } from '@notopia-uit/ui/components/shadcn/popover';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { TooltipProvider } from '@notopia-uit/ui/components/shadcn/tooltip';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { fetchAccessTokenClientSide } from '@notopia-uit/ui/lib/get-access-token-client-side';
import { useQuery } from '@tanstack/react-query';
import type { UIMessage } from 'ai';
import { DefaultChatTransport } from 'ai';
import { BotMessageSquare, Check, FileText, StickyNote, X } from 'lucide-react';
import { nanoid } from 'nanoid';
import { useMemo, useState } from 'react';

interface AttachedNote {
  noteId: string;
  name: string;
  folderPath: string;
}

interface FlatNote extends AttachedNote {
  key: string;
}

function flattenNotes(folder: NoteWorkspaceTreeFolder, parentPath: string): FlatNote[] {
  const path = parentPath ? `${parentPath} / ${folder.name}` : folder.name;
  const notes: FlatNote[] = folder.notes.map((note) => ({
    key: note.id,
    noteId: note.id,
    name: note.name,
    folderPath: path,
  }));
  for (const child of folder.children) {
    notes.push(...flattenNotes(child, path));
  }
  return notes;
}

function AttachedNoteChip({ note, onRemove }: { note: AttachedNote; onRemove?: () => void }) {
  return (
    <Badge variant="secondary" className="max-w-55 gap-1 truncate">
      <StickyNote className="size-3 shrink-0" />
      <span className="truncate">{note.name}</span>
      {onRemove && (
        <button
          type="button"
          onClick={onRemove}
          aria-label={`Remove ${note.name}`}
          className="hover:text-foreground ml-1 shrink-0 cursor-pointer"
        >
          <X className="size-3" />
        </button>
      )}
    </Badge>
  );
}

function NotePicker({
  workspaceId,
  attached,
  onToggle,
}: {
  workspaceId: string;
  attached: AttachedNote[];
  onToggle: (note: AttachedNote) => void;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const { retry } = useQueryErrorHandler();
  const { data, isPending, isError, error } = useQuery({
    ...getWorkspaceTreeOptions({ path: { workspaceId } }),
  });

  const notes = data ? flattenNotes(data, '') : [];
  const query = search.trim().toLowerCase();
  const filtered = query
    ? notes.filter(
        (note) =>
          note.name.toLowerCase().includes(query) || note.folderPath.toLowerCase().includes(query)
      )
    : notes;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <PromptInputButton tooltip="Attach a workspace note" aria-label="Attach a workspace note">
          <StickyNote className="size-4" />
        </PromptInputButton>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="start">
        <Command>
          <CommandInput placeholder="Search notes…" value={search} onValueChange={setSearch} />
          <CommandList>
            {isPending ? (
              <div className="flex items-center justify-center py-6">
                <Spinner />
              </div>
            ) : isError ? (
              <div className="p-2">
                <QueryErrorFallback
                  error={error}
                  onRetry={retry}
                  title="Failed to load notes"
                  compact
                />
              </div>
            ) : (
              <>
                <CommandEmpty>No notes found.</CommandEmpty>
                <CommandGroup>
                  {filtered.map((note) => {
                    const isAttached = attached.some((a) => a.noteId === note.noteId);
                    return (
                      <CommandItem
                        key={note.key}
                        value={`${note.name} ${note.folderPath}`}
                        onSelect={() => onToggle(note)}
                      >
                        <FileText className="size-4 shrink-0" />
                        <span className="min-w-0 flex-1 truncate">{note.name}</span>
                        <span className="text-muted-foreground max-w-30 truncate text-xs">
                          {note.folderPath}
                        </span>
                        {isAttached && <Check className="size-4 shrink-0" />}
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

export function AssistantView({
  workspaceId,
  aiApiUrl,
}: {
  workspaceId: string;
  aiApiUrl: string;
}) {
  const transport = useMemo(
    () =>
      new DefaultChatTransport({
        api: aiApiUrl,
        headers: async () => ({
          Authorization: `Bearer ${await fetchAccessTokenClientSide()}`,
        }),
      }),
    [aiApiUrl]
  );
  const { messages, sendMessage, status, error } = useChat({ transport });
  const isTyping = status === 'streaming' || status === 'submitted';
  const [input, setInput] = useState('');
  const [attachedNotes, setAttachedNotes] = useState<AttachedNote[]>([]);

  const toggleNote = (note: AttachedNote) => {
    setAttachedNotes((prev) =>
      prev.some((a) => a.noteId === note.noteId)
        ? prev.filter((a) => a.noteId !== note.noteId)
        : [...prev, note]
    );
  };

  const handleSubmit = (message: PromptInputMessage) => {
    const text = message.text.trim();
    if ((!text && attachedNotes.length === 0) || isTyping) {
      return;
    }
    const context =
      attachedNotes.length > 0
        ? `\n\nAttached notes:\n${attachedNotes.map((note) => `- ${note.name} (id: ${note.noteId}, folder: ${note.folderPath})`).join('\n')}`
        : '';
    const userMessage: UIMessage = {
      id: nanoid(),
      role: 'user',
      parts: [
        ...attachedNotes.map((note) => ({
          type: 'data-attached-note' as const,
          data: { noteId: note.noteId, name: note.name },
        })),
        ...(text || context ? [{ type: 'text' as const, text: `${text}${context}` }] : []),
      ],
    };
    void sendMessage(userMessage);
    setAttachedNotes([]);
    setInput('');
  };

  return (
    <TooltipProvider>
      <div className="flex h-full min-h-0 flex-col">
        <Conversation className="min-h-0 flex-1">
          <ConversationContent>
            {messages.length === 0 ? (
              <ConversationEmptyState
                icon={<BotMessageSquare className="size-12" />}
                title="Ask your workspace"
                description="Attach workspace notes, then ask."
              />
            ) : (
              messages.map((message) => (
                <Message from={message.role} key={message.id}>
                  <MessageContent>
                    {message.parts.map((part, index) => {
                      if (part.type === 'text') {
                        return (
                          <MessageResponse key={`${message.id}-${index}`}>
                            {part.text}
                          </MessageResponse>
                        );
                      }
                      if (part.type === 'data-attached-note') {
                        const data = part.data as { noteId: string; name: string };
                        return (
                          <AttachedNoteChip
                            key={`${message.id}-${index}`}
                            note={{ noteId: data.noteId, name: data.name, folderPath: '' }}
                          />
                        );
                      }
                      return null;
                    })}
                  </MessageContent>
                </Message>
              ))
            )}
          </ConversationContent>
          <ConversationScrollButton />
        </Conversation>

        <div className="mx-auto w-full max-w-3xl p-4">
          <PromptInput onSubmit={handleSubmit}>
            {error && (
              <p className="text-destructive px-3 pt-2 text-sm">Request failed: {error.message}</p>
            )}
            <PromptInputHeader>
              {attachedNotes.length > 0 && (
                <div className="flex flex-wrap gap-1.5 px-3 pt-2">
                  {attachedNotes.map((note) => (
                    <AttachedNoteChip
                      key={note.noteId}
                      note={note}
                      onRemove={() => toggleNote(note)}
                    />
                  ))}
                </div>
              )}
            </PromptInputHeader>
            <PromptInputBody>
              <PromptInputTextarea
                value={input}
                onChange={(event) => setInput(event.currentTarget.value)}
                placeholder="Ask about your notes…"
              />
            </PromptInputBody>
            <PromptInputFooter>
              <PromptInputTools>
                <NotePicker
                  workspaceId={workspaceId}
                  attached={attachedNotes}
                  onToggle={toggleNote}
                />
              </PromptInputTools>
              <PromptInputSubmit
                status={isTyping ? 'streaming' : 'ready'}
                disabled={(!input.trim() && attachedNotes.length === 0) || isTyping}
              />
            </PromptInputFooter>
          </PromptInput>
          <p className="text-muted-foreground mt-2 text-center text-xs">
            Answers stream from the AI service; attached notes are sent as context.
          </p>
        </div>
      </div>
    </TooltipProvider>
  );
}
