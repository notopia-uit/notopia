import { MeilisearchError, type Task } from 'meilisearch';

export class SearchWorkerError extends Error {
  override name = 'SearchWorkerError';
}

export class MeiliError extends SearchWorkerError {
  override name = MeiliError.name;
  override cause: MeilisearchError;

  constructor(cause: MeilisearchError) {
    super(`Meilisearch error occurred`);

    this.cause = cause;
  }
}

export class MeiliTaskError extends SearchWorkerError {
  override name = MeiliTaskError.name;

  constructor(readonly task: Task) {
    super(
      `Meilisearch task ${task.uid} (${task.type}) ${task.status}: ${task.error?.message ?? ''}`
    );
  }
}
