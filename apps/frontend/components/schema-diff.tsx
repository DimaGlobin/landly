'use client'

import { LandingSchema } from '@/lib/types'
import { BLOCK_TYPES } from '@/lib/block-types'

interface SchemaDiffProps {
  oldSchema: LandingSchema | null
  newSchema: LandingSchema | null
}

export function SchemaDiff({ oldSchema, newSchema }: SchemaDiffProps) {
  if (!oldSchema || !newSchema) {
    return null
  }

  const oldBlocks = oldSchema.pages?.[0]?.blocks || []
  const newBlocks = newSchema.pages?.[0]?.blocks || []

  const oldBlockTypes = new Set(oldBlocks.map((b) => b.type))
  const newBlockTypes = new Set(newBlocks.map((b) => b.type))

  const added = Array.from(newBlockTypes).filter((t) => !oldBlockTypes.has(t))
  const removed = Array.from(oldBlockTypes).filter((t) => !newBlockTypes.has(t))
  const changed: string[] = []

  // Простая проверка изменений: сравниваем количество блоков каждого типа
  for (const blockType of BLOCK_TYPES) {
    const oldCount = oldBlocks.filter((b) => b.type === blockType).length
    const newCount = newBlocks.filter((b) => b.type === blockType).length
    if (oldCount !== newCount && !added.includes(blockType) && !removed.includes(blockType)) {
      changed.push(`${blockType} (${oldCount} → ${newCount})`)
    }
  }

  const totalAdded = newBlocks.length - oldBlocks.length
  const totalRemoved = oldBlocks.length - newBlocks.length

  if (added.length === 0 && removed.length === 0 && changed.length === 0 && totalAdded === 0 && totalRemoved === 0) {
    return (
      <div className="rounded-lg border border-blue-200 bg-blue-50 px-3 py-2 text-xs text-blue-800 sm:px-4 sm:text-sm">
        Схема обновлена (изменения незначительны)
      </div>
    )
  }

  return (
    <div className="space-y-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2.5 text-xs sm:px-4 sm:py-3 sm:text-sm">
      <div className="font-semibold text-emerald-900">Что изменилось:</div>
      <div className="space-y-1 text-emerald-800">
        {totalAdded > 0 && (
          <div>
            ✅ Добавлено блоков: <span className="font-medium">{totalAdded}</span>
          </div>
        )}
        {totalRemoved > 0 && (
          <div>
            ❌ Удалено блоков: <span className="font-medium">{totalRemoved}</span>
          </div>
        )}
        {added.length > 0 && (
          <div className="break-words">
            ➕ Новые типы блоков: <span className="font-medium">{added.join(', ')}</span>
          </div>
        )}
        {removed.length > 0 && (
          <div className="break-words">
            ➖ Удалённые типы: <span className="font-medium">{removed.join(', ')}</span>
          </div>
        )}
        {changed.length > 0 && (
          <div className="break-words">
            🔄 Изменено количество: <span className="font-medium">{changed.join(', ')}</span>
          </div>
        )}
      </div>
    </div>
  )
}

