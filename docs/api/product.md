# Product API

Service: shinkansen.product

## Overview

The product service provides APIs for managing product-related operations.

## RPC Methods

### ListProducts

**Request:** `ListProductsRequest`

**Response:** `ListProductsResponse`

### GetProduct

**Request:** `GetProductRequest`

**Response:** `GetProductResponse`

### CreateProduct

**Request:** `CreateProductRequest`

**Response:** `CreateProductResponse`

### UpdateProduct

**Request:** `UpdateProductRequest`

**Response:** `UpdateProductResponse`

### DeleteProduct

**Request:** `DeleteProductRequest`

**Response:** `shinkansen.common.Empty`

### SearchProducts

**Request:** `SearchProductsRequest`

**Response:** `SearchProductsResponse`

### GetProductVariants

**Request:** `GetProductVariantsRequest`

**Response:** `GetProductVariantsResponse`


## HTTP Endpoints

| Method | Path |
|--------|------|
| GET | `/v1/products` |
| GET | `/v1/products/{product_id}` |
| DELETE | `/v1/products/{product_id}` |
| GET | `/v1/products/search` |
| GET | `/v1/products/{product_id}/variants` |

## Message Types

Message types are defined in `product/product_messages.proto`

### Product

Data structure for product operations.

### ProductVariant

Data structure for product operations.

### Category

Data structure for product operations.

## Implementation

**Language:** Go
**Location:** `services/product-service/`

## Testing

```bash
# Example gRPC call using grpcurl
grpcurl -plaintext localhost:<port> shinkansen.product.ProductService/ListProducts
```

