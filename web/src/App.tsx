import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { ProtectedRoute } from './components/layout/ProtectedRoute'
import { Setup } from './pages/Setup'
import { Login } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { SecretsList } from './pages/secrets/SecretsList'
import { ApiKeysList } from './pages/apikeys/ApiKeysList'
import { AuditLog } from './pages/audit/AuditLog'
import { AccountSettings } from './pages/settings/AccountSettings'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/setup" element={<Setup />} />
          <Route path="/login" element={<Login />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          >
            <Route index element={<SecretsList />} />
            <Route path="apikeys" element={<ApiKeysList />} />
            <Route path="audit" element={<AuditLog />} />
            <Route path="settings" element={<AccountSettings />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
