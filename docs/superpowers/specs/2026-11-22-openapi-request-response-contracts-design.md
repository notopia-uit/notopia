# OpenAPI request and response contracts

## Status and handoff

Design for a later implementation session. This document is not an implementation plan.
The implementer should inspect the current worktree and existing uncommitted changes before
making changes, then write a detailed plan and implement it separately.

## Problem

During local billing UAT, `POST /note/workspaces` rejected a request containing a name and
slug because `id` was missing. The operation uses `Workspace` as its request body, while
that schema marks `id` both `readOnly` and required. The note service generates the ID
itself and ignores an incoming ID.

The service validates requests with `gin-middleware` and `kin-openapi` against the bundled
OpenAPI 3.1 spec. In an isolated comparison using the current bundled spec and request
validation options, both kin-openapi v0.147.0 and v0.153.0 rejected the body without `id`
and accepted one with an `id`. This was not a full Gin or service test. It does not prove
why the request previously worked.

The earlier OpenAPI 3.0-to-3.1 migration is a plausible contributor: the validator chooses
JSON Schema 2020 behavior for 3.1, where `readOnly` is an annotation and does not by itself
remove a property from `required`. A controlled comparison changing only the declared
specification version was not performed. Do not present the version migration or a
dependency upgrade as a proven root cause.

## Goal

Make request contracts accurately describe what callers must and may supply, independently
of fields returned in responses. Keep OpenAPI 3.1, preserve response contracts and intended
business behavior, regenerate clients and servers, and update affected consumers.

## Approaches considered

1. **Separate request and response schemas where contracts differ (chosen).** Explicit
   contracts work across validators and generated clients. They require an audit and can
   change generated types. Reuse property schemas where practical instead of copying
   constraints.
2. **Change validation centrally.** This might avoid contract changes, but requires a
   supported, tested mechanism that preserves all other validation and direction-sensitive
   behavior. Do not disable validation or silently accept invalid inputs as a shortcut.
3. **Downgrade the specification to 3.0.** This changes semantics for the entire API and
   may lose 3.1 features; it is not the chosen solution.

## Scope and known candidates

Audit request bodies across `api/note`, `api/document`, and `api/billing`, including
nested references, arrays, and compositions. Identify request schemas that reuse response
models containing required `readOnly` fields or otherwise describe server-owned fields as
caller inputs. Change only contracts whose request and response shapes actually differ;
`readOnly` by itself is not proof of a defect.

Known starting points:

| Surface                                                           | Current contract                                                                 | Status                                                                                                                                       |
| ----------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `POST /note/workspaces`                                           | `Workspace` requires read-only `id`                                              | Confirmed validator failure. Introduce a creation request containing required `name` and `slug`; retain required `id` in the response model. |
| `POST /note/notes`                                                | Reuses `Note`, which requires read-only `id`, `tags`, `updatedAt`, and `trashed` | Likely affected; verify the existing handler and intended writable inputs, including nullability and requiredness.                           |
| `POST /note/folders`                                              | Reuses `Folder`, which requires read-only `id`, `updatedAt`, and `trashed`       | Likely affected; verify the existing handler and intended writable inputs.                                                                   |
| `PUT /note/workspaces/{workspaceId}/members`                      | Uses `WorkspaceMember`, whose `name` is read-only but not required               | Audit, including whether a dedicated update input clarifies or protects the contract; do not assume it currently fails.                      |
| Document revision schemas and other models with `readOnly` fields | Some appear to be response-only                                                  | Audit their actual use before changing them.                                                                                                 |

Inspect billing and the remaining note/document operations rather than assuming this list is
exhaustive. Record the audit findings and affected operations in the implementation
handoff or accompanying documentation.

## Contract design

- Define focused request schemas, for example `CreateWorkspaceRequest`,
  `CreateNoteRequest`, and `CreateFolderRequest`, where the audit confirms they are needed.
- Derive required fields, nullability, validation constraints, and defaults from the
  existing intended public behavior and handlers; do not infer that every non-read-only
  response property must be mandatory in a create request.
- Reuse existing property definitions where safe, or extract neutral shared property/value
  schemas if doing so avoids reference cycles or inappropriate `readOnly` annotations.
  Verify the bundled result and generated code: referencing a response property that
  carries `readOnly` can unintentionally propagate that annotation.
- Do not weaken response requirements or alter response payloads to accommodate requests.
  Keep operation IDs, routes, authorization, errors, and documented business behavior stable
  unless an independently approved change is required.
- Treat server-owned input handling explicitly: ensure the schema doesn't advertise those
  fields as writable; verify actual validator behavior for unexpected fields. Do not claim
  they are rejected unless the contract and runtime actually enforce rejection.
- Resolve any discrepancy between current schemas and actual handler behavior explicitly;
  do not silently introduce breaking changes for existing clients.

## Generated code and consumers

Change source YAML under `api/`, then run the repository's generation pipeline (`nx gen
api`), including bundling and the relevant Go, NestJS, and TypeScript outputs. Do not
hand-edit generated artifacts. Review generated diffs for unintended contract changes.

- **Note backend:** generated request body types can change. Update HTTP controllers,
  mappings, tests, or constructors only where necessary. Workspace creation already
  generates its own ID and reads the name and slug; business logic should remain intact.
- **Document and billing backends:** adjust generated types and implementation only if
  the audit finds affected contracts or generation requires it. Document uses generated
  NestJS interfaces; billing and note use generated Go servers.
- **Frontend and API clients:** check `packages/api-gen`, `apps/web`, and other actual
  callers of changed operations. Update imports, payload construction, and types only
  where needed; callers already sending the correct fields may need no runtime change.
  Preserve public response types and do not assume the billing UI is implemented.
- **API viewer:** verify Scalar shows the correct create-body schema and example after
  the bundled spec updates.

## Verification and acceptance

1. For every changed operation, exercise request validation through the actual service
   middleware using the bundled OpenAPI spec. Include valid requests without server-owned
   fields and invalid requests missing genuinely required inputs. Assert the relevant
   responses and preserve authorization behavior.
2. For workspace creation, verify a valid name/slug passes validation without `id`, and
   verify the handler/service generates and returns an ID. Do not require a Stripe
   integration for this test.
3. Verify response schemas still require their existing server-owned properties, and check
   nullability and existing client payload compatibility for notes and folders.
4. Run API lint/bundle/code generation and relevant service tests, lint, and TypeScript
   typechecks/builds for affected consumers. Report commands and any failures or blocked
   checks. Follow repository Nx targets and never hand-edit generated code.
5. Manually verify `POST /note/workspaces` in Scalar against a local authenticated
   environment if available; record blockers rather than claiming an end-to-end pass.
6. Document changed API contracts and migration considerations in the existing service
   documentation as appropriate.

## Non-goals and constraints

- No blanket disabling of schema validation, no authentication bypass, and no fabricated
  client-side IDs as the intended contract.
- No unrelated database, Stripe sandbox, gateway, or Authentik cleanup.
- No speculative dependency upgrade or OpenAPI version downgrade as the primary fix.
- Do not alter other in-progress user work, remove existing data, or assume a clean working
  tree. A separate implementation session should decide on isolation before editing.
