'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { api } from '@/lib/api'
import { Project } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { SimpleGenerateForm } from '@/components/simple-generate-form'
import { SimpleLandingPreview } from '@/components/simple-landing-preview'

export default function SimpleProjectPage() {
  const router = useRouter()
  const params = useParams()
  const projectId = params.id as string

  const [project, setProject] = useState<Project | null>(null)
  const [generatedSchema, setGeneratedSchema] = useState<any>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadProject()
  }, [projectId])

  const loadProject = async () => {
    try {
      const projectData = await api.getProject(projectId)
      setProject(projectData)
    } catch (error) {
      console.error('Failed to load project', error)
      setError('Не удалось загрузить проект')
    } finally {
      setIsLoading(false)
    }
  }

  const handleGenerateSuccess = (schema: any) => {
    console.log('🎉 Генерация успешна, схема:', schema)
    setGeneratedSchema(schema)
  }

  const handleBack = () => {
    router.push('/app/projects')
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-4 text-gray-600">Загружаем проект...</p>
        </div>
      </div>
    )
  }

  if (error || !project) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <p className="text-red-600 mb-4">❌ {error || 'Проект не найден'}</p>
              <Button onClick={handleBack}>← Назад к проектам</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Button onClick={handleBack} variant="outline" className="mb-4">
          ← Назад к проектам
        </Button>
        <h1 className="text-3xl font-bold">{project.name}</h1>
        <p className="text-gray-600 mt-2">Простой генератор лендингов</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Форма генерации */}
        <div>
          <SimpleGenerateForm 
            projectId={projectId} 
            onSuccess={handleGenerateSuccess}
          />
        </div>

        {/* Предпросмотр результата */}
        <div>
          {generatedSchema ? (
            <SimpleLandingPreview schema={generatedSchema} />
          ) : (
            <Card>
              <CardContent className="pt-6">
                <div className="text-center text-gray-500">
                  <div className="text-6xl mb-4">🎨</div>
                  <p>Сгенерируйте лендинг, чтобы увидеть предпросмотр</p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      {/* Информация о проекте */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle>📋 Информация о проекте</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="font-semibold">Название:</p>
              <p className="text-gray-600">{project.name}</p>
            </div>
            <div>
              <p className="font-semibold">Ниша:</p>
              <p className="text-gray-600">{project.niche}</p>
            </div>
            <div>
              <p className="font-semibold">Статус:</p>
              <p className="text-gray-600">{project.status}</p>
            </div>
            <div>
              <p className="font-semibold">Создан:</p>
              <p className="text-gray-600">{new Date(project.created_at).toLocaleDateString()}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
