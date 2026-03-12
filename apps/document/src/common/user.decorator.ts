import { createParamDecorator, ExecutionContext } from '@nestjs/common';

import { User } from './user';

export const ReqUser = createParamDecorator(
  (_: unknown, ctx: ExecutionContext): User => {
    const request = ctx.switchToHttp().getRequest();
    return request.user;
  }
);
