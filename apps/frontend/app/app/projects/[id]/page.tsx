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
  getPublicURL,
} from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LandingPreview } from '@/components/landing-preview'
import { SchemaDiff } from '@/components/schema-diff'
import { SchemaHistory } from '@/components/schema-history'
import { Copy, ExternalLink, PauseCircle, RefreshCcw, History } from 'lucide-react'

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

export default function ProjectPage() {
  const router = useRouter()
  const params = useParams()
  const projectId = params.id as string

  const [project, setProject] = useState<Project | null>(null)
  const [schema, setSchema] = useState<LandingSchema | null>(null)
  const [previousSchema, setPreviousSchema] = useState<LandingSchema | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isPublishing, setIsPublishing] = useState(false)
  const [isUnpublishing, setIsUnpublishing] = useState(false)
  const [publishInfo, setPublishInfo] = useState<ProjectPublishInfo | null>(null)
  const [chatSession, setChatSession] = useState<ChatSession | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [isSending, setIsSending] = useState(false)
  const [confirmAction, setConfirmAction] = useState<'publish' | 'unpublish' | null>(null)
  const [showHistory, setShowHistory] = useState(false)
  const chatContainerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    loadProject()
  }, [projectId])

  // Скролл только внутри контейнера чата, не всей страницы
  useEffect(() => {
    if (chatContainerRef.current) {
      chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight
    }
  }, [messages])


  const applyChat = (chatData?: ChatHistoryResponse) => {
    if (!chatData || !chatData.session) {
      setChatSession(null)
      setMessages([])
      setSchema(null)
      return
    }

    setChatSession(chatData.session)
    setMessages(chatData.messages ?? [])
    
    const newSchema = parseSchema(chatData.session.schema_json)
    // Сохраняем предыдущую схему для diff
    if (schema && newSchema) {
      setPreviousSchema(schema)
    }
    setSchema(newSchema)

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
      // Сохраняем текущую схему перед генерацией для diff
      if (schema) {
        setPreviousSchema(schema)
      }
      const chatData = await api.sendChatMessage(projectId, trimmed)
      applyChat(chatData)
      setInput('')
      // Очищаем diff через 5 секунд
      setTimeout(() => {
        setPreviousSchema(null)
      }, 5000)
    } catch (error: any) {
      console.error('Failed to send message', error)
      alert('Не удалось отправить сообщение. Попробуйте ещё раз.')
    } finally {
      setIsSending(false)
    }
  }

  const handleRevert = async () => {
    await fetchProjectData()
    setShowHistory(false)
  }

  const executePublish = async () => {
    try {
      setIsPublishing(true)
      const result = await api.publishProject(projectId)
      
      // Устанавливаем промежуточный статус с subdomain
      // Публикация происходит в фоне, поэтому статус может быть "draft" или "published"
      setPublishInfo({
        status: result.status as 'draft' | 'published',
        subdomain: result.subdomain,
        last_published_at: result.published_at,
      })
      
      // Если статус уже "published", не нужно polling
      if (result.status === 'published') {
        await fetchProjectData()
        return
      }
      
      // Polling для проверки статуса публикации (максимум 30 секунд)
      let attempts = 0
      const maxAttempts = 30
      const pollInterval = 1000 // 1 секунда
      
      const pollStatus = async () => {
        try {
          const projectData = await api.getProject(projectId)
          const publishStatus = projectData.publish?.status
          
          if (publishStatus === 'published') {
            // Публикация завершена успешно
            setPublishInfo({
              status: 'published',
              subdomain: projectData.publish!.subdomain,
              last_published_at: projectData.publish!.last_published_at,
            })
            return
          } else if (publishStatus === 'failed') {
            // Публикация провалилась
            alert('Ошибка публикации: публикация завершилась с ошибкой')
            await fetchProjectData()
            return
          }
          
          // Продолжаем polling
          attempts++
          if (attempts < maxAttempts) {
            setTimeout(pollStatus, pollInterval)
          } else {
            // Таймаут - обновляем данные вручную
            await fetchProjectData()
          }
        } catch (error) {
          console.error('Failed to poll publish status', error)
          // При ошибке обновляем данные
          await fetchProjectData()
        }
      }
      
      // Начинаем polling через 2 секунды (даем время на старт публикации)
      setTimeout(pollStatus, 2000)
    } catch (error: any) {
      alert('Ошибка публикации: ' + (error.response?.data?.error || error.message))
      await fetchProjectData()
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
      <div className="relative z-10 mx-auto max-w-7xl px-4 py-4 space-y-4 md:px-6">
        <div className="glass-panel px-3 py-2.5 md:px-6 md:py-3">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
            <div className="flex items-center gap-2 sm:gap-4 min-w-0 flex-1">
              <Button
                variant="ghost"
                onClick={() => router.push('/app/projects')}
                className="h-8 shrink-0 rounded-full border border-white/50 bg-white/80 px-3 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-white sm:px-4 sm:text-sm"
              >
                <span className="hidden sm:inline">← Назад</span>
                <span className="sm:hidden">←</span>
              </Button>
              <div className="flex items-center gap-2 min-w-0 flex-1 sm:gap-3">
                <h1 className="text-base font-bold tracking-tight truncate sm:text-lg md:text-xl">{project.name}</h1>
                <span className="hidden text-sm text-slate-500 sm:inline">·</span>
                <span className="hidden text-sm text-slate-600 truncate sm:inline">{project.niche}</span>
              </div>
            </div>
            <div className="flex items-center gap-1.5 sm:gap-2 flex-wrap">
              <span
                className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide shrink-0 ${projectStatusClasses[project.status]}`}
              >
                <span className="h-1.5 w-1.5 rounded-full bg-current/80" />
                <span className="hidden sm:inline">{projectStatusLabels[project.status] ?? project.status}</span>
                <span className="sm:hidden">
                  {project.status === 'draft' ? 'Черн.' : project.status === 'generated' ? 'Сген.' : 'Опубл.'}
                </span>
              </span>
              {schema && (
                <>
                  <Button
                    onClick={() => setShowHistory(true)}
                    variant="ghost"
                    className="h-8 shrink-0 rounded-full border border-white/50 bg-white/80 px-2 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-white sm:px-3 sm:text-sm"
                  >
                    <History className="h-3.5 w-3.5 sm:mr-1.5" />
                    <span className="hidden sm:inline">История</span>
                  </Button>
                  <Button
                    onClick={openPublishConfirm}
                    disabled={isPublishing || isSending}
                    className={`h-8 shrink-0 rounded-full px-3 text-xs font-semibold transition sm:px-4 sm:text-sm ${
                      project.status === 'published'
                        ? 'border border-blue-400/50 bg-blue-500/10 text-blue-700 hover:bg-blue-500/20'
                        : 'bg-blue-600 text-white hover:bg-blue-700'
                    } ${isSending ? 'opacity-50 cursor-not-allowed' : ''}`}
                  >
                    {isPublishing ? (
                      '...'
                    ) : isSending ? (
                      '...'
                    ) : project.status === 'published' ? (
                      <span className="flex items-center gap-1.5">
                        <RefreshCcw className="h-3.5 w-3.5" />
                        <span className="hidden sm:inline">Обновить</span>
                        <span className="sm:hidden">Обнов.</span>
                      </span>
                    ) : (
                      <>
                        <span className="hidden sm:inline">Опубликовать</span>
                        <span className="sm:hidden">Опубл.</span>
                      </>
                    )}
                  </Button>
                  {project.status === 'published' && (
                    <Button
                      variant="ghost"
                      onClick={openUnpublishConfirm}
                      disabled={isUnpublishing}
                      className="h-8 w-8 shrink-0 rounded-full border border-slate-200 bg-white/80 p-0 text-slate-500 shadow-sm transition hover:bg-white hover:text-slate-700"
                      title="Снять с публикации"
                    >
                      <PauseCircle className="h-4 w-4" />
                    </Button>
                  )}
                </>
              )}
            </div>
          </div>
        </div>

        {publishInfo?.subdomain && (() => {
          const publicURL = getPublicURL(publishInfo.subdomain)
          return (
            <div className="glass-panel sticky top-2 z-20 border-emerald-400/40 bg-emerald-50/80 px-3 py-2 shadow-lg sm:px-4">
              <div className="flex items-center justify-between gap-2 sm:gap-4">
                <div className="flex items-center gap-2 min-w-0 flex-1 sm:gap-3">
                  <span className="text-emerald-600 shrink-0">✅</span>
                  <a
                    href={publicURL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="truncate text-xs font-medium text-emerald-800 hover:underline sm:text-sm"
                    onClick={(event) => event.stopPropagation()}
                  >
                    {publicURL}
                  </a>
                </div>
                <div className="flex shrink-0 items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 rounded-full bg-white/70 p-0 text-emerald-700 hover:bg-white"
                    onClick={(event) => handleCopyLink(event, publicURL)}
                    title="Скопировать ссылку"
                  >
                    <Copy className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 rounded-full bg-white/70 p-0 text-emerald-700 hover:bg-white"
                    onClick={(event) => handleOpenLink(event, publicURL)}
                    title="Открыть"
                  >
                    <ExternalLink className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          )
        })()}

        <div className="flex flex-col gap-6 xl:flex-row">
          <Card className="surface-card flex h-[75vh] flex-col border-white/40 text-slate-900 xl:h-[calc(100vh-140px)] xl:w-[380px]">
            <div className="border-b border-white/50 px-4 py-3 sm:px-6 sm:py-4">
              <h2 className="text-base font-semibold sm:text-lg">AI-ассистент</h2>
              <p className="mt-1 text-xs text-slate-600 sm:text-sm">
                Опишите свою идею или изменения, и я обновлю лендинг. Можно задавать вопросы и уточнять детали — всё как в чате.
              </p>
            </div>
            <div ref={chatContainerRef} className="flex-1 space-y-4 overflow-y-auto bg-white/50 px-4 py-3 sm:px-6 sm:py-4">
              {messages.length === 0 ? (
                <p className="text-xs text-slate-500 sm:text-sm">
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
                      <div className="max-w-[85%] sm:max-w-[80%]">
                        <div className={`rounded-2xl px-4 py-2.5 text-xs shadow-sm transition sm:px-5 sm:py-3.5 sm:text-sm ${bubbleClasses}`}>
                          <p className="whitespace-pre-wrap leading-relaxed text-xs sm:text-sm">{message.content}</p>
                        </div>
                        <p className={`mt-1 text-xs text-slate-500 ${isUser ? 'text-right' : ''}`}>
                          {label} · {formatMessageTime(message.created_at)}
                        </p>
                      </div>
                    </div>
                  )
                })
              )}
              {isSending && (
                <div className="flex items-center gap-2 rounded-lg border border-blue-200 bg-blue-50 px-3 py-2 text-xs text-blue-800 sm:px-4 sm:py-3 sm:text-sm">
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
                  <span>AI генерирует схему...</span>
                </div>
              )}
            </div>
            <form className="border-t border-white/50 bg-white/70 px-4 py-3 sm:px-6 sm:py-4" onSubmit={handleSubmit}>
              <div className="space-y-2 sm:space-y-3">
                <textarea
                  value={input}
                  onChange={(event) => setInput(event.target.value)}
                  onKeyDown={handleTextareaKeyDown}
                  placeholder="Опишите, что нужно добавить или изменить..."
                  className="h-24 w-full resize-none rounded-2xl border border-slate-200/70 bg-white/80 px-3 py-2.5 text-xs text-slate-900 shadow-sm focus:border-blue-400 focus:outline-none focus:ring-2 focus:ring-blue-400/40 sm:h-28 sm:px-4 sm:py-3 sm:text-sm"
                  disabled={isSending}
                />
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <span className="text-xs text-slate-500">Enter — отправить, Shift + Enter — новая строка</span>
                  <Button
                    type="submit"
                    disabled={isSending || !input.trim()}
                    className="h-9 rounded-full bg-blue-600 px-4 text-xs font-semibold text-white shadow-sm transition hover:bg-blue-700 sm:h-10 sm:px-6 sm:text-sm"
                  >
                    {isSending ? 'Генерация...' : 'Отправить'}
                  </Button>
                </div>
              </div>
            </form>
          </Card>

          <div className="flex-1 space-y-6">
            <Card className="surface-card flex h-[75vh] flex-col border-white/40 text-slate-900 xl:h-[calc(100vh-140px)]">
              <CardHeader className="flex-shrink-0 border-b border-white/50 px-4 py-3 sm:px-6 sm:py-4">
                <CardTitle className="text-base font-semibold sm:text-lg">Предпросмотр лендинга</CardTitle>
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

            {previousSchema && schema && (
              <SchemaDiff oldSchema={previousSchema} newSchema={schema} />
            )}
          </div>
        </div>
      </div>

      {showHistory && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="w-full max-w-2xl rounded-2xl bg-white p-4 shadow-2xl sm:p-6 max-h-[90vh] overflow-y-auto">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-base font-semibold sm:text-lg">История изменений схемы</h3>
              <Button
                variant="ghost"
                onClick={() => setShowHistory(false)}
                className="h-8 w-8 shrink-0 rounded-full p-0"
              >
                ×
              </Button>
            </div>
            <SchemaHistory projectId={projectId} onRevert={handleRevert} />
          </div>
        </div>
      )}

      {confirmAction && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-4 shadow-2xl sm:p-6">
            <div className="mb-4">
              <h3 className="text-base font-semibold sm:text-lg">
                {confirmAction === 'publish' ? 'Опубликовать лендинг?' : 'Снять лендинг с публикации?'}
              </h3>
              <p className="mt-2 text-xs text-gray-600 sm:text-sm">
                {confirmAction === 'publish'
                  ? 'После публикации лендинг станет доступен по публичной ссылке. Продолжить?'
                  : 'Публичная ссылка перестанет работать, но проект останется в списке. Продолжить?'}
              </p>
            </div>
            <div className="flex flex-col gap-2 sm:flex-row sm:justify-end sm:gap-3">
              <Button
                variant="outline"
                onClick={closeConfirmDialog}
                disabled={confirmAction === 'publish' ? isPublishing : isUnpublishing}
                className="h-9 text-xs sm:h-10 sm:text-sm"
              >
                Отмена
              </Button>
              <Button
                variant={confirmAction === 'publish' ? 'default' : 'destructive'}
                onClick={confirmActionHandler}
                disabled={confirmAction === 'publish' ? isPublishing : isUnpublishing}
                className="h-9 text-xs sm:h-10 sm:text-sm"
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

