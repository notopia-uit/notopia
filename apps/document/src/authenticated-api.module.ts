import { Controller, type DynamicModule, type Type, UseGuards } from '@nestjs/common';
import { ApiModule, type ApiModuleConfiguration } from '@notopia-uit/api-document-nestjs-server';

import { HttpUserGuard } from './common';

export function authenticatedApiModule(configuration: ApiModuleConfiguration): DynamicModule {
  const apiModule = ApiModule.forRoot(configuration);
  return {
    ...apiModule,
    controllers: apiModule.controllers?.map((controller) => {
      const BaseController = controller as Type;
      @Controller()
      @UseGuards(HttpUserGuard)
      class AuthenticatedController extends BaseController {}
      return AuthenticatedController;
    }),
  };
}
