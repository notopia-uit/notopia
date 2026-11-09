import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { dirname, join, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { parse } from 'yaml';

import {
  checkReleaseSecret,
  findSecrets,
  parseDocuments,
  validateSecret,
} from './secret-validation.ts';

const root = dirname(fileURLToPath(import.meta.url));
const repo = resolve(root, '../../../..');
const resources = () =>
  parseDocuments(execFileSync('kubectl', ['kustomize', join(root, 'infra')], { encoding: 'utf8' }));

test('Postgres clusters', () => {
  const clusters = resources().filter((r) => r.kind === 'Cluster');
  assert.deepEqual(
    new Set(clusters.map((r) => r.metadata.name)),
    new Set(['note', 'billing', 'authorization', 'document'].map((s) => `cnpg-${s}`))
  );
  for (const { metadata, spec } of clusters) {
    const service = metadata.name.replace(/^cnpg-/, '');
    assert.equal(spec.instances, 1);
    assert.equal(spec.storage.size, '1Gi');
    assert.equal(spec.imageCatalogRef.major, 18);
    assert.equal(spec.imageCatalogRef.name, 'postgresql-minimal-trixie');
    assert.deepEqual(spec.bootstrap.initdb, { database: service, owner: service });
    assert.deepEqual(new Set(spec.managed.services.disabledDefaultServices), new Set(['ro', 'r']));
  }
});

test('infrastructure releases and values', () => {
  const docs = resources();
  const releases = new Map(
    docs.filter((r) => r.kind === 'HelmRelease').map((r) => [r.metadata.name, r])
  );
  const configs = new Map(
    docs.filter((r) => r.kind === 'ConfigMap').map((r) => [r.metadata.name, r])
  );
  const secrets = docs.filter((r) => r.kind === 'Secret' && r.metadata?.namespace === 'notopia');
  for (const service of ['note', 'billing', 'document']) {
    const release: any = releases.get(`valkey-${service}`);
    checkReleaseSecret(release, secrets, `valkey-${service}`, 'password');
    const values = parse((configs.get(release.spec.valuesFrom[0].name) as any).data['values.yaml']);
    assert.ok(values.dataStorage.enabled);
    assert.equal(values.replica.enabled, false);
    assert.match(values.valkeyConfig, /maxmemory-policy noeviction/);
    assert.match(values.valkeyConfig, /appendonly yes/);
    assert.equal(values.auth.usersExistingSecret, `valkey-${service}`);
  }
  const release: any = releases.get('meilisearch');
  checkReleaseSecret(release, secrets, 'meilisearch', 'MEILI_MASTER_KEY');
  const values = parse((configs.get(release.spec.valuesFrom[0].name) as any).data['values.yaml']);
  assert.equal(values.environment.MEILI_ENV, 'production');
  assert.equal(values.auth.existingMasterKeySecret, 'meilisearch');
  assert.ok(values.persistence.enabled);
});

test('meilisearch route', () => {
  const route = resources().find((r) => r.kind === 'HTTPRoute');
  assert.deepEqual(route.spec.hostnames, ['notopia-meilisearch.${PUBLIC_DOMAIN}']);
  const rule = route.spec.rules[0];
  assert.equal(rule.backendRefs[0].name, 'meilisearch');
  assert.deepEqual(rule.filters.find((f: any) => f.type === 'CORS').cors.allowOrigins, [
    'https://notopia.${PUBLIC_DOMAIN}',
    'https://notopia.${LOCAL_DOMAIN}',
  ]);
});

test('repository secrets are encrypted', () => {
  const output = execFileSync(
    'git',
    [
      'ls-files',
      '-z',
      '--cached',
      '--others',
      '--exclude-standard',
      '--',
      '*.yaml',
      '*.yml',
      '*.json',
    ],
    { cwd: repo, encoding: 'utf8' }
  );
  for (const [path, secret] of findSecrets(
    output
      .split('\0')
      .filter(Boolean)
      .map((path) => join(repo, path))
  ))
    validateSecret(secret, path);
});
