import { readFileSync, statSync } from 'node:fs';
import { extname } from 'node:path';

import { parseAllDocuments } from 'yaml';

export const RECIPIENT = 'age1ytym5g6mxf0974syp2g2ecrh0y74tnl0u9ee4xm8caawhyv25pls3f2sy4';
const ENC =
  /^ENC\[AES256_GCM,data:([A-Za-z0-9+/=]*),iv:([A-Za-z0-9+/=]+),tag:([A-Za-z0-9+/=]+),type:str\]$/;

type ObjectMap = Record<string, any>;
const isObject = (value: unknown): value is ObjectMap =>
  value !== null && typeof value === 'object' && !Array.isArray(value);

function decode(value: string): Buffer | null {
  if (!/^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/.test(value)) return null;
  return Buffer.from(value, 'base64');
}

function encryptedValue(value: unknown): boolean {
  if (typeof value !== 'string') return false;
  const match = ENC.exec(value);
  if (!match) return false;
  const [data, iv, tag] = match.slice(1).map(decode);
  return !!data?.length && iv?.length === 32 && tag?.length === 16;
}

export function validateSecret(secret: ObjectMap, path: string): void {
  function require(condition: unknown, message: string): asserts condition {
    if (!condition) throw new Error(`${path}: ${message}`);
  }
  const metadata = secret.sops;
  require(isObject(metadata), 'missing SOPS metadata');
  require(metadata.encrypted_regex === '^(data|stringData)$', 'incorrect encrypted regex');
  require(/^\d+\.\d+\.\d+$/.test(String(metadata.version ?? '')), 'invalid SOPS version');
  require(typeof metadata.lastmodified === 'string' &&
    /(?:Z|[+-]\d\d:\d\d)$/.test(metadata.lastmodified) &&
    !Number.isNaN(Date.parse(metadata.lastmodified)), 'invalid SOPS timestamp');
  require(encryptedValue(metadata.mac), 'invalid SOPS MAC');
  require(Array.isArray(metadata.age), 'missing age recipients');
  const recipient = metadata.age.find(
    (entry: unknown) => isObject(entry) && entry.recipient === RECIPIENT
  );
  require(recipient, 'missing homelab age recipient');
  const match =
    typeof recipient.enc === 'string' &&
    /^-----BEGIN AGE ENCRYPTED FILE-----\s+([A-Za-z0-9+/=\s]+)-----END AGE ENCRYPTED FILE-----\s*$/.exec(
      recipient.enc
    );
  require(match, 'invalid age envelope');
  const decoded = decode(match[1].replace(/\s/g, ''));
  require(decoded?.subarray(0, 22).toString() ===
    'age-encryption.org/v1\n', 'invalid age envelope payload');
  const fields = ['data', 'stringData']
    .filter((field) => field in secret)
    .map((field) => secret[field]);
  require(fields.length &&
    fields.every((field) => isObject(field) && Object.keys(field).length), 'missing secret values');
  require(fields.every((field) =>
    Object.values(field).every(encryptedValue)
  ), 'plaintext or malformed secret value');
}

export function parseDocuments(text: string): any[] {
  return parseAllDocuments(text)
    .map((doc) => {
      if (doc.errors.length) throw doc.errors[0];
      return doc.toJS();
    })
    .filter(Boolean);
}

export function* findSecrets(paths: Iterable<string>): Generator<[string, ObjectMap]> {
  function* secretsIn(document: unknown): Generator<ObjectMap> {
    if (!isObject(document)) return;
    if (document.kind === 'Secret' && (document.apiVersion ?? 'v1') === 'v1') yield document;
    if (
      (document.kind === 'List' || document.kind === 'SecretList') &&
      Array.isArray(document.items)
    ) {
      for (const item of document.items) yield* secretsIn(item);
    }
  }
  for (const path of paths) {
    if (!['.yaml', '.yml', '.json'].includes(extname(path))) continue;
    try {
      if (!statSync(path).isFile()) continue;
    } catch (error: any) {
      if (error.code === 'ENOENT') continue;
      throw error;
    }
    const text = readFileSync(path, 'utf8');
    let documents: any[];
    try {
      documents = parseDocuments(text);
    } catch (error) {
      if (/\bkind['"]?\s*:\s*['"]?(?:Secret|SecretList)\b/.test(text))
        throw new Error(`${path}: cannot parse Secret document`, { cause: error });
      continue;
    }
    for (const document of documents)
      for (const secret of secretsIn(document)) yield [path, secret];
  }
}

export function checkReleaseSecret(
  release: ObjectMap,
  secrets: ObjectMap[],
  name: string,
  key: string
): void {
  if (release.spec.suspend) return;
  if (
    secrets.some(
      (secret) =>
        secret.metadata?.name === name &&
        (Object.hasOwn(secret.data ?? {}, key) || Object.hasOwn(secret.stringData ?? {}, key))
    )
  )
    return;
  throw new Error(
    `${release.metadata.name}: missing rendered Secret ${name} key ${key}; release must be suspended`
  );
}
