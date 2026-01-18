# Architecture Diagram

```d2
direction: right

vars: {
  d2-config: {
    layout-engine: elk
  }
}

apps: Apps {
  web: Web {
    icon: https://simpleicons.org/icons/nextdotjs.svg
  }
}

services: Services {
  gateway: Gateway {
    icon: https://simpleicons.org/icons/traefikproxy.svg
  }

  authorization_server: Authorization Server {
    icon: https://simpleicons.org/icons/keycloak.svg
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

  search: Search Group {
    meilisearch: Meilisearch {
      icon: https://simpleicons.org/icons/meilisearch.svg
    }
    meilisearch_sync_service: Sync Service {
      icon: https://simpleicons.org/icons/go.svg
    }
    meilisearch_sync_service_database: Sync Database {
      icon: https://simpleicons.org/icons/redis.svg
    }
    meilisearch_sync_service -> meilisearch
    meilisearch_sync_service -> meilisearch_sync_service_database
  }

  kafka: Kafka {
    icon: https://simpleicons.org/icons/apachekafka.svg
  }
}

apps -> services.gateway
apps -> services.authorization_server
apps -> services.object_storage

services.gateway -> services.edit.edit_service
services.gateway -> services.note.note_service
services.gateway -> services.search.meilisearch

services.edit.edit_service -> services.note.note_service
services.edit.edit_service -> services.kafka

services.note.note_service -> services.authorization_server
services.note.note_service -> services.object_storage
services.note.note_service -> services.kafka

services.kafka <- services.search.meilisearch_sync_service
```

<!-- diagram id="architecture-diagram" -->
