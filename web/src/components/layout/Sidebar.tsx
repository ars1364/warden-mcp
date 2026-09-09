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

  return (
    <aside className="flex w-56 shrink-0 flex-col border-r border-border bg-surface">
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
  )
}
