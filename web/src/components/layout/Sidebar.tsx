import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

const navItems = [
  { to: '/', label: 'Secrets', end: true },
  { to: '/hosts', label: 'Hosts', end: false },
  { to: '/apikeys', label: 'API Keys', end: false },
  { to: '/audit', label: 'Audit Log', end: false },
  { to: '/settings', label: 'Account Settings', end: false },
]

export function Sidebar() {
  const { me, logout } = useAuth()
  const [open, setOpen] = useState(false)

  return (
    <>
      {/* Mobile top bar: the sidebar below is an off-canvas drawer under the md breakpoint. */}
      <div className="flex shrink-0 items-center justify-between border-b border-border bg-surface px-4 py-3 md:hidden">
        <h1 className="text-sm font-bold text-text-bright">Warden MCP</h1>
        <button
          onClick={() => setOpen(true)}
          className="rounded-md border border-border px-2.5 py-1.5 text-sm text-text hover:bg-surface-hover"
          aria-label="Open menu"
        >
          ☰
        </button>
      </div>

      {open && (
        <div
          className="fixed inset-0 z-40 bg-black/60 md:hidden"
          onClick={() => setOpen(false)}
          aria-hidden="true"
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-64 shrink-0 -translate-x-full flex-col border-r border-border bg-surface transition-transform duration-200 md:static md:z-auto md:w-56 md:translate-x-0 ${
          open ? 'translate-x-0' : ''
        }`}
      >
        <div className="border-b border-border px-4 py-4">
          <h1 className="text-sm font-bold text-text-bright">Warden MCP</h1>
          <p className="text-xs text-text-muted">Credentials Vault</p>
        </div>

        <nav className="flex-1 space-y-1 px-2 py-3">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              onClick={() => setOpen(false)}
              className={({ isActive }) =>
                `block rounded-md px-3 py-2 text-sm font-medium ${
                  isActive
                    ? 'bg-accent/15 text-accent'
                    : 'text-text hover:bg-surface-hover hover:text-text-bright'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-border px-4 py-3">
          <p className="truncate text-xs text-text-muted">{me?.username}</p>
          <button
            onClick={logout}
            className="mt-1 text-xs font-medium text-text-muted hover:text-danger"
          >
            Sign out
          </button>
        </div>
      </aside>
    </>
  )
}
