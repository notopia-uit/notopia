import { Chat } from '@ai-sdk/react';
import { BlockNoteEditor } from '@blocknote/core';
import { withCollaboration } from '@blocknote/core/yjs';
import { AIExtension } from '@blocknote/xl-ai';
import * as Y from 'yjs';

import { AssistantAiBridge } from '../../src/lib/assistant-ai';
import { AssistantRunner } from '../../src/lib/assistant-runner';
import { AssistantEditorSession } from '../../src/lib/assistant-session';

try {
  const doc = new Y.Doc();
  const editor = BlockNoteEditor.create(
    withCollaboration({
      collaboration: {
        fragment: doc.getXmlFragment('prosemirror'),
        user: { name: 'Test', color: '#111111' },
      },
      extensions: [AIExtension({})],
    })
  );
  editor.mount(document.getElementById('editor')!);
  editor.replaceBlocks(editor.document, [{ id: 'base', type: 'paragraph', content: 'Original' }]);
  const remote = new Y.Doc();
  Y.applyUpdate(remote, Y.encodeStateAsUpdate(doc));
  const collaborator = BlockNoteEditor.create(
    withCollaboration({
      collaboration: {
        fragment: remote.getXmlFragment('prosemirror'),
        user: { name: 'Remote', color: '#222222' },
      },
    })
  );
  const second = document.createElement('div');
  document.body.append(second);
  collaborator.mount(second);
  doc.on('update', (update: Uint8Array) => Y.applyUpdate(remote, update));
  remote.on('update', (update: Uint8Array) => Y.applyUpdate(doc, update));
  const bridge = new AssistantAiBridge(editor as any);
  await bridge.prepare(new AbortController().signal);
  editor.updateBlock('base', { content: 'AI changed' });
  collaborator.insertBlocks(
    [{ id: 'remote', type: 'paragraph', content: 'Collaborator kept' }],
    'base',
    'after'
  );
  bridge.review();
  bridge.reject();
  const rejected = JSON.stringify(editor.document);
  if (
    !rejected.includes('Original') ||
    !rejected.includes('Collaborator kept') ||
    rejected.includes('AI changed')
  )
    throw new Error(`Reject lost edits: ${rejected}`);
  await bridge.prepare(new AbortController().signal);
  editor.updateBlock('base', { content: 'Accepted change' });
  bridge.review();
  bridge.accept();
  const accepted = JSON.stringify(collaborator.document);
  if (!accepted.includes('Accepted change') || !accepted.includes('Collaborator kept'))
    throw new Error(`Accept did not merge: ${accepted}`);
  let runner: AssistantRunner<any>;
  let requests = 0;
  const task = { taskId: 'test-task', expiresAt: Date.now() + 60000 };
  const session = new AssistantEditorSession<any>((target) => {
    if (target) queueMicrotask(() => session.ready(target.sessionId, editor));
  });
  const chat = new Chat({
    onData: (part) => runner.onData(part),
    sendAutomaticallyWhen: () => false,
    transport: {
      reconnectToStream: async () => null,
      sendMessages: async () => {
        requests++;
        const input = JSON.stringify({
          operations: [{ type: 'update', id: 'base$', block: '<p>Streamed change</p>' }],
        });
        const chunks: any[] =
          requests === 1
            ? [
                { type: 'data-assistant-task', data: task, transient: true },
                { type: 'start', messageId: 'open' },
                {
                  type: 'tool-input-available',
                  toolName: 'openNoteForEditing',
                  toolCallId: 'open',
                  input: { noteId: 'note' },
                },
                { type: 'finish', finishReason: 'tool-calls' },
                {
                  type: 'data-assistant-continuation',
                  data: { ...task, token: 'one' },
                  transient: true,
                },
              ]
            : requests === 2
              ? [
                  { type: 'start', messageId: 'edit' },
                  {
                    type: 'tool-input-start',
                    toolName: 'applyDocumentOperations',
                    toolCallId: 'edit',
                  },
                  {
                    type: 'tool-input-delta',
                    toolCallId: 'edit',
                    inputTextDelta: input.slice(0, 45),
                  },
                  { type: 'tool-input-delta', toolCallId: 'edit', inputTextDelta: input.slice(45) },
                  {
                    type: 'tool-input-available',
                    toolName: 'applyDocumentOperations',
                    toolCallId: 'edit',
                    input: JSON.parse(input),
                  },
                  { type: 'finish', finishReason: 'tool-calls' },
                  {
                    type: 'data-assistant-continuation',
                    data: { ...task, token: 'two' },
                    transient: true,
                  },
                ]
              : [
                  { type: 'start', messageId: 'done' },
                  { type: 'text-start', id: 'done-text' },
                  { type: 'text-delta', id: 'done-text', delta: 'Done' },
                  { type: 'text-end', id: 'done-text' },
                  { type: 'finish', finishReason: 'stop' },
                ];
        return new ReadableStream({
          async start(controller) {
            for (const chunk of chunks) {
              controller.enqueue(chunk);
              await new Promise((r) => setTimeout(r, 5));
            }
            controller.close();
          },
        });
      },
    },
  });
  runner = new AssistantRunner({
    chat,
    session,
    bridge: async (editor) => new AssistantAiBridge(editor),
    cancel: async () => {},
    onChange: () => {},
  });
  await runner.submit({ id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Edit note' }] });
  if (runner.state.error) throw runner.state.error;
  if (requests !== 3 || !JSON.stringify(editor.document).includes('Streamed change'))
    throw new Error(`Streaming did not execute: ${requests} ${JSON.stringify(editor.document)}`);
  collaborator.insertBlocks(
    [{ id: 'remote2', type: 'paragraph', content: 'Concurrent streaming collaborator' }],
    'remote',
    'after'
  );
  runner.reject();
  const streamedReject = JSON.stringify(editor.document);
  if (
    streamedReject.includes('Streamed change') ||
    !streamedReject.includes('Concurrent streaming collaborator')
  )
    throw new Error('Streamed rejection lost collaborator data');
  const stopping = new AbortController();
  const queuedRequest = await bridge.prepare(stopping.signal);
  const queuedDelete = queuedRequest.streamTools
    .find((tool: any) => tool.name === 'delete')!
    .executor();
  stopping.abort();
  bridge.reject();
  let rejectedLateDelete = false;
  try {
    await queuedDelete.execute({
      operation: { type: 'delete', id: 'remote2' },
      isUpdateToPreviousOperation: false,
      isPossiblyPartial: false,
      metadata: {},
    });
  } catch {
    rejectedLateDelete = true;
  }
  if (
    !rejectedLateDelete ||
    !JSON.stringify(editor.document).includes('Concurrent streaming collaborator')
  )
    throw new Error('Queued delete mutated after Stop/reject');
  if (chat.messages.some((message) => message.parts.length === 0))
    throw new Error('Continuation left an empty history message');
  document.getElementById('result')!.textContent = JSON.stringify({
    passed: true,
    requests,
    rejected: JSON.parse(rejected),
    accepted: JSON.parse(accepted),
    streamedReject: JSON.parse(streamedReject),
  });
} catch (error) {
  document.getElementById('result')!.textContent = String(error) + '\n' + (error as Error).stack;
}
