export interface ApiErrorBody {
  error: {
    code: string
    message: string
  }
}

export class ApiError extends Error {
  code: string
  status: number

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export interface Me {
  id: string
  username: string
  totp_enabled: boolean
}

export interface LoginResponse {
  token: string
}

export interface TotpSetupResponse {
  secret: string
  otpauth_url: string
}

export type SecretType = 'opaque' | 'structured' | 'totp' | 'reference'

export interface SecretMeta {
  id: string
  name: string
  description: string
  tags: string[]
  type: SecretType
  created_by: string
  created_at: string
  updated_at: string
}

export interface SecretField {
  key: string
  value: string
}

// GET /api/secrets/:id — value is populated for opaque secrets, fields for the rest.
export interface SecretWithValue extends SecretMeta {
  value?: string
  fields?: SecretField[]
}

// POST/PUT body: value is used for type "opaque", fields for the other three.
export interface SecretInput {
  name: string
  description: string
  tags: string[]
  type: SecretType
  value: string
  fields: SecretField[]
}

export interface SecretUpdateInput {
  description: string
  tags: string[]
  type: SecretType
  value: string
  fields: SecretField[]
}

export interface TotpCodeResponse {
  code: string
  seconds_remaining: number
}

export type ApiKeyScope = 'read' | 'write'

export interface ApiKeyMeta {
  id: string
  name: string
  scopes: ApiKeyScope[]
  created_by: string
  created_at: string
  last_used_at?: string
  last_used_ip?: string
  revoked_at?: string
}

export interface ApiKeyCreated extends ApiKeyMeta {
  key: string
}

export interface AuditEntry {
  id: string
  ts: string
  actor_type: string
  actor_id: string
  actor_label: string
  action: string
  secret_name: string
  ip: string
  detail: string
}
