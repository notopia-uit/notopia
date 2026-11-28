# Docs publishing and REST reference

## Intent

Publish one docs artifact that includes the existing Scalar application and a browsable, searchable REST reference with real Markdown for AI readers. Keep `apps/api-web` independently runnable, keep generated output out of `docs/src`, and preserve the production and pull-request site bases. The API landing page belongs to VitePress; Scalar gets its own published subroute.

Prefer the least code and configuration that meet these goals. Keep the existing sidebar plugin and use VitePress's native sidebar configuration for the REST pages; don't fork the plugin, replace the theme, or maintain a second endpoint list.

## Approaches

- **Selected: assemble in the VitePress output.** Nx builds Scalar, and VitePress copies that artifact into its own output at build end. The `docs` project owns publishing and preview cleanup; the separate `page` project is no longer needed.
- **Stage Scalar in an ignored `docs/src/public` directory.** VitePress would copy it automatically and see it during link checking, but builds would have to manage stale generated state in the source tree. Local dev could accidentally serve assets from a different `DOCS_BASE`.
- **Keep the `page` assembler.** This preserves the existing separate artifacts and deployment targets, but leaves the duplication that prompted the change. Replacing `api-web` with a different Scalar wrapper is out of scope.

## Publishing and cache

`nx build docs` builds the joined OpenAPI bundle, then `api-web`, then VitePress. A small build-end integration copies `apps/api-web/dist` into `<site outDir>/api/openapi-scalar` and fails when the expected Scalar entry point is missing. It does not spawn Nx from VitePress, write to `src/public`, or copy a second docs tree. VitePress owns `/api/` and `/api/index.html` through the existing `src/api/index.md`; link to Scalar from the landing page and site nav. The published Scalar bundle and its standalone preview use `${DOCS_BASE}api/openapi-scalar/` for assets. Standalone Scalar dev keeps its existing URL. Use a narrow dead-link exception for the new Scalar entry point, since the file is copied after page validation.

The two Nx builds remain independently cacheable. Make task ordering and the project dependency graph explicit so changes to the API spec or Scalar app affect the docs deployment. Track the joined spec and Scalar build outputs as generated dependency inputs, keep `DOCS_BASE` in both build inputs, and declare the complete docs output (`docs/.vitepress/dist`) as the docs build artifact. Include route loaders, VitePress helpers, and relevant dependency versions in the docs cache key. Avoid hashing generated output as source.

Move `deployPage` and `cleanup-previews` to `docs`, with an explicit deploy directory. Update CI affected detection as needed, the PR preview and cleanup commands, and the production publishing script invocation. Preserve the existing production publish behavior, preview prefixes and cleanup settings. Remove the tracked `page/` project and its workspace and lockfile entries once its consumers have moved; leave `apps/api-web` in place. `nx preview docs` serves the full artifact. `nx dev docs` needs the OpenAPI bundle for its reference pages, but not a Scalar build or server; `nx dev api-web` stays separate.

## REST reference

```text
/api/                           VitePress API landing page
/api/openapi-scalar/            Published Scalar application
/api/rest                       VitePress REST overview, grouped by OpenAPI tag
/api/rest/operations/<id>        Interactive endpoint page
/api/rest/operations/<id>.md     Plain Markdown for the same endpoint
/api/proto                      Existing gRPC reference, unchanged
```

Keep the existing `/api/rest` link as the overview. Generate one VitePress route per `operationId` from the existing joined spec and render it with `OAOperation`. Use a single tag-grouped overview and sidebar rather than duplicating the same endpoints in separate tag pages. Derive the endpoint titles, method and path labels, tag groups, and links from the spec; reuse the route mapping for the sidebar and Markdown output. Keep operation URLs independent of tag names so regrouping does not change them. Preserve the existing general sidebar and use Pagefind for endpoint discovery.

Show request and response schemas and examples on the pages where readers use them. No separate schema catalog, hand-maintained endpoint list, or changes to the API contract or codegen.

## LLM output

Give each generator a clear set of routes. `vitepress-plugin-llms` owns general page Markdown and the site-wide `llms.txt` and `llms-full.txt`; exclude the dynamic endpoint templates and generated endpoint routes from its output. `vitepress-openapi` owns the Markdown for `/api/rest.md` and `/api/rest/operations/*`, with `llmsTxt: false` so it cannot replace the site-wide indexes. Use its generated introduction for the overview Markdown and retain the overview's entry in the general index. Register its dev middleware before the general plugin and write the reference Markdown in VitePress `buildEnd`.

Make the REST overview discoverable from `llms.txt` and make its Markdown link to the generated endpoint Markdown under the correct site base. The general `llms-full.txt` does not need to duplicate the expanded endpoint reference; it must not contain raw endpoint component placeholders. Nothing under `docs/ai` or `docs/superpowers` is published.

## Acceptance

- Build with the production base and a representative PR base. Confirm `/api/` serves the VitePress landing page, `/api/openapi-scalar/` serves Scalar and its assets, and `/api/proto` still works.
- Verify a representative endpoint across tags: sidebar and overview links, Pagefind results, request/response schemas or examples present in the spec, and the matching Markdown without Vue placeholders.
- Verify site-wide LLM files link to real Markdown, no generator overwrites the other's files, and the generated links resolve under both bases. Check endpoint Markdown in `nx dev docs` without Scalar running.
- Verify repeat builds restore a complete Nx artifact from cache, switching `DOCS_BASE` cannot hit the wrong cache entry, and API or Scalar changes invalidate the docs build. Check PR preview, cleanup, and production workflow commands without publishing during verification.
