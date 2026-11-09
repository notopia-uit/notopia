export interface AssistantTarget {
  noteId: string;
  sessionId: string;
}

export class AssistantEditorSession<Editor> {
  target: AssistantTarget | null = null;
  private editor?: Editor;
  private pending?: {
    promise: Promise<AssistantTarget>;
    resolve(target: AssistantTarget): void;
    reject(error: Error): void;
    cleanup(): void;
  };

  constructor(private readonly onTarget: (target: AssistantTarget | null) => void) {}

  open(noteId: string, signal: AbortSignal): Promise<AssistantTarget> {
    signal.throwIfAborted();
    if (this.target?.noteId === noteId) {
      if (this.pending) return this.pending.promise;
      if (this.editor) return Promise.resolve(this.target);
    }
    if (this.target) throw new Error('Close the current target before opening another note.');
    const target = { noteId, sessionId: crypto.randomUUID() };
    const { promise, resolve, reject } = Promise.withResolvers<AssistantTarget>();
    const cancel = () => this.cancel();
    const timer = setTimeout(
      () => this.fail(target.sessionId, new Error('The note did not sync in time.')),
      30000
    );
    this.pending = {
      promise,
      resolve,
      reject,
      cleanup: () => {
        clearTimeout(timer);
        signal.removeEventListener('abort', cancel);
      },
    };
    signal.addEventListener('abort', cancel, { once: true });
    this.target = target;
    this.onTarget(target);
    return promise;
  }

  ready(sessionId: string, editor: Editor) {
    if (this.target?.sessionId !== sessionId) return;
    this.editor = editor;
    this.pending?.cleanup();
    this.pending?.resolve(this.target);
    this.pending = undefined;
  }

  getReadyEditor(sessionId: string): Editor {
    if (this.target?.sessionId !== sessionId || !this.editor)
      throw new Error('The target editor is not ready.');
    return this.editor;
  }

  unready(sessionId: string) {
    if (this.target?.sessionId === sessionId) this.editor = undefined;
  }

  fail(sessionId: string, error: Error) {
    if (this.target?.sessionId !== sessionId) return;
    this.pending?.cleanup();
    this.pending?.reject(error);
    this.pending = undefined;
    this.editor = undefined;
    this.target = null;
    this.onTarget(null);
  }

  cancel() {
    if (this.pending && this.target)
      this.fail(this.target.sessionId, new Error('Opening the note was cancelled.'));
  }

  close() {
    this.cancel();
    this.editor = undefined;
    this.target = null;
    this.onTarget(null);
  }
}
