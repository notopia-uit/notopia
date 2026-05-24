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
  apiKey?: string; // Optional - can be passed from parent after fetching
}

export function MeilisearchProvider({ children, host, apiKey }: MeilisearchProviderProps) {
  const client = useMemo(() => {
    if (!host) {
      console.warn('Meilisearch provider: host is missing');
      return null;
    }

    // If apiKey is not provided, return null (client will be initialized later)
    if (!apiKey) {
      console.warn('Meilisearch provider: apiKey is not available yet');
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
