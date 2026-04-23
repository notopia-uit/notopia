import { type ServerBlockNoteEditor } from '@blocknote/server-util';
import { marked } from 'marked';

type InlineNode = {
  type?: string;
  href?: unknown;
  props?: Record<string, unknown>;
  [key: string]: unknown;
};

type ParsedBlocks = Awaited<
  ReturnType<ServerBlockNoteEditor['tryParseHTMLToBlocks']>
>;

type ParsedBlock = ParsedBlocks[number];

type BlockView = ParsedBlock & {
  content?: unknown;
  children: ParsedBlock[];
};

function isInlineArray(content: unknown): content is InlineNode[] {
  return Array.isArray(content);
}

function toReferenceInline(noteId: string): InlineNode {
  return {
    type: 'reference',
    props: { noteId },
    content: undefined,
  };
}

function toTagInline(tag: string): InlineNode {
  return {
    type: 'tag',
    props: { tag },
    content: undefined,
  };
}

function isLinkInlineNode(node: InlineNode): node is InlineNode & { href: string } {
  return node.type === 'link' && typeof node.href === 'string';
}

function markdownToHTML(markdown: string): string {
  const parsed = marked.parse(markdown, {
    async: false,
    gfm: true,
  });

  return typeof parsed === 'string' ? parsed : '';
}

function transformInlineContent(content: InlineNode[]): InlineNode[] {
  const transformed: InlineNode[] = [];

  for (const node of content) {
    if (isLinkInlineNode(node)) {
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

    transformed.push(node);
  }

  return transformed;
}

function transformBlock(block: ParsedBlock): ParsedBlock {
  const blockView = block as BlockView;
  const transformed = {
    ...block,
    children: blockView.children.map((child) => transformBlock(child)),
  };

  if (block.type !== 'codeBlock' && isInlineArray(blockView.content)) {
    transformed.content = transformInlineContent(
      blockView.content
    ) as ParsedBlock['content'];
  }

  return transformed as ParsedBlock;
}

export async function parseSeedMarkdownToBlocks(
  editor: ServerBlockNoteEditor,
  markdown: string
): Promise<unknown> {
  const html = markdownToHTML(markdown);
  const parsedBlocks = await editor.tryParseHTMLToBlocks(html);
  return parsedBlocks.map((block) => transformBlock(block));
}
