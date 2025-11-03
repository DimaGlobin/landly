// Project types
export interface Project {
  id: string
  user_id: string
  name: string
  niche: string
  status: 'draft' | 'generated' | 'published'
  created_at: string
  updated_at: string
  publish?: ProjectPublishInfo
}

export interface ProjectPublishInfo {
  status: 'draft' | 'published' | 'failed'
  public_url: string
  subdomain: string
  last_published_at?: string
}

export interface ChatSession {
  id: string
  project_id: string
  status: 'pending' | 'completed' | 'failed'
  schema_json?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export type ChatMessageRole = 'user' | 'assistant' | 'system'

export interface ChatMessage {
  id: string
  role: ChatMessageRole
  content: string
  metadata?: string
  tokens_used: number
  created_at: string
}

export interface ChatHistoryResponse {
  session: ChatSession
  messages: ChatMessage[]
}

// Landing schema types
export interface LandingSchema {
  version: string
  pages: Page[]
  theme?: Theme
  payment?: Payment
}

export interface Page {
  path: string
  title: string
  description?: string
  blocks: Block[]
}

export interface Block {
  type: BlockType
  order: number
  props: Record<string, any>
}

export type BlockType =
  | 'hero'
  | 'features'
  | 'pricing'
  | 'testimonials'
  | 'faq'
  | 'cta'
  | 'gallery'
  | 'about'
  | 'contact'

export interface Theme {
  palette?: {
    primary?: string
    secondary?: string
    accent?: string
    background?: string
    text?: string
  }
  font?: string
  borderRadius?: string
}

export interface Payment {
  url: string
  buttonText?: string
}

// Analytics types
export interface AnalyticsStats {
  project_id: string
  total_pageviews: number
  total_cta_clicks: number
  total_pay_clicks: number
  unique_visitors: number
}

