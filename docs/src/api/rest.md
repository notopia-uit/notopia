---
aside: false
outline: false
title: REST API
order: 1
---

<script setup>
import { withBase } from 'vitepress';
import { data as groups } from './rest.data';
</script>

# REST API

Browse the endpoints by tag. Each endpoint page includes request and response details.

<section v-for="{ tag, operations } in groups" :key="tag">
  <h2>{{ tag }}</h2>
  <ul>
    <li v-for="{ operationId, title, method, path, href } in operations" :key="operationId">
      <a :href="withBase(`${href}.html`)">{{ title }}</a> — <code>{{ method }} {{ path }}</code>
    </li>
  </ul>
</section>
