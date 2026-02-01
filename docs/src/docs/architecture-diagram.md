# Architecture Diagram

```d2
direction: right

vars: {
  d2-config: {
    layout-engine: elk
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

  edit: Edit Service Group {
    edit_service: Edit Service {
      icon: https://simpleicons.org/icons/nodedotjs.svg
    }
    edit_service_main_database: Main Database {
      icon: https://simpleicons.org/icons/postgresql.svg
    }
    edit_service_sec_database: Secondary Database {
      icon: https://simpleicons.org/icons/redis.svg
    }
    edit_service -> edit_service_main_database
    edit_service -> edit_service_sec_database
  }

  note: Note Service Group {
    note_service: Note Service {
      icon: https://simpleicons.org/icons/go.svg
    }
    note_service_database: Database {
      icon: https://simpleicons.org/icons/postgresql.svg
    }
    note_service -> note_service_database
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

  authorization_database: Authorization Database {
    icon: https://simpleicons.org/icons/postgresql.svg
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
gateway -> services.edit
gateway -> services.note
gateway -> services.identity_provider
gateway -> services.object_storage
gateway -> services.search

services.edit -> services.note
services.edit <-> services.event_bus
services.edit -> services.authorization_database

services.note -> services.identity_provider
services.note -> services.object_storage
services.note <-> services.event_bus
services.note -> services.authorization_database

services.event_bus <- services.search

style.border-radius: 15
*.style.border-radius: 15
*.*.style.border-radius: 15
*.*.*.style.border-radius: 15
```

<!-- diagram id="architecture-diagram" -->
