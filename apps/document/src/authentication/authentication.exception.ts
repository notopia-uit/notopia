import { BadRequestException } from '@nestjs/common';

export class InvalidAuthenticationTokenException extends BadRequestException {
  override name = InvalidAuthenticationTokenException.name;

  readonly token: string;

  constructor(token: string, cause: unknown) {
    super(`Invalid authentication token: ${token}`, {
      cause: cause instanceof Error ? cause.message : String(cause),
      description: 'INVALID_AUTHENTICATION_TOKEN',
    });
    this.token = token;
  }
}
