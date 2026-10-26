'use client';

import { createContext, useContext, type ReactNode } from 'react';

const AiContext = createContext<string | null>(null);

export function AiProvider({ aiApiUrl, children }: { aiApiUrl: string; children: ReactNode }) {
  return <AiContext.Provider value={aiApiUrl}>{children}</AiContext.Provider>;
}

export function useAiApiUrl() {
  const aiApiUrl = useContext(AiContext);
  if (!aiApiUrl) throw new Error('AI components require an AiProvider.');
  return aiApiUrl;
}
