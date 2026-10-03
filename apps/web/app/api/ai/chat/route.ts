import { NextResponse } from 'next/server';

// TODO(nest-ai-backend): replace this stub with a proxy to the Nest AI
// service. Contract: useChat with DefaultChatTransport POSTs UIMessages;
// respond with the AI SDK v6 UI message stream. The frontend
// `AssistantView` already speaks UIMessage, so only this route needs
// swapping (transport api stays `/api/ai/chat`).
export function POST() {
  return NextResponse.json(
    {
      code: 'AI_BACKEND_NOT_IMPLEMENTED',
      message: 'AI chat backend is not implemented yet.',
    },
    { status: 501 }
  );
}
