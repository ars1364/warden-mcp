import { useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { TotpCodeResponse } from '../../types/api'
import { Button } from '../../components/common/Button'
import { CopyButton } from '../../components/common/CopyButton'

// Fetches a fresh 6-digit code on demand rather than polling — the seed's server-side
// countdown is short-lived and a stale cached code is worse than an explicit refetch.
export function TotpCodeButton({ secretId }: { secretId: string }) {
  const [code, setCode] = useState<TotpCodeResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleFetch() {
    setLoading(true)
    setError(null)
    try {
      const res = await api.get<TotpCodeResponse>(`/api/secrets/${secretId}/totp-code`)
      setCode(res)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="mb-1 flex items-center gap-2">
      <Button type="button" variant="secondary" onClick={handleFetch} disabled={loading}>
        {loading ? 'Fetching…' : 'Get code'}
      </Button>
      {code && (
        <>
          <code className="font-mono text-sm text-text-bright">{code.code}</code>
          <span className="text-xs text-text-muted">{code.seconds_remaining}s</span>
          <CopyButton value={code.code} />
        </>
      )}
      {error && <span className="text-xs text-danger">{error}</span>}
    </div>
  )
}
