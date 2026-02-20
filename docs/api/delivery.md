# Delivery API

Service: shinkansen.delivery

## Overview

The delivery service provides APIs for managing delivery-related operations.

## RPC Methods

### GetDeliverySlots

**Request:** `GetDeliverySlotsRequest`

**Response:** `GetDeliverySlotsResponse`

### ReserveDeliverySlot

**Request:** `ReserveDeliverySlotRequest`

**Response:** `ReserveDeliverySlotResponse`

### GetShipment

**Request:** `GetShipmentRequest`

**Response:** `GetShipmentResponse`

### UpdateShipmentStatus

**Request:** `UpdateShipmentStatusRequest`

**Response:** `shinkansen.common.Empty`


## HTTP Endpoints

| Method | Path |
|--------|------|
| GET | `/v1/delivery/slots` |
| GET | `/v1/shipments/{shipment_id}` |

## Message Types

Message types are defined in `delivery/delivery_messages.proto`

### DeliverySlot

Data structure for delivery operations.

### DeliveryZone

Data structure for delivery operations.

### Shipment

Data structure for delivery operations.

### TrackingEvent

Data structure for delivery operations.

## Implementation

**Language:** Go
**Location:** `services/delivery-service/`

## Testing

```bash
# Example gRPC call using grpcurl
grpcurl -plaintext localhost:<port> shinkansen.delivery.DeliveryService/GetDeliverySlots
```

