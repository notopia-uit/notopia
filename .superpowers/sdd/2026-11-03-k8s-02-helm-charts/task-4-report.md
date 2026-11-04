# Task 4 report

Implemented `deploy/k8s/charts/notopia-edge` using `notopia-common-chart` 0.1.0. The mandatory billing admin rule excludes both the root and subtree from the general authenticated ALLOW rule; custom group paths are excluded too. The front route removes all five identity headers and `groups_json` maps to the groups header. Public and protected path conflicts and unsupported wildcards are rejected at render time.

## Verification

- `helm dependency build --skip-refresh deploy/k8s/charts/notopia-edge && helm unittest deploy/k8s/charts/notopia-edge`: initial red, missing policy template, as expected.
- `node --test deploy/k8s/charts/notopia-edge/tests/edge.test.ts`: initial red for wildcard admin group; corrected validation.
- `nx run notopia-edge-chart:gen --skip-nx-cache --output-style=static`: passed.
- `nx run notopia-edge-chart:test --skip-nx-cache --output-style=static`: passed, 1 Helm unit test and 5 TypeScript tests.
- `nx run notopia-edge-chart:lint --skip-nx-cache --output-style=static`: passed, including dependency build, Helm strict lint, and strict kubeconform CRD validation of Gateway, HTTPRoute, RequestAuthentication, AuthorizationPolicy, and enabled Telemetry.
- `kubectl cluster-info --request-timeout=5s`: unavailable; configured API server 192.168.28.25:6443 returned no route to host and command timed out after 15s. No cluster modifications attempted. Live anonymous/user/admin decisions, 403 admin denial, and spoofed-header checks remain pending on a disposable Istio cluster.

## Concerns

- The default issuer and JWKS URI use nonfunctional `.invalid` examples because empty values fail the Istio CRD schema. Replace both before deployment.
- Configure `frontRoute.parentRefs` and prevent clients bypassing the front gateway: the edge relies on the front route to strip spoofed headers. Istio-disabled operation requires a separate non-bypassable JWT, header and authorization layer.
- CRD schema and rendered policy tests cannot establish actual Istio behavior, path normalization, JWT handling, or header-filter ordering; live integration remains required.
