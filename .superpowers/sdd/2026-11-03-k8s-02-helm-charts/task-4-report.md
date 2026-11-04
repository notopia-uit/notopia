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

## Review round 1

- Added a mandatory ingress NetworkPolicy for the managed gateway. It denies all ingress by default. Once both trusted front namespace and Gateway are specified, it permits only pods labeled as that Gateway in that namespace; the single front route parent must match those values. This policy is also rendered with Istio disabled. Rejects `gateway.enabled=false`: this chart cannot safely select and secure an arbitrary existing gateway.
- **Finding 1 is not fully verifiable or enforceable by this chart alone.** NetworkPolicy enforcement depends on the actual CNI, gateway pod labels, traffic source and host-network behavior; Kubernetes policies are additive, and other policies may permit bypass. A trusted front gateway could also have another route to the edge without header removal. Cluster administrators must enforce these prerequisites and perform live direct/spoofed-header tests; no such cluster is available. Istio-off additionally requires a non-bypassable external JWT/header/authorization layer whose existence this chart cannot verify. Do not consider the deployment secure based only on chart rendering.
- Reject percent-encoded paths, dot segments, repeated separators and unsupported wildcards. The front backend uses the configured listener port. Security switches have boolean schemas; public/group rules have strict typed fields, required nonempty lists, and HTTP method enum. Claim mappings are fixed to the four stripped headers; roles remains stripped but unset.

### Round 1 verification

- `node --test deploy/k8s/charts/notopia-edge/tests/edge.test.ts`: initial regression run failed on missing NetworkPolicy, incorrect backend port and unguarded gateway mode. Further red runs caught missing front-route guard and parent match. A separate run failed due to the new test assuming an incorrect rule order; the test was corrected to locate rules by contents.
- `nx run notopia-edge-chart:gen --skip-nx-cache --output-style=static`: PASS, regenerated schema and README.
- `oxfmt deploy/k8s/charts/notopia-edge/tests/edge.test.ts`: PASS.
- `nx run notopia-edge-chart:test --skip-nx-cache --output-style=static`: PASS, 1 Helm unit test and 9 TypeScript tests.
- `nx run notopia-edge-chart:lint --skip-nx-cache --output-style=static`: PASS, including strict CRD schema validation with Istio and telemetry enabled.
- `helm-schema -p -c deploy/k8s/charts/notopia-edge --check`: PASS.
- `bash -o pipefail -c 'helm template test deploy/k8s/charts/notopia-edge --set istio.enabled=false --set gateway.listenerPort=8080 --set frontRoute.trustedNamespace=front,frontRoute.trustedGateway=front-gateway,frontRoute.parentRefs[0].name=front-gateway,frontRoute.parentRefs[0].namespace=front | kubeconform -strict -summary -schema-location default -schema-location "https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json"'`: PASS, 3 valid resources, zero invalid/errors/skipped.
- Live cluster integration remains pending as instructed; no live enforcement claim is made.
