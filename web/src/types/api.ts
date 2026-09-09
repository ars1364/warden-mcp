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
  expires_at?: string
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
// expires_at is a plain "YYYY-MM-DD" date, or "" for no expiry.
export interface SecretInput {
  name: string
  description: string
  tags: string[]
  type: SecretType
  value: string
  fields: SecretField[]
  expires_at: string
}

export interface SecretUpdateInput {
  description: string
  tags: string[]
  type: SecretType
  value: string
  fields: SecretField[]
  expires_at: string
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

export type HostType = 'physical' | 'vm' | 'vps' | 'container' | 'switch' | 'router' | 'other'
export type HostStatus = 'active' | 'maintenance' | 'decommissioned'

export interface HostAddress {
  label: string
  address: string
}

// GET /api/hosts (list) and /api/hosts/:id (detail) both return this shape — addresses are
// included on both since, unlike secret values, they aren't sensitive.
export interface HostMeta {
  id: string
  name: string
  host_type: HostType
  status: HostStatus
  description: string
  tags: string[]
  parent_host_id?: string
  parent_host_name?: string
  location_kind: string
  cloud_provider: string
  cloud_account: string
  physical_location: string
  ssh_port: number
  ssh_username: string
  ssh_secret_name: string
  ssh_jump_host_id?: string
  ssh_jump_host_name?: string
  addresses?: HostAddress[]
  created_by: string
  created_at: string
  updated_at: string
}

// POST/PUT body. Parent/jump hosts are referenced by name. addresses replaces the full set.
export interface HostInput {
  name: string
  host_type: HostType
  status: HostStatus
  description: string
  tags: string[]
  parent_host_name: string
  location_kind: string
  cloud_provider: string
  cloud_account: string
  physical_location: string
  ssh_port: number
  ssh_username: string
  ssh_secret_name: string
  ssh_jump_host_name: string
  addresses: HostAddress[]
}
