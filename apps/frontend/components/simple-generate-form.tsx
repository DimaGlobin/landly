'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface SimpleGenerateFormProps {
  projectId: string
  onSuccess: (schema: any) => void
}

interface FormData {
  prompt: string
  paymentURL: string
}

export function SimpleGenerateForm({ projectId, onSuccess }: SimpleGenerateFormProps) {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const onSubmit = async (data: FormData) => {
    setIsLoading(true)
    setError(null)

    try {
      const token = localStorage.getItem('access_token')
      if (!token) {
        throw new Error('Токен авторизации не найден')
      }

      const response = await fetch(`http://localhost:8080/v1/projects/${projectId}/generate-simple`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          prompt: data.prompt,
          payment_url: data.paymentURL || undefined,
        }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Ошибка генерации')
      }

      if (result.success) {
        console.log('✅ Генерация успешна:', result)
        onSuccess(result.schema)
      } else {
        throw new Error(result.error || 'Неизвестная ошибка')
      }
    } catch (err) {
      console.error('❌ Ошибка генерации:', err)
      setError(err instanceof Error ? err.message : 'Неизвестная ошибка')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>🚀 Простая генерация лендинга</CardTitle>
        <CardDescription>
          Новый простой API - результат сразу!
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label htmlFor="prompt" className="block text-sm font-medium mb-2">
              Описание проекта *
            </label>
            <textarea
              {...register('prompt', { required: 'Описание обязательно' })}
              id="prompt"
              rows={4}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Опишите ваш проект, что вы предлагаете, кто ваша целевая аудитория..."
            />
            {errors.prompt && (
              <p className="text-red-500 text-sm mt-1">{errors.prompt.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="paymentURL" className="block text-sm font-medium mb-2">
              Ссылка на оплату (необязательно)
            </label>
            <Input
              {...register('paymentURL')}
              id="paymentURL"
              type="url"
              placeholder="https://example.com/pay"
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <p className="text-red-600 text-sm">❌ {error}</p>
            </div>
          )}

          <Button
            type="submit"
            disabled={isLoading}
            className="w-full"
          >
            {isLoading ? '🔄 Генерируем...' : '🚀 Сгенерировать лендинг'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
