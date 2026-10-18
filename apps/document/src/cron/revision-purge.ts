import { isUnlimited } from '../billing';

const DAY_MS = 86_400_000;

export interface RevisionRef {
  id: string;
  documentId: string;
  createdAt: Date;
}

export const planRevisionDeletes = (
  revisions: RevisionRef[],
  retentionDaysByDocument: Map<string, number>,
  now: Date
): string[] => {
  const revisionsByDocument = new Map<string, RevisionRef[]>();
  for (const revision of revisions) {
    const list = revisionsByDocument.get(revision.documentId) ?? [];
    list.push(revision);
    revisionsByDocument.set(revision.documentId, list);
  }
  const deletes: string[] = [];
  for (const [documentId, list] of revisionsByDocument) {
    const days = retentionDaysByDocument.get(documentId);
    if (days === undefined || isUnlimited(days)) continue;
    const cutoff = now.getTime() - days * DAY_MS;
    const newestFirst = [...list].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    for (const revision of newestFirst.slice(1)) {
      if (revision.createdAt.getTime() < cutoff) deletes.push(revision.id);
    }
  }
  return deletes;
};
