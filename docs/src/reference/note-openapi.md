---
aside: false
outline: false
title: Note OpenAPI
---

<script setup>
import { onBeforeMount, onBeforeUnmount } from 'vue'
import { useTheme } from 'vitepress-openapi/client'
import noteApiSpec from "notopia-api-note" with { type: "json" };

onBeforeMount(() => {
    useTheme({
        server: {
            allowCustomServer: true,
        },
    })
})
</script>

<OASpec :spec="noteApiSpec" />
