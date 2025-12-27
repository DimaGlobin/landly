'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowLeft, Save, CreditCard, Check } from 'lucide-react'
import { api } from '@/lib/api'
import { UserProfile, Subscription } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

// Моки для подписок
const mockSubscription: Subscription = {
  plan: 'free',
  status: 'active',
  features: ['До 5 проектов', 'Базовые шаблоны', 'Публикация на поддомене'],
}

const proSubscription: Subscription = {
  plan: 'pro',
  status: 'active',
  expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
  features: ['Неограниченное количество проектов', 'Все шаблоны', 'Кастомные домены', 'Приоритетная поддержка'],
}

export default function ProfilePage() {
  const router = useRouter()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [subscription, setSubscription] = useState<Subscription>(mockSubscription)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [passwordConfirm, setPasswordConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  useEffect(() => {
    loadProfile()
  }, [])

  const loadProfile = async () => {
    try {
      const data = await api.getUserProfile()
      setProfile(data)
      setEmail(data.email)
      setIsLoading(false)
    } catch (error: any) {
      console.error('Failed to load profile', error)
      if (error instanceof Error && 'shouldRedirectToLogin' in error && (error as any).shouldRedirectToLogin()) {
        router.push('/auth/login')
      } else {
        setError('Не удалось загрузить профиль')
        setIsLoading(false)
      }
    }
  }

  const handleSave = async () => {
    if (password && password !== passwordConfirm) {
      setError('Пароли не совпадают')
      return
    }

    if (password && password.length < 8) {
      setError('Пароль должен быть не менее 8 символов')
      return
    }

    setIsSaving(true)
    setError(null)
    setSuccess(null)

    try {
      const updated = await api.updateUserProfile(
        email !== profile?.email ? email : undefined,
        password || undefined
      )
      setProfile(updated)
      setPassword('')
      setPasswordConfirm('')
      setSuccess('Профиль успешно обновлен')
    } catch (error: any) {
      console.error('Failed to update profile', error)
      setError(error.message || 'Не удалось обновить профиль')
    } finally {
      setIsSaving(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
    })
  }

  if (isLoading) {
    return (
      <div className="app-shell min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="app-shell min-h-screen">
      <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="mb-6">
          <Button
            variant="ghost"
            onClick={() => router.back()}
            className="mb-4 text-sm text-slate-600 hover:text-slate-900"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Назад
          </Button>
          <h1 className="text-2xl font-bold text-slate-900 sm:text-3xl">Личный кабинет</h1>
          <p className="mt-2 text-sm text-slate-600">Управляйте своим профилем и подпиской</p>
        </div>

        <div className="grid gap-6 md:grid-cols-2">
          {/* Профиль */}
          <Card>
            <CardHeader>
              <CardTitle>Профиль</CardTitle>
              <CardDescription>Основная информация о вашем аккаунте</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {error && (
                <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">
                  {error}
                </div>
              )}
              {success && (
                <div className="rounded-lg bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">
                  {success}
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Email</label>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Новый пароль</label>
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Оставьте пустым, чтобы не менять"
                  className="w-full"
                />
              </div>
              {password && (
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">Подтвердите пароль</label>
                  <Input
                    type="password"
                    value={passwordConfirm}
                    onChange={(e) => setPasswordConfirm(e.target.value)}
                    className="w-full"
                  />
                </div>
              )}
              {profile && (
                <div className="pt-2 border-t border-slate-200">
                  <p className="text-xs text-slate-500">
                    Аккаунт создан: {formatDate(profile.created_at)}
                  </p>
                </div>
              )}
              <Button onClick={handleSave} disabled={isSaving} className="w-full">
                <Save className="mr-2 h-4 w-4" />
                {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
              </Button>
            </CardContent>
          </Card>

          {/* Подписка */}
          <Card>
            <CardHeader>
              <CardTitle>Подписка</CardTitle>
              <CardDescription>Текущий план и возможности</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-slate-700">Текущий план</span>
                  <span
                    className={`rounded-full px-3 py-1 text-xs font-semibold ${
                      subscription.plan === 'pro'
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-slate-100 text-slate-700'
                    }`}
                  >
                    {subscription.plan === 'pro' ? 'PRO' : 'FREE'}
                  </span>
                </div>
                {subscription.expires_at && (
                  <p className="text-xs text-slate-500">
                    Действует до: {formatDate(subscription.expires_at)}
                  </p>
                )}
              </div>
              <div className="pt-2 border-t border-slate-200">
                <p className="text-sm font-medium text-slate-700 mb-3">Включено в план:</p>
                <ul className="space-y-2">
                  {subscription.features.map((feature, index) => (
                    <li key={index} className="flex items-start text-sm text-slate-600">
                      <Check className="mr-2 h-4 w-4 text-green-600 flex-shrink-0 mt-0.5" />
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>
              {subscription.plan === 'free' && (
                <Button className="w-full" onClick={() => setSubscription(proSubscription)}>
                  <CreditCard className="mr-2 h-4 w-4" />
                  Перейти на PRO
                </Button>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

