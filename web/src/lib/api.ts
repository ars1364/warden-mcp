import { ApiError } from '../types/api'

const TOKEN_KEY = 'warden_token'

let inMemoryToken: string | null = localStorage.getItem(TOKEN_KEY)
let onUnauthorized: (() => void) | null = null

export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler
}

export function getToken(): string | null {
  return inMemoryToken
}

export function setToken(token: string | null) {
  inMemoryToken = token
  if (token) {
    localStorage.setItem(TOKEN_KEY, token)
  } else {
    localStorage.removeItem(TOKEN_KEY)
  }
}

interface RequestOptions {
  method?: string
  body?: unknown
  skipAuth?: boolean
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {}
  let body: string | undefined

  if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(opts.body)
  }

  if (!opts.skipAuth && inMemoryToken) {
    headers['Authorization'] = `Bearer ${inMemoryToken}`
  }

  const res = await fetch(path, {
    method: opts.method ?? 'GET',
    headers,
    body,
  })

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  const data: unknown = text ? JSON.parse(text) : undefined

  if (!res.ok) {
    const errBody = data as { error?: { code?: string; message?: string } } | undefined
    const code = errBody?.error?.code ?? 'UNKNOWN_ERROR'
    const message = errBody?.error?.message ?? res.statusText
    if (res.status === 401 && !opts.skipAuth) {
      setToken(null)
      onUnauthorized?.()
    }
    throw new ApiError(res.status, code, message)
  }

  return data as T
}

export const api = {
  get: <T>(path: string, skipAuth = false) => request<T>(path, { method: 'GET', skipAuth }),
  post: <T>(path: string, body?: unknown, skipAuth = false) =>
    request<T>(path, { method: 'POST', body, skipAuth }),
  put: <T>(path: string, body?: unknown, skipAuth = false) =>
    request<T>(path, { method: 'PUT', body, skipAuth }),
  del: <T>(path: string, skipAuth = false) => request<T>(path, { method: 'DELETE', skipAuth }),
}
