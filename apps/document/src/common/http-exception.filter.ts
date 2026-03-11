import {
  ArgumentsHost,
  Catch,
  ExceptionFilter,
  HttpException,
  HttpStatus,
  Logger,
} from '@nestjs/common';
import { HttpAdapterHost } from '@nestjs/core';
import { ModelError } from '@notopia-uit/api-document-nestjs-server';

/**
 * Global exception filter that transforms any thrown exception into a
 * ModelError response body, conforming to the OpenAPI error schema.
 *
 * NestJS HttpExceptions are mapped with their status code and message.
 * Unexpected errors fall back to 500 Internal Server Error.
 */
@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  private readonly logger = new Logger(HttpExceptionFilter.name);

  constructor(private readonly httpAdapterHost: HttpAdapterHost) {}

  catch(exception: unknown, host: ArgumentsHost): void {
    const { httpAdapter } = this.httpAdapterHost;
    const ctx = host.switchToHttp();

    let status: number;
    let code: string;
    let message: string;

    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const res = exception.getResponse();
      if (typeof res === 'string') {
        message = res;
      } else if (typeof res === 'object' && res !== null) {
        const obj = res as Record<string, unknown>;
        message = Array.isArray(obj['message'])
          ? (obj['message'] as string[]).join(', ')
          : String(obj['message'] ?? exception.message);
      } else {
        message = exception.message;
      }
      // reverse-lookup enum name: HttpStatus[404] => 'NOT_FOUND'
      code =
        (HttpStatus as unknown as Record<number, string>)[status] ??
        'HTTP_ERROR';
    } else {
      status = HttpStatus.INTERNAL_SERVER_ERROR;
      code = 'INTERNAL_SERVER_ERROR';
      message = 'An unexpected error occurred';
      this.logger.error(
        `Unhandled exception`,
        exception instanceof Error ? exception.stack : String(exception)
      );
    }

    const body: ModelError = { code, message };
    httpAdapter.reply(ctx.getResponse(), body, status);
  }
}
