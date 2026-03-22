# Notopia

<div align=center>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_backend">
    <img alt="SonarQube Quality Gate - Note" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=alert_status"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_note">
    <img alt="SonarQube Quality Bug - Note" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=bugs"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_note">
    <img alt="SonarQube Quality Code Smells - Note" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=code_smells"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_note">
    <img alt="SonarQube Quality Maintainability Rating - Note" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=sqale_rating"/>
  </a>

  <br/>

  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_web">
  <img alt="SonarQube Quality Gate - Web" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=alert_status"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_web">
    <img alt="SonarQube Quality Bug - Web" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=bugs"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_web">
    <img alt="SonarQube Quality Code Smells - Web" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=code_smells"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_web">
    <img alt="SonarQube Quality Maintainability Rating - Web" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=sqale_rating"/>
  </a>

  <br/>

  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_document">
    <img alt="SonarQube Quality Gate - Document" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=alert_status"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_document">
    <img alt="SonarQube Quality Bug - Document" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=bugs"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_document">
    <img alt="SonarQube Quality Code Smells - Document" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=code_smells"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_document">
    <img alt="SonarQube Quality Maintainability Rating - Document" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=sqale_rating"/>
  </a>

  <br/>

  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization">
    <img alt="SonarQube Quality Gate - Authorization" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=alert_status"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization">
    <img alt="SonarQube Quality Bug - Authorization" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=bugs"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization">
    <img alt="SonarQube Quality Code Smells - Authorization" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=code_smells"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization">
    <img alt="SonarQube Quality Maintainability Rating - Authorization" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=sqale_rating"/>
  </a>

  <br/>

  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker">
    <img alt="SonarQube Quality Gate - Search-worker" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=alert_status"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker">
    <img alt="SonarQube Quality Bug - Search-worker" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=bugs"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker">
    <img alt="SonarQube Quality Code Smells - Search-worker" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=code_smells"/>
  </a>
  <a href="https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker">
    <img alt="SonarQube Quality Maintainability Rating - Search-worker" src="https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=sqale_rating"/>
  </a>

<br/>

  <a href="https://wakatime.com/badge/github/notopia-uit/notopia">
    <img alt="Wakatime" src="https://wakatime.com/badge/github/notopia-uit/notopia.svg"/>
  </a>
</div>

## OTEL

<https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/>

## Dev

- Delete broken symlinks in the current directory and its subdirectories
  ```bash
  find . -type l ! -exec test -e {} \; -print -delete
  ```

## TO-DO

- [ ] Document service
  - [ ] Currently only hocuspocus guard the document, other like create/update/delete revision not check, bcs I'm lazy
    > If do, should create a guard outside of it
  - [ ] Health check
- [ ] Authorization service
  - [ ] Health check
