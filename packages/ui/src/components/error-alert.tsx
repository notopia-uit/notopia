import { AlertCircleIcon } from 'lucide-react';

import { Alert, AlertDescription, AlertTitle } from './shadcn/alert';

export const ErrorAlert = ({ message, title }: { message: string; title: string }) => (
  <Alert variant="destructive" className="max-w-md">
    <AlertCircleIcon />
    <AlertTitle>{title}</AlertTitle>
    <AlertDescription>{message}</AlertDescription>
  </Alert>
);
