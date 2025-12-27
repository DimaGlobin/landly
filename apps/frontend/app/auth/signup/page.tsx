"use client"

import Link from 'next/link'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { ArrowRight, Sparkles } from 'lucide-react'

import { api, ApiError } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ErrorMessage } from '@/components/error-message'

interface FormData {
  email: string
  password: string
  confirmPassword: string
}

export default function SignUpPage() {
  const router = useRouter()
  const [error, setError] = useState<ApiError | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isCheckingAuth, setIsCheckingAuth] = useState(true)
  const { register, handleSubmit, watch, formState: { errors } } = useForm<FormData>()

  const password = watch('password')

  // Redirect to projects if already logged in (but verify token first)
  useEffect(() => {
    let mounted = true
    
    const checkAuth = async () => {
      const accessToken = localStorage.getItem('access_token')
      if (!accessToken) {
        if (mounted) setIsCheckingAuth(false)
        return
      }

      // Verify token is valid by making a test request
      try {
        await api.getProjects()
        // Token is valid, redirect
        if (mounted) {
          router.replace('/app/projects')
        }
      } catch (error: any) {
        // Token is invalid, clear it and stay on signup page
        if (mounted) {
          if (error instanceof ApiError && error.shouldRedirectToLogin()) {
            api.logout()
          }
          setIsCheckingAuth(false)
        }
      }
    }
    
    checkAuth()
    
    return () => {
      mounted = false
    }
  }, [router])

  // Show loading while checking auth
  if (isCheckingAuth) {
    return (
      <div className="app-shell min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    )
  }

  const onSubmit = async (data: FormData) => {
    try {
      setError(null)
      setIsLoading(true)
      await api.signUp(data.email, data.password)
      router.push('/app/projects')
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err)
      } else {
        setError(new ApiError('UNKNOWN', 'Ошибка регистрации'))
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="app-shell">
      <div className="relative z-10 mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-6 px-4 pb-12 pt-6 sm:gap-10 sm:px-6 sm:pb-16 sm:pt-10 md:flex-row md:items-center md:justify-between md:px-10">
        <div className="hidden w-full max-w-md rounded-3xl border border-white/50 bg-slate-900/90 p-8 text-slate-200 shadow-2xl backdrop-blur-2xl md:block md:p-10">
          <div className="inline-flex items-center gap-2 rounded-full border border-white/30 bg-white/10 px-4 py-1.5 text-xs font-semibold uppercase tracking-wide text-slate-200">
            <Sparkles className="h-4 w-4" /> AI workspace
          </div>
          <h2 className="mt-6 text-3xl font-bold leading-snug text-white">
            Запустите рабочее пространство, похожее на Cursor и ChatGPT, но для лендингов.
          </h2>
          <ul className="mt-8 space-y-4 text-sm text-slate-300">
            <li>• Генерируйте, редактируйте и сохраняйте историю промптов.</li>
            <li>• Предпросмотр и продакшн синхронизированы, поэтому результат никогда не расходится.</li>
            <li>• Каждый проект можно опубликовать в один клик — сразу в CDN.</li>
          </ul>
          <div className="mt-10 inline-flex items-center gap-2 text-sm font-semibold text-white">
            Уже есть аккаунт?{' '}
            <Link href="/auth/login" className="inline-flex items-center gap-1 underline-offset-4 hover:underline">
              Войти <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>

        <Card className="surface-card w-full border-white/40 p-4 sm:p-6 md:max-w-md">
          <CardHeader className="space-y-1 px-0 pt-0">
            <CardTitle className="text-xl font-bold text-slate-900 sm:text-2xl">Регистрация в Landly</CardTitle>
            <CardDescription className="text-xs text-slate-600 sm:text-sm">
              Создайте аккаунт и начните генерировать лендинги в диалоге с AI. Первые проекты — бесплатно.
            </CardDescription>
          </CardHeader>
          <CardContent className="px-0 pb-0">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 sm:space-y-5">
              <ErrorMessage error={error} onDismiss={() => setError(null)} />

              <div className="space-y-2">
                <label htmlFor="email" className="text-xs font-medium text-slate-700 sm:text-sm">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  className="h-10 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50 sm:h-11"
                  {...register('email', { required: 'Email обязателен' })}
                />
                {errors.email && <p className="text-xs text-red-500">{errors.email.message}</p>}
              </div>

              <div className="space-y-2">
                <label htmlFor="password" className="text-xs font-medium text-slate-700 sm:text-sm">
                  Пароль
                </label>
                <Input
                  id="password"
                  type="password"
                  className="h-10 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50 sm:h-11"
                  {...register('password', {
                    required: 'Пароль обязателен',
                    minLength: { value: 8, message: 'Минимум 8 символов' },
                  })}
                />
                {errors.password && <p className="text-xs text-red-500">{errors.password.message}</p>}
              </div>

              <div className="space-y-2">
                <label htmlFor="confirmPassword" className="text-xs font-medium text-slate-700 sm:text-sm">
                  Подтвердите пароль
                </label>
                <Input
                  id="confirmPassword"
                  type="password"
                  className="h-10 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50 sm:h-11"
                  {...register('confirmPassword', {
                    required: 'Подтвердите пароль',
                    validate: (value) => value === password || 'Пароли не совпадают',
                  })}
                />
                {errors.confirmPassword && <p className="text-xs text-red-500">{errors.confirmPassword.message}</p>}
              </div>

              <Button
                type="submit"
                className="h-10 w-full rounded-full bg-blue-600 text-xs font-semibold text-white shadow-lg transition hover:bg-blue-700 sm:h-11 sm:text-sm"
                disabled={isLoading}
              >
                {isLoading ? 'Регистрация...' : 'Зарегистрироваться'}
              </Button>

              <div className="text-center text-xs text-slate-600 sm:text-sm">
                Уже есть аккаунт?{' '}
                <Link href="/auth/login" className="font-semibold text-blue-600 hover:underline">
                  Войдите в Landly
                </Link>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
