import '@blocknote/core';
import { createBlockNoteSchema } from './block-note';

declare global {
  declare namespace NodeJS {
    interface ProcessEnv {
      BETTER_AUTH_SECRET?: string;
      BETTER_AUTH_URL?: string;
      AUTHENTIK_CLIENT_ID?: string;
      AUTHENTIK_CLIENT_SECRET?: string;
      AUTHENTIK_CLIENT_DISCOVERY_URL?: string;
      AUTHENTIK_REDIRECT_URI?: string;
      NEXT_PUBLIC_RUSTFS_ENDPOINT?: string;
      NEXT_PUBLIC_RUSTFS_BUCKET?: string;
      NEXT_PUBLIC_RUSTFS_REGION?: string;
      RUSTFS_ACCESS_KEY?: string;
      RUSTFS_SECRET_KEY?: string;
    }
  }
}

declare module '@blocknote/core' {
  export type MySchema = ReturnType<typeof createBlockNoteSchema>;
  export type DefaultStyleSchema = MySchema['styleSchema'];
  export type DefaultBlockSchema = MySchema['blockSchema'];
  export type DefaultInlineContentSchema = MySchema['inlineContentSchema'];

  export type MyEditor = MySchema['BlockNoteEditor'];
}
