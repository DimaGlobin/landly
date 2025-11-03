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
  draft: 'bg-white/70 border border-slate-200 text-slate-700',
  generated: 'bg-blue-500/10 border border-blue-400/40 text-blue-700',
  published: 'bg-emerald-500/10 border border-emerald-400/40 text-emerald-700',
}

const publishStatusLabels: Record<string, string> = {
  draft: 'Черновик публикации',
  published: 'Опубликовано',
  failed: 'Ошибка публикации',
}

const AUTO_REDIRECT_KEY = 'landly:auto-opened'

export default function ProjectsPage() {
  const router = useRouter()
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newProject, setNewProject] = useState({ name: '', niche: '' })
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [projectToDelete, setProjectToDelete] = useState<Project | null>(null)

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

  useEffect(() => {
    if (typeof window === 'undefined') return
    if (isLoading) return
    if (showCreateForm) return

    const alreadyRedirected = sessionStorage.getItem(AUTO_REDIRECT_KEY)
    if (!alreadyRedirected && projects.length > 0) {
      sessionStorage.setItem(AUTO_REDIRECT_KEY, '1')
      router.replace(`/app/projects/${projects[0].id}`)
    }
  }, [isLoading, showCreateForm, projects, router])

  const loadProjects = async () => {
    try {
      const data = await api.getProjects()
      const sorted = (data.projects || []).slice().sort((a: Project, b: Project) => {
        const left = new Date(b.updated_at || b.created_at).getTime()
        const right = new Date(a.updated_at || a.created_at).getTime()
        return left - right
      })
      setProjects(sorted)
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
      sessionStorage.setItem(AUTO_REDIRECT_KEY, '1')
      router.push(`/app/projects/${project.id}`)
    } catch (error) {
      console.error('Failed to create project', error)
    }
  }

  const handleDeleteProject = (event: MouseEvent<HTMLButtonElement>, project: Project) => {
    event.stopPropagation()
    setProjectToDelete(project)
  }

  const cancelDeleteProject = () => {
    if (deletingId) return
    setProjectToDelete(null)
  }

  const confirmDeleteProject = async () => {
    if (!projectToDelete) return

    try {
      setDeletingId(projectToDelete.id)
      await api.deleteProject(projectToDelete.id)
      setProjects((prev) => prev.filter((project) => project.id !== projectToDelete.id))
      setProjectToDelete(null)
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
    if (typeof window !== 'undefined') {
      sessionStorage.removeItem(AUTO_REDIRECT_KEY)
    }
    router.push('/')
  }

  if (isLoading) {
    return <div className="min-h-screen flex items-center justify-center">Загрузка...</div>
  }

  return (
    <div className="app-shell">
      <div className="relative z-10 mx-auto max-w-7xl px-6 py-10 space-y-8">
        <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div className="glass-panel px-6 py-5 md:px-8 md:py-6">
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
              <div>
                <p className="text-sm font-semibold uppercase tracking-wide text-blue-600/80">Landly Workspace</p>
                <h1 className="mt-1 text-3xl font-bold tracking-tight text-slate-900 md:text-4xl">
                  Мои проекты
                </h1>
                <p className="mt-2 max-w-2xl text-sm text-slate-600">
                  Управляйте проектами, запускайте генерацию лендингов в чате и публикуйте результат в один клик.
                </p>
              </div>
              <div className="flex items-center gap-3">
                <Button
                  variant="ghost"
                  onClick={handleLogout}
                  className="h-10 rounded-full border border-white/50 bg-white/80 px-5 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-white"
                >
                  Выйти
                </Button>
                <Button
                  onClick={() => setShowCreateForm(true)}
                  className="h-10 rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-lg transition hover:bg-blue-700"
                >
                  + Новый проект
                </Button>
              </div>
            </div>
          </div>
        </div>

        {showCreateForm && (
          <Card className="surface-card border-white/40 text-slate-900">
            <CardHeader className="border-b border-white/50">
              <CardTitle className="text-lg font-semibold">Создать новый проект</CardTitle>
              <CardDescription className="text-sm text-slate-600">
                Мы спросим только название и нишу. После создания откроется чат, где можно описать идею и получить лендинг.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 py-6">
              <Input
                placeholder="Название проекта"
                value={newProject.name}
                onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
                className="h-11 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50"
              />
              <Input
                placeholder="Ниша (например: онлайн-курсы)"
                value={newProject.niche}
                onChange={(e) => setNewProject({ ...newProject, niche: e.target.value })}
                className="h-11 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50"
              />
              <div className="flex gap-2">
                <Button
                  onClick={handleCreateProject}
                  className="h-10 rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-lg transition hover:bg-blue-700"
                >
                  Создать
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => setShowCreateForm(false)}
                  className="h-10 rounded-full border border-slate-200 bg-white/80 px-6 text-sm font-semibold text-slate-700 shadow-sm hover:bg-white"
                >
                  Отмена
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {projects.length === 0 ? (
          <Card className="surface-card border-white/40 text-slate-900">
            <CardContent className="py-12 text-center">
              <p className="mb-4 text-slate-600">
                У вас пока нет проектов. Создайте первый и сразу начните диалог с AI, чтобы описать лендинг.
              </p>
              <Button
                onClick={() => setShowCreateForm(true)}
                className="h-10 rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-lg hover:bg-blue-700"
              >
                Создать первый проект
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            {projects.map((project) => (
              <Card
                key={project.id}
                className="surface-card group relative cursor-pointer border-white/40 text-slate-900 transition hover:-translate-y-1 hover:shadow-2xl"
                onClick={() => router.push(`/app/projects/${project.id}`)}
              >
                <Button
                  variant="ghost"
                  size="sm"
                  className="absolute right-4 top-4 inline-flex items-center gap-2 rounded-full border border-transparent bg-white/80 px-3 py-1.5 text-xs font-medium text-slate-500 shadow-sm transition hover:border-red-200 hover:bg-red-50 hover:text-red-500"
                  onClick={(event) => handleDeleteProject(event, project)}
                  disabled={deletingId === project.id}
                >
                  <Trash2 className="h-4 w-4" />
                  {deletingId === project.id ? 'Удаляем…' : 'Удалить'}
                </Button>
                <CardHeader className="pr-16">
                  <CardTitle className="text-xl font-semibold text-slate-900">{project.name}</CardTitle>
                  <CardDescription className="text-sm text-slate-600">{project.niche}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4 pb-6">
                  <div className="flex items-center justify-between">
                    <span
                      className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-wide ${projectStatusClasses[project.status]}`}
                    >
                      <span className="h-2 w-2 rounded-full bg-current/80" />
                      {projectStatusLabels[project.status] ?? project.status}
                    </span>
                    <span className="text-sm text-slate-500">Обновлён {formatDate(project.updated_at || project.created_at)}</span>
                  </div>

                  {project.publish && (
                    <div className="rounded-2xl border border-emerald-400/40 bg-emerald-50/60 px-4 py-3">
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <p className="text-[11px] font-semibold uppercase tracking-wide text-emerald-700">
                            {publishStatusLabels[project.publish!.status] ?? project.publish!.status}
                          </p>
                          <a
                            href={project.publish!.public_url}
                            onClick={(event) => event.stopPropagation()}
                            className="mt-1 block break-all text-sm font-medium text-emerald-800 hover:underline"
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            {project.publish!.public_url}
                          </a>
                          {project.publish!.last_published_at && (
                            <p className="mt-1 text-xs text-emerald-900/70">
                              Обновлено {formatDateTime(project.publish!.last_published_at)}
                            </p>
                          )}
                        </div>
                        <div className="flex shrink-0 items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 rounded-full bg-white/80 p-0 text-emerald-700 hover:bg-white"
                            onClick={(event) => handleCopyLink(event, project.publish!.public_url)}
                            title="Скопировать ссылку"
                          >
                            <Copy className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 rounded-full border border-emerald-300 bg-white/80 px-3 text-xs font-semibold text-emerald-700 shadow-sm hover:bg-white"
                            onClick={(event) => handleOpenLink(event, project.publish!.public_url)}
                          >
                            <ExternalLink className="mr-2 h-3.5 w-3.5" />
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
      </div>

      {projectToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <div className="mb-4">
              <h3 className="text-lg font-semibold">Удалить проект?</h3>
              <p className="mt-2 text-sm text-gray-600">
                Это действие нельзя отменить. Проект «{projectToDelete.name}» и все его данные будут удалены навсегда.
              </p>
            </div>
            <div className="flex justify-end gap-3">
              <Button
                variant="outline"
                onClick={cancelDeleteProject}
                disabled={deletingId === projectToDelete.id}
              >
                Отмена
              </Button>
              <Button
                variant="destructive"
                onClick={confirmDeleteProject}
                disabled={deletingId === projectToDelete.id}
              >
                {deletingId === projectToDelete.id ? 'Удаляем…' : 'Удалить навсегда'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

