import { NextResponse } from 'next/server';

export function POST() {
  return NextResponse.json(
    {
      code: 'AI_BACKEND_NOT_IMPLEMENTED',
      message: 'BlockNote AI backend is not implemented yet.',
    },
    { status: 501 }
  );
}
