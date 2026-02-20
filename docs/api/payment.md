# Payment API

Service: shinkansen.payment

## Overview

The payment service provides APIs for managing payment-related operations.

## RPC Methods

### CreatePayment

**Request:** `CreatePaymentRequest`

**Response:** `shinkansen.common.Empty`

### GetPayment

**Request:** `GetPaymentRequest`

**Response:** `shinkansen.common.Empty`

### ProcessPayment

**Request:** `ProcessPaymentRequest`

**Response:** `shinkansen.common.Empty`

### RefundPayment

**Request:** `RefundPaymentRequest`

**Response:** `shinkansen.common.Empty`


## HTTP Endpoints

| Method | Path |
|--------|------|
| GET | `/v1/payments/{payment_id}` |

## Message Types

Message types are defined in `payment/payment_messages.proto`

### Payment

Data structure for payment operations.

## Implementation

**Language:** Go
**Location:** `services/payment-service/`

## Testing

```bash
# Example gRPC call using grpcurl
grpcurl -plaintext localhost:<port> shinkansen.payment.PaymentService/CreatePayment
```

