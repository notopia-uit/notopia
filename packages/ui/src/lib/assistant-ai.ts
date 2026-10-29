import type { Chat } from '@ai-sdk/react';
import type { MyEditor } from '@blocknote/core';
import { ForkYDocExtension } from '@blocknote/core/yjs';
import {
  AIExtension,
  aiDocumentFormats,
  buildAIRequest,
  sendMessageWithAIRequest,
  type AIRequest,
} from '@blocknote/xl-ai';
import type { UIMessage } from 'ai';

export class AssistantAiBridge {
  private reviewing = false;
  private revision = 0;

  constructor(private readonly editor: MyEditor) {}

  async prepare(signal: AbortSignal): Promise<AIRequest> {
    signal.throwIfAborted();
    const extension = this.editor.getExtension(AIExtension);
    if (!extension) throw new Error('The editor AI extension is unavailable.');
    if (!this.reviewing) {
      extension.openAIMenuAtBlock(this.editor.document[0].id);
      this.editor.getExtension(ForkYDocExtension)?.fork();
      this.reviewing = true;
    }
    extension.setAIResponseStatus('thinking');
    const revision = this.revision;
    const checkActive = () => {
      signal.throwIfAborted();
      if (revision !== this.revision) throw new Error('The editing session has ended.');
    };
    const provider = aiDocumentFormats.html.getStreamToolsProvider({ withDelays: false });
    const request = await buildAIRequest({
      editor: this.editor,
      useSelection: false,
      deleteEmptyCursorBlock: false,
      streamToolsProvider: {
        getStreamTools: (...args) =>
          provider.getStreamTools(...args).map((tool) => ({
            ...tool,
            executor: () => {
              const executor = tool.executor();
              return {
                execute: (chunk, abortSignal) => {
                  checkActive();
                  abortSignal?.throwIfAborted();
                  return executor.execute(chunk, abortSignal);
                },
              };
            },
          })) as ReturnType<typeof provider.getStreamTools>,
      },
      onStart: () => extension.setAIResponseStatus('ai-writing'),
    });
    signal.throwIfAborted();
    return request;
  }

  async send(
    chat: Chat<UIMessage>,
    request: AIRequest,
    message: Parameters<Chat<UIMessage>['sendMessage']>[0],
    options: Parameters<Chat<UIMessage>['sendMessage']>[1],
    signal: AbortSignal
  ) {
    signal.throwIfAborted();
    // BlockNote subscribes to the last message before sending. A completed tool
    // from the preceding continuation must not become a new executor stream.
    const placeholderId = !message ? crypto.randomUUID() : undefined;
    if (placeholderId)
      chat.messages = [...chat.messages, { id: placeholderId, role: 'assistant', parts: [] }];
    try {
      const result = await sendMessageWithAIRequest(chat, request, message, options, signal);
      signal.throwIfAborted();
      if (chat.error) throw chat.error;
      if (!result.ok) throw result.error;
    } finally {
      if (placeholderId)
        chat.messages = chat.messages.filter((entry) => entry.id !== placeholderId);
    }
  }

  review() {
    this.editor.getExtension(AIExtension)?.setAIResponseStatus('user-reviewing');
  }

  accept() {
    if (!this.reviewing) return;
    this.revision++;
    this.editor.getExtension(AIExtension)?.acceptChanges();
    this.reviewing = false;
  }

  reject() {
    if (!this.reviewing) return;
    this.revision++;
    this.editor.getExtension(AIExtension)?.rejectChanges();
    this.reviewing = false;
  }
}
