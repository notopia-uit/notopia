import { execFileSync } from 'node:child_process';
import path from 'node:path';

interface ChartMetadata {
  version?: unknown;
  dependencies?: { repository?: string }[];
}

const git = (...args: string[]) => execFileSync('git', args, { encoding: 'utf8' });
const root = 'deploy/k8s/charts/';
const [baseInput, headInput, ...extra] = process.argv.slice(2);
if (!baseInput || !headInput || extra.length) {
  console.error('Usage: node scripts/check-chart-versions.ts <base-sha> <head-sha>');
  process.exit(1);
}

function commit(ref: string, label: string) {
  try {
    return git('rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`).trim();
  } catch {
    throw new Error(
      `Invalid or missing ${label} commit: ${ref}; fetch full history before checking chart versions`
    );
  }
}

function charts(ref: string) {
  const result = new Map<string, ChartMetadata>();
  for (const file of git('ls-tree', '-r', '--name-only', '-z', ref, '--', root).split('\0')) {
    if (!/^deploy\/k8s\/charts\/[^/]+\/Chart\.yaml$/.test(file)) continue;
    const content = git('show', `${ref}:${file}`);
    const metadata: ChartMetadata = JSON.parse(
      execFileSync('yq', ['-o=json', '.', '-'], { input: content, encoding: 'utf8' })
    );
    if (!metadata || typeof metadata !== 'object')
      throw new Error(`Invalid chart metadata: ${file} at ${ref}`);
    result.set(path.posix.dirname(file), metadata);
  }
  return result;
}

const semver =
  /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$/;
function version(value: unknown) {
  return typeof value === 'string' ? semver.exec(value) : null;
}

function higher(next: RegExpExecArray, previous: RegExpExecArray) {
  for (let i = 1; i <= 3; i++) {
    if (BigInt(next[i]) !== BigInt(previous[i])) return BigInt(next[i]) > BigInt(previous[i]);
  }
  if (!previous[4]) return false;
  if (!next[4]) return true;
  const a = next[4].split('.');
  const b = previous[4].split('.');
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    if (a[i] === undefined) return false;
    if (b[i] === undefined) return true;
    if (a[i] === b[i]) continue;
    const aNumeric = /^\d+$/.test(a[i]);
    const bNumeric = /^\d+$/.test(b[i]);
    if (aNumeric && bNumeric) return BigInt(a[i]) > BigInt(b[i]);
    if (aNumeric !== bNumeric) return !aNumeric;
    return a[i] > b[i];
  }
  return false;
}

try {
  const base = commit(baseInput, 'base');
  const head = commit(headInput, 'head');
  const before = charts(base);
  const after = charts(head);
  const changed = new Set<string>();
  for (const file of git('diff', '--no-renames', '--name-only', '-z', base, head, '--', root).split(
    '\0'
  )) {
    const match = /^(deploy\/k8s\/charts\/[^/]+)\//.exec(file);
    if (match) changed.add(match[1]);
  }

  // Include both graphs so removing a dependency cannot bypass a required bump.
  let expanded;
  do {
    expanded = false;
    for (const graph of [before, after]) {
      for (const [directory, metadata] of graph) {
        for (const dependency of metadata.dependencies ?? []) {
          if (!dependency.repository?.startsWith('file://')) continue;
          const target = path.posix.normalize(
            path.posix.join(directory, dependency.repository.slice(7))
          );
          if (changed.has(target) && !changed.has(directory)) {
            changed.add(directory);
            expanded = true;
          }
        }
      }
    }
  } while (expanded);

  const errors: string[] = [];
  for (const directory of [...changed].sort()) {
    const name = directory.slice(root.length);
    const current = after.get(directory);
    const old = before.get(directory);
    if (!current) {
      errors.push(
        `${name}: deleted or missing Chart.yaml; chart removal needs an explicit migration policy`
      );
      continue;
    }
    const next = version(current.version);
    if (!next) {
      errors.push(`${name}: invalid SemVer version ${JSON.stringify(current.version)}`);
      continue;
    }
    if (!old) continue;
    const previous = version(old.version);
    if (!previous || !higher(next, previous)) {
      errors.push(
        `${name}: version must be higher than base ${JSON.stringify(old.version)} (head ${JSON.stringify(current.version)})`
      );
    }
  }
  if (errors.length) throw new Error(`Chart version check failed:\n${errors.join('\n')}`);
  console.log(`Chart versions valid (${changed.size} changed charts including local dependents)`);
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
}
