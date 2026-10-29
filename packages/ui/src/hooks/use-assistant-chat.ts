'use client';

import { Chat, useChat } from '@ai-sdk/react';
import type { MyEditor } from '@blocknote/core';
import type { UIMessage } from 'ai';
import { useEffect, useState } from 'react';

import { useAiApiUrl } from '../contexts/ai-context';
import { AssistantRunner, type AssistantRunState } from '../lib/assistant-runner';
import { AssistantEditorSession, type AssistantTarget } from '../lib/assistant-session';
import { cancelAssistantTask, createAssistantTransport } from '../lib/assistant-transport';

export function useAssistantChat(workspaceId: string) {
  const api = useAiApiUrl();
  const [target, setTarget] = useState<AssistantTarget | null>(null);
  const [state, setState] = useState<AssistantRunState>({ busy: false, review: false });
  const [controller] = useState(() => {
    const session = new AssistantEditorSession<MyEditor>(setTarget);
    const chat = new Chat<UIMessage>({
      id: `workspace-assistant-${workspaceId}`,
      transport: createAssistantTransport(api, workspaceId),
      sendAutomaticallyWhen: () => false,
      onData: (part) => runner.onData(part),
    });
    const runner = new AssistantRunner({
      chat,
      session,
      bridge: async (editor) => {
        const { AssistantAiBridge } = await import('../lib/assistant-ai');
        return new AssistantAiBridge(editor);
      },
      cancel: (taskId) => cancelAssistantTask(api, workspaceId, taskId),
      onChange: setState,
    });
    return { chat, session, runner };
  });
  const chatState = useChat({ chat: controller.chat });
  useEffect(() => {
    controller.runner.activate();
    return () => {
      void controller.runner.dispose();
    };
  }, [controller]);
  return { ...controller, ...state, target, messages: chatState.messages };
}
