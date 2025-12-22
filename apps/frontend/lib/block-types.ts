// Shared block types - must match backend BlockType constants
export const BLOCK_TYPES = [
  'hero',
  'features',
  'pricing',
  'testimonials',
  'faq',
  'cta',
  'gallery',
  'about',
  'contact',
] as const

export type BlockType = typeof BLOCK_TYPES[number]

export function isValidBlockType(type: string): type is BlockType {
  return BLOCK_TYPES.includes(type as BlockType)
}

