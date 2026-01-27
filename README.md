# Notopia

<div align=center>
  <a href="https://wakatime.com/badge/github/notopia-uit/notopia">
    <img alt="Wakatime" src="https://wakatime.com/badge/github/notopia-uit/notopia.svg"/>
  </a>
</div>

## Dev

- Delete broken symlinks in the current directory and its subdirectories
  ```bash
  find . -type l ! -exec test -e {} \; -print -delete
  ```
