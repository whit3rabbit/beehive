import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// Paths that don't require authentication
const publicPaths = ['/login']

export function middleware(request: NextRequest) {
  const token = request.cookies.get('auth_token')
  const isPublicPath = publicPaths.some(path => request.nextUrl.pathname.startsWith(path))

  if (!token && !isPublicPath) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('from', request.nextUrl.pathname)
    return NextResponse.redirect(loginUrl)
  }

  if (token && isPublicPath) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}



// Configure paths that trigger the middleware
export const config = {
    matcher: [
      /*
       * Match all request paths except for the ones starting with:
       * - api (API routes)
       * - _next/static (static files)
       * - _next/image (image optimization files)
       * - favicon.ico (favicon file)
       * - public folder
       */
      '/((?!api|_next/static|_next/image|favicon.ico|public/).*)',
    ],
};