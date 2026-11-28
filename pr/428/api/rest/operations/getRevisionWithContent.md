# Get revision with content

`GET /document/revisions/{revisionId}`

## Servers

- `http://api.notopia.localhost`

## Authorization

Use one of the following options:

- Option 1:
  - `document_Oauth2`: OAuth 2.0. Required scopes: `openid`. OAuth2 Security Scheme
    - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
    - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/
- Option 2:
  - `document_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). Required scopes: `openid`. OpenID Connect Security Scheme

## Parameters

### Path parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `revisionId` | `string<uuid>` | Yes | Example: `000c2575-1847-5293-8117-2415a2328ef8`. |

## Responses

### 200

Revision retrieved successfully

Content type: `application/json`

Type: `object` (schema `document_RevisionWithContent`).

- `id` (`string<uuid>`, required): Example: `000c2575-1847-5293-8117-2415a2328ef8`.
- `name` (`string | null`, required): Constraints: minLength: 1, maxLength: 255. Example: `Initial Draft`.
- `createdAt` (`string<date-time>`, required): Example: `2024-01-01T12:00:00Z`.
- `authorId` (`string | null`, required): Who the revision is attributed to. Null for revisions created before authors were recorded. Example: `42`.
- `trigger` (`string`, schema `document_RevisionTrigger`, required): Possible values: `manual`, `auto`, `sweeper`. Example: `manual`.
- `contributorIds` (`array<string>`, required): Users who edited the document since the previous revision, most recent last. Example: `["41","42"]`.
- `content` (`array<object>`, schema `document_DocumentContent`, required): BlockNote model. Example: `[{"type":"paragraph","content":"Welcome to this demo!"},{"type":"paragraph"},{"type":"paragraph","content":[{"type":"text","text":"Blocks:","styles":{"bold":true}}]},{"type":"paragraph","content":"Paragraph"},{"type":"heading","content":"Heading"},{"id":"toggle-heading","type":"heading","props":{"isToggleable":true},"content":"Toggle Heading"},{"type":"quote","content":"Quote"},{"type":"bulletListItem","content":"Bullet List Item"},{"type":"numberedListItem","content":"Numbered List Item"},{"type":"checkListItem","content":"Check List Item"},{"id":"toggle-list-item","type":"toggleListItem","content":"Toggle List Item"},{"type":"codeBlock","props":{"language":"javascript"},"content":"console.log('Hello, world!');"},{"type":"table","content":{"type":"tableContent","rows":[{"cells":["Table Cell","Table Cell","Table Cell"]},{"cells":["Table Cell","Table Cell","Table Cell"]},{"cells":["Table Cell","Table Cell","Table Cell"]}]}},{"type":"file"},{"type":"image","props":{"url":"https://interactive-examples.mdn.mozilla.net/media/cc0-images/grapefruit-slice-332-332.jpg","caption":"From https://interactive-examples.mdn.mozilla.net/media/cc0-images/grapefruit-slice-332-332.jpg"}},{"type":"video","props":{"url":"https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.webm","caption":"From https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.webm"}},{"type":"audio","props":{"url":"https://interactive-examples.mdn.mozilla.net/media/cc0-audio/t-rex-roar.mp3","caption":"From https://interactive-examples.mdn.mozilla.net/media/cc0-audio/t-rex-roar.mp3"}},{"type":"paragraph"},{"type":"paragraph","content":[{"type":"text","text":"Inline Content:","styles":{"bold":true}}]},{"type":"paragraph","content":[{"type":"text","text":"Styled Text","styles":{"bold":true,"italic":true,"textColor":"red","backgroundColor":"blue"}},{"type":"text","text":" ","styles":{}},{"type":"link","content":"Link","href":"https://www.blocknotejs.org"}]},{"type":"paragraph"}]`.

Example:

```json
{
  "id": "000c2575-1847-5293-8117-2415a2328ef8",
  "name": null,
  "content": [
    {
      "id": "d74b509e-84f4-4cc8-a7a5-86f4b50ddcac",
      "type": "heading",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left",
        "level": 1,
        "isToggleable": false
      },
      "content": [
        {
          "type": "text",
          "text": "PowerUp.ps1",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "e8c23d73-b072-4226-979e-76ebde0f7ad9",
      "type": "quote",
      "props": {
        "backgroundColor": "default",
        "textColor": "default"
      },
      "content": [
        {
          "type": "text",
          "text": " [!Note] Assume we're using the same scenario as in ",
          "styles": {}
        },
        {
          "type": "reference",
          "props": {
            "noteId": "156390b6-4b24-54c5-ae7f-77e462d7f106"
          }
        },
        {
          "type": "text",
          "text": " and the other notes in this section. ",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "cb8066a6-ee55-4432-9e78-a89dd6cf4d6c",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "link",
          "href": "https://github.com/PowerShellMafia/PowerSploit/tree/master/Privesc",
          "content": [
            {
              "type": "text",
              "text": "PowerUp.ps1",
              "styles": {
                "bold": true
              }
            }
          ]
        },
        {
          "type": "text",
          "text": " is a tool which we can use to detect ",
          "styles": {}
        },
        {
          "type": "reference",
          "props": {
            "noteId": "3961c9dc-b75d-5fe5-bac4-ff8d20cc3c96"
          }
        },
        {
          "type": "text",
          "text": " privesc vectors. To use it, we just need to copy it to our ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "attacking machine",
          "styles": {
            "italic": true
          }
        },
        {
          "type": "text",
          "text": ", then use a python HTTP server and ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "iwr",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " to infiltrate it onto the victim Windows machine.",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "4572d125-29ab-4966-a3c5-ffef57e59ab7",
      "type": "heading",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left",
        "level": 2,
        "isToggleable": false
      },
      "content": [
        {
          "type": "text",
          "text": "Use",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "792e4288-4c00-4c5e-b4df-16880773f899",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "After getting the script onto our target box, we need to start ",
          "styles": {}
        },
        {
          "type": "reference",
          "props": {
            "noteId": "8442c808-cf47-560b-9ab6-c7537e739c80"
          }
        },
        {
          "type": "text",
          "text": " with the ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "ExecutionPolicy Bypass",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " to allow our current ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "dave",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " user to execute the script:",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "ae1800a9-7de7-445d-9842-85030702929d",
      "type": "codeBlock",
      "props": {
        "language": "powershell"
      },
      "content": [
        {
          "type": "text",
          "text": "PS C:\\Users\\dave> powershell -ep bypass\n...\nPS C:\\Users\\dave>  . .\\PowerUp.ps1\n",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "30b179e3-c502-4e9a-8ff4-4224f6744c4e",
      "type": "heading",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left",
        "level": 3,
        "isToggleable": false
      },
      "content": [
        {
          "type": "text",
          "text": "Get-ModifiableServiceFile",
          "styles": {
            "code": true
          }
        }
      ],
      "children": []
    },
    {
      "id": "e1cd04b5-e439-4cab-bf6a-014fe2f573c2",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "After importing the script, we can use ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "Get-ModifiableServiceFile",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " which is a function that displays ",
          "styles": {}
        },
        {
          "type": "reference",
          "props": {
            "noteId": "fc7da9b5-5ad9-56e8-a97d-56fa7a67435e"
          }
        },
        {
          "type": "text",
          "text": " the current user ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "is allowed to modify",
          "styles": {
            "italic": true
          }
        },
        {
          "type": "text",
          "text": ":",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "76d87037-63ed-4224-b560-6e77aa95f89e",
      "type": "codeBlock",
      "props": {
        "language": "powershell"
      },
      "content": [
        {
          "type": "text",
          "text": "PS C:\\Users\\dave> Get-ModifiableServiceFile\n\n...\n\nServiceName                     : mysql\nPath                            : C:\\xampp\\mysql\\bin\\mysqld.exe --defaults-file=c:\\xampp\\mysql\\bin\\my.ini mysql\nModifiableFile                  : C:\\xampp\\mysql\\bin\\mysqld.exe\nModifiableFilePermissions       : {WriteOwner, Delete, WriteAttributes, Synchronize...}\nModifiableFileIdentityReference : BUILTIN\\Users\nStartName                       : LocalSystem\nAbuseFunction                   : Install-ServiceBinary -Name 'mysql'\nCanRestart                      : False\n",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "d6112827-9de6-445b-8e9a-c33655911d12",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "PowerUp identified ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "mysqld.exe",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " in the ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "\\xampp",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " directory as vulnerable (along with some others). It also tells us that we ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "don't have permission to restart the service",
          "styles": {
            "italic": true
          }
        },
        {
          "type": "text",
          "text": ".",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "52bbc039-761d-4b06-aea9-ac69c5007314",
      "type": "heading",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left",
        "level": 3,
        "isToggleable": false
      },
      "content": [
        {
          "type": "text",
          "text": "AbuseFunction",
          "styles": {
            "code": true
          }
        }
      ],
      "children": []
    },
    {
      "id": "86451bc9-04ea-49b7-b1af-d653faa92fbc",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "Listed in the above output is the function ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "AbuseFunction",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": ". We can use this to replace the binary (if our current user has permission to). This is a built in function, and it use it, we just have to call the command listed to the right of it.",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "a8722f26-bfdb-4211-a14e-9da18f193d96",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "Calling this function results in its default action: create a new user named ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "john",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " with the password ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "Password123!",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " and add him to the local ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "Administrators",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " group. To call it, our current user account also needs permissions to reboot the machine. Unfortunately, our current user ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "dave",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " does not have those permissions, so calling it would result in:",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "fe60d5a6-558a-4a76-b8dd-687d483637ec",
      "type": "codeBlock",
      "props": {
        "language": "powershell"
      },
      "content": [
        {
          "type": "text",
          "text": "PS C:\\Users\\dave> Install-ServiceBinary -Name 'mysql'\n\nService binary 'C:\\xampp\\mysql\\bin\\mysqld.exe --defaults-file=c:\\xampp\\mysql\\bin\\my.ini mysql' for service mysql not\nmodifiable by the current user.\nAt C:\\Users\\dave\\PowerUp.ps1:2178 char:13\n+             throw \"Service binary '$($ServiceDetails.PathName)' for s ...\n+             ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n    + CategoryInfo          : OperationStopped: (Service binary ...e current user.:String) [], RuntimeException\n    + FullyQualifiedErrorId : Service binary 'C:\\xampp\\mysql\\bin\\mysqld.exe --defaults-file=c:\\xampp\\mysql\\bin\\my.ini\n   mysql' for service mysql not modifiable by the current user.\n",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "84d0d005-cb05-4cde-85ca-708aa14a9d21",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "Notice that the above error isn't regarding whether ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "dave",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " has reboot permissions. Instead, it describes the the path for the ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "mysql",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " service is ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "not modifiable by the current user",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": ". This is a bug you can track down manually in the ",
          "styles": {}
        },
        {
          "type": "link",
          "href": "https://github.com/PowerShellMafia/PowerSploit/blob/master/Privesc/PowerUp.ps1",
          "content": [
            {
              "type": "text",
              "text": "PowerUp.ps1 code",
              "styles": {}
            }
          ]
        },
        {
          "type": "text",
          "text": ". Basically, when you give a path to a PS function defined in the code as ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "Get-ModifiablePath",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": ", it breaks. ",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "09391e37-f22b-4282-a056-11e845b1cdc4",
      "type": "paragraph",
      "props": {
        "backgroundColor": "default",
        "textColor": "default",
        "textAlignment": "left"
      },
      "content": [
        {
          "type": "text",
          "text": "Since this function is called when you run the ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "AbuseFunction",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " (",
          "styles": {}
        },
        {
          "type": "text",
          "text": "Install-ServiceBinary",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": "), it triggers this error, even though we already know the current user ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "dave",
          "styles": {
            "code": true
          }
        },
        {
          "type": "text",
          "text": " ",
          "styles": {}
        },
        {
          "type": "text",
          "text": "can modify the binary",
          "styles": {
            "italic": true
          }
        },
        {
          "type": "text",
          "text": ". This is all to say that you, again, should never blindly trust an automated tool. Offsec gives the following manual test steps to establish this:",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "09edacc0-f0cc-459a-98c4-29190df2befa",
      "type": "codeBlock",
      "props": {
        "language": "powershell"
      },
      "content": [
        {
          "type": "text",
          "text": "PS C:\\Users\\dave> $ModifiableFiles = echo 'C:\\xampp\\mysql\\bin\\mysqld.exe' | Get-ModifiablePath -Literal\n\nPS C:\\Users\\dave> $ModifiableFiles\n\nModifiablePath                IdentityReference Permissions\n--------------                ----------------- -----------\nC:\\xampp\\mysql\\bin\\mysqld.exe BUILTIN\\Users     {WriteOwner, Delete, WriteAttributes, Synchronize...}\n\nPS C:\\Users\\dave> $ModifiableFiles = echo 'C:\\xampp\\mysql\\bin\\mysqld.exe argument' | Get-ModifiablePath -Literal\n\nPS C:\\Users\\dave> $ModifiableFiles\n\nModifiablePath     IdentityReference                Permissions\n--------------     -----------------                -----------\nC:\\xampp\\mysql\\bin NT AUTHORITY\\Authenticated Users {Delete, WriteAttributes, Synchronize, ReadControl...}\nC:\\xampp\\mysql\\bin NT AUTHORITY\\Authenticated Users {Delete, GenericWrite, GenericExecute, GenericRead}\n\nPS C:\\Users\\dave> $ModifiableFiles = echo 'C:\\xampp\\mysql\\bin\\mysqld.exe argument -conf=C:\\test\\path' | Get-ModifiablePath -Literal \n\nPS C:\\Users\\dave> $ModifiableFiles\n",
          "styles": {}
        }
      ],
      "children": []
    },
    {
      "id": "51754503-482a-4a72-a375-8ab3696cc474",
      "type": "quote",
      "props": {
        "backgroundColor": "default",
        "textColor": "default"
      },
      "content": [
        {
          "type": "text",
          "text": " [!Resources]  ",
          "styles": {}
        },
        {
          "type": "link",
          "href": "https://github.com/PowerShellMafia/PowerSploit/tree/master/Privesc",
          "content": [
            {
              "type": "text",
              "text": "PowerUp.ps1",
              "styles": {
                "bold": true
              }
            }
          ]
        },
        {
          "type": "text",
          "text": " My ",
          "styles": {}
        },
        {
          "type": "link",
          "href": "https://github.com/trshpuppy/obsidian-notes",
          "content": [
            {
              "type": "text",
              "text": "own notes",
              "styles": {}
            }
          ]
        },
        {
          "type": "text",
          "text": " linked throughout the text.  ",
          "styles": {}
        }
      ],
      "children": []
    }
  ],
  "createdAt": "2026-04-27T10:16:34.910Z",
  "authorId": "42",
  "trigger": "manual",
  "contributorIds": [
    "41",
    "42"
  ]
}
```

### 400

Bad Request Error response

Content type: `application/json`

Type: `object` (schema `document_Error`).

- `code` (`string`, required): Error code. Example: `EXAMPLE_CODE`.
- `message` (`string`, required): Human-readable error message. Example: `Example error`.
- `more_info` (`string`): URL with more information about the error. Example: `https://api.notopia.com/docs/errors/EXAMPLE_CODE`.

Example:

```json
{
  "code": "BAD_REQUEST",
  "message": "The request could not be understood by the server due to malformed syntax",
  "more_info": "https://api.notopia.com/docs/errors/BAD_REQUEST"
}
```

### 401

Unauthorized Error response

Content type: `application/json`

Type: `object`.

Example:

```json
{}
```

### 403

Forbidden Error response

Content type: `application/json`

Type: `object` (see schema `document_Error` above).

Example:

```json
{
  "code": "FORBIDDEN",
  "message": "You do not have permission to access this resource",
  "more_info": "https://api.notopia.com/docs/errors/FORBIDDEN"
}
```

### 404

Not Found Error response

Content type: `application/json`

Type: `object` (see schema `document_Error` above).

Example:

```json
{
  "code": "NOT_FOUND",
  "message": "The requested resource was not found",
  "more_info": "https://api.notopia.com/docs/errors/NOT_FOUND"
}
```

### 500

Internal Server Error response

Content type: `application/json`

Type: `object` (see schema `document_Error` above).

Example:

```json
{
  "code": "INTERNAL_SERVER_ERROR",
  "message": "An internal server error occurred. Please try again later",
  "more_info": "https://api.notopia.com/docs/errors/INTERNAL_SERVER_ERROR"
}
```

## Code samples

**cURL**

```bash
curl http://api.notopia.localhost/document/revisions/000c2575-1847-5293-8117-2415a2328ef8 \
  --header 'Authorization: Bearer document_Oauth2'
```
