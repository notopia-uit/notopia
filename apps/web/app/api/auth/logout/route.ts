import { NextResponse } from 'next/server';

export function GET() {
  const endSessionUrl = process.env.AUTHENTIK_CLIENT_DISCOVERY_URL?.replace(
    '/.well-known/openid-configuration',
    '/end-session/'
  );

  const redirectUri = encodeURIComponent(`${process.env.BETTER_AUTH_URL}/signin`);

  return NextResponse.redirect(`${endSessionUrl}?post_logout_redirect_uri=${redirectUri}`);
}
