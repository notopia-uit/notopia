import { client } from '@notopia-uit/api-gen';
import { fetchAccessTokenServerSide } from '@notopia-uit/ui/lib/get-access-token';
import { betterAuth } from 'better-auth';
import { createAuthMiddleware } from 'better-auth/api';
import { genericOAuth } from 'better-auth/plugins';

export const auth = betterAuth({
  baseURL: process.env.BETTER_AUTH_URL,
  secret: process.env.BETTER_AUTH_SECRET,
  appName: 'Notopia',
  logger: {
    level: 'debug',
  },
  plugins: [
    genericOAuth({
      config: [
        {
          providerId: 'authentik',
          clientId: process.env.AUTHENTIK_CLIENT_ID as string,
          clientSecret: process.env.AUTHENTIK_CLIENT_SECRET as string,
          discoveryUrl: process.env.AUTHENTIK_CLIENT_DISCOVERY_URL as string,
          redirectURI: process.env.AUTHENTIK_REDIRECT_URI as string,
        },
      ],
    }),
  ],
  hooks: {
    after: createAuthMiddleware(async (ctx) => {
      const newSession = ctx.context.session ?? ctx.context.newSession;
      if (newSession) {
        client.setConfig({
          auth: fetchAccessTokenServerSide,
        });
      }
      return Promise.resolve();
    }),
  },
});
