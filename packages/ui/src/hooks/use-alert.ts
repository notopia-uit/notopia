import { useState, useRef, useEffect, useCallback } from 'react';

type AlertType = 'success' | 'error';

interface AlertState {
  type: AlertType;
  title: string;
  message: string;
}

export const useAlert = (duration = 3000) => {
  const [alert, setAlert] = useState<AlertState | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const showAlert = useCallback(
    (type: AlertType, title: string, message: string) => {
      setAlert({ type, title, message });
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => {
        setAlert(null);
      }, duration);
    },
    [duration]
  );
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  return { alert, showAlert };
};
