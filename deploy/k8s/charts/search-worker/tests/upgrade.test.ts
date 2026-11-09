import { test } from 'node:test';

import { assertRecreateUpgrade } from '../../notopia-common-chart/tests/upgrade-assertions.ts';

test('search-worker explicitly removes defaulted rollingUpdate in a strategic patch', () => {
  assertRecreateUpgrade('search-worker');
});
