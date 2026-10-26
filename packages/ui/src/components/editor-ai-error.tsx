'use client';

import type { MyEditor } from '@blocknote/core';
import { AIExtension } from '@blocknote/xl-ai';
import { useCallback, useSyncExternalStore } from 'react';

import { Alert, AlertDescription, AlertTitle } from './shadcn/alert';
import { Button } from './shadcn/button';

export function EditorAiError({ editor }: { editor: MyEditor }) {
  const ai = editor.getExtension(AIExtension);
  const subscribe = useCallback(
    (listener: () => void) => ai?.store.subscribe(listener) ?? (() => {}),
    [ai]
  );
  const getSnapshot = useCallback(() => ai?.store.state.aiMenuState ?? 'closed', [ai]);
  const state = useSyncExternalStore(subscribe, getSnapshot, () => 'closed' as const);
  if (state === 'closed' || state.status !== 'error') return null;
  const message =
    state.error instanceof Error ? state.error.message : 'The AI request failed. Please try again.';
  return (
    <Alert variant="destructive" className="my-2">
      <AlertTitle>AI request failed</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="mt-2 w-fit"
        onClick={() => ai?.closeAIMenu()}
      >
        Dismiss
      </Button>
    </Alert>
  );
}
