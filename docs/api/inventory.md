# Inventory API

Service: shinkansen.inventory

## Overview

The inventory service provides APIs for managing inventory-related operations.

## RPC Methods

### GetStock

**Request:** `GetStockRequest`

**Response:** `GetStockResponse`

### UpdateStock

**Request:** `UpdateStockRequest`

**Response:** `shinkansen.common.Empty`

### ReserveStock

**Request:** `ReserveStockRequest`

**Response:** `ReserveStockResponse`

### ReleaseStock

**Request:** `ReleaseStockRequest`

**Response:** `shinkansen.common.Empty`

### GetStockMovements

**Request:** `GetStockMovementsRequest`

**Response:** `GetStockMovementsResponse`


## HTTP Endpoints

| Method | Path |
|--------|------|
| GET | `/v1/inventory/stock` |
| GET | `/v1/inventory/stock/{stock_item_id}/movements` |

## Message Types

Message types are defined in `inventory/inventory_messages.proto`

### StockItem

Data structure for inventory operations.

### StockMovement

Data structure for inventory operations.

### StockReservationItem

Data structure for inventory operations.

## Implementation

**Language:** Rust
**Location:** `services/inventory-service/`

## Testing

```bash
# Example gRPC call using grpcurl
grpcurl -plaintext localhost:<port> shinkansen.inventory.InventoryService/GetStock
```

