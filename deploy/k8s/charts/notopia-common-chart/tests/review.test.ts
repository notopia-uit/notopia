import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';

const fixture = path.join(import.meta.dirname, 'fixture');
const project = JSON.parse(readFileSync(path.join(import.meta.dirname, '../project.json'), 'utf8'));

type Resource = {
  kind: string;
  metadata: { name: string };
  spec: {
    template: {
      spec: {
        containers: Array<{
          env: Array<{ name: string; value?: string }>;
          resources: { limits: { memory: string } };
        }>;
      };
    };
  };
};

function render(values: Record<string, unknown> = {}) {
  return spawnSync('helm', ['template', 'fixture', fixture, '-f', '-'], {
    input: JSON.stringify(values),
    encoding: 'utf8',
  });
}

function resources(output: string): Resource[] {
  return JSON.parse(
    execFileSync('yq', ['eval-all', '-o=json', '-I=0', '[select(.kind != null)]', '-'], {
      input: output,
      encoding: 'utf8',
    })
  );
}

test('lint pipeline fails when Helm rendering fails', () => {
  const command: string = project.targets.lint.options.commands.find((item: string) =>
    item.includes('kubeconform')
  );
  assert.match(command, /bash -o pipefail -c/);
  const pipeline = command.split("bash -o pipefail -c '")[1].slice(0, -1);
  const broken = pipeline.replace(
    'helm template fixture',
    'helm template --set resources.limits.memory=1.5Gi fixture'
  );
  const result = spawnSync('bash', ['-o', 'pipefail', '-c', broken], { encoding: 'utf8' });
  assert.notEqual(result.status, 0, `render failure was hidden: ${result.stderr}`);
  assert.match(result.stderr, /unsupported or missing memory limit/);
});

for (const [runtime, name, expected] of [
  ['go', 'GOMEMLIMIT', '120795955'],
  ['node', 'NODE_OPTIONS', '--max-old-space-size=96'],
]) {
  test(`${runtime} hook memory env uses its resource override`, () => {
    const result = render({
      runtime,
      hooks: { migrate: { resources: { limits: { memory: '128Mi' } } } },
      env: { EXTRA: 'included' },
    });
    assert.equal(result.status, 0, result.stderr);
    const job = resources(result.stdout).find(
      (item) => item.kind === 'Job' && item.metadata.name.endsWith('-migrate')
    );
    assert.ok(job);
    const container = job.spec.template.spec.containers[0];
    assert.equal(container.resources.limits.memory, '128Mi');
    assert.equal(container.env.find((item) => item.name === name)?.value, expected);
    assert.equal(container.env.find((item) => item.name === 'EXTRA')?.value, 'included');
    const deployment = resources(result.stdout).find((item) => item.kind === 'Deployment');
    assert.ok(deployment);
    assert.equal(
      deployment.spec.template.spec.containers[0].env.find((item) => item.name === name)?.value,
      runtime === 'go' ? '241591910' : '--max-old-space-size=192'
    );
  });
}

test('long names retain distinct migrate and seed suffixes', () => {
  const result = render({ fullnameOverride: 'a'.repeat(63) });
  assert.equal(result.status, 0, result.stderr);
  const names = resources(result.stdout)
    .filter((item) => item.kind === 'Job')
    .map((job) => job.metadata.name);
  assert.equal(names.length, 2);
  assert.equal(new Set(names).size, 2);
  assert.ok(names.some((name) => name.endsWith('-migrate') && name.length <= 63));
  assert.ok(names.some((name) => name.endsWith('-seed') && name.length <= 63));
});

for (const [name, istio] of [
  ['empty principal', { enabled: true, allowedPrincipals: [''], ports: ['8080'] }],
  ['whitespace principal', { enabled: true, allowedPrincipals: ['   '], ports: ['8080'] }],
  [
    'principal containing whitespace',
    { enabled: true, allowedPrincipals: ['valid principal'], ports: ['8080'] },
  ],
  [
    'blank principal among valid principals',
    { enabled: true, allowedPrincipals: ['cluster.local/ns/default/sa/edge', ''], ports: ['8080'] },
  ],
  [
    'empty port',
    { enabled: true, allowedPrincipals: ['cluster.local/ns/default/sa/edge'], ports: [''] },
  ],
  [
    'whitespace port',
    { enabled: true, allowedPrincipals: ['cluster.local/ns/default/sa/edge'], ports: ['   '] },
  ],
  [
    'out of range port',
    { enabled: true, allowedPrincipals: ['cluster.local/ns/default/sa/edge'], ports: ['65536'] },
  ],
  [
    'zero port',
    { enabled: true, allowedPrincipals: ['cluster.local/ns/default/sa/edge'], ports: ['0'] },
  ],
  [
    'blank port among valid ports',
    { enabled: true, allowedPrincipals: ['cluster.local/ns/default/sa/edge'], ports: ['8080', ''] },
  ],
] as const) {
  test(`enabled policy rejects ${name}`, () => {
    const result = render({ istio });
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /istio workload policy/);
  });
}

test('valid enabled policy renders and disabled policy does not', () => {
  const allowedPrincipals = ['cluster.local/ns/default/sa/edge'];
  const ports = ['8080'];
  const enabled = render({ istio: { enabled: true, allowedPrincipals, ports } });
  assert.equal(enabled.status, 0, enabled.stderr);
  assert.ok(resources(enabled.stdout).some((item) => item.kind === 'AuthorizationPolicy'));
  const disabled = render({ istio: { enabled: false, allowedPrincipals: [''], ports: [''] } });
  assert.equal(disabled.status, 0, disabled.stderr);
  assert.ok(!resources(disabled.stdout).some((item) => item.kind === 'AuthorizationPolicy'));
});

test('schema permits supported extra environment, resource and telemetry settings', () => {
  const result = render({
    env: { ANOTHER: 'value' },
    extraEnv: [{ name: 'POD', valueFrom: { fieldRef: { fieldPath: 'metadata.name' } } }],
    envFrom: [{ configMapRef: { name: 'settings' } }],
    resources: { limits: { cpu: '1' }, requests: { memory: '64Mi' } },
    otel: { resourceAttributes: { region: 'test' } },
    hooks: { seed: { activeDeadlineSeconds: 60, resources: { limits: { memory: '128Mi' } } } },
  });
  assert.equal(result.status, 0, result.stderr);
});

test('schema validation remains enabled', () => {
  const result = render({ runtime: 42 });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /schema/);
});

test('hook memory override without a limit fails instead of using the deployment limit', () => {
  const result = render({ hooks: { migrate: { resources: { requests: { memory: '64Mi' } } } } });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /unsupported or missing memory limit/);
});
