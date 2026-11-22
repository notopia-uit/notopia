#!/usr/bin/env bash
set -euo pipefail

chart="$1"
chart_dir="deploy/k8s/charts/$chart"
helm_chart="$chart_dir"
if [[ "$chart" == "notopia-common-chart" ]]; then
  helm_chart="$chart_dir/tests/fixture"
fi

reports="coverage/charts/$chart"
mkdir -p tmp/chart-lock "$reports"
rm -f "$reports/helm.xml" "$reports/node.xml"
flock "tmp/chart-lock/$chart.lock" helm dependency build --skip-refresh "$helm_chart"

status=0
flock -s "tmp/chart-lock/$chart.lock" helm unittest -t JUnit -o "$reports/helm.xml" "$helm_chart" || status=$?

tests=("$chart_dir"/tests/*.test.ts)
if [[ -f "${tests[0]}" ]]; then
  flock -s "tmp/chart-lock/$chart.lock" node --test \
    --test-reporter=spec --test-reporter-destination=stdout \
    --test-reporter=node-test-junit-reporter --test-reporter-destination="$reports/node.xml" \
    "${tests[@]}" || status=$?
fi

exit "$status"
