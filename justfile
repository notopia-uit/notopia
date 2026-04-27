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
