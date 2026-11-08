# lib-server (`@notopia-uit/lib-server`)

Node-only shared code for NestJS services.

| Entry | Contents |
| --- | --- |
| `./block-note` | BlockNote server schema (`createSchema`) |
| `./locks` | `LocksSetupModule.forRoot({ redisUrl?, ttl? })` over `@nestjs/locks`, `RedisLockStore` |

## Locks

- With `redisUrl`, locks use `RedisLockStore`. Redis must use `noeviction` and
  persistence; app must call `app.enableShutdownHooks()`.
- Without `redisUrl`, uses in-memory store and logs warning. **Intentional, also in
  production**: avoids extra Redis for single-instance deployments and local
  development. Do not make startup error. Multiple instances without `redisUrl`
  each run same cron.