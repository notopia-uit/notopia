import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import path from 'node:path';
import { test } from 'node:test';

type Resource = {
  kind: string;
  env: Array<{
    name: string;
    value?: string;
    valueFrom?: { secretKeyRef?: { name: string } };
  }>;
  envFrom: Array<{ secretRef?: { name: string } }>;
};

const fixture = path.join(import.meta.dirname, 'fixture');

for (const runtime of ['go', 'node']) {
  test(`hook Job and Deployment render identical environment and secrets for ${runtime}`, () => {
    const manifests = execFileSync(
      'helm',
      ['template', 'fixture', fixture, '--set', `runtime=${runtime}`],
      { encoding: 'utf8' }
    );
    const parsed = spawnSync(
      'yq',
      [
        'eval-all',
        '-o=json',
        '-I=0',
        '[select(.kind == "Deployment" or .kind == "Job") | {"kind": .kind, "env": .spec.template.spec.containers[0].env, "envFrom": .spec.template.spec.containers[0].envFrom}]',
        '-',
      ],
      { input: manifests, encoding: 'utf8' }
    );
    assert.equal(parsed.status, 0, parsed.stderr);
    const resources: Resource[] = JSON.parse(parsed.stdout);
    const deployment = resources.find((resource) => resource.kind === 'Deployment');
    const job = resources.find((resource) => resource.kind === 'Job');
    assert.ok(deployment && job, 'both the Deployment and hook Job must render');
    assert.deepEqual(job.env, deployment.env);
    assert.deepEqual(job.envFrom, deployment.envFrom);
    assert.ok(deployment.env.some((entry) => entry.name === 'FOO' && entry.value === 'bar'));
    assert.ok(
      deployment.env.some(
        (entry) => entry.name === 'SECRET' && entry.valueFrom?.secretKeyRef?.name === 'external'
      )
    );
    assert.ok(deployment.envFrom.some((entry) => entry.secretRef?.name === 'external'));
  });
}
