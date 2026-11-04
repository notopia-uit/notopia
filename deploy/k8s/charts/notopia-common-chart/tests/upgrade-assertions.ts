import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';

export function assertRecreateUpgrade(chart: string) {
  const rendered = execFileSync(
    'helm',
    [
      'template',
      'test',
      `deploy/k8s/charts/${chart}`,
      '--show-only',
      'templates/deployment.yaml',
      '--set',
      'istio.enabled=false',
    ],
    { encoding: 'utf8' }
  );
  const deployment = JSON.parse(
    execFileSync('yq', ['-o=json'], { input: rendered, encoding: 'utf8' })
  );
  const strategy = deployment.spec.strategy;
  assert.equal(strategy.type, 'Recreate');
  assert.ok(Object.hasOwn(strategy, 'rollingUpdate'));
  assert.equal(strategy.rollingUpdate, null);
  const live = {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: { name: 'test' },
    spec: {
      strategy: {
        type: 'RollingUpdate',
        rollingUpdate: { maxSurge: '25%', maxUnavailable: '25%' },
      },
    },
  };
  const result = execFileSync(
    'kubectl',
    [
      'patch',
      '--local',
      '-f',
      '-',
      '--type=strategic',
      '-p',
      JSON.stringify({ spec: { strategy } }),
      '-o',
      'json',
    ],
    { input: JSON.stringify(live), encoding: 'utf8' }
  );
  assert.deepEqual(JSON.parse(result).spec.strategy, { type: 'Recreate' });
}
