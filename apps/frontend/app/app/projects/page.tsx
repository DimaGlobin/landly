'use client'

import { useEffect, useState, MouseEvent } from 'react'
import { useRouter } from 'next/navigation'
import { Trash2, ExternalLink, Copy } from 'lucide-react'
import { api } from '@/lib/api'
import { Project, getPublicURL } from '@/lib/types'
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
    } catch (error: any) {
      console.error('Failed to load projects', error)
      // Only redirect if it's an auth error and we're not already on auth page
      if (error instanceof Error && 'shouldRedirectToLogin' in error && (error as any).shouldRedirectToLogin()) {
        // API client will handle logout and redirect
        return
      }
      // For other errors, show error state instead of redirecting
      setProjects([])
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
      <div className="relative z-10 mx-auto max-w-5xl px-4 py-6 space-y-6 sm:px-6 sm:py-10 sm:space-y-8">
        <div className="glass-panel px-4 py-4 sm:px-6 sm:py-5 md:px-8 md:py-6">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="min-w-0 flex-1">
              <p className="text-xs font-semibold uppercase tracking-wide text-blue-600/80 sm:text-sm">Landly Workspace</p>
              <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl md:text-4xl">
                Мои проекты
              </h1>
              <p className="mt-2 max-w-2xl text-xs text-slate-600 sm:text-sm">
                Управляйте проектами, запускайте генерацию лендингов в чате и публикуйте результат в один клик.
              </p>
            </div>
            <div className="flex shrink-0 items-center gap-2 sm:gap-3 flex-wrap">
              <Button
                variant="ghost"
                onClick={() => router.push('/app/profile')}
                className="h-9 whitespace-nowrap rounded-full border border-white/50 bg-white/80 px-4 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-white sm:h-10 sm:px-5 sm:text-sm"
              >
                Профиль
              </Button>
              <Button
                variant="ghost"
                onClick={handleLogout}
                className="h-9 whitespace-nowrap rounded-full border border-white/50 bg-white/80 px-4 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-white sm:h-10 sm:px-5 sm:text-sm"
              >
                Выйти
              </Button>
              <Button
                onClick={() => setShowCreateForm(true)}
                className="h-9 whitespace-nowrap rounded-full bg-blue-600 px-4 text-xs font-semibold text-white shadow-lg transition hover:bg-blue-700 sm:h-10 sm:px-6 sm:text-sm"
              >
                <span className="hidden sm:inline">+ Новый проект</span>
                <span className="sm:hidden">+ Новый</span>
              </Button>
            </div>
          </div>
        </div>

        {showCreateForm && (
          <Card className="surface-card border-white/40 text-slate-900">
            <CardHeader className="border-b border-white/50 px-4 py-4 sm:px-6 sm:py-6">
              <CardTitle className="text-base font-semibold sm:text-lg">Создать новый проект</CardTitle>
              <CardDescription className="text-xs text-slate-600 sm:text-sm">
                Мы спросим только название и нишу. После создания откроется чат, где можно описать идею и получить лендинг.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 px-4 py-4 sm:px-6 sm:py-6">
              <Input
                placeholder="Название проекта"
                value={newProject.name}
                onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
                className="h-10 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50 sm:h-11"
              />
              <Input
                placeholder="Ниша (например: онлайн-курсы)"
                value={newProject.niche}
                onChange={(e) => setNewProject({ ...newProject, niche: e.target.value })}
                className="h-10 rounded-xl border border-slate-200/70 bg-white/90 text-sm shadow-sm focus-visible:ring-blue-400/50 sm:h-11"
              />
              <div className="flex flex-col gap-2 sm:flex-row">
                <Button
                  onClick={handleCreateProject}
                  className="h-10 rounded-full bg-blue-600 px-4 text-xs font-semibold text-white shadow-lg transition hover:bg-blue-700 sm:px-6 sm:text-sm"
                >
                  Создать
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => setShowCreateForm(false)}
                  className="h-10 rounded-full border border-slate-200 bg-white/80 px-4 text-xs font-semibold text-slate-700 shadow-sm hover:bg-white sm:px-6 sm:text-sm"
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
          <div className="grid gap-6 md:grid-cols-2">
            {projects.map((project) => (
              <Card
                key={project.id}
                className="surface-card group relative cursor-pointer border-white/40 text-slate-900 transition hover:-translate-y-1 hover:shadow-2xl"
                onClick={() => router.push(`/app/projects/${project.id}`)}
              >
                <Button
                  variant="ghost"
                  size="sm"
                  className="absolute right-2 top-2 inline-flex items-center gap-1 rounded-full border border-transparent bg-white/80 px-2 py-1 text-xs font-medium text-slate-500 shadow-sm transition hover:border-red-200 hover:bg-red-50 hover:text-red-500 sm:right-4 sm:top-4 sm:gap-2 sm:px-3 sm:py-1.5"
                  onClick={(event) => handleDeleteProject(event, project)}
                  disabled={deletingId === project.id}
                >
                  <Trash2 className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
                  <span className="hidden sm:inline">{deletingId === project.id ? 'Удаляем…' : 'Удалить'}</span>
                </Button>
                <CardHeader className="pr-20 sm:pr-16 px-4 pt-4 sm:px-6 sm:pt-6">
                  <CardTitle className="text-lg font-semibold text-slate-900 truncate sm:text-xl">{project.name}</CardTitle>
                  <CardDescription className="text-xs text-slate-600 truncate sm:text-sm">{project.niche}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4 px-4 pb-4 sm:px-6 sm:pb-6">
                  <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                    <span
                      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wide shrink-0 sm:gap-2 sm:px-3 sm:text-[11px] ${projectStatusClasses[project.status]}`}
                    >
                      <span className="h-1.5 w-1.5 rounded-full bg-current/80 sm:h-2 sm:w-2" />
                      <span className="hidden sm:inline">{projectStatusLabels[project.status] ?? project.status}</span>
                      <span className="sm:hidden">
                        {project.status === 'draft' ? 'Черн.' : project.status === 'generated' ? 'Сген.' : 'Опубл.'}
                      </span>
                    </span>
                    <span className="text-xs text-slate-500 sm:text-sm">Обновлён {formatDate(project.updated_at || project.created_at)}</span>
                  </div>

                  {project.publish && (() => {
                    const publicURL = getPublicURL(project.publish.subdomain)
                    return (
                      <div className="rounded-2xl border border-emerald-400/40 bg-emerald-50/60 px-3 py-2.5 sm:px-4 sm:py-3">
                        <div className="flex items-start justify-between gap-2 sm:gap-3">
                          <div className="min-w-0 flex-1">
                            <p className="text-[10px] font-semibold uppercase tracking-wide text-emerald-700 sm:text-[11px]">
                              {publishStatusLabels[project.publish.status] ?? project.publish.status}
                            </p>
                            <a
                              href={publicURL}
                              onClick={(event) => event.stopPropagation()}
                              className="mt-1 block break-all text-xs font-medium text-emerald-800 hover:underline sm:text-sm"
                              target="_blank"
                              rel="noopener noreferrer"
                            >
                              {publicURL}
                            </a>
                            {project.publish.last_published_at && (
                              <p className="mt-1 text-xs text-emerald-900/70">
                                Обновлено {formatDateTime(project.publish.last_published_at)}
                              </p>
                            )}
                          </div>
                          <div className="flex shrink-0 items-center gap-1 sm:gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 w-7 rounded-full bg-white/80 p-0 text-emerald-700 hover:bg-white sm:h-8 sm:w-8"
                              onClick={(event) => handleCopyLink(event, publicURL)}
                              title="Скопировать ссылку"
                            >
                              <Copy className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 rounded-full border border-emerald-300 bg-white/80 px-2 text-[10px] font-semibold text-emerald-700 shadow-sm hover:bg-white sm:h-8 sm:px-3 sm:text-xs"
                              onClick={(event) => handleOpenLink(event, publicURL)}
                            >
                              <ExternalLink className="mr-1 h-3 w-3 sm:mr-2 sm:h-3.5 sm:w-3.5" />
                              <span className="hidden sm:inline">Открыть</span>
                              <span className="sm:hidden">Откр.</span>
                            </Button>
                          </div>
                        </div>
                      </div>
                    )
                  })()}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      {projectToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-4 shadow-2xl sm:p-6">
            <div className="mb-4">
              <h3 className="text-base font-semibold sm:text-lg">Удалить проект?</h3>
              <p className="mt-2 text-xs text-gray-600 sm:text-sm">
                Это действие нельзя отменить. Проект «{projectToDelete.name}» и все его данные будут удалены навсегда.
              </p>
            </div>
            <div className="flex flex-col gap-2 sm:flex-row sm:justify-end sm:gap-3">
              <Button
                variant="outline"
                onClick={cancelDeleteProject}
                disabled={deletingId === projectToDelete.id}
                className="h-9 text-xs sm:h-10 sm:text-sm"
              >
                Отмена
              </Button>
              <Button
                variant="destructive"
                onClick={confirmDeleteProject}
                disabled={deletingId === projectToDelete.id}
                className="h-9 text-xs sm:h-10 sm:text-sm"
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

