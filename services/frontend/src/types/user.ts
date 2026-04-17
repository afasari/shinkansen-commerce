export interface User {
  id: string
  email: string
  name: string
  phone: string
  created_at: string
  updated_at: string
  active: boolean
  role: string | number
}

export interface Address {
  id: string
  user_id: string
  name: string
  phone: string
  postal_code: string
  prefecture: string
  city: string
  address_line1: string
  address_line2: string
  is_default: boolean
  created_at: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
  phone: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  user_id: string
  access_token: string
  refresh_token: string
  role: string | number
}

export interface UpdateUserRequest {
  name?: string
  phone?: string
}

export interface AddAddressRequest {
  name: string
  phone: string
  postal_code: string
  prefecture: string
  city: string
  address_line1: string
  address_line2: string
  is_default: boolean
}

export interface UpdateAddressRequest extends AddAddressRequest {}
