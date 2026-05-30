'use client';

import { getWorkspaceEvents, type GetWorkspaceEventsResponse } from '@notopia-uit/api-gen';
import { createContext, useCallback, useContext, useEffect, useRef } from 'react';

type WorkspaceEventPayload = GetWorkspaceEventsResponse;

interface WorkspaceEventsContextValue {
  subscribe: (handler: (event: WorkspaceEventPayload) => void) => () => void;
}

const WorkspaceEventsContext = createContext<WorkspaceEventsContextValue | null>(null);

export function WorkspaceEventsProvider({
  children,
  workspaceId,
}: {
  children: React.ReactNode;
  workspaceId: string;
}) {
  const handlersRef = useRef<Set<(event: WorkspaceEventPayload) => void>>(new Set());

  const subscribe = useCallback((handler: (event: WorkspaceEventPayload) => void) => {
    handlersRef.current.add(handler);
    return () => {
      handlersRef.current.delete(handler);
    };
  }, []);

  useEffect(() => {
    const abortController = new AbortController();

    const connect = async () => {
      try {
        const result = getWorkspaceEvents({
          path: { workspaceId },
          signal: abortController.signal,
          onSseEvent: (raw) => {
            handlersRef.current.forEach((handler) => handler(raw.data));
          },
          onSseError: (error) => {
            console.error('Workspace events SSE error:', error);
          },
        });

        const { stream } = await result;
        for await (const _ of stream) {
          // consume stream to keep connection alive
        }
      } catch (error) {
        if (!abortController.signal.aborted) {
          console.error('Workspace events stream error:', error);
        }
      }
    };

    void connect();

    return () => {
      abortController.abort();
    };
  }, [workspaceId]);

  return (
    <WorkspaceEventsContext.Provider value={{ subscribe }}>
      {children}
    </WorkspaceEventsContext.Provider>
  );
}

export function useWorkspaceEvents() {
  const context = useContext(WorkspaceEventsContext);
  if (!context) {
    throw new Error('useWorkspaceEvents must be used within a WorkspaceEventsProvider');
  }
  return context;
}
