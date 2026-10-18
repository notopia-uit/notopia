import { Injectable } from '@nestjs/common';

import { mergeContributors } from '../document/commit-policy';

@Injectable()
export class ContributorTracker {
  private readonly pending = new Map<string, string[]>();

  add(documentId: string, userId: string): void {
    this.pending.set(documentId, mergeContributors(this.pending.get(documentId) ?? [], [userId]));
  }

  peek(documentId: string): string[] {
    return [...(this.pending.get(documentId) ?? [])];
  }

  clear(documentId: string, flushed: string[]): void {
    // keep ids added while the flush was running
    const rest = (this.pending.get(documentId) ?? []).filter((id) => !flushed.includes(id));
    if (rest.length === 0) this.pending.delete(documentId);
    else this.pending.set(documentId, rest);
  }

  forget(documentId: string): void {
    this.pending.delete(documentId);
  }
}
