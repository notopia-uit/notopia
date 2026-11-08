# Documentation

The public documentation site uses VitePress. Its source pages live in `docs/src/`; the sidebar
is generated from that tree. Keep pages focused, link related material, and verify builds after
structural changes.

## Layout

- `src/`: rendered VitePress pages, including architecture, API, service, deployment, convention,
  and story content.
- `.vitepress/`: site configuration, theme, and plugins. Do not treat these files as content pages.
- `ai/` and `superpowers/`: internal agent guidance, plans, and working material. Keep these outside
  `src/` so VitePress does not render them.

## Page metadata

Use YAML frontmatter when a page needs metadata. `authorship` is optional and must be a non-empty
list containing `ai`, `human`, or both. For example:

```yaml
---
authorship:
  - ai
---
```

Do not use `custom VitePress config` as an authorship value; configuration belongs under
`.vitepress/`, not in page metadata.

## Editing

- Preserve technical accuracy and existing authorship context.
- Update links when moving pages; use VitePress routes for links between rendered pages.
- Keep AI/superpowers process notes outside `src/`.
- Run the docs build for structural changes and check for broken links.
