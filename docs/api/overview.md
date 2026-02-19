# API Overview

The Shinkansen Commerce API follows RESTful principles with gRPC backend services. All endpoints are served through API Gateway at `/`.

## Base URL

```
/
```

## Authentication

All endpoints (except `/v1/users/register` and `/v1/users/login`) require authentication via JWT bearer token.

### Request Header
```
Authorization: Bearer <access_token>
```

## Rate Limiting

- **Limit**: 100 requests per minute per user
- **Header**: `X-RateLimit-Limit: 100`
- **Header**: `X-RateLimit-Remaining: 85`
- **Header**: `X-RateLimit-Reset: 1676332800`

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

### Error Response Format
```json
{
  "error": {
    "code": "OUT_OF_STOCK",
    "message": "Product quantity exceeds available stock",
    "details": {
      "requested": 10,
      "available": 5
    }
  }
}
```

## Pagination

List endpoints support pagination via query parameters:

### Parameters
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `limit` | integer | 20 | Items per page (max: 100) |

### Response
```json
{
  "items": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

## API Endpoints by Category

| Category | Endpoints | Description |
|----------|-----------|-------------|
| [Products](/api/products) | 8 | Product catalog management |
| [Orders](/api/orders) | 7 | Order lifecycle management |
| [Users](/api/users) | 8 | User authentication & profiles |
| [Payments](/api/payments) | 4 | Payment processing |
| [Inventory](/api/inventory) | 5 | Stock management |
| [Delivery](/api/delivery) | 3 | Delivery logistics |

## Interactive Documentation

Explore API interactively using Swagger UI: [swagger.yaml](api/swagger.yaml)

## Testing the API

### Using cURL

```bash
# List products
curl /v1/products

# Register user
curl -X POST /v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User",
    "phone": "090-1234-5678"
  }'

# Create order (with JWT)
TOKEN="your-jwt-token"
curl -X POST /v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delivery_address_id": "addr-123",
    "items": [{"product_id": "prod-1", "quantity": 2}]
  }'
```

### Using Go Client

```go
package main

import (
    "context"
    "log"
    productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
    sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
    "google.golang.org/grpc"
)

func main() {
    conn, err := grpc.Dial("product-service:9091", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := productpb.NewProductServiceClient(conn)
    
    resp, err := client.ListProducts(context.Background(), &productpb.ListProductsRequest{
        Pagination: &sharedpb.Pagination{
            Page:  1,
            Limit: 20,
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, product := range resp.Products {
        log.Printf("Product: %s", product.Name)
    }
}
```

## API Versioning

Current API version: **v1**

Future versions will be path-based (e.g., `/v2/orders`) to maintain backward compatibility.
