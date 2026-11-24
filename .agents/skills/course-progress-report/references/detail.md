# Detail report

Write an English technical record the user can review and filter before requesting the VN report. Cover material changes and their significance without turning the report into a commit dump.

## 1. Resolve the reporting range

- Use the dates or commits the user supplied. An explicit start commit is an exclusive baseline; an explicit end commit is inclusive.
- If no start is provided, find the latest `docs/src/reports/YYYY-MM-dd-detail.md` and use its recorded end commit as the exclusive baseline. If that report only records an end date, infer the corresponding cutoff from the reporting branch's history and disclose the inference. If neither is available, ask for a starting point.
- For date inputs, use calendar-day boundaries in the reporting timezone. Treat the start date as inclusive and the end date as inclusive through the end of that day. Resolve dates against branch landing history rather than feature-branch author dates. If no end is specified, use the branch tip at the start of the review and today's date for the filename.
- Default to `main`. Use another branch only if the user asks to cover it, and label its scope separately. Record the resolved branch, period, and full baseline and end hashes so later reports can resume unambiguously.

## 2. Review changes and evidence

- Start with the selected branch's first-parent log over the resolved range, not `git log --all`. Inspect the landed diff or code when the commit title is incomplete or a claim is disputed. Cite the squash or merge commit when the original PR commits are not in the reporting branch. If later work reverted a change, do not describe it as still present.
- Group related changes, but keep distinct outcomes visible. Include material implementation changes, API or dependency updates, tooling, deployment work, documentation, and design work when relevant. Explain why meaningful changes matter; filter out routine churn.
- Use user-provided notes, UAT evidence, and links as supporting material. Respect exclusions for work already reported. Distinguish repository evidence from user-reported status, and keep open PRs and unmerged work out of completed results. Do not fetch references or broaden the research by default.
- For verification, distinguish a code change or added test from an observed test result. Preserve meaningful partial, blocked, skipped, and deferred results. Separate deployment preparation from a verified deployment. Link the UAT workbook and progress log if the user supplies a shareable URL; do not copy credentials, tokens, or raw sensitive logs into the report.

## 3. Write and check

- Start with the period and commit bounds. Organize completed changes by topic, then identify testing gaps, ongoing work, next-stage plans, and longer-term plans for the following semester where applicable. Keep useful PR, issue, demo, and document references with the claims they support. Bare commit hashes are acceptable in the detail report.
- Use `docs/src/reports/YYYY-MM-dd-detail.md` and add or update its entry in `docs/src/reports/index.md`. Preserve any existing VN report and user edits. Do not create a VN draft unless asked.
- Check that cited landed commits are in scope, internal links resolve, and the summary distinguishes implemented work from tested behavior and future plans. Do not run the docs build.
