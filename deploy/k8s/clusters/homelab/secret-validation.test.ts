import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { before, test } from 'node:test';

import { parse, stringify } from 'yaml';

import { checkReleaseSecret, findSecrets, RECIPIENT, validateSecret } from './secret-validation.ts';

let secret: any;
before(() => {
  const plaintext =
    'apiVersion: v1\nkind: Secret\nmetadata:\n  name: valkey-note\nstringData:\n  password: synthetic-only\n';
  secret = parse(
    execFileSync(
      'sops',
      [
        '--encrypt',
        '--age',
        RECIPIENT,
        '--encrypted-regex',
        '^(data|stringData)$',
        '--input-type',
        'yaml',
        '--output-type',
        'yaml',
        '--filename-override',
        'deploy/k8s/clusters/homelab/fixture.sops.yaml',
        '/dev/stdin',
      ],
      { input: plaintext, encoding: 'utf8' }
    )
  );
});

test('SOPS-generated fixture', () => validateSecret(secret, 'fixture'));

test('plaintext and missing or invalid metadata are rejected', () => {
  const cases = [];
  for (const value of ['synthetic-only', 'ENC[not-encrypted]']) {
    const copy = structuredClone(secret);
    copy.stringData.password = value;
    cases.push(copy);
  }
  for (const field of ['mac', 'version', 'lastmodified', 'age', 'encrypted_regex']) {
    const copy = structuredClone(secret);
    delete copy.sops[field];
    cases.push(copy);
  }
  const wrongRecipient = structuredClone(secret);
  wrongRecipient.sops.age[0].recipient = 'age1wrong';
  cases.push(wrongRecipient);
  const wrongEnvelope = structuredClone(secret);
  wrongEnvelope.sops.age[0].enc = 'not an age envelope';
  cases.push(wrongEnvelope);
  for (const item of cases) assert.throws(() => validateSecret(item, 'fixture'));
});

test('scan multidocument YAML, both extensions and JSON lists outside infra', () => {
  const base = 'tmp/homelab-wiring';
  mkdirSync(base, { recursive: true });
  const root = mkdtempSync(join(base, 'secrets-'));
  try {
    mkdirSync(join(root, 'outside'));
    for (const name of ['other.yml', 'outside/unexpected.yaml'])
      writeFileSync(join(root, name), `kind: ConfigMap\n---\n${stringify(secret)}`);
    writeFileSync(
      join(root, 'list.json'),
      JSON.stringify({ apiVersion: 'v1', kind: 'List', items: [secret] })
    );
    const paths = () =>
      readdirSync(root, { recursive: true }).map((name) => join(root, String(name)));
    assert.equal([...findSecrets(paths())].length, 3);
    writeFileSync(
      join(root, 'plaintext.yml'),
      'kind: Secret\nstringData:\n  password: synthetic-only\n'
    );
    assert.throws(() => {
      for (const [path, found] of findSecrets(paths())) validateSecret(found, path);
    });
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test('suspension depends on rendered secret and key', () => {
  const release = { metadata: { name: 'valkey-note' }, spec: { suspend: true } };
  checkReleaseSecret(release, [], 'valkey-note', 'password');
  release.spec.suspend = false;
  assert.throws(() => checkReleaseSecret(release, [], 'valkey-note', 'password'));
  const missingKey = structuredClone(secret);
  missingKey.stringData = { other: missingKey.stringData.password };
  assert.throws(() => checkReleaseSecret(release, [missingKey], 'valkey-note', 'password'));
  checkReleaseSecret(release, [secret], 'valkey-note', 'password');
});
