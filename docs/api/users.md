# Users API

## Overview

The Users API provides endpoints for user authentication and profile management.

## Register User

```bash
curl -X POST http://localhost:8080/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User",
    "phone": "090-1234-5678"
  }'
```

## Login

```bash
curl -X POST http://localhost:8080/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## Get User

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/users/me
```

## Add Address

```bash
curl -X POST http://localhost:8080/v1/users/me/addresses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Home",
    "postal_code": "100-0001",
    "prefecture": "Tokyo",
    "city": "Chiyoda-ku",
    "street_address": "1-1-1 Marunouchi"
  }'
```
