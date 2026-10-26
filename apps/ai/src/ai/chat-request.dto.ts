import { validateUIMessages, type UIMessage } from 'ai';
import { z } from 'zod';

import { AiError } from '#/common/ai-error';

const jsonSchema = z.object({ type: z.literal('object') }).passthrough();
const requestSchema = z.object({
  workspaceId: z.uuid().optional(),
  noteId: z.uuid().optional(),
  messages: z
    .array(
      z
        .object({
          id: z.string().min(1).max(200),
          role: z.enum(['user', 'assistant']),
          parts: z.array(z.unknown()).min(1).max(200),
        })
        .passthrough()
    )
    .min(1)
    .max(100),
  toolDefinitions: z
    .partialRecord(
      z.enum(['applyDocumentOperations']),
      z.object({
        description: z.string().max(10000).optional(),
        inputSchema: jsonSchema,
        outputSchema: z.record(z.string(), z.unknown()),
      })
    )
    .optional(),
});

export type BlockNoteChatBody = Omit<z.infer<typeof requestSchema>, 'messages'> & {
  messages: UIMessage[];
};

export async function parseChatRequest(input: unknown): Promise<BlockNoteChatBody> {
  try {
    if (Buffer.byteLength(JSON.stringify(input) ?? '') > 512000) {
      throw new AiError(413, 'requestTooLarge', 'AI requests must be at most 500 KB.');
    }
    const body = requestSchema.parse(input);
    const messages = await validateUIMessages({ messages: body.messages });
    if (messages.some((message) => message.parts.some((part) => part.type === 'file'))) {
      throw new Error('Remote file attachments are not supported');
    }
    const last = messages.at(-1);
    if (
      last?.role !== 'user' ||
      !last.parts.some((part) => part.type === 'text' && part.text.trim())
    ) {
      throw new Error('The final message must be a user prompt');
    }
    return { ...body, messages };
  } catch (error) {
    if (error instanceof AiError) throw error;
    throw new AiError(
      400,
      'invalidAiRequest',
      'Invalid messages, context, or tool definitions. End with a user prompt.'
    );
  }
}
