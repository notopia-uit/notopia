'use client';

import { useMemo } from 'react';

import { useAiApiUrl } from '../contexts/ai-context';
import { createAiTransport } from '../lib/ai-transport';

export function useAiTransport({ workspaceId, noteId }: { workspaceId?: string; noteId?: string }) {
  const aiApiUrl = useAiApiUrl();
  return useMemo(
    () => createAiTransport(aiApiUrl, { workspaceId, noteId }),
    [aiApiUrl, workspaceId, noteId]
  );
}
