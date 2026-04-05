import { User } from './user';
import { ExecutionContext, createParamDecorator } from '@nestjs/common';

export const ReqUser = createParamDecorator(
  (_: unknown, ctx: ExecutionContext): User => {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    return ctx.switchToHttp().getRequest().user as User;
  }
);
