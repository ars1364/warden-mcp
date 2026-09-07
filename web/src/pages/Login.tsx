import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { errorMessage } from '../lib/errors'
import { useAuth } from '../context/AuthContext'
import { ApiError, type LoginResponse } from '../types/api'
import { Button } from '../components/common/Button'
import { ErrorBanner } from '../components/common/ErrorBanner'
import { FormField, TextInput } from '../components/common/FormField'

export function Login() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [needsTotp, setNeedsTotp] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const body: { username: string; password: string; totp_code?: string } = {
        username,
        password,
      }
      if (needsTotp) {
        body.totp_code = totpCode
      }
      const result = await api.post<LoginResponse>('/api/login', body, true)
      await login(result.token)
      navigate('/', { replace: true })
    } catch (err) {
      if (err instanceof ApiError && err.status === 400 && err.code === 'TOTP_REQUIRED') {
        setNeedsTotp(true)
        setError(null)
        return
      }
      if (err instanceof ApiError && err.status === 401 && err.code === 'INVALID_CREDENTIALS') {
        setError('Invalid username or password')
        return
      }
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-bg px-4">
      <div className="w-full max-w-sm rounded-lg border border-border bg-surface p-6 shadow-xl">
        <h1 className="mb-1 text-lg font-semibold text-text-bright">Warden MCP</h1>
        <p className="mb-5 text-sm text-text-muted">Sign in to your vault.</p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <FormField label="Username">
            <TextInput
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              disabled={needsTotp}
              required
            />
          </FormField>
          <FormField label="Password">
            <TextInput
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              disabled={needsTotp}
              required
            />
          </FormField>

          {needsTotp && (
            <FormField label="6-digit authentication code">
              <TextInput
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                autoFocus
                required
                className="font-mono tracking-widest"
              />
            </FormField>
          )}

          <ErrorBanner message={error} />

          <Button type="submit" disabled={submitting} className="w-full justify-center">
            {submitting ? 'Signing in…' : needsTotp ? 'Verify' : 'Sign in'}
          </Button>
        </form>
      </div>
    </div>
  )
}
