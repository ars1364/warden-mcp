import { useAuth } from '../../context/AuthContext'
import { ChangePasswordForm } from './ChangePasswordForm'
import { TotpSetup } from './TotpSetup'

export function AccountSettings() {
  const { me } = useAuth()

  return (
    <div className="max-w-2xl space-y-10">
      <div>
        <h1 className="text-xl font-semibold text-text-bright">Account Settings</h1>
        <p className="text-sm text-text-muted">Signed in as {me?.username}</p>
      </div>

      <section>
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-text-muted">
          Change password
        </h2>
        <ChangePasswordForm />
      </section>

      <section>
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-text-muted">
          Two-factor authentication
        </h2>
        <TotpSetup />
      </section>
    </div>
  )
}
