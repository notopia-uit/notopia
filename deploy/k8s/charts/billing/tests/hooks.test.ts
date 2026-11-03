import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

test('renders only migration and seed Jobs without Istio', () => {
  const output = execFileSync(
    'helm',
    [
      'template',
      'test',
      'deploy/k8s/charts/billing',
      '--set',
      'istio.enabled=false',
      '--set',
      'redis.addr=redis:6379',
      '--set',
      'redis.passwordSecret.name=redis',
    ],
    { encoding: 'utf8' }
  );
  assert.equal((output.match(/^kind: Job$/gm) ?? []).length, 2);
  assert.doesNotMatch(output, /security\.istio\.io/);
});
