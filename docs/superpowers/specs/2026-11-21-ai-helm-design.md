# AI Helm and homelab design

Approved by the user on 2026-11-21.

Add an AI Helm chart following the existing common-library service charts and a suspended homelab release. AI has no PostgreSQL, Kafka, migration, or seed dependency; leave CNPG unchanged. Deploy a separate suspended standalone Valkey with persistent AOF and no eviction for accounting and continuation state. Secrets are provisioned separately and never include local development credentials.

AI listens on 8086 with `/ai/health/{startup,live,ready}` probes; readiness checks Redis. Configure note, authorization and billing gRPC at `note:18081`, `authorization:18089`, `billing:18084`, and Meilisearch at `http://meilisearch:7700`. Provide the Redis URL and model and search API keys from externally provisioned secrets. LLM base URL and model are operator configuration.

Route `/ai` through the existing edge with configurable streaming-safe timeouts exceeding the default 120-second application deadline. Keep generation authenticated; optionally expose health. Allow only the edge principal to AI and add the AI service account to the three gRPC services' allowed principals. Preserve the edge identity boundary blocker and all suspended states. Validate the full streaming path, external gateway timeouts, persistence, credentials, and security independently before deployment.

Use the existing Flux and chart generation conventions; release scripts already recognize AI. Test chart renders, timeouts, probes, security, cluster wiring and isolated release tag updates. Do not deploy, access a cluster, publish, commit, or create real secrets.
