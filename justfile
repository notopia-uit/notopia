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

helm-setup:
    set -e; if ! helm plugin list | grep -qE '^unittest[[:space:]]+1\.2\.1([[:space:]]|$)'; then \
      if helm plugin list | grep -qE '^unittest[[:space:]]'; then helm plugin uninstall unittest; fi; \
      helm plugin install https://github.com/helm-unittest/helm-unittest --version 1.2.1; \
    fi

seed: seed-authorization seed-search seed-billing

seed-authorization:
    nx run authorization:seed

seed-search:
    nx run search-worker:seed

seed-billing:
    nx run billing:migrate
    nx run billing:seed

fixtures: seed fixtures-authorization fixtures-note fixtures-document fixtures-search fixtures-billing

fixtures-note db-connection=note-db-connection: prepare-fixtures-note
    psql {{ db-connection }} \
      -f ./internal/notefixturegen/fixtures.sql \
      -f ./internal/notefixturegen/fixtures.gen.sql

prepare-fixtures-note:
    nx run note:goose:up
    go run ./cmd/notefixturegen/...

fixtures-document: prepare-fixtures-document
    nx run document:db:sync
    nx run document:fixtures

prepare-fixtures-document:
    go run ./apps/document/database/transform-data.go

fixtures-search: prepare-fixtures-search
    nx run search-worker:fixtures

prepare-fixtures-search:
    nx run search-worker:fixtures:transform

fixtures-billing:
    nx run billingfixtures:run

fixtures-authorization:
    nx run authorizationfixtures:run

stripe-listen:
    stripe listen \
      --all-snapshot \
      --forward-to http://api.notopia.localhost/billing/webhooks/stripe
