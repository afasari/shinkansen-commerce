# Payment API

Service: shinkansen.payment

## Overview

The payment service provides APIs for managing payment-related operations.

## RPC Methods

### CreatePayment

**Request:** `CreatePaymentRequest`

**Response:** `CreatePaymentResponse`

### GetPayment

**Request:** `GetPaymentRequest`

**Response:** `GetPaymentResponse`

### ProcessPayment

**Request:** `ProcessPaymentRequest`

**Response:** `ProcessPaymentResponse`

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

