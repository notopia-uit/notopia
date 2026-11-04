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

## Review round 2: header-boundary security blocker

The chart cannot by itself establish a portable, non-bypassable header boundary. The Gateway API `RequestHeaderModifier` is attached to a route, so another route on the same trusted gateway can bypass it. Kubernetes NetworkPolicy is additive, requires an implementing network plugin, and selects labels rather than authenticated workload identities; another matching-label pod, other policy, host/node traffic, or another access path may bypass the assumed boundary. Istio-specific proxy filter ordering would be implementation/version dependent, would not protect the Istio-off case, and was not introduced without a live integration environment.

The chart now fails Helm rendering/install/upgrade unless the operator explicitly sets the boolean `security.externalBoundaryAcknowledged=true` after establishing and verifying an external non-bypassable header boundary. Istio-off additionally requires the boolean `security.externalIdentityLayerAcknowledged=true` after establishing external JWT, header and authorization enforcement. Template-level boolean checks also reject string acknowledgements when schema validation is disabled. **These are assertions only, not mechanisms, and do not make an insecure deployment safe.** A failed upgrade does not repair an already running release. CI lint/test commands use acknowledgements solely to render and validate manifests; they do not assert real controls exist.

Required external controls are detailed in the chart README: enforce sanitization on every path including public requests and missing claims; prevent direct edge and downstream service access; control alternate routes, pod/label creation, gateway and policy changes; and verify effective enforcement for CNI, additive policies, host/node traffic and any ingress path. With Istio disabled, independently provide non-bypassable JWT verification, trusted claim-to-header mapping and public/authenticated/admin authorization. Live tests must cover alternate routes, forged-label pods, direct addresses, all five spoofed headers, missing claims and admin versus ordinary identities. These checks remain **pending**: the configured cluster was previously unreachable, and no live cluster is available. Until an operator implements and validates those controls, this is a security blocker, not a claim of a fully fixed data plane.

References consulted: https://gateway-api.sigs.k8s.io/guides/user-guides/http-header-modifier/ ; https://kubernetes.io/docs/concepts/services-networking/network-policies/ ; https://istio.io/latest/docs/reference/config/networking/envoy-filter/ ; https://istio.io/latest/docs/reference/config/security/request_authentication/ .

### Round 2 verification

- `node --test deploy/k8s/charts/notopia-edge/tests/boundary.test.ts`: initial RED, 3 failures: default rendered without acknowledgement and acknowledgement values were unsupported. A subsequent RED exposed string acknowledgements passing when schema validation is disabled; template validation was tightened.
- `oxfmt deploy/k8s/charts/notopia-edge/tests/edge.test.ts deploy/k8s/charts/notopia-edge/tests/boundary.test.ts deploy/k8s/charts/notopia-edge/project.json`: PASS.
- `nx run notopia-edge-chart:gen --skip-nx-cache --output-style=static`: PASS.
- `nx run notopia-edge-chart:test --skip-nx-cache --output-style=static`: PASS, 1 Helm unit test and 12 TypeScript tests.
- `nx run notopia-edge-chart:lint --skip-nx-cache --output-style=static`: PASS, including strict kubeconform of Istio-enabled resources with a test-only acknowledgement.
- `helm-schema -p -c deploy/k8s/charts/notopia-edge --check`: PASS.
- `bash -o pipefail -c 'helm template test deploy/k8s/charts/notopia-edge --set security.externalBoundaryAcknowledged=true,security.externalIdentityLayerAcknowledged=true,istio.enabled=false --set gateway.listenerPort=8080 --set frontRoute.trustedNamespace=front,frontRoute.trustedGateway=front-gateway,frontRoute.parentRefs[0].name=front-gateway,frontRoute.parentRefs[0].namespace=front | kubeconform -strict -summary -schema-location default -schema-location "https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json"'`: PASS, 3 valid resources, 0 invalid/errors/skipped. This is schema validation only, not live security validation.
