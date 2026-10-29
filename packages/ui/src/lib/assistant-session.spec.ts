import { describe, expect, it } from 'vitest';

import { AssistantEditorSession } from './assistant-session';

describe('assistant editor readiness', () => {
  it('invalidates readiness while disconnected until resynchronized', async () => {
    const session = new AssistantEditorSession<object>(() => {});
    const opening = session.open('note', new AbortController().signal);
    const target = session.target!;
    const editor = {};
    session.ready(target.sessionId, editor);
    await opening;
    session.unready(target.sessionId);
    expect(() => session.getReadyEditor(target.sessionId)).toThrow();
    session.ready(target.sessionId, editor);
    expect(session.getReadyEditor(target.sessionId)).toBe(editor);
  });
  it('waits for sync and ignores callbacks from an old target', async () => {
    const session = new AssistantEditorSession<object>(() => {});
    const abort = new AbortController();
    const opening = session.open('note', abort.signal);
    const target = session.target!;
    const editor = {};
    session.ready('stale', editor);
    expect(() => session.getReadyEditor(target.sessionId)).toThrow();
    session.ready(target.sessionId, editor);
    expect(await opening).toEqual(target);
    expect(session.getReadyEditor(target.sessionId)).toBe(editor);
    session.close();
    expect(() => session.getReadyEditor(target.sessionId)).toThrow();
  });
  it('deduplicates opening and disposes on cancellation', async () => {
    const session = new AssistantEditorSession<object>(() => {});
    const abort = new AbortController();
    const opening = session.open('note', abort.signal);
    expect(session.open('note', abort.signal)).toBe(opening);
    const rejected = expect(opening).rejects.toThrow();
    abort.abort();
    await rejected;
    expect(session.target).toBeNull();
  });
  it('surfaces auth failure and cannot retain a failed writable target', async () => {
    const session = new AssistantEditorSession<object>(() => {});
    const opening = session.open('note', new AbortController().signal);
    const rejected = expect(opening).rejects.toThrow('Access denied');
    session.fail(session.target!.sessionId, new Error('Access denied'));
    await rejected;
    expect(session.target).toBeNull();
  });
});
