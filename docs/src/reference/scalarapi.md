---
layout: page
aside: false
outline: false
sidebar: false
navbar: false
title: Scalar OpenAPI
---

<script setup>
import '@scalar/api-reference/style.css'
import { ApiReference } from '@scalar/api-reference'
import openapi from '@notopia-uit/api/openapi' with { type: 'json' };
</script>

<ClientOnly>
  <ApiReference
    :configuration="{
      sources: [
        {
          content: openapi,
          default: true,
          title: 'Notopia API',
        },
      ],
      persistAuth: true,
      telemetry: false,
    }"
  />
</ClientOnly>
