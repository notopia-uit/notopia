import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

test('renders the expected hooks without Istio', () => {
  const output = execFileSync(
    'helm',
    ['template', 'test', 'deploy/k8s/charts/search-worker', '--set', 'istio.enabled=false'],
    { encoding: 'utf8' }
  );
  assert.equal((output.match(/^kind: Job$/gm) ?? []).length, 1);
  assert.doesNotMatch(output, /security\.istio\.io/);
  assert.match(output, /helm.sh\/hook: "post-install,post-upgrade"/);
  assert.match(output, /- seed.cjs/);
});

test('seed and deployment receive the same Redis environment', () => {
  const output = execFileSync(
    'helm',
    [
      'template',
      'test',
      'deploy/k8s/charts/search-worker',
      '--set',
      'istio.enabled=false',
      '--set',
      'redis.urlSecret.name=redis',
    ],
    { encoding: 'utf8' }
  );
  assert.equal((output.match(/name: NOTOPIA_SEARCH_WORKER_REDIS_URL/g) ?? []).length, 2);
});
