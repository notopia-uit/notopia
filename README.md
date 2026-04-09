# Notopia

<div align="center">

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

- Delete broken symlinks in the current directory and its subdirectories
  ```bash
  find . -type l ! -exec test -e {} \; -print -delete
  ```

## TO-DO

- [ ] Note service
  - [ ] Routing kafka message based on metadata workspace id if partitioning or listening
  - [ ] Some handlers should rename the UserID to actor ID, and... so do the domain event?
  - [ ] Some query handler inject domain repo for getting workspace id
- [ ] Document service
  - [ ] Currently only hocuspocus guard the document, other like create/update/delete revision not check, bcs I'm lazy
    > If do, should create a guard outside of it
  - [ ] Health check
- [ ] Authorization service
  - [ ] Health check
- [ ] Cannot deal with `domain.com:8080/api/v1` base path
- [ ] no env validation for document, search-worker
- [ ] Health check to other services (api service, meili, postgres...)
- [ ] Connection pool max connections, idle, timeout for database, meili
