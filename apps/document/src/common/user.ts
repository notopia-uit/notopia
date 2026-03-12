export type User = {
  id: string;
  email: string;
  groups?: string[];
  roles?: string[];
};

export function unmarshalHeader(headerValue: string | undefined): string[] {
  if (!headerValue) return [];
  const s = headerValue.replace(/^\[|\]$/g, '').trim();
  return s === '' ? [] : s.split(/\s+/);
}
