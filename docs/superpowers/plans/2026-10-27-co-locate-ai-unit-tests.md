# Co-locate Unit Tests in apps/ai Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move unit test files in `apps/ai` from `apps/ai/tests/` into `apps/ai/src/` co-located next to their source files, merge `access.spec.ts` and `stream.spec.ts` into `src/ai/ai.service.spec.ts`, and configure `vitest.config.mts` and `tsconfig.spec.json`.

**Architecture:** Adopt the project configuration standard used by `apps/document` and `apps/search-worker`: tests sit adjacent to their subject in `src/`, excluded from `tsconfig.app.json`, typed via a dedicated `tsconfig.spec.json` referenced in `tsconfig.json`, and consumed by Vitest's SWC plugin.

**Tech Stack:** Vitest, unplugin-swc, TypeScript project references (`tsc --build`), Nx.

**Spec:** `docs/superpowers/specs/2026-10-27-ai-test-co-location-design.md`

## Global Constraints

- Never alter existing test assertions, mock behaviors, or test semantics.
- Tests in `src/` must be named `*.spec.ts` matching their subject filename (`subject.ts` -> `subject.spec.ts`).
- `tsconfig.app.json` must exclude `src/**/*.spec.ts` to keep app declaration emits free of test artifacts.
- `tsconfig.spec.json` must extend `../../tsconfig.base.json`, reference `./tsconfig.app.json`, and set `outDir` to `./out-tsc/vitest`.
- All imports between tests and adjacent modules must use clean relative paths (e.g. `./ai.service`, `../config`).
- Run `oxfmt <changed js/ts files>` and `nx lint ai --fix` before each commit. Use conventional commits (`refactor(ai): ...`).

## Review Focus

- Hoisted Vitest mocks: `vi.mock('@ai-sdk/openai-compatible')` in the merged `ai.service.spec.ts` must not interfere with workspace authorization tests.
- Relative import paths: every moved spec must resolve imports locally without broken relative `../src/` prefixes.
- TypeScript declaration emit: `nx typecheck ai` (`tsc --build --emitDeclarationOnly`) must succeed without emitting `.d.ts` for specs into `out-tsc/app`.
- Vitest SWC transpilation: `vitest` must successfully transpile specs using `tsconfig.spec.json` and resolve `#/*` path aliases.
- Directory cleanup: `apps/ai/tests/` must be completely removed with no lingering files or empty folders.

---

### Task 1: TypeScript and Vitest Configuration for Co-located Tests

**Files:**
- Create: `apps/ai/tsconfig.spec.json`
- Modify: `apps/ai/tsconfig.app.json`
- Modify: `apps/ai/tsconfig.json`
- Modify: `apps/ai/vitest.config.mts`

**Interfaces:**
- Consumes: `../../tsconfig.base.json`, `apps/ai/tsconfig.app.json`
- Produces: `tsconfig.spec.json` project for Vitest and typechecking

- [ ] **Step 1: Create `apps/ai/tsconfig.spec.json`**

Create `apps/ai/tsconfig.spec.json` with the following configuration:

```json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./out-tsc/vitest",
    "types": ["vitest/globals", "node", "vitest", "vitest/importMeta"],
    "experimentalDecorators": true,
    "emitDecoratorMetadata": true,
    "paths": {
      "#/*": ["./src/*"]
    }
  },
  "include": [
    "vitest.config.mts",
    "src/**/*.spec.ts"
  ],
  "exclude": ["out-tsc/**/*"],
  "references": [
    {
      "path": "./tsconfig.app.json"
    }
  ]
}
```

- [ ] **Step 2: Update `apps/ai/tsconfig.app.json` to exclude spec files**

Add `"exclude": ["src/**/*.spec.ts"]` to `apps/ai/tsconfig.app.json`:

```json
{
  "$schema": "https://json.schemastore.org/tsconfig",
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "module": "esnext",
    "moduleResolution": "bundler",
    "types": ["node", "@rspack/core/module"],
    "experimentalDecorators": true,
    "emitDecoratorMetadata": true,
    "target": "esnext",
    "rootDir": "src",
    "outDir": "out-tsc/app",
    "paths": {
      "#/*": ["./src/*"]
    }
  },
  "include": ["src/**/*"],
  "references": [
    {
      "path": "../../packages/pb/tsconfig.lib.json"
    }
  ],
  "exclude": ["src/**/*.spec.ts"]
}
```

- [ ] **Step 3: Update `apps/ai/tsconfig.json` to reference `tsconfig.spec.json`**

Update `apps/ai/tsconfig.json` to reference both `tsconfig.app.json` and `tsconfig.spec.json`:

```json
{
  "extends": "../../tsconfig.base.json",
  "files": [],
  "include": [],
  "compilerOptions": {},
  "references": [
    {
      "path": "./tsconfig.app.json"
    },
    {
      "path": "./tsconfig.spec.json"
    }
  ]
}
```

- [ ] **Step 4: Update `apps/ai/vitest.config.mts` to use `tsconfig.spec.json` and `src/**/*.spec.ts`**

Update `apps/ai/vitest.config.mts`:
- Change `swc.vite` option `tsconfigFile` from `'tsconfig.app.json'` to `'tsconfig.spec.json'`.
- Change `test.include` from `['tests/**/*.spec.ts']` to `['src/**/*.spec.ts']`.

```ts
import { fileURLToPath } from 'node:url';

import swc from 'unplugin-swc';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  resolve: { alias: { '#': fileURLToPath(new URL('./src', import.meta.url)) } },
  plugins: [swc.vite({ tsconfigFile: 'tsconfig.spec.json', module: { type: 'es6' } }) as any],
  test: { name: 'ai', environment: 'node', include: ['src/**/*.spec.ts'], watch: false },
});
```

- [ ] **Step 5: Verify configuration with typecheck**

Run: `pnpm exec nx typecheck ai`
Expected: PASS (builds `tsconfig.app.json` and `tsconfig.spec.json` cleanly).

- [ ] **Step 6: Commit**

```bash
git add apps/ai/tsconfig.spec.json apps/ai/tsconfig.app.json apps/ai/tsconfig.json apps/ai/vitest.config.mts
git commit -m "refactor(ai): configure vitest and tsconfig for co-located unit tests"
```

---

### Task 2: Co-locate `chat-request.dto.spec.ts`, `billing.service.spec.ts`, and `health.controller.spec.ts`

**Files:**
- Create: `apps/ai/src/ai/chat-request.dto.spec.ts` (moved from `apps/ai/tests/request.spec.ts`)
- Delete: `apps/ai/tests/request.spec.ts`
- Create: `apps/ai/src/billing/billing.service.spec.ts` (moved from `apps/ai/tests/billing.spec.ts`)
- Delete: `apps/ai/tests/billing.spec.ts`
- Create: `apps/ai/src/health.controller.spec.ts` (moved from `apps/ai/tests/health.spec.ts`)
- Delete: `apps/ai/tests/health.spec.ts`

**Interfaces:**
- Consumes: `src/ai/chat-request.dto.ts`, `src/billing/billing.service.ts`, `src/health.controller.ts`, `src/usage/usage.service.ts`
- Produces: Co-located unit tests for chat request DTO, billing service, and health controller.

- [ ] **Step 1: Move and update `apps/ai/tests/request.spec.ts` to `apps/ai/src/ai/chat-request.dto.spec.ts`**

Move the file and update the import:
- `import { parseChatRequest } from '../src/ai/chat-request.dto';` → `import { parseChatRequest } from './chat-request.dto';`

- [ ] **Step 2: Move and update `apps/ai/tests/billing.spec.ts` to `apps/ai/src/billing/billing.service.spec.ts`**

Move the file and update the import:
- `import { BillingService } from '../src/billing/billing.service';` → `import { BillingService } from './billing.service';`

- [ ] **Step 3: Move and update `apps/ai/tests/health.spec.ts` to `apps/ai/src/health.controller.spec.ts`**

Move the file and update the imports:
- `import { HealthController } from '../src/health.controller';` → `import { HealthController } from './health.controller';`
- `import type { UsageService } from '../src/usage/usage.service';` → `import type { UsageService } from './usage/usage.service';`

- [ ] **Step 4: Run Vitest on the migrated test files**

Run: `pnpm exec vitest run --config apps/ai/vitest.config.mts apps/ai/src/ai/chat-request.dto.spec.ts apps/ai/src/billing/billing.service.spec.ts apps/ai/src/health.controller.spec.ts`
Expected: PASS (all 19 tests pass: 3 request, 14 billing, 2 health).

- [ ] **Step 5: Commit**

```bash
git add apps/ai/src/ai/chat-request.dto.spec.ts apps/ai/src/billing/billing.service.spec.ts apps/ai/src/health.controller.spec.ts
git rm apps/ai/tests/request.spec.ts apps/ai/tests/billing.spec.ts apps/ai/tests/health.spec.ts
git commit -m "refactor(ai): co-locate request, billing, and health unit tests"
```

---

### Task 3: Co-locate `usage.service.spec.ts`

**Files:**
- Create: `apps/ai/src/usage/usage.service.spec.ts` (moved from `apps/ai/tests/usage.spec.ts`)
- Delete: `apps/ai/tests/usage.spec.ts`

**Interfaces:**
- Consumes: `src/config.ts`, `src/usage/usage.scripts.ts`, `src/usage/usage.service.ts`
- Produces: Co-located unit and Redis integration test for usage accounting.

- [ ] **Step 1: Move and update `apps/ai/tests/usage.spec.ts` to `apps/ai/src/usage/usage.service.spec.ts`**

Move the file and update the relative imports:
- `import { usageConfigSchema } from '../src/config';` → `import { usageConfigSchema } from '../config';`
- `import { RELEASE, RESERVE, START, THROTTLE } from '../src/usage/usage.scripts';` → `import { RELEASE, RESERVE, START, THROTTLE } from './usage.scripts';`
- `import { usagePeriod } from '../src/usage/usage.service';` → `import { usagePeriod } from './usage.service';`

- [ ] **Step 2: Run Vitest on `usage.service.spec.ts`**

Run: `pnpm exec vitest run --config apps/ai/vitest.config.mts apps/ai/src/usage/usage.service.spec.ts`
Expected: PASS (2 unit tests pass, 7 Redis integration tests skipped when `AI_TEST_REDIS_URL` is unset).

- [ ] **Step 3: Commit**

```bash
git add apps/ai/src/usage/usage.service.spec.ts
git rm apps/ai/tests/usage.spec.ts
git commit -m "refactor(ai): co-locate usage service tests"
```

---

### Task 4: Merge `access.spec.ts` and `stream.spec.ts` into `ai.service.spec.ts` and Clean Up `apps/ai/tests/`

**Files:**
- Create: `apps/ai/src/ai/ai.service.spec.ts` (combining `apps/ai/tests/access.spec.ts` and `apps/ai/tests/stream.spec.ts`)
- Delete: `apps/ai/tests/access.spec.ts`
- Delete: `apps/ai/tests/stream.spec.ts`

**Interfaces:**
- Consumes: `src/ai/ai.service.ts`, `src/ai/chat-request.dto.ts`, `src/ai/tools/custom-tools.ts`, `src/common/ai-error.ts`
- Produces: Comprehensive co-located test suite for `AiService` (both authorization checks and stream lifecycle).

- [ ] **Step 1: Create `apps/ai/src/ai/ai.service.spec.ts` merging both suites**

Combine `access.spec.ts` and `stream.spec.ts` with updated imports:
- `AiService`: `import { AiService } from './ai.service';`
- `BlockNoteChatBody`: `import type { BlockNoteChatBody } from './chat-request.dto';`
- `createCustomTools`: `import { createCustomTools } from './tools/custom-tools';`
- `AiError`: `import { AiError } from '../common/ai-error';`
- Preserve the hoisted `@ai-sdk/openai-compatible` mock at top:
  ```ts
  const provider = vi.hoisted(() => ({ model: undefined as any }));
  vi.mock('@ai-sdk/openai-compatible', () => ({
    createOpenAICompatible: () => () => provider.model,
  }));
  ```
- Include both suites:
  1. `describe('AI workspace authorization', () => { ... })` (all 7 tests from `access.spec.ts`)
  2. `describe('AI stream lifecycle', () => { ... })` (all 5 tests from `stream.spec.ts`)

- [ ] **Step 2: Delete `apps/ai/tests/access.spec.ts` and `apps/ai/tests/stream.spec.ts`, and remove directory `apps/ai/tests/`**

Remove:
- `apps/ai/tests/access.spec.ts`
- `apps/ai/tests/stream.spec.ts`
- Verify `apps/ai/tests/` directory is empty and remove it (`rmdir apps/ai/tests`).

- [ ] **Step 3: Run Vitest on `apps/ai/src/ai/ai.service.spec.ts`**

Run: `pnpm exec vitest run --config apps/ai/vitest.config.mts apps/ai/src/ai/ai.service.spec.ts`
Expected: PASS (all 12 tests pass: 7 authorization tests + 5 stream lifecycle tests).

- [ ] **Step 4: Commit**

```bash
git add apps/ai/src/ai/ai.service.spec.ts
git rm apps/ai/tests/access.spec.ts apps/ai/tests/stream.spec.ts
git commit -m "refactor(ai): merge access and stream specs into co-located ai.service.spec.ts"
```

---

### Task 5: Full Verification and Cleanup

**Files:**
- Verify all modified files across `apps/ai/`

- [ ] **Step 1: Run full test suite via Nx**

Run: `pnpm exec nx test ai --skip-nx-cache`
Expected: PASS (5 test files, 39 passed, 7 skipped).

- [ ] **Step 2: Run typecheck via Nx**

Run: `pnpm exec nx typecheck ai --skip-nx-cache`
Expected: PASS (clean declaration emit and spec typecheck).

- [ ] **Step 3: Run linting via Nx**

Run: `pnpm exec nx lint ai --fix`
Expected: PASS with 0 errors.

- [ ] **Step 4: Run build via Nx to verify bundle integrity**

Run: `pnpm exec nx build ai --skip-nx-cache`
Expected: PASS (rspack builds application bundle; verify `apps/ai/dist` does not contain test files).

- [ ] **Step 5: Verify working tree clean and tests directory absent**

Run: `ls -la apps/ai/tests 2>&1`
Expected: "No such file or directory"
Run: `git status`
Expected: clean working directory.
