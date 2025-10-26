'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface FormData {
  email: string
  password: string
  confirmPassword: string
}

export default function SignUpPage() {
  const router = useRouter()
  const [error, setError] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)
  const { register, handleSubmit, watch, formState: { errors } } = useForm<FormData>()

  const password = watch('password')

  const onSubmit = async (data: FormData) => {
    try {
      setError('')
      setIsLoading(true)
      await api.signUp(data.email, data.password)
      router.push('/app/projects')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка регистрации')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Регистрация в Landly</CardTitle>
          <CardDescription>Создайте аккаунт, чтобы начать создавать лендинги</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <div className="bg-destructive/10 text-destructive px-4 py-2 rounded">
                {error}
              </div>
            )}

            <div>
              <label htmlFor="email" className="block text-sm font-medium mb-2">
                Email
              </label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                {...register('email', { required: 'Email обязателен' })}
              />
              {errors.email && (
                <p className="text-sm text-destructive mt-1">{errors.email.message}</p>
              )}
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium mb-2">
                Пароль
              </label>
              <Input
                id="password"
                type="password"
                {...register('password', {
                  required: 'Пароль обязателен',
                  minLength: { value: 8, message: 'Минимум 8 символов' },
                })}
              />
              {errors.password && (
                <p className="text-sm text-destructive mt-1">{errors.password.message}</p>
              )}
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium mb-2">
                Подтвердите пароль
              </label>
              <Input
                id="confirmPassword"
                type="password"
                {...register('confirmPassword', {
                  required: 'Подтвердите пароль',
                  validate: (value) => value === password || 'Пароли не совпадают',
                })}
              />
              {errors.confirmPassword && (
                <p className="text-sm text-destructive mt-1">{errors.confirmPassword.message}</p>
              )}
            </div>

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Регистрация...' : 'Зарегистрироваться'}
            </Button>

            <div className="text-center text-sm">
              <span className="text-gray-600">Уже есть аккаунт? </span>
              <a href="/auth/login" className="text-primary hover:underline">
                Войти
              </a>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

