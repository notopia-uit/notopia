'use client';

import { Meilisearch } from 'meilisearch';
import { createContext, useContext, useMemo } from 'react';

interface MeilisearchContextType {
  client: Meilisearch | null;
}

const MeilisearchContext = createContext<MeilisearchContextType>({
  client: null,
});

interface MeilisearchProviderProps {
  children: React.ReactNode;
  host: string;
  apiKey: string;
}

export function MeilisearchProvider({ children, host, apiKey }: MeilisearchProviderProps) {
  const client = useMemo(() => {
    if (!host || !apiKey) {
      console.warn('Meilisearch provider: host or apiKey is missing');
      return null;
    }

    return new Meilisearch({
      host,
      apiKey,
    });
  }, [host, apiKey]);

  return (
    <MeilisearchContext.Provider value={{ client }}>
      {children}
    </MeilisearchContext.Provider>
  );
}

export function useMeilisearch() {
  const context = useContext(MeilisearchContext);
  if (!context) {
    throw new Error('useMeilisearch must be used within MeilisearchProvider');
  }
  return context.client;
}
