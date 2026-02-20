# Order API

Service: shinkansen.order

## Overview

The order service provides APIs for managing order-related operations.

## RPC Methods

### CreateOrder

**Request:** `CreateOrderRequest`

**Response:** `shinkansen.common.Empty`

### GetOrder

**Request:** `GetOrderRequest`

**Response:** `shinkansen.common.Empty`

### ListOrders

**Request:** `ListOrdersRequest`

**Response:** `shinkansen.common.Empty`

### UpdateOrderStatus

**Request:** `UpdateOrderStatusRequest`

**Response:** `shinkansen.common.Empty`

### CancelOrder

**Request:** `CancelOrderRequest`

**Response:** `shinkansen.common.Empty`

### ApplyPoints

**Request:** `ApplyPointsRequest`

**Response:** `shinkansen.common.Empty`

### ReserveDeliverySlot

**Request:** `ReserveDeliverySlotRequest`

**Response:** `shinkansen.common.Empty`


## HTTP Endpoints

| Method | Path |
|--------|------|
| GET | `/v1/orders/{order_id}` |
| GET | `/v1/orders` |
| POST | `/v1/orders/{order_id}/cancel` |

## Message Types

Message types are defined in `order/order_messages.proto`

### Order

Data structure for order operations.

### OrderItem

Data structure for order operations.

### ShippingAddress

Data structure for order operations.

### CreateOrderItem

Data structure for order operations.

## Implementation

**Language:** Go
**Location:** `services/order-service/`

## Testing

```bash
# Example gRPC call using grpcurl
grpcurl -plaintext localhost:<port> shinkansen.order.OrderService/CreateOrder
```

