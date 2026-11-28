---
authorship:
  - ai
---

# The Docs Build Waiting on Git

_Why a fast local build took minutes in CI._

Written with AI, November 2026.

The Docs job took almost seven minutes, though publishing itself took six seconds. A later
diagnostic run took nearly fourteen minutes. The diagram cache restored successfully, and Kroki
was healthy; neither explained the delay.

We streamed the Nx task output to locate the pause. VitePress spent over ten minutes before
its first Vite build hook, while page rendering and Pagefind indexing each took seconds or less.
The same uncached docs build took under ten seconds locally.

The culprit was the interaction between two choices:

- The Docs checkout used `fetch-depth: 0` and `filter: tree:0`: it had full commit history,
  but omitted historical Git trees.
- `vitepress-plugin-pagefind` scans Markdown Git history during configuration to provide
  search-result date metadata, even when the dates are not displayed.

Walking that history caused Git to fetch missing trees on demand. In a controlled local clone,
the same history scan with `tree:0` made 2,922 fetches and timed out after 90 seconds. With
`blob:none`, it finished in 0.18 seconds and returned the same history as the full clone.

We changed only the Docs checkout to `blob:none`. It retains commits and historical trees but
omits historical file contents until needed. In the [verification run][run], VitePress built in
7.06 seconds and the Docs job completed in 2 minutes 45 seconds; the diagram cache still hit.
Most remaining time was setup and pulling the Kroki container.

Future docs builds should avoid the repeated tree fetches while retaining history for Pagefind.
The Main job still uses `tree:0`; if another tool there starts walking file history, it may
encounter the same tradeoff.

[run]: https://github.com/notopia-uit/notopia/actions/runs/46140110306
