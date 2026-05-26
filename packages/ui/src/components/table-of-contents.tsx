'use client';

import { type BlockNoteEditor } from '@blocknote/core';
import { useEffect, useState } from 'react';

import { cn } from './../lib/shadcn/utils';
import { Popover, PopoverContent, PopoverTrigger } from './shadcn/popover';

type HeadingItem = {
  id: string;
  text: string;
  level: 1 | 2 | 3;
};

interface TableOfContentsProps {
  editor: BlockNoteEditor | null;
}

export function TableOfContents({ editor }: TableOfContentsProps) {
  const [headings, setHeadings] = useState<HeadingItem[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const [canScroll, setCanScroll] = useState(false);
  const [scrollProgress, setScrollProgress] = useState(0);
  const [tocTop, setTocTop] = useState(65);
  const SEGMENTS = 10;
  const filledSegments = Math.round(scrollProgress * SEGMENTS);

  useEffect(() => {
    const container = document.querySelector('main');
    if (!container) return;

    const handleScroll = () => {
      const scrollTop = container.scrollTop;
      const docHeight = container.scrollHeight - container.clientHeight;
      setCanScroll(docHeight > 0);
      const progress = docHeight > 0 ? scrollTop / docHeight : 0;
      setScrollProgress(progress);
      const top = Math.max(50, 65 - progress * 50);
      setTocTop(top);
    };

    container.addEventListener('scroll', handleScroll, { passive: true });
    const resizeObserver = new ResizeObserver(handleScroll);
    resizeObserver.observe(container);

    return () => {
      container.removeEventListener('scroll', handleScroll);
      resizeObserver.disconnect();
    };
  }, []);

  useEffect(() => {
    if (!editor) return;

    const extractHeadings = () => {
      const items: HeadingItem[] = [];
      editor.forEachBlock((block) => {
        if (
          block.type === 'heading' &&
          (block.props.level === 1 ||
            block.props.level === 2 ||
            block.props.level === 3)
        ) {
          const text = block.content
            .filter((c) => c.type === 'text')
            .map((c) => c.text)
            .join('');
          if (text.trim()) {
            items.push({
              id: block.id,
              text,
              level: block.props.level as 1 | 2 | 3,
            });
          }
        }
        return true;
      });
      setHeadings(items);
    };

    extractHeadings();
    const unsubscribe = editor.onChange(extractHeadings);
    return () => unsubscribe();
  }, [editor]);

  const handleClick = (id: string) => {
    setActiveId(id);
    const el = document.querySelector(`[data-id="${id}"]`);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  };

  if (!headings.length || !canScroll) return null;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        onMouseEnter={() => setOpen(true)}
        onMouseLeave={() => setOpen(false)}
        className="fixed left-0 top-[65px] z-50 flex flex-col items-center gap-[3px] p-2  transition-all duration-300"
        style={{ top: `${tocTop}px` }}
      >
        {Array.from({ length: SEGMENTS }).map((_, i) => (
          <div
            key={i}
            className={cn(
              'size-[3px]  rounded-full transition-all duration-300',
              i < filledSegments ? 'bg-foreground/40' : 'bg-foreground/10'
            )}
          />
        ))}
      </PopoverTrigger>
      <PopoverContent
        onMouseEnter={() => setOpen(true)}
        onMouseLeave={() => setOpen(false)}
        side="right"
        align="start"
        sideOffset={10}
        className="flex max-h-[400px] w-64 flex-col overflow-y-auto p-2"
      >
        <h4 className="mb-2 px-3 text-sm font-medium text-muted-foreground">
          Page Navigation
        </h4>
        {headings.map((heading) => (
          <button
            key={heading.id}
            onClick={() => handleClick(heading.id)}
            className={cn(
              'hover:bg-accent hover:text-accent-foreground text-foreground/80 w-full truncate rounded-sm px-3 py-1 text-left text-sm transition-colors',
              activeId === heading.id &&
                'bg-accent text-accent-foreground font-semibold'
            )}
            style={{ paddingLeft: `${12 + (heading.level - 1) * 12}px` }}
          >
            {heading.text}
          </button>
        ))}
      </PopoverContent>
    </Popover>
  );
}
