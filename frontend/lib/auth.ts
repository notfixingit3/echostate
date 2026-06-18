import { fetchApi } from "@/lib/api"

export type AuthUser = {
  id: string
  display_name: string
  role: "admin" | "scanner"
  disabled: boolean
}

export type AuthConfig = {
  auth_required: boolean
  auth_enabled: boolean
  enrollment_code_ttl_hours: number
  enrollment_code_length: number
  recovery_code_length: number
  session_ttl_hours: number
  max_code_attempts: number
  code_attempt_window_minutes: number
  webauthn_rp_id: string
  webauthn_rp_origin: string
}

export type AuthCredential = {
  id: string
  nickname: string
  created_at: string
  last_used_at?: string
}

export type AuthSession = {
  authenticated: boolean
  user?: AuthUser
  credentials?: AuthCredential[]
}

const USER_ID_KEY = "echostate_user_id"

export function getStoredUserId(): string | null {
  if (typeof window === "undefined") return null
  return localStorage.getItem(USER_ID_KEY)
}

export function storeUserId(userId: string) {
  localStorage.setItem(USER_ID_KEY, userId)
}

export function clearStoredUserId() {
  localStorage.removeItem(USER_ID_KEY)
}

export async function fetchAuthConfig(): Promise<AuthConfig> {
  return fetchApi<AuthConfig>("/api/auth/config")
}

export async function fetchAuthSession(): Promise<AuthSession> {
  return fetchApi<AuthSession>("/api/auth/session")
}

export async function verifyEnrollmentCode(code: string) {
  return fetchApi<{
    user: AuthUser
    needs_passkey: boolean
    expires_at: string
  }>("/api/auth/enroll/verify", {
    method: "POST",
    body: JSON.stringify({ code }),
  })
}

function bufferToBase64url(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ""
  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "")
}

function base64urlToBuffer(value: string): ArrayBuffer {
  const padded = value.replace(/-/g, "+").replace(/_/g, "/")
  const pad = padded.length % 4 === 0 ? "" : "=".repeat(4 - (padded.length % 4))
  const binary = atob(padded + pad)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

function prepareCreationOptions(options: PublicKeyCredentialCreationOptions): PublicKeyCredentialCreationOptions {
  return {
    ...options,
    challenge: base64urlToBuffer(options.challenge as unknown as string),
    user: {
      ...options.user,
      id: base64urlToBuffer(options.user.id as unknown as string),
    },
    excludeCredentials: options.excludeCredentials?.map((cred) => ({
      ...cred,
      id: base64urlToBuffer(cred.id as unknown as string),
    })),
  }
}

function prepareRequestOptions(options: PublicKeyCredentialRequestOptions): PublicKeyCredentialRequestOptions {
  return {
    ...options,
    challenge: base64urlToBuffer(options.challenge as unknown as string),
    allowCredentials: options.allowCredentials?.map((cred) => ({
      ...cred,
      id: base64urlToBuffer(cred.id as unknown as string),
    })),
  }
}

function credentialToJSON(credential: PublicKeyCredential): Record<string, unknown> {
  const response = credential.response as AuthenticatorAttestationResponse | AuthenticatorAssertionResponse
  const payload: Record<string, unknown> = {
    id: credential.id,
    rawId: bufferToBase64url(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64url(response.clientDataJSON),
    },
  }

  if ("attestationObject" in response) {
    ;(payload.response as Record<string, unknown>).attestationObject = bufferToBase64url(
      response.attestationObject
    )
  }
  if ("authenticatorData" in response) {
    ;(payload.response as Record<string, unknown>).authenticatorData = bufferToBase64url(
      response.authenticatorData
    )
  }
  if ("signature" in response) {
    ;(payload.response as Record<string, unknown>).signature = bufferToBase64url(response.signature)
  }
  if ("userHandle" in response && response.userHandle) {
    ;(payload.response as Record<string, unknown>).userHandle = bufferToBase64url(response.userHandle)
  }

  return payload
}

type WebAuthnCreationPayload = {
  publicKey: PublicKeyCredentialCreationOptions
}

type WebAuthnRequestPayload = {
  publicKey: PublicKeyCredentialRequestOptions
}

export async function registerPasskey(nickname?: string) {
  const begin = await fetchApi<{
    options: WebAuthnCreationPayload
    challenge_id: string
  }>("/api/auth/webauthn/register/begin", { method: "POST", body: "{}" })

  const credential = (await navigator.credentials.create({
    publicKey: prepareCreationOptions(begin.options.publicKey),
  })) as PublicKeyCredential | null

  if (!credential) {
    throw new Error("Passkey registration was cancelled")
  }

  const body = {
    challenge_id: begin.challenge_id,
    nickname: nickname?.trim() || "",
    ...credentialToJSON(credential),
  }

  return fetchApi<{ user: AuthUser; registered: boolean }>(
    "/api/auth/webauthn/register/finish",
    {
      method: "POST",
      body: JSON.stringify(body),
    }
  )
}

export async function loginWithPasskey(userId: string) {
  const begin = await fetchApi<{
    options: WebAuthnRequestPayload
    challenge_id: string
    user_id: string
    needs_passkey?: boolean
  }>("/api/auth/webauthn/login/begin", {
    method: "POST",
    body: JSON.stringify({ user_id: userId }),
  })

  if (begin.needs_passkey) {
    return { needsPasskey: true as const }
  }

  const credential = (await navigator.credentials.get({
    publicKey: prepareRequestOptions(begin.options.publicKey),
  })) as PublicKeyCredential | null

  if (!credential) {
    throw new Error("Passkey sign-in was cancelled")
  }

  const body = {
    challenge_id: begin.challenge_id,
    user_id: begin.user_id,
    ...credentialToJSON(credential),
  }

  const result = await fetchApi<{ user: AuthUser }>("/api/auth/webauthn/login/finish", {
    method: "POST",
    body: JSON.stringify(body),
  })

  return { needsPasskey: false as const, user: result.user }
}

export async function logout() {
  await fetchApi("/api/auth/logout", { method: "POST", body: "{}" })
  clearStoredUserId()
}