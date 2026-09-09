import { Outlet } from 'react-router-dom'
import { Sidebar } from '../components/layout/Sidebar'

export function Dashboard() {
  return (
    <div className="flex min-h-screen flex-col bg-bg md:flex-row">
      <Sidebar />
      <main className="min-w-0 flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
        <Outlet />
      </main>
    </div>
  )
}
