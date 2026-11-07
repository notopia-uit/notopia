export const AUTHORSHIP_VALUES = ['human', 'ai'] as const;

export type Authorship = (typeof AUTHORSHIP_VALUES)[number];

export function validateAuthorship(value: unknown, path: string): void {
  if (value === undefined) return;

  if (
    !Array.isArray(value) ||
    value.length === 0 ||
    value.some(
      (item: unknown) => typeof item !== 'string' || !AUTHORSHIP_VALUES.includes(item as Authorship)
    )
  ) {
    throw new Error(
      `${path}: authorship must be a non-empty list containing only: ${AUTHORSHIP_VALUES.join(', ')}`
    );
  }
}
