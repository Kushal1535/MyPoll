
const API_BASE_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8081/api'

export interface PollOption {
  id: string
  text: string
  votes: number
}

export interface Poll {
  id: string
  question: string
  description?: string
  options: PollOption[]
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface PollResponse {
  poll: Poll
  totalVotes: number
}

export interface MyPollItem {
  poll: Poll
  totalVotes: number
}

interface AuthResponse {
  message?: string
  token: string
  user?: {
    id: string
    name: string
    email: string
  }
}

function getToken(): string | null {
  return localStorage.getItem('mypoll_token')
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken()

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })

  const data = await response.json().catch(() => ({}))

  if (!response.ok) {
    throw new Error(data.error || 'Something went wrong')
  }

  return data as T
}

export const api = {
  register: (name: string, email: string, password: string) =>
    request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    }),

  login: (email: string, password: string) =>
    request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  getMe: () =>
    request<{ user: { id: string; name: string; email: string } }>(
      '/auth/me',
    ),

  createPoll: (
    question: string,
    description: string,
    options: string[],
  ) =>
    request<{ message: string; poll: Poll }>('/polls', {
      method: 'POST',
      body: JSON.stringify({ question, description, options }),
    }),

  getMyPolls: () =>
    request<{ polls: MyPollItem[]; count: number }>('/polls/mine'),

  getPoll: (pollId: string) =>
    request<PollResponse>(`/polls/${pollId}`),

  vote: (pollId: string, optionId: string, voterId: string) =>
    request<{ message: string; votes: number }>(
      `/polls/${pollId}/vote`,
      {
        method: 'POST',
        body: JSON.stringify({ optionId, voterId }),
      },
    ),
}

export function getWebSocketUrl(pollId: string): string {
  const wsBase = API_BASE_URL.replace(/^http/, 'ws')
  return `${wsBase}/polls/${pollId}/ws`
}
