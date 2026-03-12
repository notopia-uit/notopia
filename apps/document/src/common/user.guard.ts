import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';

import { User } from './user';

@Injectable()
export class UserGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest();
    const id = request.headers['x-forwarded-id'] as string;
    if (!id) {
      throw new UnauthorizedException('Missing Gateway Headers');
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

    request['user'] = user;

    return true;
  }
}
