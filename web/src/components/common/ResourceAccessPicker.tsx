import { useMemo, useState } from 'react'
import { Button } from './Button'
import { TextInput } from './FormField'

export interface GrantState {
  [name: string]: { canRead: boolean; canWrite: boolean }
}

interface ResourceAccessPickerProps {
  secretNames: string[]
  hostNames: string[]
  secretState: GrantState
  hostState: GrantState
  onChangeSecretState: (next: GrantState) => void
  onChangeHostState: (next: GrantState) => void
}

type Tab = 'secret' | 'host'

function countChecked(state: GrantState) {
  return Object.values(state).filter((s) => s.canRead || s.canWrite).length
}

export function ResourceAccessPicker({
  secretNames, hostNames, secretState, hostState, onChangeSecretState, onChangeHostState,
}: ResourceAccessPickerProps) {
  const [tab, setTab] = useState<Tab>('secret')
  const [search, setSearch] = useState('')

  const names = tab === 'secret' ? secretNames : hostNames
  const state = tab === 'secret' ? secretState : hostState
  const setState = tab === 'secret' ? onChangeSecretState : onChangeHostState

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return q ? names.filter((n) => n.toLowerCase().includes(q)) : names
  }, [names, search])

  function setRow(name: string, patch: Partial<{ canRead: boolean; canWrite: boolean }>) {
    const current = state[name] ?? { canRead: false, canWrite: false }
    setState({ ...state, [name]: { ...current, ...patch } })
  }

  function bulkApply(patch: { canRead: boolean; canWrite: boolean }) {
    const next = { ...state }
    for (const name of filtered) next[name] = { ...patch }
    setState(next)
  }

  return (
    <div className="space-y-3">
      <div className="flex gap-1 border-b border-border">
        <button
          type="button"
          onClick={() => setTab('secret')}
          className={`px-3 py-2 text-sm font-medium ${tab === 'secret' ? 'border-b-2 border-accent text-accent' : 'text-text-muted hover:text-text'}`}
        >
          Secrets ({countChecked(secretState)}/{secretNames.length})
        </button>
        <button
          type="button"
          onClick={() => setTab('host')}
          className={`px-3 py-2 text-sm font-medium ${tab === 'host' ? 'border-b-2 border-accent text-accent' : 'text-text-muted hover:text-text'}`}
        >
          Hosts ({countChecked(hostState)}/{hostNames.length})
        </button>
      </div>

      <TextInput
        placeholder={`Filter ${tab === 'secret' ? 'secrets' : 'hosts'}…`}
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />

      <div className="flex flex-wrap gap-2">
        <Button type="button" variant="secondary" onClick={() => bulkApply({ canRead: true, canWrite: true })}>
          Select all
        </Button>
        <Button type="button" variant="secondary" onClick={() => bulkApply({ canRead: false, canWrite: false })}>
          Deselect all
        </Button>
        <Button type="button" variant="ghost" onClick={() => bulkApply({ canRead: true, canWrite: false })}>
          Read only
        </Button>
        <Button type="button" variant="ghost" onClick={() => bulkApply({ canRead: false, canWrite: true })}>
          Write only
        </Button>
      </div>

      <div className="max-h-64 overflow-y-auto rounded-md border border-border">
        {filtered.length === 0 ? (
          <div className="p-4 text-center text-sm text-text-muted">No matches.</div>
        ) : (
          filtered.map((name) => {
            const row = state[name] ?? { canRead: false, canWrite: false }
            return (
              <div
                key={name}
                className="flex items-center gap-3 border-b border-border px-3 py-2 text-sm last:border-0 odd:bg-surface-hover/40"
              >
                <span className="min-w-0 flex-1 truncate font-mono text-xs text-text" title={name}>
                  {name}
                </span>
                <label className="flex shrink-0 items-center gap-1.5 text-xs text-text-muted">
                  <input type="checkbox" checked={row.canRead} onChange={(e) => setRow(name, { canRead: e.target.checked })} />
                  read
                </label>
                <label className="flex shrink-0 items-center gap-1.5 text-xs text-text-muted">
                  <input type="checkbox" checked={row.canWrite} onChange={(e) => setRow(name, { canWrite: e.target.checked })} />
                  write
                </label>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}

// initGrantState builds the starting per-name checked state: fully checked when the
// resource type is currently unrestricted (no saved grants — or a brand-new key with
// nothing saved yet), so "do nothing" == "stays unrestricted" when computing what to save.
// When restricted, only the names present in savedGrants start checked, matching what's
// actually allowed today; everything else starts unchecked.
export function initGrantState(allNames: string[], savedGrants: { resource_name: string; can_read: boolean; can_write: boolean }[]): GrantState {
  const unrestricted = savedGrants.length === 0
  const saved = new Map(savedGrants.map((g) => [g.resource_name, g]))
  const state: GrantState = {}
  for (const name of allNames) {
    const g = saved.get(name)
    if (g) {
      state[name] = { canRead: g.can_read, canWrite: g.can_write }
    } else {
      state[name] = { canRead: unrestricted, canWrite: unrestricted }
    }
  }
  return state
}

// grantStateToPayload converts checked state back into what to send the API: if every name
// is fully checked (read+write), it's still unrestricted — send an empty list so this type
// keeps auto-including resources created later, rather than freezing today's snapshot.
// Otherwise send one entry per name that has read and/or write checked; unchecked names are
// simply omitted (denied).
export function grantStateToPayload(allNames: string[], state: GrantState): { resource_name: string; can_read: boolean; can_write: boolean }[] {
  const allFullyChecked = allNames.every((n) => state[n]?.canRead && state[n]?.canWrite)
  if (allFullyChecked) return []
  return allNames
    .filter((n) => state[n]?.canRead || state[n]?.canWrite)
    .map((n) => ({ resource_name: n, can_read: !!state[n]?.canRead, can_write: !!state[n]?.canWrite }))
}
