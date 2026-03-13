# Notopia

<div align=center>
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

### Document

- Currently only hocuspocus guard the document, other like create/update/delete revision not check, bcs I'm lazy
  > If do, should create a guard outside of it
