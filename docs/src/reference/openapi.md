---
aside: false
outline: false
title: Notopia OpenAPI
---

<script setup>
import { onBeforeMount, onBeforeUnmount } from 'vue'
import { useTheme } from 'vitepress-openapi/client'
import spec from "@notopia-uit/api/openapi" with { type: "json" };

onBeforeMount(() => {
    useTheme({
        server: {
            allowCustomServer: true,
        },
    })
})
</script>

<OASpec :spec="spec" />
