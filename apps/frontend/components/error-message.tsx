'use client'

import { useTranslations } from 'next-intl'
import { ApiError, ApiErrorCode } from '@/lib/api'
import { AlertCircle, XCircle, AlertTriangle, Info } from 'lucide-react'

interface ErrorMessageProps {
  error: ApiError | Error | string | null
  className?: string
  variant?: 'inline' | 'banner' | 'toast'
  onDismiss?: () => void
}

// Map error codes to severity for styling
function getErrorSeverity(code: ApiErrorCode): 'error' | 'warning' | 'info' {
  switch (code) {
    case 'INTERNAL_ERROR':
    case 'GENERATION_FAILED':
    case 'PUBLISH_FAILED':
    case 'RENDER_FAILED':
      return 'error'
    case 'NETWORK_ERROR':
    case 'TOKEN_EXPIRED':
    case 'INVALID_TOKEN':
      return 'warning'
    default:
      return 'info'
  }
}

export function ErrorMessage({ error, className = '', variant = 'inline', onDismiss }: ErrorMessageProps) {
  const t = useTranslations('errors')

  if (!error) return null

  // Get error code and message
  let code: ApiErrorCode = 'UNKNOWN'
  let fallbackMessage = ''
  let isApiError = false

  if (error instanceof ApiError) {
    code = error.code
    fallbackMessage = error.originalMessage
    isApiError = true
  } else if (error instanceof Error) {
    fallbackMessage = error.message
  } else if (typeof error === 'string') {
    fallbackMessage = error
  }

  // Try to get translated message, fallback to original
  let message: string
  if (!isApiError && fallbackMessage) {
    // For non-API errors, use fallback message directly
    message = fallbackMessage
  } else {
    try {
      message = t(code)
      // If translation returns the key itself, it's not translated
      if (message === code) {
        message = fallbackMessage || t('UNKNOWN')
      }
    } catch {
      message = fallbackMessage || 'An error occurred'
    }
  }

  const severity = getErrorSeverity(code)

  // Styling based on severity and variant
  const severityStyles = {
    error: 'bg-red-50 border-red-200 text-red-800',
    warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
    info: 'bg-blue-50 border-blue-200 text-blue-800',
  }

  const icons = {
    error: <XCircle className="h-5 w-5 text-red-500" />,
    warning: <AlertTriangle className="h-5 w-5 text-yellow-500" />,
    info: <AlertCircle className="h-5 w-5 text-blue-500" />,
  }

  if (variant === 'inline') {
    return (
      <div className={`flex items-center gap-2 text-sm ${severityStyles[severity]} px-3 py-2 rounded-md border ${className}`}>
        {icons[severity]}
        <span>{message}</span>
      </div>
    )
  }

  if (variant === 'banner') {
    return (
      <div className={`flex items-center justify-between gap-4 ${severityStyles[severity]} px-4 py-3 rounded-lg border ${className}`}>
        <div className="flex items-center gap-3">
          {icons[severity]}
          <span className="font-medium">{message}</span>
        </div>
        {onDismiss && (
          <button
            onClick={onDismiss}
            className="text-current opacity-60 hover:opacity-100 transition-opacity"
            aria-label="Dismiss"
          >
            <XCircle className="h-5 w-5" />
          </button>
        )}
      </div>
    )
  }

  // Toast variant
  return (
    <div className={`flex items-center gap-3 ${severityStyles[severity]} px-4 py-3 rounded-lg border shadow-lg ${className}`}>
      {icons[severity]}
      <span className="flex-1">{message}</span>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="text-current opacity-60 hover:opacity-100 transition-opacity"
          aria-label="Dismiss"
        >
          <XCircle className="h-5 w-5" />
        </button>
      )}
    </div>
  )
}

// Hook for using error translations in components
export function useErrorMessage() {
  const t = useTranslations('errors')

  return {
    getErrorMessage: (error: ApiError | Error | string | null): string => {
      if (!error) return ''

      if (error instanceof ApiError) {
        try {
          const translated = t(error.code)
          return translated === error.code ? error.originalMessage : translated
        } catch {
          return error.originalMessage
        }
      }

      if (error instanceof Error) {
        return error.message
      }

      return error
    },
  }
}

