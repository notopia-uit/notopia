---
aside: false
outline: false
title: Common OpenAPI
---

<script setup>
import { onBeforeMount, onBeforeUnmount } from 'vue'
import { useTheme } from 'vitepress-openapi/client'
import commonApiSpec from "@notopia/common-api" with { type: "json" };

onBeforeMount(() => {
    useTheme({
        server: {
            allowCustomServer: true,
        },
    })
})
</script>

<OASpec :spec="commonApiSpec" />
