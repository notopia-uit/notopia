declare namespace NodeJS {
  interface ProcessEnv {
    BETTER_AUTH_SECRET?: string;
    BETTER_AUTH_URL?: string;
    AUTHENTIK_CLIENT_ID?: string;
    AUTHENTIK_CLIENT_SECRET?: string;
    AUTHENTIK_CLIENT_DISCOVERY_URL?: string;
    AUTHENTIK_REDIRECT_URI?: string;
  }
}
