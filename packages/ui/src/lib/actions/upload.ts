'use server';

import { getDocumentAttachmentUploadUrl } from '@notopia-uit/api-gen';

/**
 * Server Action for uploading files to document attachments
 * 
 * This runs on the server and handles:
 * 1. Getting presigned upload URL from backend API
 * 2. Uploading the file to the presigned URL
 * 3. Returning the final file URL
 * 
 * Benefits:
 * - API credentials never exposed to browser
 * - Server controls all authentication/authorization
 * - Can add validation, logging, monitoring on server
 */
export async function uploadDocumentAttachment(
  documentId: string,
  file: File
): Promise<string> {
  try {
    // Get presigned URLs from backend API
    const response = await getDocumentAttachmentUploadUrl({
      path: {
        documentId,
      },
    });

    if (response.error || !response.data) {
      throw new Error('Failed to get upload URL from server');
    }

    const { uploadUrl, url } = response.data;

    // Convert File to Buffer for server-side processing
    const arrayBuffer = await file.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);

    // Upload file using presigned URL
    const uploadResponse = await fetch(uploadUrl, {
      method: 'PUT',
      body: buffer,
      headers: {
        'Content-Type': file.type || 'application/octet-stream',
      },
    });

    if (!uploadResponse.ok) {
      throw new Error(
        `Failed to upload file: ${uploadResponse.status} ${uploadResponse.statusText}`
      );
    }

    // Return the final accessible URL
    return url;
  } catch (error) {
    console.error('Upload error:', error);
    throw new Error(
      error instanceof Error
        ? error.message
        : 'Unknown error during file upload'
    );
  }
}
