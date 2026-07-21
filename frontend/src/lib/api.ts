const SESSION_KEY = 'myntra-session-id'

function getSessionId(): string {
  let id = localStorage.getItem(SESSION_KEY)
  if (!id) {
    id = crypto.randomUUID()
    localStorage.setItem(SESSION_KEY, id)
  }
  return id
}

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-Session-Id': getSessionId(),
      ...init.headers,
    },
  })
  const body = await res.json()
  if (!res.ok || body.success === false) {
    throw new Error(body?.message ?? `Request failed: ${res.status}`)
  }
  return body.data as T
}
