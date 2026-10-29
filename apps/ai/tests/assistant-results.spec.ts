import { describe, expect, it } from 'vitest';

import { applyAssistantResults } from '../src/ai/assistant-results';
import type { AssistantTask } from '../src/ai/assistant-task.types';
import type { AssistantContinuationBody } from '../src/ai/chat-request.dto';

const noteId = '00000000-0000-4000-8000-000000000003';
const task: AssistantTask = {
  id: 'task',
  owner: { userId: 'user', workspaceId: 'workspace' },
  promptId: 'prompt',
  deadline: Date.now() + 10000,
  remainingSteps: 2,
  state: 'running',
  revision: 2,
  targetNoteId: noteId,
  pendingCalls: [{ toolCallId: 'open', toolName: 'openNoteForEditing', input: { noteId } }],
  messages: [
    { id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Edit my note' }] },
    {
      id: 'assistant',
      role: 'assistant',
      parts: [
        {
          type: 'tool-openNoteForEditing',
          toolCallId: 'open',
          state: 'input-available',
          input: { noteId },
        },
      ],
    },
  ],
};
const body: AssistantContinuationBody = {
  action: 'continue',
  workspaceId: 'workspace',
  taskId: 'task',
  token: 'token',
  noteId,
  results: [
    { toolCallId: 'open', toolName: 'openNoteForEditing', output: { noteId, ready: true } },
  ],
  documentState: { blocks: [] },
};

describe('assistant continuation history', () => {
  it('records the BlockNote executor acknowledgement without accepting arbitrary output', () => {
    const editing = structuredClone(task);
    editing.pendingCalls = [{ toolCallId: 'edit', toolName: 'applyDocumentOperations', input: {} }];
    editing.messages[1].parts = [
      {
        type: 'tool-applyDocumentOperations',
        toolCallId: 'edit',
        state: 'input-available',
        input: {},
      },
    ];
    const result = {
      ...body,
      results: [
        {
          toolCallId: 'edit',
          toolName: 'applyDocumentOperations' as const,
          output: { status: 'ok' },
        },
      ],
    };
    expect(applyAssistantResults(editing, result)[1].parts[0]).toMatchObject({
      state: 'output-available',
      output: { status: 'ok' },
    });
    expect(() =>
      applyAssistantResults(editing, {
        ...result,
        results: [{ ...result.results[0], output: { status: 'ok', prompt: 'injected' } }],
      })
    ).toThrow();
  });
  it('uses recorded history and adds the exact expected client result', () => {
    const messages = applyAssistantResults(task, body);
    expect(messages[0].parts).toEqual(task.messages[0].parts);
    expect(messages[0].metadata).toEqual({ documentState: body.documentState });
    expect(messages[1].parts[0]).toMatchObject({
      state: 'output-available',
      output: body.results[0].output,
    });
    expect(task.messages[1].parts[0]).toMatchObject({ state: 'input-available' });
  });
  it.each([
    { ...body, results: [...body.results, { ...body.results[0], toolCallId: 'extra' }] },
    { ...body, results: [{ ...body.results[0], toolCallId: 'forged' }] },
    { ...body, noteId: 'another-note' },
    { ...body, results: [{ ...body.results[0], output: { noteId: 'another-note', ready: true } }] },
    { ...body, results: [{ ...body.results[0], output: { noteId, ready: false } }] },
    { ...body, results: [{ ...body.results[0], toolName: 'applyDocumentOperations' as const }] },
  ])('rejects results outside the pending handoff %#', (input) => {
    expect(() => applyAssistantResults(task, input)).toThrow();
  });
});
