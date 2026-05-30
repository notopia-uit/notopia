'use client';

import { parseHighlight } from '@notopia-uit/ui/block-note';

interface HighlightedTextProps {
  text: string;
  className?: string;
}

export function HighlightedText({ text, className }: HighlightedTextProps) {
  return (
    <span className={className}>
      {parseHighlight(text).map((segment, index) =>
        segment.highlighted ? (
          <strong key={index} className="text-primary font-semibold">
            {segment.text}
          </strong>
        ) : (
          <span key={index}>{segment.text}</span>
        )
      )}
    </span>
  );
}
