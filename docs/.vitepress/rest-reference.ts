import type notopiaSpec from '@notopia-uit/api/openapi';
import { createOpenApiSpec, httpVerbs } from 'vitepress-openapi';

export const REST_OPERATION_ROUTE = 'api/rest/operations/{operationId}';

export interface RestOperation {
  operationId: string;
  title: string;
  method: string;
  path: string;
  href: string;
}

export interface RestReference {
  operations: RestOperation[];
  groups: { tag: string; operations: RestOperation[] }[];
}

export function getRestReference(spec: typeof notopiaSpec): RestReference {
  const ids = new Set<string>();
  for (const [path, pathItem] of Object.entries(spec.paths)) {
    const item = pathItem as Partial<Record<(typeof httpVerbs)[number], { operationId?: string }>>;
    for (const method of httpVerbs) {
      const operation = item[method];
      if (!operation) continue;
      const id = operation.operationId;
      if (!id) throw new Error(`Missing operationId for ${method.toUpperCase()} ${path}`);
      if (!/^[A-Za-z0-9_-]+$/.test(id)) throw new Error(`Unsafe operationId: ${id}`);
      if (ids.has(id)) throw new Error(`Duplicate operationId: ${id}`);
      ids.add(id);
    }
  }

  const groups = new Map<string, RestOperation[]>(spec.tags?.map(({ name }) => [name, []]));
  const untagged: RestOperation[] = [];
  const api = createOpenApiSpec({
    spec: spec as unknown as NonNullable<Parameters<typeof createOpenApiSpec>[0]>['spec'],
  });
  const operations = api.getPathsByVerbs().map(({ operationId, summary, verb, path, tags }) => {
    const operation: RestOperation = {
      operationId,
      title: summary || operationId,
      method: verb.toUpperCase(),
      path,
      href: `/${REST_OPERATION_ROUTE.replace('{operationId}', operationId)}`,
    };
    if (!tags.length) untagged.push(operation);
    for (const tag of new Set<string>(tags)) {
      if (!groups.has(tag)) groups.set(tag, []);
      groups.get(tag)!.push(operation);
    }
    return operation;
  });
  if (untagged.length) groups.set('Other', [...(groups.get('Other') ?? []), ...untagged]);
  return {
    operations,
    groups: Array.from(groups, ([tag, operations]) => ({ tag, operations })).filter(
      ({ operations }) => operations.length
    ),
  };
}
