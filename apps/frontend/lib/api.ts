import axios, { AxiosInstance } from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Добавляем токен к запросам
    this.client.interceptors.request.use((config) => {
      const token = this.getToken()
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    })
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
}

export const api = new ApiClient()

