# notopia-edge

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../notopia-common-chart | notopia-common-chart | 0.1.0 |

## Security and deployment

**Security blocker:** the chart does not establish a portable, non-bypassable identity-header boundary. Do not deploy or expose it until the operator has independently implemented and verified that boundary. Helm rendering/install/upgrade fails by default. After verification, set `security.externalBoundaryAcknowledged=true`; this is an operator assertion, not a control or evidence of security. A failed upgrade does not remove or secure an already-running release.

Replace the nonfunctional example JWT issuer and JWKS URI. Configure `frontRoute.parentRefs` and both `frontRoute.trustedNamespace` and `frontRoute.trustedGateway` for the gateway that strips all five identity headers. The chart creates a NetworkPolicy selecting the managed gateway's pods; by default it denies all ingress and, when configured, permits only pods labeled for that front gateway in that namespace. Existing gateways (`gateway.enabled=false`) are unsupported. `gateway.listenerPort` controls both the listener and front backend port.

**Required external controls:** all traffic reaching the edge, including public paths and requests with absent JWT claims, must have the five client-supplied identity headers removed before trusted claims are set; no alternate front route may reach the edge without that removal. Prevent direct access to the edge service, pod IPs, host ports, load balancer and all other entry points. Restrict who can create or modify routes, pods, labels, gateways and network policies, including in the trusted namespace; a matching pod label is not authenticated identity. Ensure downstream services cannot be reached directly with spoofed identity headers. Use enforceable platform controls, such as a verified authenticated proxy-to-proxy boundary and centrally controlled routing; the precise mechanism is deployment-specific.

**NetworkPolicy is defense in depth, not a security proof.** Enforcement depends on the CNI, selected pod labels, traffic path, host-network and node behavior. Policies are additive and another policy can widen access. A different route on the same front gateway or another pod with the same label may bypass the intended filter. Cluster and network administrators must verify these controls and their continuing operation; privileged operators can bypass them.

The billing admin group is mandatory; custom group paths are also excluded from the general JWT allow rule. Public and protected paths must not overlap. Path wildcards are supported only at the end; percent escapes, dot segments and repeated slashes are rejected. Header claim mappings are fixed to the stripped identity header contract.

When Istio is disabled, the NetworkPolicy remains, but Gateway API objects alone do not authenticate or authorize requests. Supply an external non-bypassable JWT verification and authorization layer that strips all five identity headers, sets trusted identity claims including `groups_json`, and enforces the public, authenticated and billing admin group rules. After independently verifying that layer, also set `security.externalIdentityLayerAcknowledged=true`. The chart cannot verify that this layer exists or that a non-Istio gateway implementation uses the expected pod labels.

Before exposure and after configuration or infrastructure changes, test direct edge access, an alternate front route, a spoofed-label pod, and all five forged headers with anonymous requests and tokens with missing claims. Test ordinary and admin users against admin paths, and repeat with Istio disabled when applicable. These checks require a representative cluster and do not follow from successful Helm rendering or acknowledgement.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| frontRoute.enabled | bool | `true` | Create a route from the front gateway |
| frontRoute.hostnames | list | `[]` | Route hostnames |
| frontRoute.parentRefs | list | `[]` | Parent gateway references |
| frontRoute.trustedGateway | string | `""` | Gateway API name of the only front gateway allowed to reach the edge |
| frontRoute.trustedNamespace | string | `""` | Namespace of the only front gateway allowed to reach the edge |
| fullnameOverride | string | `""` | Override the full release name |
| gateway.annotations | object | `{}` | Gateway annotations |
| gateway.className | string | `"istio"` | GatewayClass name |
| gateway.enabled | bool | `true` | Create the managed gateway; disabling is unsupported |
| gateway.hostname | string | `""` | Listener hostname, empty for all hosts |
| gateway.listenerPort | int | `80` | Listener port |
| istio.adminGroup | string | `"admin"` | Group required for billing administration |
| istio.enabled | bool | `true` | Enable Istio authentication and authorization |
| istio.groupRules | list | `[]` | Additional group-protected paths and permitted groups |
| istio.jwt.issuer | string | `"https://auth.example.invalid/"` | JWT issuer; replace the nonfunctional example before deployment |
| istio.jwt.jwksUri | string | `"https://auth.example.invalid/jwks"` | JWKS endpoint; replace the nonfunctional example before deployment |
| istio.publicRules | list | `[]` | Additional public rules; paths must not overlap protected group paths |
| istio.telemetry.accessLogProviders | list | `[]` | Access logging provider names |
| istio.telemetry.enabled | bool | `false` | Enable Istio telemetry |
| istio.telemetry.randomSamplingPercentage | int | `100` | Trace sampling percentage |
| istio.telemetry.tracingProviders | list | `[]` | Tracing provider names |
| nameOverride | string | `""` | Override the chart name |
| notopia-common-chart | object | `{"global":{}}` | Common library configuration |
| security.externalBoundaryAcknowledged | bool | `false` | Operator acknowledges an externally enforced non-bypassable identity-header boundary |
| security.externalIdentityLayerAcknowledged | bool | `false` | Operator acknowledges an external JWT, header, and authorization layer when Istio is disabled |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
