import { getDocumentAttachmentUploadUrl } from '@notopia-uit/api-gen';

export async function uploadDocumentAttachment(documentId: string, file: File): Promise<string> {
  const response = await getDocumentAttachmentUploadUrl({
    path: {
      documentId,
    },
  });

  if (response.error || !response.data) {
    throw new Error('Failed to get upload URL from server');
  }

  const { uploadUrl, url } = response.data;

  const uploadResponse = await fetch(uploadUrl, {
    method: 'PUT',
    body: file,
    headers: {
      'Content-Type': file.type || 'application/octet-stream',
    },
  });

  if (!uploadResponse.ok) {
    throw new Error(
      `Failed to upload file: ${uploadResponse.status} ${uploadResponse.statusText}`
    );
  }

  return url;
}
