all parallel="4":
    pnpm exec nx run-many -t lint build gen bundle --parallel={{ parallel }}
