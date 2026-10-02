'use client';

import { createContext, useContext } from 'react';

interface NavigationContextValue {
  workspaceId: string;
  onNavigate: (href: string) => void;
}

const NavigationContext = createContext<NavigationContextValue | null>(null);

export function NavigationProvider({
  workspaceId,
  onNavigate,
  children,
}: NavigationContextValue & { children: React.ReactNode }) {
  return (
    <NavigationContext.Provider value={{ workspaceId, onNavigate }}>
      {children}
    </NavigationContext.Provider>
  );
}

export function useNavigationContext() {
  const ctx = useContext(NavigationContext);
  if (!ctx) {
    throw new Error('useNavigationContext must be used within a NavigationProvider');
  }
  return ctx;
}
