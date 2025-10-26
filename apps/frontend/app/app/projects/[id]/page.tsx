'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { api } from '@/lib/api'
import { Project, LandingSchema, AnalyticsStats } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { AIBriefForm } from '@/components/ai-brief-form'
import { LandingPreview } from '@/components/landing-preview'

export default function ProjectPage() {
  const router = useRouter()
  const params = useParams()
  const projectId = params.id as string

  const [project, setProject] = useState<Project | null>(null)
  const [schema, setSchema] = useState<LandingSchema | null>(null)
  const [stats, setStats] = useState<AnalyticsStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isGenerating, setIsGenerating] = useState(false)
  const [isPublishing, setIsPublishing] = useState(false)
  const [publishResult, setPublishResult] = useState<any>(null)
  const [activeTab, setActiveTab] = useState<'generate' | 'preview' | 'analytics'>('generate')

  useEffect(() => {
    loadProject()
  }, [projectId])

  const loadProject = async () => {
    try {
      const projectData = await api.getProject(projectId)
      setProject(projectData)

      // Всегда пытаемся загрузить preview, если есть схема
      try {
        const previewData = await api.getPreview(projectId)
        if (previewData && previewData.schema) {
          setSchema(previewData.schema)
          // Переключаемся на preview только если есть данные
          if (projectData.status !== 'draft') {
            setActiveTab('preview')
          }
        }
      } catch (e) {
        // Схема еще не сгенерирована - это нормально
        console.log('No preview data yet')
      }

      // Загружаем статистику (если доступна)
      if (projectData.status !== 'draft') {
        try {
          const statsData = await api.getStats(projectId)
          setStats(statsData)
        } catch (e) {
          // Статистика может быть недоступна
        }
      }
    } catch (error) {
      console.error('Failed to load project', error)
      router.push('/app/projects')
    } finally {
      setIsLoading(false)
    }
  }

  const handleGenerate = async (prompt: string, paymentURL?: string) => {
    try {
      setIsGenerating(true)
      
      // Используем новый простой API
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
          prompt: prompt,
          payment_url: paymentURL || undefined,
        }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Ошибка генерации')
      }

      if (result.success) {
        console.log('✅ Генерация успешна:', result)
        // Обновляем проект
        await loadProject()
        
        // Принудительно загружаем превью
        try {
          const previewData = await api.getPreview(projectId)
          setSchema(previewData.schema)
          console.log('✅ Превью загружено:', previewData.schema)
        } catch (previewError) {
          console.error('❌ Ошибка загрузки превью:', previewError)
        }
        
        // Переключаемся на превью
        setActiveTab('preview')
      } else {
        throw new Error(result.error || 'Неизвестная ошибка')
      }
    } catch (error: any) {
      console.error('❌ Ошибка генерации:', error)
      alert('Ошибка генерации: ' + error.message)
    } finally {
      setIsGenerating(false)
    }
  }

  const handlePublish = async () => {
    if (!confirm('Опубликовать лендинг?')) return

    try {
      setIsPublishing(true)
      const result = await api.publishProject(projectId)
      setPublishResult(result)
      await loadProject()
    } catch (error: any) {
      alert('Ошибка публикации: ' + (error.response?.data?.error || error.message))
    } finally {
      setIsPublishing(false)
    }
  }

  if (isLoading) {
    return <div className="min-h-screen flex items-center justify-center">Загрузка...</div>
  }

  if (!project) {
    return <div className="min-h-screen flex items-center justify-center">Проект не найден</div>
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Button variant="ghost" onClick={() => router.push('/app/projects')}>
                ← Назад
              </Button>
              <div>
                <h1 className="text-2xl font-bold">{project.name}</h1>
                <p className="text-sm text-gray-600">{project.niche}</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span
                className={`px-3 py-1 rounded-full text-sm ${
                  project.status === 'published'
                    ? 'bg-green-100 text-green-800'
                    : project.status === 'generated'
                    ? 'bg-blue-100 text-blue-800'
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                {project.status}
              </span>
              {schema && (
                <Button onClick={handlePublish} disabled={isPublishing}>
                  {isPublishing ? 'Публикация...' : 'Опубликовать'}
                </Button>
              )}
            </div>
          </div>

          {/* Tabs */}
          <div className="flex gap-4 mt-4 border-b">
            <button
              className={`pb-2 px-1 ${
                activeTab === 'generate'
                  ? 'border-b-2 border-primary font-semibold'
                  : 'text-gray-600'
              }`}
              onClick={() => setActiveTab('generate')}
            >
              Генерация
            </button>
            {schema && (
              <>
                <button
                  className={`pb-2 px-1 ${
                    activeTab === 'preview'
                      ? 'border-b-2 border-primary font-semibold'
                      : 'text-gray-600'
                  }`}
                  onClick={() => setActiveTab('preview')}
                >
                  Предпросмотр
                </button>
                <button
                  className={`pb-2 px-1 ${
                    activeTab === 'analytics'
                      ? 'border-b-2 border-primary font-semibold'
                      : 'text-gray-600'
                  }`}
                  onClick={() => setActiveTab('analytics')}
                >
                  Аналитика
                </button>
              </>
            )}
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        {publishResult && (
          <Card className="mb-6 bg-green-50 border-green-200">
            <CardContent className="py-4">
              <p className="font-semibold mb-2">✅ Лендинг опубликован!</p>
              <p className="text-sm">
                URL:{' '}
                <a
                  href={publishResult.public_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {publishResult.public_url}
                </a>
              </p>
            </CardContent>
          </Card>
        )}

        {activeTab === 'generate' && (
          <div className="max-w-2xl">
            <AIBriefForm onGenerate={handleGenerate} isLoading={isGenerating} />
          </div>
        )}

        {activeTab === 'preview' && schema && (
          <div>
            <LandingPreview schema={schema} />
          </div>
        )}

        {activeTab === 'analytics' && (
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <CardHeader>
                <CardDescription>Просмотры</CardDescription>
                <CardTitle className="text-3xl">{stats?.total_pageviews || 0}</CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader>
                <CardDescription>Клики CTA</CardDescription>
                <CardTitle className="text-3xl">{stats?.total_cta_clicks || 0}</CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader>
                <CardDescription>Клики оплаты</CardDescription>
                <CardTitle className="text-3xl">{stats?.total_pay_clicks || 0}</CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader>
                <CardDescription>Уникальные посетители</CardDescription>
                <CardTitle className="text-3xl">{stats?.unique_visitors || 0}</CardTitle>
              </CardHeader>
            </Card>
          </div>
        )}
      </main>
    </div>
  )
}

