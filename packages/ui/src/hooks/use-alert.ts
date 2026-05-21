'use client';

import { useContext } from 'react';
import { AlertContext, type AlertContextType } from '../components/alert-context';

export function useAlert(): AlertContextType {
  const context = useContext(AlertContext);
  if (!context) {
    throw new Error('useAlert must be used within AlertProvider');
  }
  return context;
}
