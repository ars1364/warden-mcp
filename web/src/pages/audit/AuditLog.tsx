import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { AuditEntry } from '../../types/api'
import { ErrorBanner } from '../../components/common/ErrorBanner'

const LIMIT_OPTIONS = [25, 50, 100, 200]

export function AuditLog() {
  const [entries, setEntries] = useState<AuditEntry[]>([])
  const [limit, setLimit] = useState(100)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .get<AuditEntry[]>(`/api/audit?limit=${limit}`)
      .then((result) => {
        if (!cancelled) setEntries(result)
      })
      .catch((err) => {
        if (!cancelled) setError(errorMessage(err))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [limit])

  return (
    <div>
      <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-xl font-semibold text-text-bright">Audit Log</h1>
        <label className="flex items-center gap-2 text-sm text-text-muted">
          Show
          <select
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            className="rounded-md border border-border bg-bg px-2 py-1 text-text outline-none"
          >
            {LIMIT_OPTIONS.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
          entries
        </label>
      </div>

      <ErrorBanner message={error} />

      <div className="mt-3 overflow-x-auto rounded-lg border border-border bg-surface">
        {loading ? (
          <div className="p-6 text-sm text-text-muted">Loading…</div>
        ) : entries.length === 0 ? (
          <div className="p-6 text-sm text-text-muted">No audit entries.</div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border text-xs uppercase tracking-wide text-text-muted">
                <th className="px-4 py-2 font-medium">Time</th>
                <th className="px-4 py-2 font-medium">Actor</th>
                <th className="px-4 py-2 font-medium">Action</th>
                <th className="px-4 py-2 font-medium">Secret</th>
                <th className="px-4 py-2 font-medium">IP</th>
                <th className="px-4 py-2 font-medium">Detail</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry) => (
                <tr key={entry.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 whitespace-nowrap text-xs text-text-muted">
                    {new Date(entry.ts).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-text">
                    <span className="text-text-bright">{entry.actor_label}</span>
                    <span className="ml-1 text-xs text-text-muted">({entry.actor_type})</span>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-accent">{entry.action}</td>
                  <td className="px-4 py-3 text-text">{entry.secret_name || '—'}</td>
                  <td className="px-4 py-3 font-mono text-xs text-text-muted">{entry.ip}</td>
                  <td className="px-4 py-3 text-xs text-text-muted">{entry.detail}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
