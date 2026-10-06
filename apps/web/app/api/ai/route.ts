import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export const maxDuration = 30;

const AI_SERVICE_URL = process.env.NOTOPIA_AI_URL ?? 'http://api.notopia.localhost/ai';

export async function POST(req: NextRequest) {
  const authorization = req.headers.get('authorization');
  if (!authorization) {
    return NextResponse.json(
      { code: 'UNAUTHORIZED', message: 'Missing authorization header.' },
      { status: 401 }
    );
  }

  const upstream = await fetch(AI_SERVICE_URL, {
    method: 'POST',
    headers: {
      'content-type': req.headers.get('content-type') ?? 'application/json',
      authorization,
    },
    body: req.body,
    duplex: 'half',
  } as RequestInit);

  const contentType = upstream.headers.get('content-type') ?? 'text/event-stream';
  return new Response(upstream.body, {
    status: upstream.status,
    headers: {
      'content-type': contentType,
      'cache-control': 'no-cache',
    },
  });
}
