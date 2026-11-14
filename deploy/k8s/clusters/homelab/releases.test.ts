import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { appendFileSync, cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { parse } from 'yaml';

import { parseDocuments } from './secret-validation.ts';

const root = dirname(fileURLToPath(import.meta.url));
const repo = resolve(root, '../../../..');
const services = ['web', 'note', 'authorization', 'billing', 'document', 'search-worker'];
const readYaml = (path: string) => parse(readFileSync(path, 'utf8'));
const build = (directory: string) =>
  parseDocuments(execFileSync('kubectl', ['kustomize', directory], { encoding: 'utf8' }));
const render = (name: string, extra: string[] = []) =>
  parseDocuments(
    execFileSync(
      'helm',
      [
        'template',
        name,
        join(root, '../../charts', name),
        '-n',
        'notopia',
        '-f',
        join(root, name === 'notopia-edge' ? 'edge/values.yaml' : `apps/${name}/values.yaml`),
        ...extra,
      ],
      { encoding: 'utf8' }
    )
  );
const edgeRender = () =>
  render('notopia-edge', ['--set', 'security.externalBoundaryAcknowledged=true']);
const set = (values: string[]) => new Set(values);

test('broker values', () => {
  for (const name of services.filter((s) => s !== 'web')) {
    const values = readYaml(join(root, 'apps', name, 'values.yaml'));
    assert.equal(
      values.env[`NOTOPIA_${name.toUpperCase().replaceAll('-', '_')}_KAFKA_BROKERS`],
      'redpanda.redpanda.svc.cluster.local:9093',
      name
    );
  }
});

test('exact principals and rendered ports', () => {
  const edge = 'cluster.local/ns/notopia/sa/notopia-edge-istio';
  const principal = (name: string) => `cluster.local/ns/notopia/sa/${name}`;
  const expected: Record<string, [string[], string[]]> = {
    web: [['cluster.local/ns/istio-system/sa/default-istio'], ['3000']],
    note: [
      [edge, principal('document')],
      ['8081', '18081', '28081'],
    ],
    authorization: [
      [edge, principal('note'), principal('document'), principal('billing')],
      ['18089', '28089'],
    ],
    billing: [
      [edge, principal('note'), principal('document')],
      ['8084', '18084', '28084'],
    ],
    document: [[edge], ['8082']],
    'search-worker': [[edge], ['28083']],
  };
  for (const [name, [principals, ports]] of Object.entries(expected)) {
    const docs = render(name);
    const rule = docs.find((d) => d.kind === 'AuthorizationPolicy').spec.rules[0];
    assert.deepEqual(set(rule.from[0].source.principals), set(principals), name);
    assert.deepEqual(set(rule.to[0].operation.ports), set(ports), name);
    assert.deepEqual(
      set(docs.find((d) => d.kind === 'Service').spec.ports.map((p: any) => String(p.port))),
      set(ports),
      name
    );
  }
});

test('rendered public rules and header removal', () => {
  const docs = edgeRender();
  const rules = docs.find((d) => d.kind === 'AuthorizationPolicy').spec.rules;
  const publicRules = rules
    .filter((r: any) => !('from' in r))
    .flatMap((r: any) =>
      r.to[0].operation.paths.map((path: string) =>
        JSON.stringify([path, r.to[0].operation.methods ?? []])
      )
    );
  const expected = [
    '/note/health/*',
    '/authorization/health/*',
    '/billing/health/*',
    '/document/health/*',
    '/search-worker/health/*',
    '/billing/webhooks/stripe',
    '/document/ws/document',
  ].map((p) => JSON.stringify([p, []]));
  expected.push(JSON.stringify(['/billing/plans', ['GET']]));
  assert.deepEqual(set(publicRules), set(expected));
  const filters = docs.find((d) => d.kind === 'HTTPRoute').spec.rules[0].filters;
  assert.deepEqual(
    set(filters.find((f: any) => f.type === 'RequestHeaderModifier').requestHeaderModifier.remove),
    set([
      'X-Forwarded-ID',
      'X-Forwarded-Email',
      'X-Forwarded-Name',
      'X-Forwarded-Groups',
      'X-Forwarded-Roles',
    ])
  );
});

test('releases and hashed values', () => {
  for (const [directory, names] of [
    ['edge', ['notopia-edge']],
    ['apps', services],
  ] as const) {
    const docs = build(join(root, directory));
    const releases = docs.filter((d) => d.kind === 'HelmRelease');
    const configmaps = new Map(
      docs.filter((d) => d.kind === 'ConfigMap').map((d) => [d.metadata.name, d])
    );
    assert.deepEqual(set(releases.map((r) => r.metadata.name)), set([...names]));
    for (const { metadata, spec } of releases) {
      const name = metadata.name;
      assert.equal(spec.suspend, true);
      assert.equal(spec.interval, '12h');
      assert.equal(spec.timeout, '10m');
      assert.equal(spec.install.remediation.retries, 0);
      assert.equal(spec.upgrade.remediation.strategy, 'rollback');
      assert.equal(spec.upgrade.remediation.retries, 0);
      assert.equal(spec.upgrade.remediation.remediateLastFailure, name !== 'authorization');
      assert.equal(spec.chart.spec.chart, `./deploy/k8s/charts/${name}`);
      assert.deepEqual(spec.chart.spec.sourceRef, {
        kind: 'GitRepository',
        name: 'notopia',
        namespace: 'flux-system',
      });
      const ref = spec.valuesFrom[0].name;
      assert.ok(configmaps.has(ref));
      assert.match(ref, new RegExp(`^${name}-helm-values-[a-z0-9]{10}$`));
      assert.deepEqual(
        parse((configmaps.get(ref) as any).data['values.yaml']),
        readYaml(
          join(root, directory, directory === 'edge' ? 'values.yaml' : `${name}/values.yaml`)
        )
      );
    }
  }
});

test('service wiring', () => {
  const values = Object.fromEntries(
    services.map((name) => [name, readYaml(join(root, 'apps', name, 'values.yaml'))])
  );
  for (const [name, value] of Object.entries(values)) {
    assert.ok('tag' in value.image);
    assert.equal(
      value.otel.endpoint,
      `http://alloy-alloy.monitoring.svc:${name === 'web' ? 4318 : 4317}`
    );
  }
  assert.equal(values.web.otel.protocol, 'http/protobuf');
  assert.ok(values.web.env.DOCUMENT_WS_URL.startsWith('wss://'));
  for (const name of ['note', 'billing', 'authorization'])
    assert.equal(values[name].database.urlSecret.name, `cnpg-${name}-app`);
  for (const name of ['note', 'billing']) {
    assert.equal(values[name].redis.addr, `valkey-${name}:6379`);
    assert.equal(values[name].redis.passwordSecret.name, `valkey-${name}`);
  }
  assert.deepEqual(values['search-worker'].envFrom, [{ secretRef: { name: 'search-worker' } }]);
  assert.equal(
    values.note.env.NOTOPIA_NOTE_MEILISEARCH_NOTE_SEARCH_KEY_UID,
    values['search-worker'].env.NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID
  );
  assert.deepEqual(
    set(values.document.extraEnv.map((e: any) => e.valueFrom.secretKeyRef.key)),
    set(['host', 'port', 'username', 'password', 'dbname'])
  );
  const addresses: Record<string, Record<string, string>> = {
    note: {
      NOTOPIA_NOTE_SERVICES_AUTHORIZATION_URL: 'authorization:18089',
      NOTOPIA_NOTE_SERVICES_AUTHORIZATION_LIVE_URL:
        'http://authorization:28089/authorization/health/live',
      NOTOPIA_NOTE_SERVICES_BILLING_URL: 'billing:18084',
      NOTOPIA_NOTE_SERVICES_BILLING_LIVE_URL: 'http://billing:28084/billing/health/live',
      NOTOPIA_NOTE_MEILISEARCH_HOST: 'http://meilisearch:7700',
    },
    billing: {
      NOTOPIA_BILLING_SERVICES_AUTHORIZATION_URL: 'authorization:18089',
      NOTOPIA_BILLING_SERVICES_AUTHORIZATION_LIVE_URL:
        'http://authorization:28089/authorization/health/live',
    },
    document: {
      NOTOPIA_DOCUMENT_SERVICES_NOTE_GRPC_URL: 'note:18081',
      NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL: 'authorization:18089',
      NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL: 'billing:18084',
    },
    'search-worker': { NOTOPIA_SEARCH_WORKER_MEILI_HOST: 'http://meilisearch:7700' },
  };
  for (const [name, entries] of Object.entries(addresses))
    for (const [key, address] of Object.entries(entries))
      assert.equal(values[name].env[key], address);
  for (const [name, callers] of Object.entries({
    note: ['document'],
    billing: ['note', 'document'],
    authorization: ['note', 'document', 'billing'],
  }))
    for (const caller of callers)
      assert.ok(
        values[name].istio.allowedPrincipals.includes(`cluster.local/ns/notopia/sa/${caller}`)
      );
  const edge = readYaml(join(root, 'edge/values.yaml'));
  assert.equal(edge.security.externalBoundaryAcknowledged, false);
  assert.equal(edge.istio.adminGroup, 'notopia-admins');
  for (const rule of edge.istio.publicRules)
    for (const path of rule.paths) {
      assert.ok(!path.slice(0, -1).includes('*'));
      assert.ok(!path.startsWith('/billing/admin'));
    }
});

test('rendered edge excludes admin from general JWT', () => {
  const rules = edgeRender().find((d) => d.kind === 'AuthorizationPolicy').spec.rules;
  const general = rules.find((r: any) => r.to?.[0]?.operation?.notPaths);
  assert.deepEqual(
    set(general.to[0].operation.notPaths),
    set(['/billing/admin', '/billing/admin/*'])
  );
  assert.deepEqual(rules.find((r: any) => r.when).when[0].values, ['notopia-admins']);
});

test('values changes rename configmaps', () => {
  const base = join(repo, 'tmp/homelab-wiring');
  mkdirSync(base, { recursive: true });
  const temp = mkdtempSync(join(base, 'values-'));
  try {
    for (const [directory, name, valuesPath] of [
      ['edge', 'notopia-edge', 'values.yaml'],
      ['apps', 'note', 'note/values.yaml'],
    ]) {
      const target = join(temp, directory);
      cpSync(join(root, directory), target, { recursive: true });
      const reference = () =>
        build(target).find((d) => d.kind === 'HelmRelease' && d.metadata.name === name).spec
          .valuesFrom[0].name;
      const before = reference();
      appendFileSync(join(target, valuesPath), '\n# changed values\n');
      assert.notEqual(before, reference());
    }
  } finally {
    rmSync(temp, { recursive: true, force: true });
  }
});
