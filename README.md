# Notopia

<div align="center">

  <a href="https://codecov.io/gh/notopia-uit/notopia">
    <img alt="Codecov" src="https://img.shields.io/codecov/c/github/notopia-uit/notopia"/>
  </a>

| Service           |                                                                                          Quality Gate                                                                                          |                                                                                      Bugs                                                                                      |                                                                                         Code Smells                                                                                          |                                                                                          Maintainability                                                                                          |
| :---------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: |
| **Note**          |          [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |          [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |          [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |          [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |
| **Web**           |           [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |           [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |           [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |           [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |
| **Document**      |      [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |      [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |      [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |      [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |
| **Authorization** | [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) | [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) | [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) | [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) |
| **Search Worker** | [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) | [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) | [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) | [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) |

<br/>

[![Wakatime](https://wakatime.com/badge/github/notopia-uit/notopia.svg)](https://wakatime.com/badge/github/notopia-uit/notopia)

</div>

## OTEL

<https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/>

## Dev

- Do not source `.env`, because Nx will not override env that already exist. So, let Nx source itself
- Delete broken symlinks in the current directory and its subdirectories
  ```bash
  find . -type l ! -exec test -e {} \; -print -delete
  ```

## TO-DO

### Backend

- [ ] Note service
  - [ ] Routing kafka message based on metadata workspace id if partitioning or listening
  - [ ] Some handlers should rename the UserID to actor ID, and... so do the domain event?
  - [ ] Some query handler inject domain repo for getting workspace id
- [ ] Document service
  - [ ] Currently only hocuspocus guard the document, other like create/update/delete revision not check, bcs I'm lazy
    > If do, should create a guard outside of it
  - [ ] Health check
  - [ ] Consider merging blocknote & hocuspocus into editor module
- [ ] Authorization service
  - [ ] Health check
- [ ] Cannot deal with `domain.com:8080/api/v1` base path
- [ ] no env validation for document, search-worker
- [ ] Health check to other services (api service, meili, postgres...)
- [ ] Connection pool max connections, idle, timeout for database, meili
- [ ] gin should be protected with `SetTrustedProxies`
- [ ] Event is tracked by either otel or correlation id.
      But, currently use wotel + kafka tracer, and partially correlation id but not really connected.
- [ ] Revision endpoints aren't protected with authorization

### Frontend

- [ ] use suspend with skeleton, use suspend query for streamming
- [ ] Manage server state by tanstack, not always useState
- [ ] Clean architecture is considerable? `https://www.freecodecamp.org/news/reusable-architecture-for-large-nextjs-applications/`
- [ ] Add custom theme (or not)
- [ ] Set up logger

### Both

- [ ] yjs isn't typesafety, like getting Ymap, and set value.
      May try to see other libs, how do they do
- [ ] Those NestJS logging, we need to find a better way to wrap those controller log. NestJS Pino only http? not microservice.
      And guess that we should either using middleware or interceptor
- [ ] Mutating `update, add, delete` workspace member doesn't send event to user client
