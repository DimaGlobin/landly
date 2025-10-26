'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface AIBriefFormProps {
  onGenerate: (prompt: string, paymentURL?: string) => Promise<void>
  isLoading?: boolean
}

interface FormData {
  niche: string
  offer: string
  audience: string
  style: string
  cta: string
  benefits: string
  paymentURL: string
}

export function AIBriefForm({ onGenerate, isLoading }: AIBriefFormProps) {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>()

  const onSubmit = async (data: FormData) => {
    // Формируем промпт из полей формы
    const prompt = `
Создай лендинг для ${data.niche}.
Оффер: ${data.offer}
Целевая аудитория: ${data.audience}
Стиль: ${data.style}
Призыв к действию: ${data.cta}
Ключевые преимущества: ${data.benefits}
    `.trim()

    await onGenerate(prompt, data.paymentURL || undefined)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>AI Brief для генерации лендинга</CardTitle>
        <CardDescription>
          Опишите ваш проект, и AI создаст профессиональный лендинг
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label htmlFor="niche" className="block text-sm font-medium mb-2">
              Ниша / Тематика *
            </label>
            <Input
              id="niche"
              placeholder="Например: онлайн-курс по программированию"
              {...register('niche', { required: 'Это поле обязательно' })}
            />
            {errors.niche && (
              <p className="text-sm text-destructive mt-1">{errors.niche.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="offer" className="block text-sm font-medium mb-2">
              Оффер (что вы предлагаете) *
            </label>
            <Input
              id="offer"
              placeholder="Например: научим программировать за 3 месяца"
              {...register('offer', { required: 'Это поле обязательно' })}
            />
            {errors.offer && (
              <p className="text-sm text-destructive mt-1">{errors.offer.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="audience" className="block text-sm font-medium mb-2">
              Целевая аудитория *
            </label>
            <Input
              id="audience"
              placeholder="Например: новички без опыта в IT"
              {...register('audience', { required: 'Это поле обязательно' })}
            />
            {errors.audience && (
              <p className="text-sm text-destructive mt-1">{errors.audience.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="style" className="block text-sm font-medium mb-2">
              Стиль / Тон *
            </label>
            <Input
              id="style"
              placeholder="Например: профессиональный, дружелюбный"
              {...register('style', { required: 'Это поле обязательно' })}
            />
            {errors.style && (
              <p className="text-sm text-destructive mt-1">{errors.style.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="cta" className="block text-sm font-medium mb-2">
              Призыв к действию (CTA) *
            </label>
            <Input
              id="cta"
              placeholder="Например: Записаться на курс"
              {...register('cta', { required: 'Это поле обязательно' })}
            />
            {errors.cta && (
              <p className="text-sm text-destructive mt-1">{errors.cta.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="benefits" className="block text-sm font-medium mb-2">
              Ключевые преимущества *
            </label>
            <Input
              id="benefits"
              placeholder="Например: опытные преподаватели, практика, трудоустройство"
              {...register('benefits', { required: 'Это поле обязательно' })}
            />
            {errors.benefits && (
              <p className="text-sm text-destructive mt-1">{errors.benefits.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="paymentURL" className="block text-sm font-medium mb-2">
              Ссылка на оплату (необязательно)
            </label>
            <Input
              id="paymentURL"
              type="url"
              placeholder="https://pay.prodamus.ru/..."
              {...register('paymentURL')}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Укажите URL вашей платёжной формы (Prodamus, ЮКassa и т.д.)
            </p>
          </div>

          <Button type="submit" size="lg" className="w-full" disabled={isLoading}>
            {isLoading ? 'Генерация...' : 'Сгенерировать лендинг'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}

