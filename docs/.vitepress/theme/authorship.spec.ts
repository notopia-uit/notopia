import { describe, expect, it } from 'vitest';

import { validateAuthorship } from './authorship';

describe('validateAuthorship', () => {
  it.each([undefined, ['human'], ['ai'], ['human', 'ai']])('accepts %o', (value) => {
    expect(() => validateAuthorship(value, 'page.md')).not.toThrow();
  });

  it.each([null, 'human', [], ['humna'], ['human', 1]])('rejects %o', (value) => {
    expect(() => validateAuthorship(value, 'page.md')).toThrow(
      'page.md: authorship must be a non-empty list containing only: human, ai'
    );
  });
});
