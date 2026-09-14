// Base error
export class AppError extends Error {
  readonly code: string;
  readonly moreInfo?: string;
  readonly statusCode?: number;

  constructor(
    message: string,
    opts: { code?: string; moreInfo?: string; statusCode?: number; cause?: unknown } = {}
  ) {
    super(message, { cause: opts.cause });
    this.name = 'AppError';
    this.code = opts.code ?? 'UNKNOWN';
    this.moreInfo = opts.moreInfo;
    this.statusCode = opts.statusCode;
  }
}

// Error from NoteService
export class NoteServiceError extends AppError {
  override readonly name = 'NoteServiceError';

  constructor(
    message: string,
    opts: { code?: string; moreInfo?: string; statusCode?: number; cause?: unknown } = {}
  ) {
    super(message, opts);
  }
}

// Error from DocumentService
export class DocumentServiceError extends AppError {
  override readonly name = 'DocumentServiceError';

  constructor(
    message: string,
    opts: { code?: string; moreInfo?: string; statusCode?: number; cause?: unknown } = {}
  ) {
    super(message, opts);
  }
}

// Unauthorized error
export class UnauthorizedError extends AppError {
  override readonly name = 'UnauthorizedError';

  constructor(message = 'Unauthorized', opts: { cause?: unknown } = {}) {
    super(message, { code: 'UNAUTHORIZED', statusCode: 401, ...opts });
  }
}

// Notfound error
export class NotFoundError extends AppError {
  override readonly name = 'NotFoundError';

  constructor(message = 'Not found', opts: { cause?: unknown } = {}) {
    super(message, { code: 'NOT_FOUND', statusCode: 404, ...opts });
  }
}

interface GeneratedError {
  code?: string;
  message?: string;
  more_info?: string;
}

function isGeneratedError(obj: unknown): obj is GeneratedError {
  return typeof obj === 'object' && obj !== null && 'message' in obj;
}

export function toAppError(err: unknown, fallback?: string): AppError {
  if (err instanceof AppError) return err;

  if (err instanceof Error) {
    return new AppError(err.message, { cause: err });
  }

  if (typeof err === 'string') {
    return new AppError(err);
  }

  if (isGeneratedError(err)) {
    return new AppError(err.message ?? 'Unknown service error', {
      code: err.code ?? 'SERVICE_ERROR',
      moreInfo: err.more_info,
    });
  }

  return new AppError(fallback ?? 'An unexpected error occurred');
}

export function is404(err: unknown): boolean {
  if (err instanceof NotFoundError) return true;
  if (err instanceof AppError && err.statusCode === 404) return true;
  if (isGeneratedError(err) && err.code === 'NOT_FOUND') return true;
  return false;
}
