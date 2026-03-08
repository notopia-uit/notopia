# Architecture Diagram

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
    icon: https://simpleicons.org/icons/googlechrome.svg
  }
}

gateway: Gateway {
  icon: https://simpleicons.org/icons/traefikproxy.svg
}

apps: Apps {
  web: Web {
    icon: https://simpleicons.org/icons/nextdotjs.svg
  }
}

services: Services {
  identity_provider: Identity Provider {
    icon: https://simpleicons.org/icons/authentik.svg
  }

  object_storage: Object Storage {
    icon: https://simpleicons.org/icons/minio.svg
  }

  document: Document Service Group {
    document_service: Document Service {
      icon: https://simpleicons.org/icons/nodedotjs.svg
    }
    main_database: Main Database {
      icon: https://simpleicons.org/icons/postgresql.svg
    }
    document_service -> main_database
  }

  note: Note Service Group {
    note_service: Note Service {
      icon: https://simpleicons.org/icons/go.svg
    }
    database: Database {
      icon: https://simpleicons.org/icons/postgresql.svg
    }
    note_service -> database
  }

  authorization: Authorization Service Group {
    authorization_service: Authorization Service {
      icon: https://simpleicons.org/icons/go.svg
    }
    database: Database {
      icon: https://simpleicons.org/icons/postgresql.svg
    }
    authorization_service -> database
  }

  search: Search Service Group {
    search_service: Meilisearch {
      icon: https://simpleicons.org/icons/meilisearch.svg
    }
    search_sync_service: Sync Service {
      icon: https://simpleicons.org/icons/go.svg
    }
    search_sync_service_database: Sync Database {
      icon: https://simpleicons.org/icons/redis.svg
    }
    search_sync_service -> search_service
    search_sync_service -> search_sync_service_database
  }

  event_bus: Event Bus {
    message_broker: Message Broker {
      icon: https://simpleicons.org/icons/apachekafka.svg
    }

    pub_sub: Pub/Sub {
      icon: https://simpleicons.org/icons/redis.svg
    }
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
gateway -> services.search

services.document -> services.note
services.document <-> services.event_bus
services.document -> services.authorization

services.note -> services.identity_provider
services.note -> services.object_storage
services.note <-> services.event_bus
services.note -> services.authorization

services.authorization <- services.event_bus

services.search <- services.event_bus

style.border-radius: 15
*.style.border-radius: 15
*.*.style.border-radius: 15
*.*.*.style.border-radius: 15
```

<!-- diagram id="architecture-diagram" -->
