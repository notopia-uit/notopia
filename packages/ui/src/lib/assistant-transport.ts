import { DefaultChatTransport } from 'ai';

import { fetchAi } from './ai-transport';
import { fetchAccessTokenClientSide } from './get-access-token-client-side';

export function createAssistantTransport(api: string, workspaceId: string) {
  return new DefaultChatTransport({
    api,
    headers: async () => ({ Authorization: `Bearer ${await fetchAccessTokenClientSide()}` }),
    fetch: fetchAi,
    prepareSendMessagesRequest: ({ messages, body }) => {
      if (body?.action === 'continue') {
        const { taskId, token, results, noteId, documentState, toolDefinitions } = body;
        return {
          body: {
            action: 'continue',
            workspaceId,
            taskId,
            token,
            results,
            noteId,
            documentState,
            toolDefinitions,
          },
        };
      }
      return { body: { ...body, workspaceId, assistantMode: true, messages } };
    },
  });
}

export async function cancelAssistantTask(api: string, workspaceId: string, taskId: string) {
  await fetchAi(api, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${await fetchAccessTokenClientSide()}`,
    },
    body: JSON.stringify({ action: 'cancel', workspaceId, taskId }),
    keepalive: true,
  });
}
