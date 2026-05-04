import { InlineNode, type MyBlock } from '@blocknote/core';
import { type ServerBlockNoteEditor } from '@blocknote/server-util';
import { ReferenceInline, TagInline } from '@notopia-uit/lib/block-note';
import { marked } from 'marked';

function toReferenceInline(noteId: string): ReferenceInline {
  return {
    type: 'reference',
    props: { noteId },
    content: undefined,
  };
}

function toTagInline(tag: string): TagInline {
  return {
    type: 'tag',
    props: { tag },
    content: undefined,
  };
}

function markdownToHTML(markdown: string): string {
  const parsed = marked.parse(markdown, {
    async: false,
    gfm: true,
  });

  return typeof parsed === 'string' ? parsed : '';
}

function transformInlineContent(content: MyBlock['content']): MyBlock['content'] {
  const transformed: InlineNode[] = [];
  if (!Array.isArray(content)) {
    return content;
  }

  for (const node of content) {
    if (node.type === 'link' && typeof node.href === 'string') {
      const href = node.href;

      if (href.startsWith('@') && href.length > 1) {
        transformed.push(toReferenceInline(href.slice(1)));
        continue;
      }

      if (href.startsWith('#') && href.length > 1) {
        transformed.push(toTagInline(href.slice(1)));
        continue;
      }
    }

    transformed.push(node as InlineNode);
  }

  return transformed as MyBlock['content'];
}

function transformBlock(block: MyBlock): MyBlock {
  const transformed = {
    ...block,
    children: block.children.map((child) => transformBlock(child)),
  };
  if (block.type !== 'codeBlock') {
    transformed.content = transformInlineContent(block.content);
  }

  return transformed;
}

export async function parseSeedMarkdownToBlocks(
  editor: ServerBlockNoteEditor,
  markdown: string
): Promise<MyBlock[]> {
  const html = markdownToHTML(markdown);
  const parsedBlocks = (await editor.tryParseHTMLToBlocks(html)) as MyBlock[];
  return parsedBlocks.map((block) => transformBlock(block));
}
