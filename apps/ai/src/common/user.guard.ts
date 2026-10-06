import { CanActivate, ExecutionContext, Injectable, UnauthorizedException } from '@nestjs/common';

import { User } from './user';

@Injectable()
export abstract class UserGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const headers = context.switchToHttp().getRequest().headers as Record<string, unknown>;

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

    context.switchToHttp().getRequest().user = user;

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
