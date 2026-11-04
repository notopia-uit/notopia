import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';

test('renders only the seed Job without Istio', () => {
  const output = execFileSync(
    'helm',
    ['template', 'test', 'deploy/k8s/charts/authorization', '--set', 'istio.enabled=false'],
    {
      encoding: 'utf8',
    }
  );
  assert.equal((output.match(/^kind: Job$/gm) ?? []).length, 1);
  assert.doesNotMatch(output, /security\.istio\.io/);
});

test('documents the external seeding and upgrade maintenance prerequisite', () => {
  const readme = readFileSync('deploy/k8s/charts/authorization/README.md', 'utf8');
  assert.match(readme, /before every install and upgrade/i);
  assert.match(readme, /stop all authorization pods/i);
  assert.match(readme, /same immutable image/i);
  assert.match(readme, /does not verify/i);
});
