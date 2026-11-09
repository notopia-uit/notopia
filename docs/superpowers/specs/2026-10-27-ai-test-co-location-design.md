# Design: Co-locate Unit Tests with Source in apps/ai

## Goal

Refactor `apps/ai` to move unit test files from the standalone `tests/` directory into `src/` co-located next to their corresponding source files, aligning with conventions established in `apps/search-worker` and `apps/document`.

## Scope and Decisions

1. **Test Co-location Mapping**:
   - `apps/ai/tests/request.spec.ts` → `apps/ai/src/ai/chat-request.dto.spec.ts`
   - `apps/ai/tests/billing.spec.ts` → `apps/ai/src/billing/billing.service.spec.ts`
   - `apps/ai/tests/health.spec.ts` → `apps/ai/src/health.controller.spec.ts`
   - `apps/ai/tests/usage.spec.ts` → `apps/ai/src/usage/usage.service.spec.ts`
   - `apps/ai/tests/access.spec.ts` + `apps/ai/tests/stream.spec.ts` → merged into `apps/ai/src/ai/ai.service.spec.ts`
   - Delete `apps/ai/tests/` once all tests are migrated.

2. **TypeScript & Build Configuration**:
   - Create `apps/ai/tsconfig.spec.json` extending `../../tsconfig.base.json` with types `["vitest/globals", "node", "vitest", "vitest/importMeta"]`, path `#/*` mapping to `./src/*`, includes `["vitest.config.mts", "src/**/*.spec.ts"]`, outDir `./out-tsc/vitest`, and project reference to `./tsconfig.app.json`.
   - Update `apps/ai/tsconfig.app.json` to exclude `["src/**/*.spec.ts"]` so that production declaration builds (`tsc --build --emitDeclarationOnly`) do not emit type declarations for test specs into `out-tsc/app`.
   - Update `apps/ai/tsconfig.json` to include `./tsconfig.spec.json` under `references`.
   - Update `apps/ai/vitest.config.mts` to use `tsconfigFile: 'tsconfig.spec.json'` in `swc.vite(...)` and set `include: ['src/**/*.spec.ts']`.

3. **Constraints**:
   - Zero changes to actual test logic or assertions.
   - All relative imports within test files must be updated to reflect their new locations relative to `src/`.
   - `pnpm exec nx test ai`, `pnpm exec nx typecheck ai`, `pnpm exec nx lint ai`, and `pnpm exec nx build ai` must all pass cleanly.
