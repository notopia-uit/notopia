---
aside: false
outline: false
title: Event OpenAPI
---

<script setup>
import { onBeforeMount, onBeforeUnmount } from 'vue'
import { useTheme } from 'vitepress-openapi/client'
import eventApiSpec from "@notopia/event-api" with { type: "json" };

onBeforeMount(() => {
    useTheme({
        server: {
            allowCustomServer: true,
        },
    })
})
</script>

<OASpec :spec="eventApiSpec" />
