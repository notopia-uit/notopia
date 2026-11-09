# Workspace assistant note editing

## Intent and agreed scope

Complete the workspace assistant so a user can ask “Find Meeting notes and shorten its introduction” and have the assistant open and edit that Notopia note. The user chose assistant-driven target selection and one target note at a time. Other notes may provide reference material, but only the active target can be edited.

Reuse the existing BlockNote editor, its AI document operations, and the authenticated collaborative document connection. Deliver the complete search → open → live read → edit → review flow, with explanatory documentation for future development.

## Approaches

1. User-selected target: simpler integration, but requires manually locating a note before asking for edits.
2. Assistant-selected target (chosen): the assistant locates a note and opens its live editor beside chat, then continues the request using that editor's current document state. This requires client-tool execution and backend continuation support.

A server-side document rewriter is not the chosen architecture: it would bypass the existing client editing and review machinery.

## User experience

Chat remains scoped to the workspace. An editor panel shows the active target's identity, connection/loading status, edit activity, existing change-review controls, and commit action. On narrow screens the same chat and editor remain reachable without mounting a second editor session.

The assistant can search and read reference notes without changing the target. An explicit open-target tool loads a note into the editor. A completed open operation supplies live document context for the next model step; attaching a note label alone does not grant editing context.

Only one target is active. A switch waits until generation has stopped and pending edits have been accepted or discarded. Closing or switching cleans up the old editor and collaboration connection. Ambiguous searches can be resolved conversationally rather than guessing a target.

## Architecture

### Browser

- Extend `AssistantView` with a note-session owner and client-tool handling.
- Reuse the authenticated Hocuspocus provider setup used by note pages/previews; wait for initial synchronization before building document context or allowing edits.
- Connect the active editor's BlockNote AI request builders and streamed document-operation executor to assistant chat. Do not implement a second block-operation language.
- Identify each session by workspace, note, and a local generation/session token. Late operations from an old request must never reach a newly opened note.
- Execute each tool call at most once within the live chat lifecycle, including React rerenders. Do not replay mutations from rendered history or a restored transcript.
- Display tool errors and connection failures in chat. An open failure must not silently fall back to editing the previous target.

### AI service

- Add a narrowly defined open-target client tool alongside the existing allowlisted `applyDocumentOperations` tool. Backend-owned schema and authorization must govern target opening; arbitrary client tool definitions remain disallowed.
- Resolve a target through the user-aware note APIs, verifying its workspace and accessibility. Require write permission before offering or admitting editing operations; Hocuspocus continues enforcing document access.
- Editing requests carry the active note identity and fresh document state. Search-index content remains reference material, never the basis for mutation against live block IDs.
- Permit bounded client-tool continuations. The server must correlate them to a server-recorded admitted task, its user, workspace, expected tool call, and target. Client-supplied history or a generation identifier alone is insufficient evidence.
- Preserve the policy of one credit per started user generation across its open/edit continuations. Continuations acquire active-stream capacity, remain throttled, and have bounded total steps and lifetime. Duplicate or concurrent continuations cannot double-execute or obtain unlimited free generations.
- Keep task correlation in shared Redis, with atomic claiming and expiry. Cancellation, expiry, or terminal failure closes continuation eligibility; provider errors cannot leave unlimited reusable continuation state.
- Validate client outputs structurally and treat document text and tool outputs as untrusted content. A client report is not proof of persistence or authorization.

### Saving and review

Reuse BlockNote's supported review/accept/reject behavior. Before implementation, verify its actual collaboration semantics: the design does not assume proposed changes are private or unsynced. Explain those semantics in the UI/documentation as needed.

Collaborative persistence and committing stay separate. The existing Commit action publishes the updated revision used by indexing; ordinary AI edits do not silently force a commit. Reject/cancel must use supported operation-scoped rollback, never replace the entire document with an old snapshot that could erase collaborators' edits.

## Cancellation and failure behavior

Stop aborts model generation and pending client execution. Once stopped, no subsequent stream chunk may mutate the note. Edits already applied remain visible for review under the established BlockNote semantics; cancellation is not a claim that all work was rolled back.

Read-only access, removed/trashed notes, workspace mismatch, connection/authentication failures, malformed operations, expired continuations, and failed model steps produce actionable errors. A success label requires successful tool execution; saving status follows the collaboration provider, not the model's prose.

## Boundaries

This change edits existing Notopia note content. Creating/deleting notes, renaming or moving notes, filesystem access, multi-note mutation, background editing after leaving the page, and durable chat restoration are outside this design.

Retain existing editor AI commands and workspace search/read behavior. Preserve current unrelated working-tree changes. Billing setup is deferred by the user; do not change billing service ownership or pricing policy to implement this feature.

## Implementation decisions and discovery

The exact BlockNote shared-chat integration API and provider ownership must be confirmed against the installed versions and current documentation before coding. Prefer supported shared-chat/request helpers; if these cannot support the specified flow, surface the conflict before substituting a materially different interaction.

The implementer chooses component boundaries, Redis record/key layout, continuation token format, and concrete bounded limits consistent with existing service conventions. These are proposed engineering choices, not separately user-selected requirements.

## Verification and acceptance

- A chat request can find and open an initially unopened note, obtain synced live content, and edit that note through BlockNote operations.
- Reference-note reads cannot redirect mutations. One target stays active, with switching guarded during streaming/review.
- Permission, wrong-workspace, stale-session, duplicate tool-call, cancellation, disconnect, and malformed-operation cases are tested at the relevant boundaries.
- Backend tests cover valid continuations, forged/replayed/expired continuations, concurrent claims, step limits, and exactly one debit across a multi-request task. Exercise atomic accounting against isolated test Redis where available.
- Verify normal editor commands and workspace read-only chat still work. Run affected project tests, typechecks, formatting, and lint.
- Browser verification covers search/open/edit/review/commit and stopping or switching targets. Use controlled model responses for deterministic coverage where practical; report live-model or full-stack checks blocked by deferred billing setup explicitly.
- Add a plain-language walkthrough explaining the browser executor, backend tools, continuation accounting, saving/review semantics, and how to test and extend the feature.
