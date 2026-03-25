import { User } from './user';
import { ExecutionContext, createParamDecorator } from '@nestjs/common';

export const ReqUser = createParamDecorator(
  (_: unknown, ctx: ExecutionContext): User => {
    const request = ctx.switchToHttp().getRequest();
    return request.user;
  }
);
