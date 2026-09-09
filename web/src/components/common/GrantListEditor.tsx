import { Button } from './Button'

export interface GrantRow {
  name: string
  canWrite: boolean
}

interface GrantListEditorProps {
  label: string
  options: string[]
  grants: GrantRow[]
  onChange: (grants: GrantRow[]) => void
}

export function GrantListEditor({ label, options, grants, onChange }: GrantListEditorProps) {
  const usedElsewhere = (except: number) => new Set(grants.filter((_, i) => i !== except).map((g) => g.name))

  function update(index: number, patch: Partial<GrantRow>) {
    onChange(grants.map((g, i) => (i === index ? { ...g, ...patch } : g)))
  }

  function remove(index: number) {
    onChange(grants.filter((_, i) => i !== index))
  }

  function add() {
    const taken = new Set(grants.map((g) => g.name))
    const next = options.find((o) => !taken.has(o)) ?? ''
    onChange([...grants, { name: next, canWrite: false }])
  }

  return (
    <div className="space-y-2">
      {grants.length === 0 && <p className="text-xs text-text-muted">Unrestricted — this key can access every {label.toLowerCase()}.</p>}
      {grants.map((g, i) => {
        const taken = usedElsewhere(i)
        const rowOptions = options.filter((o) => !taken.has(o) || o === g.name)
        return (
          <div key={i} className="flex items-center gap-2">
            <select
              value={g.name}
              onChange={(e) => update(i, { name: e.target.value })}
              className="flex-1 rounded-md border border-border bg-bg px-3 py-1.5 text-sm text-text outline-none focus:border-accent"
            >
              {rowOptions.length === 0 && <option value="">— no {label.toLowerCase()}s available —</option>}
              {rowOptions.map((o) => (
                <option key={o} value={o}>{o}</option>
              ))}
            </select>
            <label className="flex shrink-0 items-center gap-1.5 text-xs text-text">
              <input type="checkbox" checked={g.canWrite} onChange={(e) => update(i, { canWrite: e.target.checked })} />
              write
            </label>
            <Button type="button" variant="ghost" onClick={() => remove(i)} aria-label={`Remove ${label} grant`}>
              ×
            </Button>
          </div>
        )
      })}
      <Button type="button" variant="secondary" onClick={add} disabled={grants.length >= options.length}>
        Add {label.toLowerCase()} grant
      </Button>
    </div>
  )
}
