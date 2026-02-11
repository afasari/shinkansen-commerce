# Orders API

## Overview

The Orders API provides endpoints for managing the order lifecycle.

**Base URL**: `/v1/orders`

## List Orders

Retrieve a paginated list of user orders.

### Request

```bash
curl "http://localhost:8080/v1/orders?page=1&limit=20&status=confirmed"
```

### Response (200 OK)

```json
{
  "orders": [
    {
      "id": "ord-123",
      "user_id": "user-456",
      "status": "confirmed",
      "total_cents": 15000,
      "created_at": "2024-02-11T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50
  }
}
```

## Create Order

Create a new order.

### Request

```bash
curl -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delivery_address_id": "addr-123",
    "items": [
      {"product_id": "prod-1", "quantity": 2}
    ]
  }'
```

### Response (201 Created)

```json
{
  "order_id": "ord-456"
}
```

## Get Order

Retrieve order details.

### Request

```bash
curl http://localhost:8080/v1/orders/ord-456
```

### Response (200 OK)

```json
{
  "order": {
    "id": "ord-456",
    "status": "confirmed",
    "total_cents": 30000
  }
}
```

## Update Order Status

Update order status.

### Request

```bash
curl -X POST http://localhost:8080/v1/orders/ord-456/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'
```

## Cancel Order

Cancel an order.

### Request

```bash
curl -X POST http://localhost:8080/v1/orders/ord-456/cancel \
  -H "Authorization: Bearer $TOKEN"
```
