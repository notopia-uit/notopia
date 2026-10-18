import { contentHash, mergeContributors, selectAuthor, shouldCommit } from './commit-policy';

describe('commit policy', () => {
  const now = new Date('2026-10-11T10:00:00Z');

  it('respects the min interval unless forced', () => {
    const last = new Date('2026-10-11T09:58:00Z');
    expect(
      shouldCommit({
        modified: true,
        lastCommittedAt: last,
        now,
        force: false,
        minIntervalSeconds: 300,
      })
    ).toBe(false);
    expect(
      shouldCommit({
        modified: true,
        lastCommittedAt: last,
        now,
        force: true,
        minIntervalSeconds: 300,
      })
    ).toBe(true);
    expect(
      shouldCommit({
        modified: true,
        lastCommittedAt: last,
        now,
        force: false,
        minIntervalSeconds: 60,
      })
    ).toBe(true);
    expect(
      shouldCommit({
        modified: false,
        lastCommittedAt: null,
        now,
        force: true,
        minIntervalSeconds: 0,
      })
    ).toBe(false);
  });

  it('hashes equal content equally', () => {
    expect(contentHash([{ type: 'paragraph' }])).toBe(contentHash([{ type: 'paragraph' }]));
  });

  it('keeps the most recent editor last without duplicates', () => {
    expect(mergeContributors(['a', 'b'], ['c', 'a'])).toEqual(['b', 'c', 'a']);
  });

  it('picks the explicit author, else the latest contributor, else the service account', () => {
    expect(
      selectAuthor({ authorId: 'http-user', contributors: ['a', 'b'], serviceAccountUserId: 'svc' })
    ).toBe('http-user');
    expect(selectAuthor({ contributors: ['a', 'b'], serviceAccountUserId: 'svc' })).toBe('b');
    expect(selectAuthor({ contributors: [], serviceAccountUserId: 'svc' })).toBe('svc');
  });
});
