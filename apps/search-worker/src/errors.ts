import { MeilisearchError } from 'meilisearch';

export class SearchWorkerError extends Error {
  override name = 'SearchWorkerError';
}

export class MeiliError extends SearchWorkerError {
  override name = 'MeiliError';
  override cause: MeilisearchError;

  constructor(cause: MeilisearchError) {
    super(`Meilisearch error occurred`);

    this.cause = cause;
  }
}
