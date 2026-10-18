import { ServiceUnavailableException } from '@nestjs/common';

export class BillingUnavailableException extends ServiceUnavailableException {
  constructor(cause: unknown) {
    super('Billing service unavailable', { cause });
  }
}
