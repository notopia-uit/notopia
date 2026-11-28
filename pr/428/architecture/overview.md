---
url: /notopia/pr/428/architecture/overview.md
---

# System Architecture

Notopia is a set of small services behind a single gateway (Traefik). Go services (note,
authorization, billing) and NestJS services (document, search worker, AI) use gRPC for
synchronous service calls and Redpanda for events. The diagram shows logical dependencies;
local Compose provides infrastructure, while application services can run on the host.

Compose includes dedicated Redis instances for note, document, billing and AI. Document
Redis is optional in configuration and supports collaborative scaling and distributed locks.
Search-worker Redis is optional and not provisioned by Compose: without it, deletion
tombstones use in-memory storage suitable only for a single worker. See
[Search Worker](../services/search-worker/) for the limitations.

See [Messaging](./messaging) for the event backbone and [Auth and Gateway](./auth-and-gateway)
for how requests are authenticated.

```d2
direction: right

vars: {
  d2-config: {
    layout-engine: elk
    theme-overrides: {
      N1: "#4c4f69"
      N2: "#5c5f77"
      N4: "#acb0be"
      N5: "#ccd0da"
      N7: "#eff1f5"
      B1: "#4c4f69"
      B2: "#6c6f85"
      B3: "#bcc0cc"
      B4: "#ccd0da"
      B5: "#dce0e8"
      B6: "#eff1f5"
      AA4: "#1e66f5"
      AA5: "#7287fd"
      AB4: "#8839ef"
      AB5: "#dc8a78"
    }
  }
}

user: User {
  browser: Browser {
    icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/googlechrome.svg
  }
}

payment_provider: Payment Provider (Stripe sandbox) {
  icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/stripe.svg
}

model_provider: Model Provider {
  label: LLM provider (remote or optional local Ollama)
}

gateway: Gateway {
  icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/traefikproxy.svg
}

apps: Apps {
  web: Web {
    icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/nextdotjs.svg
  }
}

services: Services {
  identity_provider: Identity Provider {
    icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/authentik.svg
  }

  object_storage: Object Storage {
    icon: https://cdn.jsdelivr.net/gh/selfhst/icons@main/svg/rustfs.svg
  }

  document: Document Service Group {
    document_service: Document Service {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/nodedotjs.svg
    }
    database: Database {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/postgresql.svg
    }
    pub_sub: Redis (optional scaling and locks) {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/redis.svg
    }
    document_service -> database
    document_service -> pub_sub
  }

  note: Note Service Group {
    note_service: Note Service {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/go.svg
    }
    database: Database {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/postgresql.svg
    }
    pub_sub: Pub/Sub {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/redis.svg
    }
    note_service -> database
    note_service -> pub_sub
  }

  authorization: Authorization Service Group {
    authorization_service: Authorization Service {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/go.svg
    }
    database: Database {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/postgresql.svg
    }
    authorization_service -> database
  }

  billing: Billing Service Group {
    billing_service: Billing Service {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/go.svg
    }
    database: Database {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/postgresql.svg
    }
    cron_lock: Cron Lock {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/redis.svg
    }
    billing_service -> database
    billing_service -> cron_lock
  }

  ai: AI Service Group {
    ai_service: AI Service {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/nodedotjs.svg
    }
    redis: Redis (quota and continuations) {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/redis.svg
    }
    ai_service -> redis
  }

  search: Search Service Group {
    search_service: Meilisearch {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/meilisearch.svg
    }
    search_sync_worker: Sync Worker {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/nodedotjs.svg
    }
    search_sync_worker_database: Optional Redis tombstones {
      icon: https://cdn.jsdelivr.net/npm/simple-icons@16/icons/redis.svg
    }
    search_sync_worker -> search_service
    search_sync_worker -> search_sync_worker_database: optional
  }

  event_bus: Event Bus {
    message_broker: Message Broker {
      icon: https://raw.githubusercontent.com/redpanda-data/redpanda/d17f36726370db7c7044cb12ecc1c46841e8bfa6/docs/icon-redpanda.svg
    }
  }

  monitoring: Monitoring {
    alloy: Alloy {
      icon: https://raw.githubusercontent.com/grafana/alloy/refs/heads/main/docs/sources/assets/alloy_icon_orange.svg
    }
    prometheus: Prometheus {
      icon: https://raw.githubusercontent.com/prometheus/prometheus/refs/heads/main/documentation/images/prometheus-logo.svg
    }
    loki: Loki {
      icon: https://cdn.jsdelivr.net/gh/selfhst/icons@main/svg/loki.svg
    }
    tempo: Tempo {
      icon: https://cdn.jsdelivr.net/gh/selfhst/icons@main/svg/grafana-tempo.svg
    }
    grafana: Grafana {
      icon: https://raw.githubusercontent.com/grafana/grafana/refs/heads/main/public/img/grafana_icon.svg
    }
    alloy -> prometheus
    alloy -> loki
    alloy -> tempo
    prometheus <- grafana
    loki <- grafana
    tempo <- grafana
  }
}

user.browser -> gateway

apps -> services.identity_provider
apps -> services.object_storage

gateway -> apps
gateway -> services.document
gateway -> services.note
gateway -> services.identity_provider
gateway -> services.object_storage
gateway -> services.billing
gateway -> services.ai
gateway -> services.search
gateway -> services.monitoring

services.document -> services.note
services.document <-> services.event_bus
services.document -> services.authorization
services.document -> services.object_storage
services.document -> services.billing

services.note <-> services.event_bus
services.note -> services.authorization
services.note -> services.billing

services.authorization <- services.event_bus

services.billing <-> services.event_bus
services.billing -> services.authorization
services.billing -> payment_provider: checkout, refunds
payment_provider -> gateway: webhooks

services.search <- services.event_bus
services.ai -> services.note
services.ai -> services.authorization
services.ai -> services.billing
services.ai -> services.search.search_service
services.ai -> model_provider

style.border-radius: 15
*.style.border-radius: 15
*.*.style.border-radius: 15
*.*.*.style.border-radius: 15
```
