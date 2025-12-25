import axios, { AxiosInstance, AxiosError } from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// Error codes that match backend domain.Error codes
export type ApiErrorCode =
  | 'UNAUTHORIZED'
  | 'INVALID_TOKEN'
  | 'TOKEN_EXPIRED'
  | 'INVALID_CREDENTIALS'
  | 'USER_ALREADY_EXISTS'
  | 'USER_NOT_FOUND'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'PROJECT_NOT_FOUND'
  | 'ALREADY_EXISTS'
  | 'VALIDATION_ERROR'
  | 'INVALID_INPUT'
  | 'BAD_REQUEST'
  | 'GENERATION_FAILED'
  | 'RENDER_FAILED'
  | 'PUBLISH_FAILED'
  | 'INTERNAL_ERROR'
  | 'NETWORK_ERROR'
  | 'UNKNOWN'

// API error response format from backend
export interface ApiErrorResponse {
  error: {
    code: ApiErrorCode
    message: string
  }
}

// Custom error class for API errors
export class ApiError extends Error {
  public readonly code: ApiErrorCode
  public readonly originalMessage: string

  constructor(code: ApiErrorCode, message?: string) {
    super(message || code)
    this.code = code
    this.originalMessage = message || code
    this.name = 'ApiError'
  }

  // Check if error is authentication-related
  isAuthError(): boolean {
    return ['UNAUTHORIZED', 'INVALID_TOKEN', 'TOKEN_EXPIRED', 'INVALID_CREDENTIALS'].includes(this.code)
  }

  // Check if user should be redirected to login
  shouldRedirectToLogin(): boolean {
    return ['UNAUTHORIZED', 'INVALID_TOKEN', 'TOKEN_EXPIRED'].includes(this.code)
  }
}

// Parse error from axios response
function parseApiError(error: AxiosError<ApiErrorResponse>): ApiError {
  // Network error (no response)
  if (!error.response) {
    return new ApiError('NETWORK_ERROR', 'No connection to server')
  }

  // Try to parse structured error response
  const data = error.response.data
  if (data?.error?.code) {
    return new ApiError(data.error.code, data.error.message)
  }

  // Fallback based on HTTP status
  const status = error.response.status
  switch (status) {
    case 401:
      return new ApiError('UNAUTHORIZED', 'Authorization required')
    case 403:
      return new ApiError('FORBIDDEN', 'Access denied')
    case 404:
      return new ApiError('NOT_FOUND', 'Resource not found')
    case 409:
      return new ApiError('ALREADY_EXISTS', 'Resource already exists')
    case 400:
      return new ApiError('VALIDATION_ERROR', 'Validation error')
    case 500:
    default:
      return new ApiError('INTERNAL_ERROR', 'Internal server error')
  }
}

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Add auth token to requests
    // NOTE: Using X-Landly-Authorization instead of Authorization
    // because Yandex Serverless Containers intercept Authorization header
    // and return 403 before request reaches our backend
    this.client.interceptors.request.use((config) => {
      const token = this.getToken()
      if (token) {
        config.headers['X-Landly-Authorization'] = `Bearer ${token}`
      }
      return config
    })

    // Handle errors uniformly
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiErrorResponse>) => {
        const apiError = parseApiError(error)

        // Auto logout on auth errors (except invalid credentials which is a login error)
        if (apiError.shouldRedirectToLogin()) {
          this.removeToken()
          // Redirect to login if in browser and not already on auth page
          if (typeof window !== 'undefined' && !window.location.pathname.includes('/login') && !window.location.pathname.includes('/signup')) {
            window.location.href = '/login'
          }
        }

        throw apiError
      }
    )
  }

  private getToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('access_token')
    }
    return null
  }

  private setToken(token: string): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem('access_token', token)
    }
  }

  private removeToken(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    }
  }

  // Auth
  async signUp(email: string, password: string) {
    const { data } = await this.client.post('/v1/auth/signup', { email, password })
    this.setToken(data.access_token)
    return data
  }

  async signIn(email: string, password: string) {
    const { data } = await this.client.post('/v1/auth/login', { email, password })
    this.setToken(data.access_token)
    return data
  }

  logout() {
    this.removeToken()
  }

  // Projects
  async getProjects() {
    const { data } = await this.client.get('/v1/projects')
    return data
  }

  async getProject(id: string) {
    const { data } = await this.client.get(`/v1/projects/${id}`)
    return data
  }

  async createProject(name: string, niche: string) {
    const { data } = await this.client.post('/v1/projects', { name, niche })
    return data
  }

  async deleteProject(id: string) {
    await this.client.delete(`/v1/projects/${id}`)
  }

  // Generate
  async generateLanding(projectId: string, prompt: string, paymentURL?: string) {
    const { data } = await this.client.post(`/v1/projects/${projectId}/generate`, {
      prompt,
      payment_url: paymentURL,
    })
    return data
  }

  async getPreview(projectId: string) {
    const { data } = await this.client.get(`/v1/projects/${projectId}/preview`)
    return data
  }

  async publishProject(projectId: string) {
    const { data } = await this.client.post(`/v1/projects/${projectId}/publish`)
    return data
  }

  async getChatHistory(projectId: string) {
    const { data } = await this.client.get(`/v1/projects/${projectId}/chat`)
    return data
  }

  async sendChatMessage(projectId: string, content: string) {
    const { data } = await this.client.post(`/v1/projects/${projectId}/chat`, { content })
    return data
  }

  async unpublishProject(projectId: string) {
    await this.client.delete(`/v1/projects/${projectId}/publish`)
  }

  // Analytics
  async getStats(projectId: string) {
    const { data } = await this.client.get(`/v1/analytics/${projectId}/stats`)
    return data
  }

  async trackEvent(projectId: string, eventType: string, path: string, referrer?: string) {
    await this.client.post(`/v1/analytics/${projectId}/event`, {
      event_type: eventType,
      path,
      referrer,
    })
  }

  // Schema versions
  async getSchemaVersions(projectId: string, limit?: number) {
    const params = limit ? { limit } : {}
    const { data } = await this.client.get(`/v1/projects/${projectId}/schema/versions`, { params })
    return data
  }

  async revertSchema(projectId: string, versionId: string) {
    const { data } = await this.client.post(`/v1/projects/${projectId}/schema/revert`, {
      version_id: versionId,
    })
    return data
  }
}

export const api = new ApiClient()
