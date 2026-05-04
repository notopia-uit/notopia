note-db-connection := 'postgresql://notopiauit:notopiauit@localhost:5433/note?sslmode=disable'

all parallel="4" exclude="tag:scope:docs":
    pnpm exec nx run-many -t lint typecheck build gen bundle --parallel={{ parallel }} --exclude={{ exclude }}

prod-svcs *args="--append":
    tmuxinator start -p ./.tmuxinator/services-prod.yaml {{ args }}

dev-svcs *args="--append":
    tmuxinator start -p ./.tmuxinator/services-dev.yaml {{ args }}

lazydocker COMPOSE_PROFILES="*":
    COMPOSE_PROFILES={{ COMPOSE_PROFILES }} lazydocker

docker-up:
    docker compose \
      up \
      -d

docker-up-monitoring:
    docker compose \
      --profile="monitoring" \
      up \
      -d

docker-stop-all:
    docker compose \
      --profile="*" \
      stop

export-authentik-blueprint:
    docker exec notopia-authentik_worker ak export_blueprint

shadcnadd component:
    pnpm exec shadcn add {{ component }} -c ./packages/ui/

seed-init: seed-init-note seed-init-document seed-init-search

[parallel]
seed-init-note: prepare-seed-note seed-note

[parallel]
seed-init-document: prepare-seed-document seed-document

[parallel]
seed-init-search: prepare-seed-search seed-search

seed-note db-connection=note-db-connection:
    psql {{ db-connection }} \
      -f ./internal/notecreateseed/seed.sql \
      -f ./internal/notecreateseed/seed.gen.sql

prepare-seed-note:
    nx run note:goose:up
    go run ./cmd/notecreateseed/...

seed-document:
    nx run document:typeorm-extension:seed

prepare-seed-document:
    go run ./apps/document/database/transform-data.go

seed-search:
    nx run search-worker:seed

prepare-seed-search:
    nx run search-worker:seed:transform
