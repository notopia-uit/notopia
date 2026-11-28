import spec from '@notopia-uit/api/openapi' with { type: 'json' };

import { getRestReference } from '../../../../.vitepress/rest-reference';

export default {
  paths: () =>
    getRestReference(spec).operations.map(({ operationId, title }) => ({
      params: { operationId, title },
    })),
};
