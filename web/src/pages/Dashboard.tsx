import { Outlet } from 'react-router-dom'
import { Sidebar } from '../components/layout/Sidebar'

export function Dashboard() {
  return (
    <div className="flex min-h-screen bg-bg">
      <Sidebar />
      <main className="flex-1 overflow-y-auto p-8">
        <Outlet />
      </main>
    </div>
  )
}
