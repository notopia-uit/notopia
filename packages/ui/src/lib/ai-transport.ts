import { DefaultChatTransport } from 'ai';

import { AppError } from './errors';
import { fetchAccessTokenClientSide } from './get-access-token-client-side';

export class AiRequestError extends AppError {
  override readonly name = 'AiRequestError';

  constructor(message: string, statusCode: number, code: string) {
    super(message, { statusCode, code });
  }
}

export async function fetchAi(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const response = await fetch(input, init);
  if (response.ok) return response;
  const body = await response.json().catch(() => null);
  const code = typeof body?.code === 'string' ? body.code : 'aiRequestFailed';
  let message = 'The AI request failed. Please try again.';
  if (response.status === 401) message = 'Your session expired. Sign in again to use AI.';
  else if (response.status === 402) {
    message = 'This workspace has no AI credits available.';
    if (typeof body?.resetAt === 'string' && !Number.isNaN(Date.parse(body.resetAt))) {
      message += ` Credits reset on ${new Date(body.resetAt).toLocaleDateString(undefined, { timeZone: 'UTC', dateStyle: 'medium' })} (UTC).`;
    }
  } else if (response.status === 403)
    message = 'You do not have permission to use AI for this note or workspace.';
  else if (response.status === 429)
    message = 'Too many AI requests. Wait a minute, then try again.';
  else if (response.status === 503)
    message = 'AI is temporarily unavailable. Please try again later.';
  else if (response.status === 504) message = 'The AI request timed out. Try a shorter prompt.';
  else if (typeof body?.message === 'string') message = body.message;
  throw new AiRequestError(message, response.status, code);
}

export function createAiTransport(api: string, context: { workspaceId?: string; noteId?: string }) {
  return new DefaultChatTransport({
    api,
    body: context,
    headers: async () => ({ Authorization: `Bearer ${await fetchAccessTokenClientSide()}` }),
    fetch: fetchAi,
  });
}
