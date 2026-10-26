import { createHash } from 'node:crypto';

export const REVISION_TRIGGERS = ['manual', 'auto', 'sweeper'] as const;
export type RevisionTrigger = (typeof REVISION_TRIGGERS)[number];

export const shouldCommit = (p: {
  modified: boolean;
  lastCommittedAt: Date | null;
  now: Date;
  force: boolean;
  minIntervalSeconds: number;
}): boolean => {
  if (!p.modified) return false;
  if (p.force || !p.lastCommittedAt) return true;
  return (p.now.getTime() - p.lastCommittedAt.getTime()) / 1000 >= p.minIntervalSeconds;
};

export const contentHash = (blocks: unknown): string =>
  createHash('sha256').update(JSON.stringify(blocks)).digest('hex');

export const mergeContributors = (existing: string[], incoming: string[]): string[] => {
  const kept = existing.filter((id) => !incoming.includes(id));
  return [...kept, ...incoming];
};

export const selectAuthor = (p: {
  authorId?: string;
  contributors: string[];
  serviceAccountUserId: string;
}): string => p.authorId ?? p.contributors.at(-1) ?? p.serviceAccountUserId;
