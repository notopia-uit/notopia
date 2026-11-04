import { test } from 'node:test';

import { assertRecreateUpgrade } from '../../notopia-common-chart/tests/upgrade-assertions.ts';

test('document explicitly removes defaulted rollingUpdate in a strategic patch', () => {
  assertRecreateUpgrade('document');
});
