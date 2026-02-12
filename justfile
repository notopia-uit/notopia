all parallel="4" exclude="docs":
    pnpm exec nx run-many -t lint build gen bundle --parallel={{ parallel }} --exclude={{ exclude }}
