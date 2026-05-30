import { auth } from '@lib/auth';
import { toNextJsHandler } from 'better-auth/next-js';

// eslint-disable-next-line @typescript-eslint/no-unsafe-argument
export const { GET, POST } = toNextJsHandler(auth);
