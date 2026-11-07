<script setup lang="ts">
import { useData } from 'vitepress';

import type { Authorship } from './authorship';

const { frontmatter } = useData();

const labels: Record<Authorship, string> = {
  human: '👤 Human',
  ai: '🤖 AI',
};
</script>

<template>
  <div v-if="frontmatter.authorship?.length" class="authorship">
    <span v-for="author in frontmatter.authorship" :key="author" class="badge">
      {{ labels[author as Authorship] }}
    </span>
  </div>
</template>

<style scoped>
.authorship {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.badge {
  border-radius: 1rem;
  background: var(--vp-c-default-soft);
  color: var(--vp-c-text-2);
  padding: 0.2rem 0.6rem;
  font-size: 0.8rem;
  font-weight: 500;
}
</style>
