# Authentication

## Overview

The API uses JWT (JSON Web Tokens) for authentication. Tokens are generated during login and must be included in `Authorization` header for authenticated requests.

## Register User

Create a new user account.

### Request

```bash
curl -X POST http://localhost:8080/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "name": "John Doe",
    "phone": "090-1234-5678"
  }'
```

### Response (201 Created)

```json
{
  "user_id": "usr-abc123",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Login

Authenticate with email and password to receive JWT tokens.

### Request

```bash
curl -X POST http://localhost:8080/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

### Response (200 OK)

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

## Using Access Token

Include access token in `Authorization` header:

```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/orders
```

## Error Responses

### Invalid Credentials

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password"
  }
}
```

### Missing Token

```json
{
  "error": {
    "code": "MISSING_TOKEN",
    "message": "Authorization token required"
  }
}
```

## Security Best Practices

- Store tokens securely
- Use HTTPS in production
- Clear tokens on logout
