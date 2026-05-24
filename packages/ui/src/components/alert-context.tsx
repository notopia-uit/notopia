'use client';

import React, { createContext, useCallback, useState, useRef } from 'react';

export type AlertType = 'success' | 'error';

export interface Alert {
  id: string;
  type: AlertType;
  title: string;
  message: string;
  duration?: number;
}

export interface AlertContextType {
  alerts: Alert[];
  showAlert: (alert: Omit<Alert, 'id'>) => void;
  removeAlert: (id: string) => void;
}

export const AlertContext = createContext<AlertContextType | undefined>(undefined);

export function AlertProvider({ children }: { children: React.ReactNode }) {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const idRef = useRef(0);
  const timeoutRef = useRef<Record<string, NodeJS.Timeout>>({});

  const removeAlert = useCallback((id: string) => {
    if (timeoutRef.current[id]) {
      clearTimeout(timeoutRef.current[id]);
      delete timeoutRef.current[id];
    }
    setAlerts((prev) => prev.filter((alert) => alert.id !== id));
  }, []);

  const showAlert = useCallback((alert: Omit<Alert, 'id'>) => {
    const id = `alert-${idRef.current++}`;
    const duration = alert.duration ?? 5000;

    setAlerts((prev) => [...prev, { ...alert, id }]);

    if (duration > 0) {
      timeoutRef.current[id] = setTimeout(() => {
        removeAlert(id);
      }, duration);
    }
  }, [removeAlert]);

  const value: AlertContextType = {
    alerts,
    showAlert,
    removeAlert,
  };

  return <AlertContext.Provider value={value}>{children}</AlertContext.Provider>;
}
