import { client, setTokens, clearTokens } from './client'
import type { RegisterRequest, LoginRequest, AuthResponse, User } from '@/types'

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await client.post<AuthResponse>('/v1/users/register', data)
  setTokens(res.data)
  return res.data
}

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await client.post<AuthResponse>('/v1/users/login', data)
  setTokens(res.data)
  return res.data
}

export function logout() {
  clearTokens()
}

export async function getCurrentUser(): Promise<User> {
  const res = await client.get<{ user: User }>('/v1/users/me')
  return res.data.user
}

export async function getUser(userId: string): Promise<User> {
  const res = await client.get<{ user: User }>(`/v1/users/${userId}`)
  return res.data.user
}

export async function updateCurrentUser(data: { name?: string; phone?: string }): Promise<User> {
  const userId = localStorage.getItem('user_id')
  const res = await client.put<{ user: User }>(`/v1/users/${userId}`, data)
  return res.data.user
}
