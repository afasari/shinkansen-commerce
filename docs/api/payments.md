# Payments API

## Overview

The Payments API provides endpoints for payment processing.

## Create Payment

Create a new payment.

```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ord-123",
    "method": "credit_card"
  }'
```

## Payment Methods

| Method | Description |
|--------|-------------|
| credit_card | Credit or debit card |
| konbini | Convenience store payment |
| bank_transfer | Bank transfer |
| points | Pay with loyalty points |
| paypay | PayPay mobile payment |
| line_pay | LINE Pay |

## Get Payment

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/payments/pay-123
```

## Process Payment

```bash
curl -X POST http://localhost:8080/v1/payments/pay-123/process \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Refund Payment

```bash
curl -X POST http://localhost:8080/v1/payments/pay-123/refund \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 1000}'
```
