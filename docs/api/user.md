# User API

Service: shinkansen.user

## Overview

The user service provides APIs for managing user-related operations.

## RPC Methods

### RegisterUser

**Request:** `RegisterUserRequest`

**Response:** `RegisterUserResponse`

### LoginUser

**Request:** `LoginUserRequest`

**Response:** `LoginUserResponse`

### GetUser

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

### UpdateUser

**Request:** `UpdateUserRequest`

**Response:** `UpdateUserResponse`

### AddAddress

**Request:** `AddAddressRequest`

**Response:** `AddAddressResponse`

### ListAddresses

**Request:** `ListAddressesRequest`

**Response:** `ListAddressesResponse`

### UpdateAddress

**Request:** `UpdateAddressRequest`

**Response:** `UpdateAddressResponse`

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

