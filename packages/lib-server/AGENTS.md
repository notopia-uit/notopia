# lib-server (`@notopia-uit/lib-server`)

Node-only shared code for NestJS services.

| Entry | Contents |
| --- | --- |
| `./block-note` | BlockNote server schema (`createSchema`) |
| `./locks` | `LocksSetupModule.forRoot({ redisUrl?, ttl? })` over `@nestjs/locks`, `RedisLockStore` |

## Locks

- With `redisUrl`, locks use `RedisLockStore`. That Redis must run with `noeviction` and
  persistence, and the app must call `app.enableShutdownHooks()`.
- Without `redisUrl`, an in-memory store is used and a warning is logged. **This is
  intended, also in production**: it keeps single-instance deployments and local
  development free of an extra Redis. Do not turn it into a startup error. Running more
  than one instance without `redisUrl` lets every instance run the same cron.
