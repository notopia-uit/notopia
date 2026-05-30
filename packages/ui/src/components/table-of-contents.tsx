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
  const SEGMENTS = 10;
  const filledSegments = Math.round(scrollProgress * SEGMENTS);

  useEffect(() => {
    const handleScroll = () => {
      const scrollTop = window.scrollY || document.documentElement.scrollTop;
      const docHeight =
        document.documentElement.scrollHeight -
        document.documentElement.clientHeight;
      setCanScroll(docHeight > 0);
      const progress = docHeight > 0 ? scrollTop / docHeight : 0;
      setScrollProgress(progress);
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll();

    return () => {
      window.removeEventListener('scroll', handleScroll);
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
    <div
      className="fixed right-5 top-1/2 z-50 -translate-y-1/2"
    >
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <div
            className="flex cursor-default flex-col items-center gap-4 px-1.5 py-3"
            onMouseEnter={() => setOpen(true)}
            onMouseLeave={() => setOpen(false)}
          >
            {Array.from({ length: SEGMENTS }).map((_, i) => (
              <div
                key={i}
                className={cn(
                  'h-1 w-5 rounded-full transition-all duration-200',
                  i < filledSegments
                    ? 'bg-foreground/40 scale-115'
                    : 'bg-foreground/10 scale-100'
                )}
              />
            ))}
          </div>
        </PopoverTrigger>
        <PopoverContent
          side="left"
          align="center"
          sideOffset={-30}
          className="w-56 px-1 py-3"
          onMouseEnter={() => setOpen(true)}
          onMouseLeave={() => setOpen(false)}
        >
          <h4 className="text-muted-foreground mb-2 px-2 text-[10px] font-semibold tracking-widest uppercase">
            Page Navigation
          </h4>
          <div className="max-h-[50vh] space-y-0.5 overflow-y-auto">
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
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
