import type { Chat } from '@ai-sdk/react';
import type { UIMessage } from 'ai';

import type { AssistantAiBridge } from './assistant-ai';
import {
  completedAssistantResults,
  type AssistantContinuation,
  type AssistantTaskData,
} from './assistant-protocol';
import type { AssistantEditorSession } from './assistant-session';

export interface AssistantRunState {
  busy: boolean;
  review: boolean;
  error?: Error;
}
type Bridge = Pick<AssistantAiBridge, 'prepare' | 'send' | 'review' | 'accept' | 'reject'>;

export class AssistantRunner<Editor> {
  state: AssistantRunState = { busy: false, review: false };
  private task?: AssistantTaskData;
  private continuation?: AssistantContinuation;
  private abort?: AbortController;
  private bridge?: Bridge;
  private deadlineTimer?: ReturnType<typeof setTimeout>;
  private disposed = false;
  private currentTask(): AssistantTaskData | undefined {
    return this.task;
  }

  constructor(
    private readonly options: {
      chat: Chat<UIMessage>;
      session: AssistantEditorSession<Editor>;
      bridge(editor: Editor): Promise<Bridge>;
      cancel(taskId: string): Promise<void>;
      onChange(state: AssistantRunState): void;
    }
  ) {}

  private update(value: Partial<AssistantRunState>) {
    this.state = { ...this.state, ...value };
    this.options.onChange(this.state);
  }

  activate() {
    this.disposed = false;
  }

  onData(part: { type: string; data: unknown }) {
    if (this.disposed || this.abort?.signal.aborted) return;
    if (part.type === 'data-assistant-task') {
      this.task = part.data as AssistantTaskData;
      clearTimeout(this.deadlineTimer);
      this.deadlineTimer = setTimeout(
        () => {
          void this.stop(new Error('The assistant task timed out.'));
        },
        Math.max(0, this.task.expiresAt - Date.now())
      );
    } else if (part.type === 'data-assistant-continuation')
      this.continuation = part.data as AssistantContinuation;
  }

  async submit(message: UIMessage) {
    if (this.disposed || this.state.busy || this.state.review) return;
    const abort = new AbortController();
    this.abort = abort;
    this.task = undefined;
    this.continuation = undefined;
    this.update({ busy: true, error: undefined });
    const { chat, session } = this.options;
    chat.clearError();
    try {
      const initialTarget = session.target;
      if (initialTarget) {
        const bridge = await this.options.bridge(session.getReadyEditor(initialTarget.sessionId));
        abort.signal.throwIfAborted();
        this.bridge = bridge;
        const request = await this.bridge.prepare(abort.signal);
        await this.bridge.send(
          chat,
          request,
          message,
          { body: { noteId: initialTarget.noteId } },
          abort.signal
        );
      } else await chat.sendMessage(message);
      if (chat.error) throw chat.error;
      while (this.continuation) {
        abort.signal.throwIfAborted();
        const continuation: AssistantContinuation = this.continuation;
        this.continuation = undefined;
        if (
          continuation.taskId !== this.currentTask()?.taskId ||
          Date.now() >= continuation.expiresAt
        )
          throw new Error('The assistant continuation expired.');
        const response = chat.lastMessage;
        const calls = (response?.parts ?? []).filter(
          (part) => part.type === 'tool-openNoteForEditing' && part.state === 'input-available'
        );
        if (calls.length > 1) throw new Error('The assistant requested multiple target notes.');
        for (const call of calls) {
          if (
            !('input' in call) ||
            !call.input ||
            typeof call.input !== 'object' ||
            !('noteId' in call.input) ||
            typeof call.input.noteId !== 'string'
          )
            throw new Error('Invalid target note.');
          const target = await session.open(call.input.noteId, abort.signal);
          abort.signal.throwIfAborted();
          if (!this.bridge) {
            const bridge = await this.options.bridge(session.getReadyEditor(target.sessionId));
            abort.signal.throwIfAborted();
            this.bridge = bridge;
          }
          await chat.addToolOutput({
            tool: 'openNoteForEditing',
            toolCallId: call.toolCallId,
            output: { noteId: target.noteId, ready: true },
          });
        }
        if (!session.target || !this.bridge)
          throw new Error('No live editing session is available.');
        const results = completedAssistantResults(chat.lastMessage);
        const request = await this.bridge.prepare(abort.signal);
        await this.bridge.send(
          chat,
          request,
          undefined,
          {
            body: {
              action: 'continue',
              ...continuation,
              expiresAt: undefined,
              results,
              noteId: session.target.noteId,
              documentState: request.documentState,
            },
          },
          abort.signal
        );
        if (chat.error) throw chat.error;
      }
      abort.signal.throwIfAborted();
    } catch (error) {
      if (!abort.signal.aborted && !this.disposed)
        this.update({
          error: error instanceof Error ? error : new Error('The assistant task failed.'),
        });
      const task = this.currentTask();
      if (task) {
        try {
          await this.options.cancel(task.taskId);
        } catch (cancelError) {
          this.update({
            error: new Error(`Could not cancel the assistant task: ${String(cancelError)}`),
          });
        }
      }
    } finally {
      if (this.abort === abort) {
        clearTimeout(this.deadlineTimer);
        this.task = undefined;
        this.continuation = undefined;
        if (!this.disposed) {
          if (this.bridge) this.bridge.review();
          this.update({ busy: false, review: !!this.bridge });
        }
      }
    }
  }

  async stop(error?: Error) {
    this.abort?.abort();
    this.continuation = undefined;
    clearTimeout(this.deadlineTimer);
    this.options.session.cancel();
    const task = this.task;
    this.task = undefined;
    await this.options.chat.stop();
    if (error) this.update({ error });
    if (task) {
      const id = task.taskId;
      try {
        await this.options.cancel(id);
      } catch (cause) {
        this.update({ error: new Error(`Could not cancel the assistant task: ${String(cause)}`) });
      }
    }
  }

  accept() {
    this.bridge?.accept();
    this.bridge = undefined;
    this.update({ review: false });
  }
  reject() {
    this.bridge?.reject();
    this.bridge = undefined;
    this.update({ review: false });
  }
  clearError() {
    this.options.chat.clearError();
    this.update({ error: undefined });
  }
  close() {
    if (this.state.busy || this.state.review) return;
    this.options.session.close();
    this.update({ busy: false, review: false });
  }
  async dispose() {
    this.disposed = true;
    const stopping = this.stop();
    this.reject();
    this.options.session.close();
    this.update({ busy: false, review: false });
    await stopping;
  }
}
