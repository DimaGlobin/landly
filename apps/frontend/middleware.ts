import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Next.js Middleware for API proxying
 * 
 * In development/docker-compose: BACKEND_URL is set, proxy /v1/* to backend
 * In production (Yandex Cloud): BACKEND_URL is not set, API Gateway handles routing
 */
export function middleware(request: NextRequest) {
  const backendUrl = process.env.BACKEND_URL

  // If BACKEND_URL is not set, let the request pass through
  // (In production, API Gateway handles /v1/* routing)
  if (!backendUrl) {
    return NextResponse.next()
  }

  const { pathname } = request.nextUrl

  // Proxy API requests to backend
  if (pathname.startsWith('/v1/')) {
    const url = new URL(pathname + request.nextUrl.search, backendUrl)
    
    // Clone headers and forward them
    const headers = new Headers(request.headers)
    
    return NextResponse.rewrite(url, {
      headers,
    })
  }

  // Proxy health/metrics endpoints
  if (pathname === '/health' || pathname === '/metrics') {
    const url = new URL(pathname, backendUrl)
    return NextResponse.rewrite(url)
  }

  return NextResponse.next()
}

export const config = {
  // Match API routes and health endpoints
  matcher: ['/v1/:path*', '/health', '/metrics'],
}


