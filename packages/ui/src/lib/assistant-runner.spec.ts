import { Chat } from '@ai-sdk/react';
import type { UIMessage, UIMessageChunk } from 'ai';
import { describe, expect, it } from 'vitest';

import { AssistantRunner } from './assistant-runner';
import { AssistantEditorSession } from './assistant-session';

describe('assistant browser orchestration', () => {
  it('can activate after Strict Mode cleanup and still stop on final unmount', async () => {
    let requests = 0;
    const chat = new Chat<UIMessage>({
      transport: {
        reconnectToStream: async () => null,
        sendMessages: async () => {
          requests++;
          return new ReadableStream({
            start(controller) {
              controller.close();
            },
          });
        },
      },
    });
    const runner = new AssistantRunner({
      chat,
      session: new AssistantEditorSession<object>(() => {}),
      cancel: async () => {},
      onChange: () => {},
      bridge: async () => {
        throw new Error('No target');
      },
    });
    await runner.dispose();
    runner.activate();
    await runner.submit({ id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Hello' }] });
    expect(requests).toBe(1);
    await runner.dispose();
    await runner.submit({ id: 'late', role: 'user', parts: [{ type: 'text', text: 'Late' }] });
    expect(requests).toBe(1);
  });
  it('does not revive review or accept late task data after navigation disposal', async () => {
    let release!: () => void;
    const loading = new Promise<void>((resolve) => {
      release = resolve;
    });
    const session = new AssistantEditorSession<object>(() => {});
    const opening = session.open('note', new AbortController().signal);
    session.ready(session.target!.sessionId, {});
    await opening;
    let reviews = 0;
    let rejects = 0;
    const runner = new AssistantRunner({
      chat: new Chat<UIMessage>({}),
      session,
      cancel: async () => {},
      onChange: () => {},
      bridge: async () => {
        await loading;
        return {
          prepare: async () => ({ documentState: {} }) as any,
          send: async () => {},
          review: () => {
            reviews++;
          },
          accept: () => {},
          reject: () => {
            rejects++;
          },
        };
      },
    });
    const submission = runner.submit({
      id: 'prompt',
      role: 'user',
      parts: [{ type: 'text', text: 'Edit' }],
    });
    await new Promise((resolve) => setTimeout(resolve, 0));
    await runner.dispose();
    release();
    await submission;
    runner.onData({
      type: 'data-assistant-task',
      data: { taskId: 'late', expiresAt: Date.now() + 60000 },
    });
    expect(reviews).toBe(0);
    expect(rejects).toBe(0);
    expect(runner.state.review).toBe(false);
    expect(session.target).toBeNull();
  });
  it('waits for the open editor and resumes the same task exactly once', async () => {
    let calls = 0;
    let runner: AssistantRunner<object>;
    const editor = {};
    const session = new AssistantEditorSession<object>(() => {});
    const chat = new Chat<UIMessage>({
      onData: (part) => runner.onData(part),
      transport: {
        reconnectToStream: async () => null,
        sendMessages: async () => {
          calls++;
          const chunks: UIMessageChunk[] =
            calls === 1
              ? [
                  {
                    type: 'data-assistant-task',
                    data: { taskId: 'task', expiresAt: Date.now() + 60000 },
                    transient: true,
                  },
                  { type: 'start', messageId: 'open-message' },
                  {
                    type: 'tool-input-available',
                    toolName: 'openNoteForEditing',
                    toolCallId: 'open',
                    input: { noteId: 'note' },
                  },
                  { type: 'finish', finishReason: 'tool-calls' },
                  {
                    type: 'data-assistant-continuation',
                    data: { taskId: 'task', token: 'token', expiresAt: Date.now() + 60000 },
                    transient: true,
                  },
                ]
              : [
                  { type: 'start', messageId: 'done' },
                  { type: 'finish', finishReason: 'stop' },
                ];
          return new ReadableStream({
            start(controller) {
              chunks.forEach((chunk) => controller.enqueue(chunk));
              controller.close();
            },
          });
        },
      },
    });
    runner = new AssistantRunner({
      chat,
      session,
      cancel: async () => {},
      onChange: () => {},
      bridge: async () => ({
        prepare: async () => ({ documentState: {} }) as any,
        send: async (chat, _request, message, options) => {
          await chat.sendMessage(message, options);
        },
        review: () => {},
        accept: () => {},
        reject: () => {},
      }),
    });
    const submission = runner.submit({
      id: 'prompt',
      role: 'user',
      parts: [{ type: 'text', text: 'Edit note' }],
    });
    await new Promise<void>((resolve) => {
      const timer = setInterval(() => {
        if (session.target) {
          clearInterval(timer);
          resolve();
        }
      }, 1);
    });
    expect(calls).toBe(1);
    session.ready(session.target!.sessionId, editor);
    await submission;
    expect(calls).toBe(2);
    expect(runner.state.review).toBe(true);
    expect(runner.state.error).toBeUndefined();
  });

  it('stops a pending open and cancels the server task', async () => {
    const session = new AssistantEditorSession<object>(() => {});
    const cancelled: string[] = [];
    const runner = new AssistantRunner({
      chat: new Chat<UIMessage>({}),
      session,
      cancel: async (id) => {
        cancelled.push(id);
      },
      onChange: () => {},
      bridge: async () => {
        throw new Error('Must not open');
      },
    });
    runner.onData({
      type: 'data-assistant-task',
      data: { taskId: 'task', expiresAt: Date.now() + 60000 },
    });
    await runner.stop();
    expect(cancelled).toEqual(['task']);
    expect(session.target).toBeNull();
  });
});
