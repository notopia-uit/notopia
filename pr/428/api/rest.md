# Notopia API

Version: `0.0.0`

Notopia API

## Servers

- `http://api.notopia.localhost`

## Security schemes

- `document_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). OpenID Connect Security Scheme
- `document_Oauth2`: OAuth 2.0. OAuth2 Security Scheme
  - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
  - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/
- `note_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). OpenID Connect Security Scheme
- `note_Oauth2`: OAuth 2.0. OAuth2 Security Scheme
  - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
  - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/
- `billing_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). OpenID Connect Security Scheme
- `billing_Oauth2`: OAuth 2.0. OAuth2 Security Scheme
  - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
  - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/

## Document

Note content for editing, collaborating

- [Get document attachment upload url](/notopia/pr/428/api/rest/operations/getDocumentAttachmentUploadUrl.md) `GET /document/documents/{documentId}/attachment-url`
- [Commit document](/notopia/pr/428/api/rest/operations/commitDocument.md) `POST /document/documents/{documentId}/commit`

## Revision

Document revisions

- [Get revisions](/notopia/pr/428/api/rest/operations/getRevisions.md) `GET /document/revisions`
- [Get revision with content](/notopia/pr/428/api/rest/operations/getRevisionWithContent.md) `GET /document/revisions/{revisionId}`
- [Delete revision](/notopia/pr/428/api/rest/operations/deleteRevision.md) `DELETE /document/revisions/{revisionId}`
- [Rename revision](/notopia/pr/428/api/rest/operations/renameRevision.md) `POST /document/revisions/{revisionId}/rename`

## Workspace

Workspace

- [Get workspace](/notopia/pr/428/api/rest/operations/getWorkspace.md) `GET /note/workspaces-by-slug/{workspaceSlug}`
- [Check workspace slug exists](/notopia/pr/428/api/rest/operations/checkWorkspaceSlugExists.md) `GET /note/workspaces-by-slug/{workspaceSlug}/exists`
- [Create workspace](/notopia/pr/428/api/rest/operations/createWorkspace.md) `POST /note/workspaces`
- [Get my workspaces](/notopia/pr/428/api/rest/operations/getMyWorkspaces.md) `GET /note/workspaces/me`
- [Delete workspace](/notopia/pr/428/api/rest/operations/deleteWorkspace.md) `DELETE /note/workspaces/{workspaceId}`
- [Change workspace slug](/notopia/pr/428/api/rest/operations/changeWorkspaceSlug.md) `POST /note/workspaces/{workspaceId}/change-slug`
- [Empty trash](/notopia/pr/428/api/rest/operations/emptyTrash.md) `DELETE /note/workspaces/{workspaceId}/empty-trash`
- [SSE workspace updates](/notopia/pr/428/api/rest/operations/getWorkspaceEvents.md) `GET /note/workspaces/{workspaceId}/events`
- [Get workspace graph](/notopia/pr/428/api/rest/operations/getWorkspaceGraph.md) `GET /note/workspaces/{workspaceId}/graph`
- [Leave workspace](/notopia/pr/428/api/rest/operations/leaveWorkspace.md) `POST /note/workspaces/{workspaceId}/leave`
- [Get workspace members](/notopia/pr/428/api/rest/operations/getWorkspaceMembers.md) `GET /note/workspaces/{workspaceId}/members`
- [Update workspace members](/notopia/pr/428/api/rest/operations/updateWorkspaceMembers.md) `PUT /note/workspaces/{workspaceId}/members`
- [Move workspace's items](/notopia/pr/428/api/rest/operations/moveWorkspaceItems.md) `POST /note/workspaces/{workspaceId}/move-items`
- [Permanently delete workspace items](/notopia/pr/428/api/rest/operations/permanentlyDeleteWorkspaceItems.md) `POST /note/workspaces/{workspaceId}/permanently-delete-items`
- [Publish workspace](/notopia/pr/428/api/rest/operations/publishWorkspace.md) `POST /note/workspaces/{workspaceId}/publish`
- [Rename workspace](/notopia/pr/428/api/rest/operations/renameWorkspace.md) `POST /note/workspaces/{workspaceId}/rename`
- [Restore trashed workspace items](/notopia/pr/428/api/rest/operations/restoreTrashedWorkspaceItems.md) `POST /note/workspaces/{workspaceId}/restore-trashed-items`
- [Get Workspace Search Token](/notopia/pr/428/api/rest/operations/getWorkspaceSearchToken.md) `GET /note/workspaces/{workspaceId}/search-token`
- [Show trash](/notopia/pr/428/api/rest/operations/showTrash.md) `GET /note/workspaces/{workspaceId}/show-trash`
- [Trash workspace's items](/notopia/pr/428/api/rest/operations/trashWorkspaceItems.md) `POST /note/workspaces/{workspaceId}/trash-items`
- [Get workspace tree](/notopia/pr/428/api/rest/operations/getWorkspaceTree.md) `GET /note/workspaces/{workspaceId}/tree`
- [Unpublish workspace](/notopia/pr/428/api/rest/operations/unpublishWorkspace.md) `POST /note/workspaces/{workspaceId}/unpublish`

## Folder

Folder in workspace

- [Create folder](/notopia/pr/428/api/rest/operations/createFolder.md) `POST /note/folders`
- [Permanently delete folder](/notopia/pr/428/api/rest/operations/permanentlyDeleteFolder.md) `DELETE /note/folders/{folderId}`
- [Rename folder](/notopia/pr/428/api/rest/operations/renameFolder.md) `POST /note/folders/{folderId}/rename`

## Note

Note in workspace's root folder or in a folder

- [Create note](/notopia/pr/428/api/rest/operations/createNote.md) `POST /note/notes`
- [Get note](/notopia/pr/428/api/rest/operations/getNote.md) `GET /note/notes/{noteId}`
- [Permanently delete note](/notopia/pr/428/api/rest/operations/permanentlyDeleteNote.md) `DELETE /note/notes/{noteId}`
- [Get note graph](/notopia/pr/428/api/rest/operations/getNoteGraph.md) `GET /note/notes/{noteId}/graph`
- [Get note links](/notopia/pr/428/api/rest/operations/getNoteLinks.md) `GET /note/notes/{noteId}/links`
- [Publish note](/notopia/pr/428/api/rest/operations/publishNote.md) `POST /note/notes/{noteId}/publish`
- [Rename note](/notopia/pr/428/api/rest/operations/renameNote.md) `POST /note/notes/{noteId}/rename`
- [Unpublish note](/notopia/pr/428/api/rest/operations/unpublishNote.md) `POST /note/notes/{noteId}/unpublish`

## NoteUser

User from Note service

- [Search users](/notopia/pr/428/api/rest/operations/searchUsers.md) `GET /note/search-users`

## Plan

Public plan catalog

- [List plans](/notopia/pr/428/api/rest/operations/listPlans.md) `GET /billing/plans`

## WorkspaceBilling

Subscription, entitlements and payer of a workspace

- [Get workspace entitlements](/notopia/pr/428/api/rest/operations/getWorkspaceEntitlements.md) `GET /billing/workspaces/{workspaceId}/entitlements`
- [Get workspace billing](/notopia/pr/428/api/rest/operations/getWorkspaceBilling.md) `GET /billing/workspaces/{workspaceId}`
- [Create checkout session](/notopia/pr/428/api/rest/operations/createCheckoutSession.md) `POST /billing/workspaces/{workspaceId}/checkout-session`
- [Create portal session](/notopia/pr/428/api/rest/operations/createPortalSession.md) `POST /billing/workspaces/{workspaceId}/portal-session`
- [Preview immediate cancellation](/notopia/pr/428/api/rest/operations/previewCancelNow.md) `GET /billing/workspaces/{workspaceId}/subscription/cancel-now/preview`
- [Cancel subscription now](/notopia/pr/428/api/rest/operations/cancelNow.md) `POST /billing/workspaces/{workspaceId}/subscription/cancel-now`
- [Claim payer role](/notopia/pr/428/api/rest/operations/claimPayer.md) `POST /billing/workspaces/{workspaceId}/payer/claim`

## BillingAdmin

Admin-only catalog management (group notopia-admins)

- [List plans](/notopia/pr/428/api/rest/operations/adminListPlans.md) `GET /billing/admin/plans`
- [Update plan](/notopia/pr/428/api/rest/operations/adminUpdatePlan.md) `PATCH /billing/admin/plans/{tier}`
- [Get billing policy](/notopia/pr/428/api/rest/operations/adminGetPolicy.md) `GET /billing/admin/policy`
- [Update billing policy](/notopia/pr/428/api/rest/operations/adminUpdatePolicy.md) `PATCH /billing/admin/policy`
- [List prices](/notopia/pr/428/api/rest/operations/adminListPrices.md) `GET /billing/admin/prices`
- [Create price](/notopia/pr/428/api/rest/operations/adminCreatePrice.md) `POST /billing/admin/prices`
- [Archive price](/notopia/pr/428/api/rest/operations/adminArchivePrice.md) `POST /billing/admin/prices/{priceId}/archive`

## BillingWebhook

Payment provider webhooks

- [Stripe webhook endpoint](/notopia/pr/428/api/rest/operations/stripeWebhook.md) `POST /billing/webhooks/stripe`
