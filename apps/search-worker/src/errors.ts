import { MeiliSearchError } from 'meilisearch';

export class SearchWorkerError extends Error {
  override name = 'SearchWorkerError';
}

export class MeiliError extends SearchWorkerError {
  override name = 'MeiliError';
  override cause: MeiliSearchError;

  constructor(cause: MeiliSearchError) {
    super(`MeiliSearch error occurred`);

    this.cause = cause;
  }
}
