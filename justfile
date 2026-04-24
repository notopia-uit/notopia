all parallel="4" exclude="tag:scope:docs":
    pnpm exec nx run-many -t lint typecheck build gen bundle --parallel={{ parallel }} --exclude={{ exclude }}

up-api configuration="development":
    pnpm exec nx run-many \
      -t dev \
      --projects=document,note,authorization,search-worker \
      --configuration={{ configuration }} \
      --parallel=99

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
      down

export-authentik-blueprint:
    docker exec notopia-authentik_worker ak export_blueprint

shadcnadd component:
    pnpm exec shadcn add {{ component }} -c ./packages/ui/
