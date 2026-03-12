all parallel="4" exclude="tag:scope:docs":
    pnpm exec nx run-many -t lint build gen bundle --parallel={{ parallel }} --exclude={{ exclude }}

up-api configuration="development":
    pnpm exec nx run-many \
      -t dev \
      --projects=document,note,authorization,searchworker \
      --configuration={{ configuration }} \
      --parallel=99

lazydocker COMPOSE_PROFILES="*":
    COMPOSE_PROFILES={{ COMPOSE_PROFILES }} lazydocker
