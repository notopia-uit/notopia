import { ArgumentsHost, Catch, ExceptionFilter, HttpException, HttpStatus } from '@nestjs/common';
import { Request, Response } from 'express';
import { Logger } from 'nestjs-pino';

export interface ModelError {
  code: string;
  message: string;
  more_info?: string;
}

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  constructor(private readonly logger: Logger) {}

  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    const status =
      exception instanceof HttpException ? exception.getStatus() : HttpStatus.INTERNAL_SERVER_ERROR;

    let message = 'Internal server error';
    let code = 'INTERNAL_SERVER_ERROR';

    if (exception instanceof HttpException) {
      const res = exception.getResponse();
      message =
        typeof res === 'object' && res !== null && 'message' in res
          ? String((res as Record<string, unknown>).message)
          : exception.message;
      code = exception.name || 'HTTP_EXCEPTION';
    } else if (exception instanceof Error) {
      message = exception.message;
      code = exception.constructor.name;
    }

    this.logger.error(
      {
        err: exception,
        path: request.url,
        statusCode: status,
      },
      `Exception caught by filter: ${message}`
    );

    const errorResponse: ModelError = {
      code: code,
      message: message,
      // more_info: 'https://api.docs.com/errors/' + code, // Optional: logic for more_info
    };

    response.status(status).json(errorResponse);
  }
}
