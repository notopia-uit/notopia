# API Reference Site (nx: `api-web`)

Vite + React page that renders the whole OpenAPI spec with Scalar. Served at the API host
root (`api.notopia.localhost`). Essentially one component: `src/app/app.tsx`.

There is almost nothing to change here. The spec is imported as
`@notopia-uit/api/openapi` (a JSON import) — so **content changes belong in `api/`**, not
in this app. `nx dev api-web` depends on `watch-bundle:json` for `api/note` and
`api/document`, so edits to the spec yaml rebuild the bundle and hot-reload the page.

Edit this app only to change Scalar itself — `configuration` options, extra `sources`,
theming, or the page shell. When a new API surface is added under `api/`, add it to
`sources` here and to the `dev` target's `dependsOn` in `project.json`.
