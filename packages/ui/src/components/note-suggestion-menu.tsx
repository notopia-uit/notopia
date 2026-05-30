'use client';

import { type SuggestionMenuProps } from '@blocknote/react';
import { type NoteSuggestionItem } from '@notopia-uit/ui/block-note';
import { HighlightedText } from '@notopia-uit/ui/components/highlighted-text';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { FileText } from 'lucide-react';

export function NoteSuggestionMenu({
  items,
  selectedIndex,
  onItemClick,
}: SuggestionMenuProps<NoteSuggestionItem>) {
  return (
    <div className="bg-popover text-popover-foreground z-50 flex max-h-72 w-72 flex-col gap-0.5 overflow-y-auto rounded-xl border p-1 shadow-md">
      {items.map((item, index) => (
        <div
          key={index}
          role="option"
          aria-selected={index === selectedIndex}
          onMouseDown={(e) => {
            e.preventDefault();
            onItemClick?.(item);
          }}
          className={cn(
            'flex cursor-default items-start gap-2 rounded-lg px-2 py-1.5 text-sm',
            index === selectedIndex && 'bg-muted text-foreground'
          )}
        >
          <FileText className="mt-0.5 size-4 shrink-0" />
          <div className="flex min-w-0 flex-col">
            <HighlightedText text={item.formattedName ?? item.title} className="truncate" />
            {item.contentSnippet ? (
              <HighlightedText
                text={item.contentSnippet}
                className="text-muted-foreground line-clamp-2 text-xs"
              />
            ) : (
              item.subtext && (
                <span className="text-muted-foreground truncate text-xs">{item.subtext}</span>
              )
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
