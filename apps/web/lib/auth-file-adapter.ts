import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';

import { createAdapterFactory } from 'better-auth/adapters';

type Row = Record<string, unknown>;
type FileDb = Record<string, Row[]>;

const ISO_DATE_RE = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2})?$/;

function revive(value: unknown): unknown {
  if (typeof value === 'string' && ISO_DATE_RE.test(value)) {
    const d = new Date(value);
    if (!Number.isNaN(d.getTime())) return d;
    return value;
  }
  if (Array.isArray(value)) return value.map(revive);
  if (value !== null && typeof value === 'object') {
    for (const [k, v] of Object.entries(value)) {
      (value as Record<string, unknown>)[k] = revive(v);
    }
  }
  return value;
}

function resolveDbPath(): string {
  if (process.env.BETTER_AUTH_DB_PATH) return resolve(process.env.BETTER_AUTH_DB_PATH);
  let dir = process.cwd();
  for (;;) {
    if (existsSync(join(dir, 'pnpm-workspace.yaml')))
      return join(dir, 'tmp', 'web', 'better-auth.json');
    const parent = dirname(dir);
    if (parent === dir) return join(process.cwd(), 'tmp', 'web', 'better-auth.json');
    dir = parent;
  }
}

function loadDb(path: string): FileDb {
  try {
    return revive(JSON.parse(readFileSync(path, 'utf8'))) as FileDb;
  } catch {
    return {};
  }
}

export const fileAdapter = () => {
  const path = resolveDbPath();
  let db = loadDb(path);
  let inTransaction = false;

  const persist = () => {
    if (inTransaction) return;
    mkdirSync(dirname(path), { recursive: true });
    writeFileSync(path, JSON.stringify(db), { mode: 0o600 });
  };
  const table = (model: string): Row[] => (db[model] ??= []);

  interface Clause {
    field: string;
    value: unknown;
    operator?: string;
    connector?: string;
    mode?: string;
  }

  const matchClause = (record: Row, clause: Clause): boolean => {
    const { field, value, operator, mode = 'sensitive' } = clause;
    const rv = record[field];
    const insensitive =
      mode === 'insensitive' &&
      (typeof value === 'string' ||
        (Array.isArray(value) && value.every((v) => typeof v === 'string')));
    const ci = (a: unknown, b: unknown) =>
      typeof a === 'string' && typeof b === 'string'
        ? a.toLowerCase() === b.toLowerCase()
        : a === b;
    switch (operator) {
      case 'in':
        if (!Array.isArray(value)) throw new Error('Value must be an array');
        return insensitive && typeof rv === 'string'
          ? value.some(
              (v) => typeof v === 'string' && (rv as string).toLowerCase() === v.toLowerCase()
            )
          : (value as unknown[]).includes(rv);
      case 'not_in':
        if (!Array.isArray(value)) throw new Error('Value must be an array');
        return !(insensitive && typeof rv === 'string'
          ? value.some(
              (v) => typeof v === 'string' && (rv as string).toLowerCase() === v.toLowerCase()
            )
          : (value as unknown[]).includes(rv));
      case 'contains':
        return typeof rv === 'string' && typeof value === 'string'
          ? insensitive
            ? rv.toLowerCase().includes(value.toLowerCase())
            : rv.includes(value)
          : false;
      case 'starts_with':
        return typeof rv === 'string' && typeof value === 'string'
          ? insensitive
            ? rv.toLowerCase().startsWith(value.toLowerCase())
            : rv.startsWith(value)
          : false;
      case 'ends_with':
        return typeof rv === 'string' && typeof value === 'string'
          ? insensitive
            ? rv.toLowerCase().endsWith(value.toLowerCase())
            : rv.endsWith(value)
          : false;
      case 'ne':
        return insensitive ? !ci(rv, value) : rv !== value;
      case 'gt':
        return value != null && (rv as number) > (value as number);
      case 'gte':
        return value != null && (rv as number) >= (value as number);
      case 'lt':
        return value != null && (rv as number) < (value as number);
      case 'lte':
        return value != null && (rv as number) <= (value as number);
      default:
        if (insensitive) return ci(rv, value);
        if (value === null) return rv == null;
        return rv === value;
    }
  };

  const buildAdapterFactory = (active: () => FileDb) =>
    createAdapterFactory({
      config: {
        adapterId: 'notopia-file',
        adapterName: 'Notopia File Adapter',
        usePlural: false,
        supportsArrays: true,
        transaction: async (cb) => {
          const snapshot = structuredClone(db);
          inTransaction = true;
          try {
            const result = await cb(buildAdapterFactory(() => db)(lazyOptions));
            inTransaction = false;
            persist();
            return result;
          } catch (err) {
            db = snapshot;
            inTransaction = false;
            persist();
            throw err;
          }
        },
      },
      adapter: ({ getFieldName, getDefaultFieldName }) => {
        const applySort = (
          records: Row[],
          sortBy: { field: string; direction: string } | undefined,
          modelKey: string
        ) => {
          if (!sortBy) return records;
          return records.sort((a, b) => {
            const field = getFieldName({ model: modelKey, field: sortBy.field });
            const av = a[field];
            const bv = b[field];
            let c: number;
            if (av == null && bv == null) c = 0;
            else if (av == null) c = -1;
            else if (bv == null) c = 1;
            else if (typeof av === 'string' && typeof bv === 'string') c = av.localeCompare(bv);
            else if (av instanceof Date && bv instanceof Date) c = av.getTime() - bv.getTime();
            else if (typeof av === 'number' && typeof bv === 'number') c = av - bv;
            else if (typeof av === 'boolean' && typeof bv === 'boolean')
              c = av === bv ? 0 : av ? 1 : -1;
            else c = String(av).localeCompare(String(bv));
            return sortBy.direction === 'asc' ? c : -c;
          });
        };

        interface JoinAttr {
          on: { from: string; to: string };
          relation?: string;
          limit?: number;
        }

        const filter = (
          model: string,
          modelKey: string,
          where: Clause[],
          select?: string[],
          join?: Record<string, JoinAttr>
        ) => {
          let rows = (active()[model] ?? []).filter((record) => {
            if (!where.length) return true;
            let result = matchClause(record, where[0]);
            for (let i = 1; i < where.length; i++) {
              const r = matchClause(record, where[i]);
              result = where[i].connector === 'OR' ? result || r : result && r;
            }
            return result;
          });
          if (select?.length) {
            rows = rows.map((record) =>
              Object.fromEntries(
                Object.entries(record).filter(([key]) =>
                  select.includes(getDefaultFieldName({ model: modelKey, field: key }))
                )
              )
            );
          }
          if (!join) return rows;
          const grouped = new Map<string, Row>();
          for (const baseRecord of rows) {
            const baseId = String(baseRecord.id);
            let nested = grouped.get(baseId);
            if (!nested) {
              nested = { ...baseRecord };
              for (const [joinModel, joinAttr] of Object.entries(join)) {
                nested[joinModel] = joinAttr.relation === 'one-to-one' ? null : [];
              }
              grouped.set(baseId, nested);
            }
            for (const [joinModel, joinAttr] of Object.entries(join)) {
              const joinTable = active()[joinModel];
              if (!joinTable) throw new Error(`Join model ${joinModel} not found`);
              const matches = joinTable.filter(
                (joinRecord) => joinRecord[joinAttr.on.to] === baseRecord[joinAttr.on.from]
              );
              if (joinAttr.relation === 'one-to-one') {
                nested[joinModel] = matches[0] ?? null;
              } else {
                const list = nested[joinModel] as Row[];
                const seen = new Set(list.map((r) => r.id));
                const limit = joinAttr.limit ?? 100;
                for (const m of matches) {
                  if (list.length >= limit) break;
                  if (!seen.has(m.id)) {
                    seen.add(m.id);
                    list.push(m);
                  }
                }
              }
            }
          }
          return [...grouped.values()];
        };

        return {
          create: async <T extends Record<string, unknown>>({
            model,
            data,
          }: {
            model: string;
            data: T;
          }): Promise<T> => {
            table(model).push(data as Row);
            persist();
            return data;
          },
          findOne: async <T>({
            model,
            modelKey = model,
            where,
            select,
            join,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
            select?: string[];
            join?: Record<string, JoinAttr>;
          }): Promise<T | null> => {
            return (filter(model, modelKey, where, select, join)[0] ?? null) as T | null;
          },
          findMany: async <T>({
            model,
            modelKey = model,
            where,
            sortBy,
            limit,
            select,
            offset,
            join,
          }: {
            model: string;
            modelKey?: string;
            where?: Clause[];
            sortBy?: { field: string; direction: 'asc' | 'desc' };
            limit?: number;
            select?: string[];
            offset?: number;
            join?: Record<string, JoinAttr>;
          }): Promise<T[]> => {
            let res = filter(model, modelKey, where ?? [], select, join);
            res = applySort(res, sortBy, modelKey);
            if (offset !== undefined) res = res.slice(offset);
            if (limit !== undefined) res = res.slice(0, limit);
            return res as T[];
          },
          count: async ({
            model,
            where,
          }: {
            model: string;
            modelKey?: string;
            where?: Clause[];
          }): Promise<number> => {
            if (where) return filter(model, model ?? '', where).length;
            return (active()[model] ?? []).length;
          },
          update: async <T>({
            model,
            modelKey = model,
            where,
            update,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
            update: T;
          }): Promise<T | null> => {
            if (!where.length) return null;
            const res = filter(model, modelKey, where);
            for (const record of res) Object.assign(record, update as Record<string, unknown>);
            if (res.length) persist();
            return (res[0] ?? null) as T | null;
          },
          updateMany: async ({
            model,
            modelKey = model,
            where,
            update,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
            update: Record<string, unknown>;
          }): Promise<number> => {
            const res = filter(model, modelKey, where);
            for (const record of res) Object.assign(record, update);
            if (res.length) persist();
            return res.length;
          },
          delete: async ({
            model,
            modelKey = model,
            where,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
          }): Promise<void> => {
            if (!where.length) return;
            const condemned = new Set(filter(model, modelKey, where));
            if (!condemned.size) return;
            db[model] = table(model).filter((r) => !condemned.has(r));
            persist();
          },
          deleteMany: async ({
            model,
            modelKey = model,
            where,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
          }): Promise<number> => {
            const condemned = new Set(filter(model, modelKey, where));
            db[model] = table(model).filter((r) => !condemned.has(r));
            if (condemned.size) persist();
            return condemned.size;
          },
          consumeOne: async <T>({
            model,
            modelKey = model,
            where,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
          }): Promise<T | null> => {
            const target = filter(model, modelKey, where)[0];
            if (!target) return null;
            db[model] = table(model).filter((r) => r !== target);
            persist();
            return target as T;
          },
          incrementOne: async <T>({
            model,
            modelKey = model,
            where,
            increment,
            set,
          }: {
            model: string;
            modelKey?: string;
            where: Clause[];
            increment: Record<string, number>;
            set?: Record<string, unknown>;
          }): Promise<T | null> => {
            const target = filter(model, modelKey, where)[0];
            if (!target) return null;
            for (const [field, delta] of Object.entries(increment ?? {})) {
              target[field] =
                (typeof target[field] === 'number' ? (target[field] as number) : 0) + delta;
            }
            if (set) Object.assign(target, set);
            persist();
            return target as T;
          },
        };
      },
    });

  let lazyOptions!: Parameters<ReturnType<typeof buildAdapterFactory>>[0];
  const creator = buildAdapterFactory(() => db);
  return (options: Parameters<typeof creator>[0]) => {
    lazyOptions = options;
    return creator(options);
  };
};
