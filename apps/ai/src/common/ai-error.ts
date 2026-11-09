import { status } from '@grpc/grpc-js';
import { ArgumentsHost, Catch, ExceptionFilter, HttpException, Logger } from '@nestjs/common';
import type { Response } from 'express';

import { isGrpcError } from './grpc';

export class AiError extends HttpException {
  constructor(
    statusCode: number,
    code: string,
    message: string,
    details?: Record<string, unknown>
  ) {
    super({ code, message, ...details }, statusCode);
  }
}

export function publicError(error: unknown): { status: number; body: Record<string, unknown> } {
  if (error instanceof AiError) {
    return { status: error.getStatus(), body: error.getResponse() as Record<string, unknown> };
  }
  if (error instanceof HttpException) {
    return { status: error.getStatus(), body: { code: 'requestFailed', message: error.message } };
  }
  if (isGrpcError(error)) {
    if (error.code === status.PERMISSION_DENIED) {
      return { status: 403, body: { code: 'forbidden', message: 'Access denied.' } };
    }
    if (error.code === status.NOT_FOUND) {
      return { status: 404, body: { code: 'notFound', message: 'Resource not found.' } };
    }
    return {
      status: 503,
      body: { code: 'serviceUnavailable', message: 'A required service is unavailable.' },
    };
  }
  if (typeof error === 'object' && error !== null && 'type' in error) {
    if (error.type === 'entity.too.large')
      return {
        status: 413,
        body: { code: 'requestTooLarge', message: 'AI requests must be at most 500 KB.' },
      };
    if (error.type === 'entity.parse.failed')
      return { status: 400, body: { code: 'invalidAiRequest', message: 'Invalid JSON request.' } };
  }
  return { status: 500, body: { code: 'aiError', message: 'The AI request failed.' } };
}

@Catch()
export class AiExceptionFilter implements ExceptionFilter {
  private readonly logger = new Logger(AiExceptionFilter.name);

  catch(error: unknown, host: ArgumentsHost) {
    const response = host.switchToHttp().getResponse<Response>();
    const mapped = publicError(error);
    if (mapped.status >= 500) this.logger.error({ err: error }, 'AI request failed');
    if (!response.headersSent) {
      if (mapped.status === 429) response.setHeader('Retry-After', '60');
      response.status(mapped.status).json(mapped.body);
    } else {
      response.end();
    }
  }
}
