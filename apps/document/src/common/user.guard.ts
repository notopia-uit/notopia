import { CanActivate, ExecutionContext, Injectable, UnauthorizedException } from '@nestjs/common';
import { WsException } from '@nestjs/websockets';

import { User } from './user';

@Injectable()
export abstract class UserGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const type = context.getType();
    let headers: Record<string, unknown> | undefined;

    switch (type) {
      case 'http': {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
        headers = context.switchToHttp().getRequest().headers as Record<string, unknown>;
        break;
      }
      case 'ws': {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
        headers = context.switchToWs().getClient().handshake as Record<string, unknown>;
        break;
      }
    }

    if (!headers) {
      this.throwException('Invalid request context');
      return false;
    }

    const id = headers['x-forwarded-id'] as string | undefined;

    if (!id) {
      this.throwException('Missing Gateway Headers');
      return false;
    }

    const parseHeaderList = (header?: string): string[] => {
      if (!header) return [];
      const cleaned = header.replace(/^\[|\]$/g, '').trim();
      return cleaned === '' ? [] : cleaned.split(/\s+/);
    };

    const groupsList = parseHeaderList(headers['x-forwarded-groups'] as string | undefined);
    const rolesList = parseHeaderList(headers['x-forwarded-roles'] as string | undefined);

    const user: User = {
      id,
      email: (headers['x-forwarded-email'] as string | undefined) || '',
      ...(groupsList.length > 0 && { groups: groupsList }),
      ...(rolesList.length > 0 && { roles: rolesList }),
    };

    switch (type) {
      case 'http': {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
        context.switchToHttp().getRequest().user = user;
        break;
      }
      case 'ws': {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
        context.switchToWs().getClient().user = user;
        break;
      }
    }

    return true;
  }

  protected abstract throwException(message: string): void;
}

@Injectable()
export class HttpUserGuard extends UserGuard {
  protected throwException(message: string): void {
    throw new UnauthorizedException(message);
  }
}

@Injectable()
export class WsUserGuard extends UserGuard {
  protected throwException(message: string): void {
    throw new WsException(message);
  }
}
