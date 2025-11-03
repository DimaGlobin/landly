'use client'

import {
  FormEvent,
  KeyboardEvent,
  MouseEvent,
  useEffect,
  useRef,
  useState,
} from 'react'
import { useRouter, useParams } from 'next/navigation'
import { api } from '@/lib/api'
import {
  ChatHistoryResponse,
  ChatMessage,
  ChatSession,
  LandingSchema,
  Project,
  ProjectPublishInfo,
} from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { LandingPreview } from '@/components/landing-preview'
import { Copy, ExternalLink, PauseCircle, RefreshCcw } from 'lucide-react'

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

const parseSchema = (schemaJSON?: string | null): LandingSchema | null => {
  if (!schemaJSON) return null
  try {
    return JSON.parse(schemaJSON)
  } catch (error) {
    console.error('Failed to parse schema JSON', error)
    return null
  }
}

const formatMessageTime = (value: string) =>
  new Date(value).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })

const truncate = (value: string, limit = 80) => {
  const runes = Array.from(value)
  if (runes.length <= limit) {
    return value
  }
  return `${runes.slice(0, limit).join('')}...`
}

export default function ProjectPage() {
  const router = useRouter()
  const params = useParams()
  const projectId = params.id as string

  const [project, setProject] = useState<Project | null>(null)
  const [schema, setSchema] = useState<LandingSchema | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isPublishing, setIsPublishing] = useState(false)
  const [isUnpublishing, setIsUnpublishing] = useState(false)
  const [publishInfo, setPublishInfo] = useState<ProjectPublishInfo | null>(null)
  const [showPublishBanner, setShowPublishBanner] = useState(false)
  const [chatSession, setChatSession] = useState<ChatSession | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [isSending, setIsSending] = useState(false)
  const [confirmAction, setConfirmAction] = useState<'publish' | 'unpublish' | null>(null)
  const bottomRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    loadProject()
  }, [projectId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (publishInfo?.public_url) {
      setShowPublishBanner(true)
      window.scrollTo({ top: 0, behavior: 'smooth' })
      const timer = window.setTimeout(() => setShowPublishBanner(false), 9000)
      return () => window.clearTimeout(timer)
    }
    return undefined
  }, [publishInfo?.public_url])

  const applyChat = (chatData?: ChatHistoryResponse) => {
    if (!chatData || !chatData.session) {
      setChatSession(null)
      setMessages([])
      setSchema(null)
      return
    }

    setChatSession(chatData.session)
    setMessages(chatData.messages ?? [])
    setSchema(parseSchema(chatData.session.schema_json))

    setProject((prev) => {
      if (!prev) return prev
      const nextStatus = chatData.session.status === 'completed'
        ? prev.status === 'published'
          ? prev.status
          : 'generated'
        : prev.status
      return {
        ...prev,
        status: nextStatus,
        updated_at: chatData.session.updated_at,
      }
    })
  }

  const fetchProjectData = async () => {
    const projectData = await api.getProject(projectId)
    setProject(projectData)
    setPublishInfo(projectData.publish ?? null)

    try {
      const chatData = await api.getChatHistory(projectId)
      applyChat(chatData)
    } catch (error) {
      console.warn('Chat history is empty', error)
      applyChat(undefined)
    }
  }

  const loadProject = async () => {
    try {
      setIsLoading(true)
      await fetchProjectData()
    } catch (error) {
      console.error('Failed to load project', error)
      router.push('/app/projects')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    await sendMessage()
  }

  const handleTextareaKeyDown = async (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      if (!isSending) {
        await sendMessage()
      }
    }
  }

  const sendMessage = async () => {
    const trimmed = input.trim()
    if (!trimmed) {
      return
    }

    try {
      setIsSending(true)
      const chatData = await api.sendChatMessage(projectId, trimmed)
      applyChat(chatData)
      setInput('')
    } catch (error: any) {
      console.error('Failed to send message', error)
      alert('Не удалось отправить сообщение. Попробуйте ещё раз.')
    } finally {
      setIsSending(false)
    }
  }

  const executePublish = async () => {
    try {
      setIsPublishing(true)
      const result = await api.publishProject(projectId)
      setPublishInfo({
        status: 'published',
        public_url: result.public_url,
        subdomain: result.subdomain,
        last_published_at: result.published_at,
      })
      await fetchProjectData()
      window.scrollTo({ top: 0, behavior: 'smooth' })
      setShowPublishBanner(true)
    } catch (error: any) {
      alert('Ошибка публикации: ' + (error.response?.data?.error || error.message))
    } finally {
      setIsPublishing(false)
    }
  }

  const executeUnpublish = async () => {
    try {
      setIsUnpublishing(true)
      await api.unpublishProject(projectId)
      setPublishInfo(null)
      await fetchProjectData()
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } catch (error: any) {
      alert('Ошибка: ' + (error.response?.data?.error || error.message))
    } finally {
      setIsUnpublishing(false)
    }
  }

  const openPublishConfirm = () => {
    setConfirmAction('publish')
  }

  const openUnpublishConfirm = (event: MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation()
    setConfirmAction('unpublish')
  }

  const closeConfirmDialog = () => {
    if (isPublishing || isUnpublishing) return
    setConfirmAction(null)
  }

  const confirmActionHandler = async () => {
    if (confirmAction === 'publish') {
      await executePublish()
    } else if (confirmAction === 'unpublish') {
      await executeUnpublish()
    }
    setConfirmAction(null)
  }

  const handleCopyLink = async (event: MouseEvent<HTMLButtonElement>, url: string) => {
    event.stopPropagation()
    try {
      await navigator.clipboard.writeText(url)
      setShowPublishBanner(true)
    } catch (error) {
      console.error('Не удалось скопировать ссылку', error)
    }
  }

  const handleOpenLink = (event: MouseEvent<HTMLButtonElement>, url: string) => {
    event.stopPropagation()
    window.open(url, '_blank', 'noopener,noreferrer')
  }

  if (isLoading) {
    return <div className="min-h-screen flex items-center justify-center">Загрузка...</div>
  }

  if (!project) {
    return <div className="min-h-screen flex items-center justify-center">Проект не найден</div>
  }

  return (
    <div className="app-shell">
      <div className="relative z-10 mx-auto max-w-7xl px-6 py-10 space-y-8">
        {publishInfo?.public_url && showPublishBanner && (
          <div className="glass-panel border-emerald-400/50 bg-emerald-500/10 px-6 py-4 text-sm text-emerald-800 shadow-2xl transition">
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
              <div className="space-y-1">
                <p className="font-semibold uppercase tracking-wide text-emerald-700">Лендинг опубликован</p>
                <a
                  href={publishInfo.public_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="break-all font-medium hover:underline"
                  onClick={(event) => event.stopPropagation()}
                >
                  {publishInfo.public_url}
                </a>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-9 w-9 rounded-full bg-white/80 p-0 text-emerald-700 hover:bg-white"
                  onClick={(event) => handleCopyLink(event, publishInfo.public_url)}
                  title="Скопировать ссылку"
                >
                  <Copy className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-9 rounded-full border border-emerald-300 bg-white/80 px-4 text-xs font-semibold text-emerald-700 shadow-sm hover:bg-white"
                  onClick={(event) => handleOpenLink(event, publishInfo.public_url)}
                >
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Открыть
                </Button>
              </div>
            </div>
          </div>
        )}
        <div className="glass-panel px-6 py-5 md:px-8 md:py-6">
          <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:gap-5">
              <Button
                variant="ghost"
                onClick={() => router.push('/app/projects')}
                className="h-10 w-fit rounded-full border border-white/50 bg-white/80 px-5 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-white"
              >
                ← Назад к проектам
              </Button>
              <div>
                <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{project.name}</h1>
                <p className="mt-1 text-sm text-slate-600">{project.niche}</p>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-3">
              <span
                className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${projectStatusClasses[project.status]}`}
              >
                <span className="h-2 w-2 rounded-full bg-current/80" />
                {projectStatusLabels[project.status] ?? project.status}
              </span>
              {schema && (
                <div className="flex flex-wrap items-center gap-3">
                  <Button
                    onClick={openPublishConfirm}
                    disabled={isPublishing}
                    className={`h-10 rounded-full px-5 text-sm font-semibold transition hover:-translate-y-0.5 focus:ring-2 focus:ring-offset-0 ${
                      project.status === 'published'
                        ? 'border border-blue-400/50 bg-blue-500/10 text-blue-700 hover:bg-blue-500/20'
                        : 'bg-blue-600 text-white hover:bg-blue-700'
                    }`}
                  >
                    {isPublishing ? (
                      'Публикация...'
                    ) : project.status === 'published' ? (
                      <span className="flex items-center gap-2">
                        <RefreshCcw className="h-4 w-4" />
                        Обновить публикацию
                      </span>
                    ) : (
                      'Опубликовать'
                    )}
                  </Button>
                  {project.status === 'published' && (
                    <Button
                      variant="ghost"
                      onClick={openUnpublishConfirm}
                      disabled={isUnpublishing}
                      className="h-10 rounded-full border border-slate-200 bg-white/80 px-5 text-sm font-semibold text-slate-700 shadow-sm transition hover:bg-white"
                    >
                      {isUnpublishing ? (
                        'Остановка...'
                      ) : (
                        <span className="flex items-center gap-2">
                          <PauseCircle className="h-4 w-4" />
                          Снять с публикации
                        </span>
                      )}
                    </Button>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {publishInfo?.public_url && (
          <div className="glass-panel sticky top-6 border-emerald-400/40 bg-emerald-50/70 px-6 py-5 shadow-xl md:px-8">
            <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
              <div>
                <p className="mb-1 text-sm font-semibold text-emerald-700">
                  ✅ {publishStatusLabels[publishInfo.status ?? 'published'] ?? 'Опубликовано'}
                </p>
                <a
                  href={publishInfo.public_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="break-all text-sm font-medium text-emerald-800 hover:underline"
                  onClick={(event) => event.stopPropagation()}
                >
                  {publishInfo.public_url}
                </a>
                {publishInfo.last_published_at && (
                  <p className="mt-2 text-xs text-emerald-900/70">
                    Обновлено {new Date(publishInfo.last_published_at).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-9 w-9 rounded-full bg-white/70 p-0 text-emerald-700 hover:bg-white"
                  onClick={(event) => handleCopyLink(event, publishInfo.public_url)}
                  title="Скопировать ссылку"
                >
                  <Copy className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-9 rounded-full border border-emerald-300 bg-white/80 px-4 text-sm font-medium text-emerald-700 shadow-sm hover:bg-white"
                  onClick={(event) => handleOpenLink(event, publishInfo.public_url)}
                >
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Открыть
                </Button>
              </div>
            </div>
          </div>
        )}

        <div className="flex flex-col gap-6 xl:flex-row">
          <Card className="surface-card flex h-[72vh] flex-col border-white/40 text-slate-900 xl:h-[calc(100vh-260px)] xl:w-[380px]">
            <div className="border-b border-white/50 px-6 py-4">
              <h2 className="text-lg font-semibold">AI-ассистент</h2>
              <p className="mt-1 text-sm text-slate-600">
                Опишите свою идею или изменения, и я обновлю лендинг. Можно задавать вопросы и уточнять детали — всё как в чате.
              </p>
            </div>
            <div className="flex-1 space-y-4 overflow-y-auto bg-white/50 px-6 py-4">
              {messages.length === 0 ? (
                <p className="text-sm text-slate-500">
                  Расскажите, какой лендинг вы хотите получить. Например: «Сделай лендинг для онлайн-курса по Data Science с акцентом на карьерный рост».
                </p>
              ) : (
                messages.map((message) => {
                  const isUser = message.role === 'user'
                  const bubbleClasses = isUser
                    ? 'bg-primary text-white'
                    : 'bg-white text-gray-900 border border-gray-200'
                  const alignClass = isUser ? 'justify-end' : 'justify-start'
                  const label = message.role === 'assistant' ? 'Landly AI' : message.role === 'system' ? 'Система' : 'Вы'

                  return (
                    <div key={message.id} className={`flex ${alignClass}`}>
                      <div className="max-w-[80%]">
                        <div className={`rounded-2xl px-5 py-3.5 text-sm shadow-sm transition ${bubbleClasses}`}>
                          <p className="whitespace-pre-wrap text-sm leading-relaxed">{message.content}</p>
                        </div>
                        <p className={`mt-1 text-xs text-slate-500 ${isUser ? 'text-right' : ''}`}>
                          {label} · {formatMessageTime(message.created_at)}
                        </p>
                      </div>
                    </div>
                  )
                })
              )}
              <div ref={bottomRef} />
            </div>
            <form className="border-t border-white/50 bg-white/70 px-6 py-4" onSubmit={handleSubmit}>
              <div className="space-y-3">
                <textarea
                  value={input}
                  onChange={(event) => setInput(event.target.value)}
                  onKeyDown={handleTextareaKeyDown}
                  placeholder="Опишите, что нужно добавить или изменить..."
                  className="h-28 w-full resize-none rounded-2xl border border-slate-200/70 bg-white/80 px-4 py-3 text-sm text-slate-900 shadow-sm focus:border-blue-400 focus:outline-none focus:ring-2 focus:ring-blue-400/40"
                  disabled={isSending}
                />
                <div className="flex items-center justify-between">
                  <span className="text-xs text-slate-500">Enter — отправить, Shift + Enter — новая строка</span>
                  <Button
                    type="submit"
                    disabled={isSending || !input.trim()}
                    className="h-10 rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700"
                  >
                    {isSending ? 'Генерация...' : 'Отправить'}
                  </Button>
                </div>
              </div>
            </form>
          </Card>

          <div className="flex-1 space-y-6">
            <Card className="surface-card flex h-[72vh] flex-col border-white/40 text-slate-900 xl:h-[calc(100vh-260px)]">
              <CardHeader className="flex-shrink-0 border-b border-white/50 px-6 py-4">
                <CardTitle className="text-lg font-semibold">Предпросмотр лендинга</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-1 items-stretch rounded-b-[1.5rem] p-0">
                {schema ? (
                  <div className="flex-1 overflow-hidden">
                    <LandingPreview schema={schema} />
                  </div>
                ) : (
                  <div className="flex flex-1 items-center justify-center border border-dashed border-slate-300/60 p-6 text-sm text-slate-500">
                    Сгенерируйте описание, чтобы увидеть лендинг.
                  </div>
                )}
              </CardContent>
            </Card>

            {chatSession && (
              <Card className="surface-card border-white/40 text-slate-900">
                <CardHeader className="border-b border-white/50">
                  <CardTitle className="text-base font-semibold">Сессия генерации</CardTitle>
                  <CardDescription className="text-xs text-slate-500">
                    Последнее обновление: {new Date(chatSession.updated_at).toLocaleString('ru-RU')}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-2 text-sm text-slate-700">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-500">Статус</span>
                    <span className="font-medium text-slate-800">
                      {chatSession.status === 'completed'
                        ? 'Готово'
                        : chatSession.status === 'failed'
                        ? 'Ошибка'
                        : 'В процессе'}
                    </span>
                  </div>
                  {chatSession.completed_at && (
                    <div className="flex items-center justify-between">
                      <span className="text-slate-500">Ответ</span>
                      <span className="font-medium text-slate-800">
                        {new Date(chatSession.completed_at).toLocaleString('ru-RU')}
                      </span>
                    </div>
                  )}
                  <div className="flex items-center justify-between">
                    <span className="text-slate-500">ID сессии</span>
                    <span className="font-mono text-xs text-slate-500">{truncate(chatSession.id, 18)}</span>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </div>

      {confirmAction && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <div className="mb-4">
              <h3 className="text-lg font-semibold">
                {confirmAction === 'publish' ? 'Опубликовать лендинг?' : 'Снять лендинг с публикации?'}
              </h3>
              <p className="mt-2 text-sm text-gray-600">
                {confirmAction === 'publish'
                  ? 'После публикации лендинг станет доступен по публичной ссылке. Продолжить?'
                  : 'Публичная ссылка перестанет работать, но проект останется в списке. Продолжить?'}
              </p>
            </div>
            <div className="flex justify-end gap-3">
              <Button
                variant="outline"
                onClick={closeConfirmDialog}
                disabled={confirmAction === 'publish' ? isPublishing : isUnpublishing}
              >
                Отмена
              </Button>
              <Button
                variant={confirmAction === 'publish' ? 'default' : 'destructive'}
                onClick={confirmActionHandler}
                disabled={confirmAction === 'publish' ? isPublishing : isUnpublishing}
              >
                {confirmAction === 'publish'
                  ? isPublishing
                    ? 'Публикуем…'
                    : 'Опубликовать'
                  : isUnpublishing
                    ? 'Останавливаем…'
                    : 'Снять с публикации'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

