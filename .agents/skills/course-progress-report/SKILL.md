---
name: course-progress-report
description: Write or revise a course progress report
---

# Course progress report

Keep the technical record for project review separate from the shorter Vietnamese report for the lecturer. Produce only the report requested.

## Inputs

The user may specify a start and end by date or commit, or ask to report through the present. For a detail report, default to the latest detail report's end commit as the next baseline and the current tip of `main` as the end. The user may also supply notes, test results, exclusions, issues, PRs, or other references. Use them where relevant; report rather than conduct new research unless asked.

## Outputs

- **Detail report:** `docs/src/reports/YYYY-MM-dd-detail.md`, in English. Follow [the detail report guide](references/detail.md).
- **VN report:** `docs/src/reports/YYYY-MM-dd.md`, in Vietnamese. Follow [the VN report guide](references/vn.md) without repeating the source review.

## Shared format

1. Keep `authorship: [ai]` in YAML frontmatter. Omit `title` and any repeated bold title.
2. Use one H1 based on the filename: `# YYYY-MM-dd` for the VN report and `# YYYY-MM-dd Detail` for the detail report. The date is the reporting end date.
3. Directly below the H1, state the reporting period, exclusive baseline commit, and inclusive end commit on the named branch. Localize the labels and follow the reference rules for that report.
4. Add a short overview of the main outcomes and current status. The rest of the structure is flexible; suggested sections are completed work, testing or ongoing work, and next steps.

Update `docs/src/reports/index.md` for the report written, preserve the user's edits, and do not run the docs build.
