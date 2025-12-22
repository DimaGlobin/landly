'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface SchemaVersion {
  id: string
  project_id: string
  created_at: string
  created_by: string
  source: 'chat' | 'generate' | 'revert' | 'system'
  tokens_used: number
}

interface SchemaHistoryProps {
  projectId: string
  onRevert?: () => void
}

const sourceLabels: Record<string, string> = {
  chat: 'Чат',
  generate: 'Генерация',
  revert: 'Откат',
  system: 'Система',
}

export function SchemaHistory({ projectId, onRevert }: SchemaHistoryProps) {
  const [versions, setVersions] = useState<SchemaVersion[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [revertingId, setRevertingId] = useState<string | null>(null)

  useEffect(() => {
    loadVersions()
  }, [projectId])

  const loadVersions = async () => {
    try {
      setIsLoading(true)
      const data = await api.getSchemaVersions(projectId, 10)
      setVersions(data.versions || [])
    } catch (error) {
      console.error('Failed to load schema versions', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleRevert = async (versionId: string) => {
    try {
      setRevertingId(versionId)
      await api.revertSchema(projectId, versionId)
      if (onRevert) {
        onRevert()
      }
      await loadVersions()
      alert('Схема успешно откачена')
    } catch (error: any) {
      console.error('Failed to revert schema', error)
      alert('Ошибка отката: ' + (error.response?.data?.error || error.message))
    } finally {
      setRevertingId(null)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <Card className="surface-card border-white/40 text-slate-900">
      <CardHeader className="border-b border-white/50">
        <CardTitle className="text-lg font-semibold">История изменений</CardTitle>
        <CardDescription className="text-sm text-slate-600">Последние 10 версий схемы</CardDescription>
      </CardHeader>
      <CardContent className="max-h-[400px] space-y-3 overflow-y-auto py-4">
        {isLoading ? (
          <div className="text-center text-sm text-slate-500">Загрузка...</div>
        ) : versions.length === 0 ? (
          <div className="text-center text-sm text-slate-500">История пуста</div>
        ) : (
          versions.map((version) => (
            <div
              key={version.id}
              className="flex items-center justify-between rounded-lg border border-slate-200 bg-white/50 px-4 py-3"
            >
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-slate-600">
                    {sourceLabels[version.source] || version.source}
                  </span>
                  <span className="text-xs text-slate-400">•</span>
                  <span className="text-xs text-slate-500">{formatDate(version.created_at)}</span>
                </div>
                <div className="mt-1 text-xs text-slate-400">
                  ID: {version.id.substring(0, 8)}...
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleRevert(version.id)}
                disabled={revertingId === version.id}
                className="h-8 rounded-full border border-slate-200 bg-white/80 px-3 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-white"
              >
                {revertingId === version.id ? 'Откат...' : 'Откатить'}
              </Button>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  )
}

