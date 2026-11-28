---
aside: false
---

<script setup>
import spec from '@notopia-uit/api/openapi' with { type: 'json' };
import { useData } from 'vitepress';
import { useTheme } from 'vitepress-openapi/client';
import { onBeforeMount } from 'vue';

const { params } = useData();
onBeforeMount(() => useTheme({ server: { allowCustomServer: true } }));
</script>

<OAOperation :spec="spec" :operation-id="params.operationId" />
