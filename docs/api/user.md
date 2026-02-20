# User API

Service: shinkansen.user

## Overview

The user service provides APIs for managing user-related operations.

## RPC Methods

### RegisterUser

**Request:** `RegisterUserRequest`

**Response:** `shinkansen.common.Empty`

### LoginUser

**Request:** `LoginUserRequest`

**Response:** `shinkansen.common.Empty`

### GetUser

**Request:** `GetUserRequest`

**Response:** `shinkansen.common.Empty`

### UpdateUser

**Request:** `UpdateUserRequest`

**Response:** `shinkansen.common.Empty`

### AddAddress

**Request:** `AddAddressRequest`

**Response:** `shinkansen.common.Empty`

### ListAddresses

**Request:** `ListAddressesRequest`

**Response:** `shinkansen.common.Empty`

### UpdateAddress

**Request:** `UpdateAddressRequest`

**Response:** `shinkansen.common.Empty`

### DeleteAddress

**Request:** `DeleteAddressRequest`

**Response:** `shinkansen.common.Empty`


## HTTP Endpoints

| Method | Path |
|--------|------|
| GET | `/v1/users/{user_id}` |
| GET | `/v1/users/{user_id}/addresses` |
| DELETE | `/v1/users/{user_id}/addresses/{address_id}` |

## Message Types

Message types are defined in `user/user_messages.proto`

### User

Data structure for user operations.

### Address

Data structure for user operations.

## Implementation

**Language:** Go
**Location:** `services/user-service/`

## Testing

```bash
# Example gRPC call using grpcurl
grpcurl -plaintext localhost:<port> shinkansen.user.UserService/RegisterUser
```

