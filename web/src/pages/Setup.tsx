import { useState, type FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../lib/api'
import { errorMessage } from '../lib/errors'
import { ApiError } from '../types/api'
import { Button } from '../components/common/Button'
import { ErrorBanner } from '../components/common/ErrorBanner'
import { FormField, TextInput } from '../components/common/FormField'

export function Setup() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    if (password !== confirm) {
      setError('Passwords do not match')
      return
    }

    setSubmitting(true)
    try {
      await api.post('/api/setup', { username, password }, true)
      navigate('/login', { replace: true })
    } catch (err) {
      if (err instanceof ApiError && err.status === 403 && err.code === 'ALREADY_INITIALIZED') {
        navigate('/login', { replace: true })
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
        <p className="mb-5 text-sm text-text-muted">Create the initial admin account.</p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <FormField label="Username">
            <TextInput
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              required
            />
          </FormField>
          <FormField label="Password">
            <TextInput
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
              required
            />
          </FormField>
          <FormField label="Confirm password">
            <TextInput
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              autoComplete="new-password"
              required
            />
          </FormField>

          <ErrorBanner message={error} />

          <Button type="submit" disabled={submitting} className="w-full justify-center">
            {submitting ? 'Creating…' : 'Create admin account'}
          </Button>
        </form>

        <p className="mt-4 text-center text-xs text-text-muted">
          Already set up? <Link to="/login" className="text-accent hover:underline">Sign in</Link>
        </p>
      </div>
    </div>
  )
}
