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

seed: seed-search seed-billing

seed-search:
    nx run search-worker:seed

seed-billing:
    nx run billingseed:run

fixtures: seed fixtures-note fixtures-document fixtures-search fixtures-billing

fixtures-note db-connection=note-db-connection: prepare-fixtures-note
    psql {{ db-connection }} \
      -f ./internal/notefixturegen/fixtures.sql \
      -f ./internal/notefixturegen/fixtures.gen.sql

prepare-fixtures-note:
    nx run note:goose:up
    go run ./cmd/notefixturegen/...

fixtures-document: prepare-fixtures-document
    nx run document:fixtures

prepare-fixtures-document:
    go run ./apps/document/database/transform-data.go

fixtures-search: prepare-fixtures-search
    nx run search-worker:fixtures

prepare-fixtures-search:
    nx run search-worker:fixtures:transform

fixtures-billing:
    nx run billingfixtures:run

# Events must match the switch in internal/billing/controller/http/webhook.go
stripe-listen:
    stripe listen \
      --events checkout.session.completed,customer.subscription.created,customer.subscription.updated,customer.subscription.deleted,invoice.paid,invoice.payment_failed,refund.created,refund.updated,refund.failed \
      --forward-to http://api.notopia.localhost/billing/webhooks/stripe
