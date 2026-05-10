import { Module } from '@nestjs/common';

import { AuthenticationService } from '#/authentication/authentication.service';

@Module({
  providers: [AuthenticationService],
  exports: [AuthenticationService],
})
export class AuthenticationModule {}
