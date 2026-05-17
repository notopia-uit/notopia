import { Module } from '@nestjs/common';

import { AuthenticationService } from '#/authentication';

@Module({
  providers: [AuthenticationService],
  exports: [AuthenticationService],
})
export class AuthenticationModule {}
