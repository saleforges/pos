import createMiddleware from 'next-intl/middleware';
import { routing } from './i18n/routing';

export default createMiddleware(routing);

export const config = {
  // Match all paths except api, static files, images, etc.
  matcher: ['/', '/(id|en)/:path*', '/((?!api|_next|_vercel|.*\\..*).*)'],
};
