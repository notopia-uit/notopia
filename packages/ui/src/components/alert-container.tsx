'use client';

import React from 'react';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { ErrorAlert } from './error-alert';
import { SuccessAlert } from './success-alert';

export function AlertContainer() {
  const { alerts, removeAlert } = useAlert();

  return (
    <div className="fixed bottom-4 left-4 z-50 space-y-2 pointer-events-none">
      {alerts.map((alert) => (
        <div
          key={alert.id}
          className="animate-in fade-in slide-in-from-left-4 duration-300 pointer-events-auto"
          onClick={() => removeAlert(alert.id)}
        >
          {alert.type === 'success' && (
            <SuccessAlert title={alert.title} message={alert.message} />
          )}
          {alert.type === 'error' && (
            <ErrorAlert title={alert.title} message={alert.message} />
          )}
        </div>
      ))}
    </div>
  );
}
