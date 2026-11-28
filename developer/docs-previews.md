---
url: /notopia/developer/docs-previews.md
---
# Documentation previews

CI builds affected documentation for pull requests. Same-repository pull requests, except
Dependabot pull requests, also receive a preview link. Fork and Dependabot pull requests are
validated without publishing.

Production and previews share the `gh-pages` branch. The dedicated cleanup workflow runs on a
schedule or manually. It removes preview
directories whose most recent Git commit author timestamp is more than seven days old, even if
the pull request is still open. A subsequent update recreates a removed preview; an unchanged
deployment does not refresh its timestamp. Cleanup attempts to update the corresponding PR
comment.

## Local docs

Run `nx dev docs` for VitePress and the REST reference. Restart it after editing an OpenAPI spec
to rebuild the bundle and refresh endpoint routes and Markdown. For the standalone Scalar
explorer, run `nx dev api-web`.

Run `nx preview docs` to build and serve the full artifact, including Scalar and the REST
Markdown. The default base is `/notopia/`; set `DOCS_BASE=/notopia/pr/123/` to preview a PR base.
