import { planRevisionDeletes } from './revision-purge';

const d = (iso: string) => new Date(iso);

describe('planRevisionDeletes', () => {
  const now = d('2026-10-11T00:00:00Z');

  it('deletes revisions older than retention but always keeps the latest', () => {
    const revisions = [
      { id: 'r1', documentId: 'a', createdAt: d('2026-09-01T00:00:00Z') },
      { id: 'r2', documentId: 'a', createdAt: d('2026-09-02T00:00:00Z') },
      { id: 'r3', documentId: 'b', createdAt: d('2026-09-01T00:00:00Z') },
    ];
    const retention = new Map([
      ['a', 7],
      ['b', 7],
    ]);
    expect(planRevisionDeletes(revisions, retention, now)).toEqual(['r1']);
  });

  it('skips unlimited and unknown documents', () => {
    const revisions = [
      { id: 'r1', documentId: 'a', createdAt: d('2020-01-01T00:00:00Z') },
      { id: 'r2', documentId: 'a', createdAt: d('2020-01-02T00:00:00Z') },
      { id: 'r3', documentId: 'gone', createdAt: d('2020-01-01T00:00:00Z') },
      { id: 'r4', documentId: 'gone', createdAt: d('2020-01-02T00:00:00Z') },
    ];
    expect(planRevisionDeletes(revisions, new Map([['a', -1]]), now)).toEqual([]);
  });
});
