export const abortReason = (signal: AbortSignal): string =>
  signal.reason instanceof Error ? signal.reason.name : String(signal.reason);
