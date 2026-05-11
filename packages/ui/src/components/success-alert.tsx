import { CheckCircle2 } from 'lucide-react';

import { Alert, AlertDescription, AlertTitle } from './shadcn/alert';

export const SuccessAlert = ({ message, title }: { message: string; title: string }) => (
  <Alert className="border-emerald-900 bg-emerald-950/50 text-emerald-400">
    <CheckCircle2 className="size-4 stroke-emerald-500" />
    <AlertTitle className="text-emerald-500">{title}</AlertTitle>
    <AlertDescription className="text-emerald-500/90">{message}</AlertDescription>
  </Alert>
);
