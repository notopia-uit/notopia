---
authorship:
  - ai
---

# The Documentation Pages That Wouldn't Disappear

_Why publishing PR previews changed how we deploy production documentation._

Written with AI, November 2026. The preview implementation requires `nx-github-pages` 1.3.1;
the dependency update supersedes PR #185.

## One Branch, Two Kinds of Content

Notopia publishes its documentation and API reference from the `gh-pages` branch. A production
build belongs at the branch root and is served under `/notopia/`. A pull request preview belongs
under `pr/<number>/` and is served under `/notopia/pr/<number>/`.

That shared branch gives us two requirements that pull in opposite directions:

- A production deployment must replace the old production site, including deleting pages that
  no longer exist in the new build.
- The same deployment must leave the `pr/` directory alone, so it does not erase other people's
  previews.

Publishing the build as a complete replacement satisfies the first requirement but loses the
previews. Keeping everything from the old branch protects previews but also keeps obsolete pages.

## The Ghost Page

We looked at `nx-github-pages` 1.3.1, which supports both preview deployments and scheduled preview
cleanup. Its documentation recommends `syncWithBaseBranch` to preserve previews during production
deployments.

The [1.3.1 deploy executor][executor] builds a fresh Git history from the output directory and
merges the existing `gh-pages` branch into it. The default merge favors new files when both
histories have a conflicting file. But a file present only in the old branch is not a conflict:
Git can retain it in the merge. The executor then uses a force-push.

For example:

```text
Old gh-pages:  index.html  retired-guide.html  pr/123/index.html
New build:     index.html
Merged result: index.html  retired-guide.html  pr/123/index.html
```

The preview survived, but the deleted guide did too. It could remain available at its old URL
until someone explicitly removes it. The plugin's [README][readme] describes synced production
deploys as merging rather than force-pushing; the published executor source still invokes
`git push -f`.

This is about what the current site serves, not about erasing Git history: removing a page from the
branch tip does not guarantee that previously published content is gone from history, caches, or
other copies.

## What We Chose

We kept the plugin for preview publishing and age-based preview cleanup. For production,
`scripts/publish-docs.sh` clones the existing `gh-pages` branch, removes the old production files
while preserving `.git` and `pr/`, copies in the new build, and pushes a normal commit. It rejects
a production build that itself contains a top-level `pr` directory, because that path is reserved
for previews.

The normal push also rejects a conflicting update instead of overwriting it. Production, preview,
and cleanup jobs share a non-cancelling concurrency group, but a writer outside that group could
still race; in that case the failed production deployment needs to be rerun.

A disposable Git remote test exercised the important case: two preview directories survived a
production redeploy, while an old production page disappeared. Live preview comments and scheduled
cleanup still need verification on GitHub.

## What Else Made Previews Work

The deployment directory is only half the problem. VitePress, diagrams, the API site, and search
must all use the preview prefix when the site is built. `DOCS_BASE` supplies that prefix, and the
docs, API, and combined page build cache keys include it. Otherwise an Nx cache hit could restore
a production build into a preview, or the reverse.

Preview cleanup is age-based rather than PR-state-based. The plugin uses the most recent Git
commit author timestamp for a preview directory, so a preview for an open PR can expire; a
subsequent PR update recreates it. An unchanged deploy does not refresh its timestamp.

## Takeaways

- A merge that preserves unrelated preview directories can also preserve deleted production files.
- A shared publishing branch needs both a reserved preview path and coordinated writers.
- A build-time URL prefix is part of the build's cache identity.
- Check the published executor's behavior as well as its documentation before relying on
  deployment guarantees.

[executor]: https://github.com/AgentEnder/nx-github-pages/blob/293f6bdec51acde86982c0ffb5fbc6d80e957c44/packages/nx-github-pages/src/executors/deploy/executor.ts
[readme]: https://github.com/AgentEnder/nx-github-pages/blob/293f6bdec51acde86982c0ffb5fbc6d80e957c44/README.md
