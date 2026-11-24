# Homelab RustFS

This guide prepares RustFS for document attachments.

## 1. Connect to RustFS

You need the RustFS **S3 API** HTTPS endpoint, its region, and an existing
non-root administrator credential with permission to manage the bucket, users,
groups and policies. Find those details in your existing credential manager or
ask your RustFS administrator. If no suitable account exists, stop and follow
your administrator's approved bootstrap or recovery procedure. Do not guess
credentials or use the default root account.

> [!NOTE]
> You can use `us-east-1` as the region if it matches your RustFS configuration.

:::tabs key:rustfs-method

== CLI

Install the development tools with `mise install`, then check `rc --version`.
On a trusted terminal, list existing aliases first. The next command replaces
an alias with the same name:

```sh
rc alias list
rc alias set homelab https://<s3-api-host> --region <region>
```

On a terminal, current `rc` versions prompt for the access and secret keys when
omitted. Confirm your version supports this; do not put credentials in command
arguments or shell history. Verify the connection and administrative access:

```sh
rc ls homelab/
rc admin user list homelab
```

`rc` stores alias credentials in `~/.config/rc/config.toml`. Restrict access
to your account and protect the machine, backups and editor. Do not commit or
share this file. When finished, `rc alias remove homelab` removes the local
alias, not the RustFS account.

== Console

Open the RustFS Console URL, which is separate from the S3 API endpoint.
Sign in with an existing authorized non-root administrator using Key Login or
your configured identity provider. If the Console cannot reach RustFS, check
its Server Configuration. See the
[RustFS Console guide](https://github.com/rustfs/docs.rustfs.com/blob/main/content/en/administration/console/index.md).

:::

## 2. Inspect and create the bucket

If `notopia-document`, `notopiaDocumentUpload`, `notopiaDocumentDeveloper`,
or `notopiaDevelopers` already exists, inspect its ownership and configuration.
Do not replace or reuse it until you understand the effects. Back up existing
bucket policies and CORS rules before changing them.

:::tabs key:rustfs-method

== CLI

Inspect existing resources and create a new bucket:

```sh
rc ls homelab/
rc admin user list homelab
rc admin policy list homelab
rc admin group ls homelab
rc mb homelab/notopia-document
```

== Console

Inspect the existing buckets and the users, policies and groups under Identity.
On **Buckets**, select **Create Bucket**, enter `notopia-document`, and select
**Create**. Verify the bucket appears in the list.

:::

## 3. Create the application identity

Create a non-secret file `notopiaDocumentUpload.json` containing:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:PutObject"],
      "Resource": ["arn:aws:s3:::notopia-document/*"]
    }
  ]
}
```

:::tabs key:rustfs-method

== CLI

Create the policy:

```sh
rc admin policy create homelab notopiaDocumentUpload notopiaDocumentUpload.json
```

The current `rc` requires you to supply the user secret and service-account
keys; it does not generate them. Generate independent random values and create
the parent user.

```sh
NOTOPIA_USER_SECRET="$(openssl rand -hex 32)"
rc admin user add homelab notopia-document "$NOTOPIA_USER_SECRET"
rc admin policy attach homelab notopiaDocumentUpload --user notopia-document
rc admin user info homelab notopia-document
```

Generate a separate service-account key pair and create it under the parent.
`--user` requires an administrator permitted to create a service account for
another user; if the server denies it, do not omit `--user` while using the
administrator alias, because that would use the wrong parent.

```sh
NOTOPIA_S3_ACCESS_KEY="$(openssl rand -hex 10)"
NOTOPIA_S3_SECRET_KEY="$(openssl rand -hex 32)"
rc admin service-account create homelab "$NOTOPIA_S3_ACCESS_KEY" "$NOTOPIA_S3_SECRET_KEY" --user notopia-document
rc admin service-account info homelab "$NOTOPIA_S3_ACCESS_KEY"
```

Confirm that the parent is `notopia-document`. Without an additional session
policy, the service account inherits the parent's upload policy. Save the
service-account key pair in a protected secret manager for step 6; do not use
the parent credential for the application.

== Console

1. Under **Identity**, create the policy `notopiaDocumentUpload` with the JSON
   above. Create the IAM user `notopia-document` with a unique strong secret,
   then attach that policy to the user.
2. Sign in as `notopia-document`. Open **Access Keys**, select **Add Access
   Key**, and submit the form. Save the generated access key and secret in a
   protected secret manager for step 6.
3. Confirm the key belongs to `notopia-document` and inherits its policy. Do not
   use the parent user credential for the application. Return to the
   administrator session for the next step.

Console menus and parent-selection support vary by version. If you cannot
verify the parent or create the policy, stop and consult
[RustFS IAM documentation](https://github.com/rustfs/docs.rustfs.com/blob/main/content/en/security-compliance/iam/index.md).

:::

## 4. Create the developers group

Create `notopiaDocumentDeveloper.json` with bucket-scoped object and configuration
permissions. This permits members to read, upload and delete objects, list the
bucket, and inspect or change its CORS and public-access policy. It does not
permit bucket deletion, IAM administration, or access to other buckets.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListBucket",
        "s3:GetBucketPolicy",
        "s3:PutBucketPolicy",
        "s3:GetBucketCORS",
        "s3:PutBucketCORS"
      ],
      "Resource": ["arn:aws:s3:::notopia-document"]
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
      "Resource": ["arn:aws:s3:::notopia-document/*"]
    }
  ]
}
```

:::tabs key:rustfs-method

== CLI

```sh
rc admin policy create homelab notopiaDocumentDeveloper notopiaDocumentDeveloper.json
rc admin group add homelab notopiaDevelopers
rc admin policy attach homelab notopiaDocumentDeveloper --group notopiaDevelopers
rc admin group add-members homelab notopiaDevelopers <existing-developer-user>
rc admin group info homelab notopiaDevelopers
```

== Console

Under **Identity**, create the `notopiaDocumentDeveloper` policy with the JSON
above. Create the `notopiaDevelopers` group, attach the policy, and add approved
existing developer users. Inspect the group to confirm its policy and members.
If your Console version lacks these controls, use the CLI tab.

:::

Repeat the membership command for approved individual users. Do not add the
`notopia-document` application user: it must not inherit developer permissions.

## 5. Allow public reads and configure CORS

:::tabs key:rustfs-method
== CLI

Inspect existing settings first:

```sh
rc anonymous get-json homelab/notopia-document
rc cors get homelab/notopia-document
```

== Console

Open `notopia-document` in **Buckets** and inspect its existing bucket policy
and CORS settings. If your Console version cannot inspect or edit these settings,
use the CLI tab. Do not replace existing settings without reviewing them.

:::

> [!WARNING]
> A missing policy or CORS configuration can be expected on a new bucket. Stop
> on permission or connection errors. Both commands below **replace** existing
> bucket configuration; merge existing rules deliberately instead of overwriting
> them.

For a new bucket without a policy, create `publicRead.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::notopia-document/*"]
    }
  ]
}
```

:::tabs key:rustfs-method
== CLI

```sh
rc anonymous set-json publicRead.json homelab/notopia-document
rc anonymous get-json homelab/notopia-document
```

== Console

If your Console supports custom bucket policies, apply the JSON above to
`notopia-document`, then inspect the saved policy. Otherwise use the CLI tab.

:::

This grants no anonymous writes or bucket listing. In particular, do not use
`anonymous set public`, which also permits uploads.

Create `cors.xml`, replacing the example with the exact browser origin. Add
other trusted origins only if needed:

```xml
<CORSConfiguration>
  <CORSRule>
    <AllowedOrigin>https://notopia.example.com</AllowedOrigin>
    <AllowedMethod>PUT</AllowedMethod>
    <AllowedMethod>GET</AllowedMethod>
    <AllowedMethod>HEAD</AllowedMethod>
    <AllowedHeader>*</AllowedHeader>
    <ExposeHeader>ETag</ExposeHeader>
    <MaxAgeSeconds>3600</MaxAgeSeconds>
  </CORSRule>
</CORSConfiguration>
```

:::tabs key:rustfs-method

== CLI

```sh
rc cors set homelab/notopia-document cors.xml
rc cors get homelab/notopia-document
```

== Console

If your Console supports bucket CORS configuration, apply the origins, methods,
headers and max age shown above, then inspect the saved settings. Otherwise use
the CLI tab.

:::

CORS only controls browser access; it does not grant S3 permissions. If RustFS
or `rc` reports a capability as unsupported, stop and consult the version's
documentation rather than forcing the operation.

## 6. Configure Notopia

The service account's access key and secret go into the existing document Secret:

- `NOTOPIA_DOCUMENT_S3_ACCESS_KEY_ID`
- `NOTOPIA_DOCUMENT_S3_SECRET_ACCESS_KEY`

Follow the secure editor and SOPS instructions in the
[homelab provisioning runbook](https://github.com/notopia-uit/notopia/blob/main/deploy/k8s/clusters/homelab/README.md#secrets).

Set these non-secret values in the document Helm values when preparing the
deployment:

- `NOTOPIA_DOCUMENT_S3_ENDPOINT`: HTTPS S3 API endpoint.
- `NOTOPIA_DOCUMENT_S3_REGION`: the bucket's region, such as `us-east-1`.
- `NOTOPIA_DOCUMENT_S3_BUCKET_NAME`: `notopia-document`.

The service uses path-style URLs and the same endpoint for signing and public
URLs. The browser must resolve and reach that endpoint, and proxies must preserve
the signed host and request path. Keep releases suspended until all deployment
prerequisites in the runbook are satisfied.

## 7. Verify

On the actual RustFS version, verify with a disposable, non-sensitive object:

1. Upload using the application child credential and confirm success. Check
   that the same credential cannot write to another bucket or administer IAM.
2. Retrieve the known object URL without credentials. Confirm anonymous upload
   and bucket listing fail. An unguessable URL is not access control.
3. From the real web origin, verify browser preflight and an upload using a
   document-service presigned URL. Check the browser can fetch the public URL.
4. Delete the test object using a developer or administrator identity; the
   application identity intentionally lacks delete permission.

Do not expose credentials or presigned URLs in logs, tickets, or shell history.
Inspect effective policies if any denied test unexpectedly succeeds. Credential
replacement is not required for initial setup; see the runbook's
[coordinated secret rotation](https://github.com/notopia-uit/notopia/blob/main/deploy/k8s/clusters/homelab/README.md#coordinated-secret-rotation)
procedure when a key needs replacement.
