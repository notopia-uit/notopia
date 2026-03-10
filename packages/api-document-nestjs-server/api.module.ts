import { DynamicModule, Module, Provider } from '@nestjs/common';
import { ApiImplementations } from './api-implementations'
import { DocumentApi } from './api';
import { DocumentApiController } from './controllers';
import { RevisionApi } from './api';
import { RevisionApiController } from './controllers';

export type ApiModuleConfiguration = {
  /**
  * your Api implementations
  */
  apiImplementations: ApiImplementations,
  /**
  * additional Providers that may be used by your implementations
  */
  providers?: Provider[],
}

@Module({})
export class ApiModule {
  static forRoot(configuration: ApiModuleConfiguration): DynamicModule {
      const providers: Provider[] = [
        {
          provide: DocumentApi,
          useClass: configuration.apiImplementations.documentApi
        },
        {
          provide: RevisionApi,
          useClass: configuration.apiImplementations.revisionApi
        },
        ...(configuration.providers || []),
      ];

      return {
        module: ApiModule,
        controllers: [
          DocumentApiController,
          RevisionApiController,
        ],
        providers: [...providers],
        exports: [...providers]
      }
    }
}