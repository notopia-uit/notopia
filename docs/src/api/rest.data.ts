import spec from '@notopia-uit/api/openapi' with { type: 'json' };
import { defineLoader } from 'vitepress';

import { type RestReference, getRestReference } from '../../.vitepress/rest-reference';

type RestGroups = RestReference['groups'];

export declare const data: RestGroups;

export default defineLoader<RestGroups>({
  load: () => getRestReference(spec).groups,
});
