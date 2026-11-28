---
url: /notopia/pr/428/developer/story/vitepress-openapi-ssr-usedata.md
---

# The Docs Build That Lost `useData`

*How a patch bump of `vitepress-openapi` broke page rendering in CI, and why the error pointed at VitePress itself.*

Written with AI, November 2026. Fixed by bundling `vitepress-openapi` for SSR in `docs/.vitepress/config.mts`.

***

## The Symptom

After a Renovate batch, `nx run docs:build:production` failed in CI. The client and server bundles
built fine, then the page-rendering step failed:

```
✗ rendering pages...
SyntaxError: The requested module 'vitepress' does not provide an export named 'useData'
    at loadESMFromCJS (node:internal/modules/cjs/loader)
```

Pagefind looked like the cause because `vitepress-plugin-pagefind` 0.4.26 → 0.4.27 was in the
same batch. It wasn't: that diff only changes scroll handling in `CommandList.vue`.

## Finding It

VitePress deletes `.vitepress/.temp` (the SSR bundle) when the build fails. We preloaded a small
script that made `fs.rm` skip `.temp` paths, so the bundle stayed on disk. `app.js` left only two
third-party imports external:

```
vitepress-mermaid-renderer   → import ok
vitepress-openapi/client     → useData is missing
```

`vitepress-openapi` had moved from 0.2.9 to 0.2.14 in the same batch. 0.2.13 ("theme: follow
VitePress isDark") added this to the client entry:

```js
import { useData } from 'vitepress';
```

Versions 0.2.9 and 0.2.12 have no `vitepress` import.

## Why It Breaks

While VitePress bundles code for SSR, it points `vitepress` at its client runtime, which exports
`useData`. Vite keeps third-party packages external during SSR, though, so Node loads
`vitepress-openapi/client` on its own at render time. Node then resolves `vitepress` through the
package's `exports` and gets the **node** entry, the config/build API, which has no `useData`.

```
bundled code      : 'vitepress' → client runtime   ✓ useData
external package  : 'vitepress' → dist/node/index  ✗ useData
```

## The Fix

Bundle the package so its `vitepress` import goes through VitePress's alias:

```ts
vite: {
  ssr: {
    noExternal: ['vitepress-openapi'],
  },
},
```

Pinning 0.2.12 also works, but it blocks future updates.

## Takeaways

* A VitePress plugin that imports client APIs (`useData`, `useRoute`, …) must be in
  `ssr.noExternal`, or it breaks during the SSR render step only.
* When the SSR bundle fails, keep `.temp` and import each external package in Node to find the
  one that fails.
* A Renovate batch breaks the build as a group. Check every bumped package, not just the one
  that looks most likely.
