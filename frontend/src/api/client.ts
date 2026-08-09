import type { ApiErrorPayload, Leaderboard, Pet, Reward, Summary, Task, TaskType, User } from './types'

const API_ROOT = '/api/v1'

let unauthorizedHandler: (() => void) | null = null

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export function setUnauthorizedHandler(handler: (() => void) | null) {
  unauthorizedHandler = handler
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_ROOT}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  })

  if (!response.ok) {
    let payload: ApiErrorPayload = {}
    try {
      payload = await response.json() as ApiErrorPayload
    } catch {
      // The status code still provides a useful fallback error.
    }
    if (response.status === 401) {
      unauthorizedHandler?.()
    }
    throw new ApiError(response.status, payload.code ?? 'request_failed', payload.message ?? 'Request failed')
  }

  if (response.status === 204) {
    return undefined as T
  }
  return response.json() as Promise<T>
}

export const api = {
  register: (input: { email: string; password: string; displayName: string }) =>
    request<User>('/auth/register', { method: 'POST', body: JSON.stringify(input) }),
  login: (input: { email: string; password: string }) =>
    request<User>('/auth/login', { method: 'POST', body: JSON.stringify(input) }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  getMe: () => request<User>('/auth/me'),
  getPet: () => request<Pet>('/pet'),
  chargePet: () => request<Pet>('/pet/actions', { method: 'POST', body: JSON.stringify('charge') }),
  getTasks: async () => (await request<{ tasks: Task[] }>('/tasks/today')).tasks,
  applyTaskAction: (action: TaskType) =>
    request<{ status: 'applied' }>('/mock-avito/actions', { method: 'POST', body: JSON.stringify(action) }),
  getRewards: async () => (await request<{ rewards: Reward[] }>('/rewards')).rewards,
  useReward: (rewardId: string) => request<Reward>(`/rewards/${rewardId}/use`, { method: 'POST' }),
  getSummary: () => request<Summary>('/summary/today'),
  getLeaderboard: () => request<Leaderboard>('/leaderboard'),
}
