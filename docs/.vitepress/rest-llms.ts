import spec from '@notopia-uit/api/openapi' with { type: 'json' };
import type { LlmsFilesOptions } from 'vitepress-openapi';

import { REST_OPERATION_ROUTE } from './rest-reference';

export const restLlmsOptions = {
  spec: spec as unknown as LlmsFilesOptions['spec'],
  operationRoute: REST_OPERATION_ROUTE,
  introductionRoute: 'api/rest',
  llmsTxt: false,
} satisfies LlmsFilesOptions;
