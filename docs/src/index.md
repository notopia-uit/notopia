---
# https://vitepress.dev/reference/default-theme-home-page
layout: home
order: 0

hero:
  name: Notopia
  text: Utopia of Notes
  actions:
    - theme: brand
      text: Getting Started
      link: /getting-started/
    - theme: alt
      text: Architecture
      link: /architecture/overview
    - theme: alt
      text: Services
      link: /services/note/
---

<script setup>
import { withBase } from 'vitepress';
</script>

# Notopia

Documentation for getting started, understanding, developing, and deploying Notopia.

## Markdown for AI tools

- <a :href="withBase('/llms.txt')"><code>llms.txt</code></a> — a site-wide index linking to individual Markdown pages.
- <a :href="withBase('/llms-full.txt')"><code>llms-full.txt</code></a> — the general documentation in one file. For endpoint details, see the [API reference](./api/).
