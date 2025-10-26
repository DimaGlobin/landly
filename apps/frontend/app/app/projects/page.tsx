'use client'

import { useEffect, useState, MouseEvent } from 'react'
import { useRouter } from 'next/navigation'
import { Trash2, ExternalLink, Copy } from 'lucide-react'
import { api } from '@/lib/api'
import { Project } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

const projectStatusLabels: Record<Project['status'], string> = {
  draft: 'Черновик',
  generated: 'Сгенерирован',
  published: 'Опубликован',
}

const projectStatusClasses: Record<Project['status'], string> = {
  draft: 'bg-gray-100 text-gray-800',
  generated: 'bg-blue-100 text-blue-800',
  published: 'bg-green-100 text-green-800',
}

const publishStatusLabels: Record<string, string> = {
  draft: 'Черновик публикации',
  published: 'Опубликовано',
  failed: 'Ошибка публикации',
}

export default function ProjectsPage() {
  const router = useRouter()
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newProject, setNewProject] = useState({ name: '', niche: '' })
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const formatDate = (value: string) => new Date(value).toLocaleDateString('ru-RU')

  const formatDateTime = (value: string) =>
    new Date(value).toLocaleString('ru-RU', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })

  useEffect(() => {
    loadProjects()
  }, [])

  const loadProjects = async () => {
    try {
      const data = await api.getProjects()
      setProjects(data.projects || [])
    } catch (error) {
      console.error('Failed to load projects', error)
      router.push('/auth/login')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateProject = async () => {
    try {
      const project = await api.createProject(newProject.name, newProject.niche)
      router.push(`/app/projects/${project.id}`)
    } catch (error) {
      console.error('Failed to create project', error)
    }
  }

  const handleDeleteProject = async (event: MouseEvent<HTMLButtonElement>, projectId: string) => {
    event.stopPropagation()
    if (!window.confirm('Удалить проект?')) {
      return
    }

    try {
      setDeletingId(projectId)
      await api.deleteProject(projectId)
      setProjects((prev) => prev.filter((project) => project.id !== projectId))
    } catch (error) {
      console.error('Failed to delete project', error)
    } finally {
      setDeletingId(null)
    }
  }

  const handleCopyLink = async (event: MouseEvent<HTMLButtonElement>, url: string) => {
    event.stopPropagation()

    try {
      await navigator.clipboard.writeText(url)
      window.alert('Ссылка скопирована в буфер обмена')
    } catch (error) {
      console.error('Failed to copy link', error)
    }
  }

  const handleOpenLink = (event: MouseEvent<HTMLButtonElement>, url: string) => {
    event.stopPropagation()
    window.open(url, '_blank', 'noopener,noreferrer')
  }

  const handleLogout = () => {
    api.logout()
    router.push('/')
  }

  if (isLoading) {
    return <div className="min-h-screen flex items-center justify-center">Загрузка...</div>
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Landly</h1>
          <Button variant="ghost" onClick={handleLogout}>
            Выйти
          </Button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-3xl font-bold">Мои проекты</h2>
          <Button onClick={() => setShowCreateForm(true)}>+ Новый проект</Button>
        </div>

        {showCreateForm && (
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>Создать новый проект</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <Input
                  placeholder="Название проекта"
                  value={newProject.name}
                  onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
                />
                <Input
                  placeholder="Ниша (например: онлайн-курсы)"
                  value={newProject.niche}
                  onChange={(e) => setNewProject({ ...newProject, niche: e.target.value })}
                />
                <div className="flex gap-2">
                  <Button onClick={handleCreateProject}>Создать</Button>
                  <Button variant="outline" onClick={() => setShowCreateForm(false)}>
                    Отмена
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {projects.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-gray-600 mb-4">У вас пока нет проектов</p>
              <Button onClick={() => setShowCreateForm(true)}>Создать первый проект</Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((project) => (
              <Card
                key={project.id}
                className="relative cursor-pointer transition hover:shadow-lg"
                onClick={() => router.push(`/app/projects/${project.id}`)}
              >
                    <Button
                      variant="ghost"
                      size="sm"
                      className="absolute right-3 top-3 h-8 w-8 text-muted-foreground hover:text-destructive"
                      onClick={(event) => handleDeleteProject(event, project.id)}
                      disabled={deletingId === project.id}
                    >
                  {deletingId === project.id ? (
                    <span className="text-xs">···</span>
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                </Button>
                <CardHeader className="space-y-2 pr-12">
                  <CardTitle>{project.name}</CardTitle>
                  <CardDescription>{project.niche}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between">
                    <span
                      className={`px-2.5 py-1 text-xs font-medium uppercase tracking-wide ${projectStatusClasses[project.status]}`}
                    >
                      {projectStatusLabels[project.status] ?? project.status}
                    </span>
                    <span className="text-sm text-gray-500">{formatDate(project.created_at)}</span>
                  </div>

                  {project.publish && (
                    <div className="rounded-lg border border-green-200 bg-green-50 px-3 py-2">
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <p className="text-xs font-semibold uppercase tracking-wide text-green-700">
                            {publishStatusLabels[project.publish!.status] ?? project.publish!.status}
                          </p>
                          <a
                            href={project.publish!.public_url}
                            onClick={(event) => event.stopPropagation()}
                            className="block break-all text-sm font-medium text-green-800 hover:underline"
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            {project.publish!.public_url}
                          </a>
                          {project.publish!.last_published_at && (
                            <p className="mt-1 text-xs text-gray-600">
                              Обновлено {formatDateTime(project.publish!.last_published_at)}
                            </p>
                          )}
                        </div>
                        <div className="flex shrink-0 items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0"
                            onClick={(event) => handleCopyLink(event, project.publish!.public_url)}
                            title="Скопировать ссылку"
                          >
                            <Copy className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={(event) => handleOpenLink(event, project.publish!.public_url)}
                          >
                            <ExternalLink className="mr-2 h-4 w-4" />
                            Открыть
                          </Button>
                        </div>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}

