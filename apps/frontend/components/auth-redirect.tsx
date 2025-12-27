'use client'

import { useEffect, useState, ReactNode } from 'react'
import { useRouter } from 'next/navigation'

interface AuthRedirectProps {
  children?: ReactNode
}

/**
 * Client component that redirects authenticated users to /app/projects
 * Use this on public pages (landing, login, signup) to auto-redirect logged-in users
 * 
 * Shows a loading state while checking to prevent flash of content
 */
export function AuthRedirect({ children }: AuthRedirectProps) {
  const router = useRouter()
  const [status, setStatus] = useState<'checking' | 'guest' | 'redirecting'>('checking')

  useEffect(() => {
    const accessToken = localStorage.getItem('access_token')
    if (accessToken) {
      setStatus('redirecting')
      router.replace('/app/projects')
    } else {
      setStatus('guest')
    }
  }, [router])

  // While checking or redirecting, show nothing (or minimal loading)
  if (status === 'checking' || status === 'redirecting') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    )
  }

  // User is a guest, show the page content
  return <>{children}</>
}

/**
 * Hook version for use in client components that already have useEffect
 */
export function useAuthRedirect() {
  const router = useRouter()

  useEffect(() => {
    const accessToken = localStorage.getItem('access_token')
    if (accessToken) {
      router.replace('/app/projects')
    }
  }, [router])
}

