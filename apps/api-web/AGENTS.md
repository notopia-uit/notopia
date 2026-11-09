# API Reference Site (nx: `api-web`)

Vite + React page renders whole OpenAPI spec with Scalar. Served at API host root (`api.notopia.localhost`). Essentially one component: `src/app/app.tsx`.

Almost nothing to change here. Spec imported as `@notopia-uit/api/openapi` (JSON import) — **content changes belong in `api/`**, not this app. `nx dev api-web` depends on `watch-bundle:json` for `api/note` and `api/document`, so spec yaml edits rebuild bundle and hot-reload page.

Edit app only for Scalar changes — `configuration` options, extra `sources`, theming, or page shell. When adding API surface under `api/`, add it to `sources` here and `dev` target's `dependsOn` in `project.json`.