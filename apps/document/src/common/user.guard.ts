import { User } from './user';
import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';
import { WsException } from '@nestjs/websockets';

@Injectable()
export abstract class UserGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const type = context.getType();
    let request: any;

    switch (type) {
      case 'http':
        request = context.switchToHttp().getRequest();
        break;
      case 'ws':
        request = context.switchToWs().getClient().handshake;
        break;
    }

    const id = request.headers['x-forwarded-id'] as string;

    if (!id) {
      this.throwException('Missing Gateway Headers');
    }

    const parseHeaderList = (header?: string): string[] => {
      if (!header) return [];
      const cleaned = header.replace(/^\[|\]$/g, '').trim();
      return cleaned === '' ? [] : cleaned.split(/\s+/);
    };

    const user: User = {
      id,
      email: request.headers['x-forwarded-email'] as string,
      groups: parseHeaderList(request.headers['x-forwarded-groups'] as string),
      roles: parseHeaderList(request.headers['x-forwarded-roles'] as string),
    };

    switch (type) {
      case 'http':
        request.user = user;
        break;
      case 'ws':
        context.switchToWs().getClient().user = user;
        break;
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
