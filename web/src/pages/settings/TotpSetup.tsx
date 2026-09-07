import { useState, type FormEvent } from 'react'
import QRCode from 'qrcode'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import { useAuth } from '../../context/AuthContext'
import type { TotpSetupResponse } from '../../types/api'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { FormField, TextInput } from '../../components/common/FormField'
import { CopyButton } from '../../components/common/CopyButton'

export function TotpSetup() {
  const { me, refreshMe } = useAuth()
  const [setupData, setSetupData] = useState<TotpSetupResponse | null>(null)
  const [qrDataUrl, setQrDataUrl] = useState<string | null>(null)
  const [code, setCode] = useState('')
  const [disablePassword, setDisablePassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleStartSetup() {
    setError(null)
    try {
      const result = await api.post<TotpSetupResponse>('/api/totp/setup')
      setSetupData(result)
      const dataUrl = await QRCode.toDataURL(result.otpauth_url)
      setQrDataUrl(dataUrl)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function handleEnable(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await api.post('/api/totp/enable', { code })
      setSetupData(null)
      setQrDataUrl(null)
      setCode('')
      await refreshMe()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDisable(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await api.post('/api/totp/disable', { password: disablePassword })
      setDisablePassword('')
      await refreshMe()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  if (me?.totp_enabled) {
    return (
      <div className="max-w-sm space-y-4">
        <p className="text-sm text-success">Two-factor authentication is enabled.</p>
        <form onSubmit={handleDisable} className="space-y-4">
          <FormField label="Confirm password to disable">
            <TextInput
              type="password"
              value={disablePassword}
              onChange={(e) => setDisablePassword(e.target.value)}
              required
            />
          </FormField>
          <ErrorBanner message={error} />
          <Button type="submit" variant="danger" disabled={submitting}>
            {submitting ? 'Disabling…' : 'Disable 2FA'}
          </Button>
        </form>
      </div>
    )
  }

  if (!setupData) {
    return (
      <div className="max-w-sm space-y-4">
        <p className="text-sm text-text-muted">
          Two-factor authentication is not enabled on this account.
        </p>
        <ErrorBanner message={error} />
        <Button onClick={handleStartSetup}>Set up two-factor authentication</Button>
      </div>
    )
  }

  return (
    <div className="max-w-sm space-y-4">
      <p className="text-sm text-text">
        Scan this QR code with your authenticator app, then enter the 6-digit code to confirm.
      </p>

      {qrDataUrl && (
        <img src={qrDataUrl} alt="TOTP QR code" className="rounded-md border border-border" />
      )}

      <div className="flex items-center gap-2 rounded-md border border-border bg-bg px-3 py-2">
        <code className="flex-1 overflow-x-auto whitespace-nowrap font-mono text-xs text-text">
          {setupData.secret}
        </code>
        <CopyButton value={setupData.secret} />
      </div>

      <form onSubmit={handleEnable} className="space-y-4">
        <FormField label="6-digit code">
          <TextInput
            value={code}
            onChange={(e) => setCode(e.target.value)}
            inputMode="numeric"
            pattern="[0-9]{6}"
            maxLength={6}
            required
            className="font-mono tracking-widest"
          />
        </FormField>
        <ErrorBanner message={error} />
        <div className="flex gap-2">
          <Button type="submit" disabled={submitting}>
            {submitting ? 'Confirming…' : 'Confirm and enable'}
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              setSetupData(null)
              setQrDataUrl(null)
            }}
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  )
}
