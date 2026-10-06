export const DEFAULT_BLOCKNOTE_AI_API_URL = '/api/ai';

export function getBlockNoteAiApiUrl(override?: string): string {
  return override ?? DEFAULT_BLOCKNOTE_AI_API_URL;
}
