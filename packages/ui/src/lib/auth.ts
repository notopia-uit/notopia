import { betterAuth } from 'better-auth';
import { genericOAuth } from 'better-auth/plugins';

export const auth = betterAuth({
  baseURL: process.env.BETTER_AUTH_URL,
  secret: process.env.BETTER_AUTH_SECRET,
  appName: 'Notopia',
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
});
