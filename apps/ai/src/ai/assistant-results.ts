import type { UIMessage } from 'ai';
import { z } from 'zod';

import { AiError } from '#/common/ai-error';

import type { AssistantTask } from './assistant-task.types';
import type { AssistantContinuationBody } from './chat-request.dto';

const openResult = z.strictObject({ noteId: z.uuid(), ready: z.literal(true) });

export function applyAssistantResults(
  task: AssistantTask,
  body: AssistantContinuationBody
): UIMessage[] {
  const invalid = (): never => {
    throw new AiError(
      400,
      'invalidAssistantResult',
      'Results do not match the pending assistant task.'
    );
  };
  if (
    body.taskId !== task.id ||
    body.workspaceId !== task.owner.workspaceId ||
    body.noteId !== task.targetNoteId ||
    body.results.length !== task.pendingCalls.length
  )
    invalid();

  const messages = structuredClone(task.messages);
  const seen = new Set<string>();
  for (const result of body.results) {
    const expected = task.pendingCalls.find((call) => call.toolCallId === result.toolCallId);
    if (!expected || expected.toolName !== result.toolName || seen.has(result.toolCallId))
      invalid();
    seen.add(result.toolCallId);
    if (result.toolName === 'openNoteForEditing') {
      const output = openResult.safeParse(result.output);
      if (!output.success || output.data.noteId !== task.targetNoteId) invalid();
    } else if (!z.strictObject({ status: z.literal('ok') }).safeParse(result.output).success)
      invalid();

    const response = messages.at(-1);
    if (response?.role !== 'assistant') invalid();
    const part = response!.parts.find(
      (part) =>
        part.type === `tool-${result.toolName}` &&
        'toolCallId' in part &&
        part.toolCallId === result.toolCallId
    );
    if (!part || !('state' in part) || part.state !== 'input-available') invalid();
    Object.assign(part!, { state: 'output-available', output: result.output });
  }
  const prompt = messages.find(
    (message) => message.id === task.promptId && message.role === 'user'
  );
  if (!prompt) invalid();
  if (body.documentState) {
    const metadata = z.record(z.string(), z.unknown()).safeParse(prompt!.metadata);
    prompt!.metadata = {
      ...(metadata.success ? metadata.data : {}),
      documentState: body.documentState,
    };
  }
  return messages;
}
