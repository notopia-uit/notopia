# Homelab deployment

Deployment readiness and live verification are deferred by the user.
Use [README.md](./README.md) for provisioning, ordered bootstrap, authorization
seeding, security checks, and recovery procedures.

## Current state

- HelmReleases remain suspended. Do not resume, provision credentials, encrypt,
  publish images, or deploy without a new user request.
- `infra/secrets/*.secret.sops.yaml` are intentionally **plaintext review drafts**,
  despite their filenames. They contain public development defaults and
  placeholders, not production-ready credentials.
- Keep drafts excluded from Kustomizations until provisioned and encrypted.
  Keep the centralized secret layout; do not partially move application Secrets.
- Leave CNPG-generated credentials and application Secret references unchanged.
- Local chart/render tests do not establish live deployment or security readiness.

## Deferred before deployment

- [ ] Provision credentials, replace development defaults and token/provider
  placeholders, and configure AI's LLM base URL and model.
- [ ] Encrypt drafts when requested, verify Flux decryption, and include only ready
  Secrets in the appropriate Kustomization. Follow the README's separate
  external-homelab procedure for the Flux Receiver Secret.
- [ ] Configure document's `NOTOPIA_DOCUMENT_S3_ENDPOINT`,
  `NOTOPIA_DOCUMENT_S3_REGION`, and `NOTOPIA_DOCUMENT_S3_BUCKET_NAME`; verify bucket
  access and browser-reachable presigned upload/download URLs.
- [ ] Publish chart-compatible service images, set immutable tags, and verify the
  manual main-only release and deployment-tag update end to end.
- [ ] Implement and independently verify the edge identity-header boundary,
  Istio caller/port enforcement, and the README's other bootstrap gates.
- [ ] Verify AI streaming and timeouts through every gateway, dedicated Valkey
  persistence, backups, and restore on the intended cluster.
- [ ] Follow the README's external authorization seeding, migration, and ordered
  bootstrap procedures before deliberately resuming each ready release.

## Validation blockers

- The repository encryption check intentionally rejects the plaintext drafts.
  Restore it by encrypting when authorized; do not weaken the check.
- The existing SOPS test fixture fails opening Node's piped `/dev/stdin` in this
  environment. Resolve that fixture issue and rerun the full homelab suite.
- Earlier passing render-test results predate copying plaintext drafts into the
  repository; do not report the current full suite as passing.
