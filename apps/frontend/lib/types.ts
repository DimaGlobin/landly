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

