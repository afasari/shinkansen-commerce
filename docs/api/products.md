# Products API

## Overview

The Products API provides endpoints for managing product catalog, including CRUD operations, search, and variant management.

**Base URL**: `/v1/products`

## List Products

Retrieve a paginated list of products with optional filtering.

### Request

```bash
curl "http://localhost:8080/v1/products?page=1&limit=20&category_id=cat-electronics&active_only=true"
```

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `limit` | integer | No | 20 | Items per page (max: 100) |
| `category_id` | string | No | - | Filter by category |
| `active_only` | boolean | No | true | Only return active products |

### Response (200 OK)

```json
{
  "products": [
    {
      "id": "prod-abc123",
      "name": "Wireless Headphones",
      "description": "Premium noise-cancelling headphones",
      "category_id": "cat-electronics",
      "price": {
        "currency": "JPY",
        "units": 15000,
        "nanos": 0
      },
      "sku": "WH-BT-001",
      "active": true,
      "image_urls": ["https://example.com/products/wh-bt-001.jpg"],
      "stock_quantity": 100,
      "created_at": "2024-02-11T12:00:00Z",
      "updated_at": "2024-02-11T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Go Client Example

```go
resp, _ := client.ListProducts(context.Background(), &productpb.ListProductsRequest{
    CategoryId: "cat-electronics",
    ActiveOnly: true,
    Pagination: &sharedpb.Pagination{Page: 1, Limit: 20},
})
for _, product := range resp.Products {
    fmt.Printf("%s - %d JPY\n", product.Name, product.Price.Units)
}
```

### Caching
Products list is cached in Redis for 5 minutes. Cache is invalidated when products are created, updated, or deleted.

## Get Product

Retrieve detailed information about a specific product.

### Request

```bash
curl http://localhost:8080/v1/products/prod-abc123
```

### Response (200 OK)

```json
{
  "product": {
    "id": "prod-abc123",
    "name": "Wireless Headphones",
    "description": "Premium noise-cancelling headphones",
    "category_id": "cat-electronics",
    "price": {
      "currency": "JPY",
      "units": 15000,
      "nanos": 0
    },
    "sku": "WH-BT-001",
    "active": true,
    "image_urls": ["https://example.com/products/wh-bt-001.jpg"],
    "stock_quantity": 100,
    "created_at": "2024-02-11T12:00:00Z",
    "updated_at": "2024-02-11T12:00:00Z"
  }
}
```

## Create Product

Create a new product in the catalog.

### Request

```bash
curl -X POST http://localhost:8080/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Mouse",
    "description": "Ergonomic wireless mouse",
    "category_id": "cat-electronics",
    "price": {"currency": "JPY", "units": 3000, "nanos": 0},
    "sku": "WM-BT-001",
    "stock_quantity": 50
  }'
```

### Response (201 Created)

```json
{
  "product_id": "prod-xyz789"
}
```

### Go Client Example

```go
resp, _ := client.CreateProduct(context.Background(), &productpb.CreateProductRequest{
    Name:        "Wireless Mouse",
    CategoryId:  "cat-electronics",
    Price: &commonpb.Money{Currency: "JPY", Units: 3000},
    Sku:          "WM-BT-001",
    StockQuantity: 50,
})
fmt.Printf("Created product: %s\n", resp.ProductId)
```

## Update Product

Update an existing product.

### Request

```bash
curl -X PUT http://localhost:8080/v1/products/prod-abc123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Headphones Pro",
    "price": {"currency": "JPY", "units": 18000, "nanos": 0},
    "active": true
  }'
```

### Response (200 OK)

```json
{
  "product": {
    "id": "prod-abc123",
    "name": "Wireless Headphones Pro",
    "price": {"currency": "JPY", "units": 18000, "nanos": 0},
    "active": true
  }
}
```

## Delete Product

Delete a product from the catalog.

### Request

```bash
curl -X DELETE http://localhost:8080/v1/products/prod-abc123 \
  -H "Authorization: Bearer $TOKEN"
```

### Response (200 OK)

```json
{}
```

## Search Products

Search products by text query and filters.

### Request

```bash
curl "http://localhost:8080/v1/products/search?query=wireless&min_price=1000&max_price=20000&in_stock_only=true"
```

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | Search query (product name, description) |
| `category_id` | string | No | - | Filter by category |
| `min_price` | integer | No | - | Minimum price (units) |
| `max_price` | integer | No | - | Maximum price (units) |
| `in_stock_only` | boolean | No | false | Only in-stock products |
| `page` | integer | No | 1 | Page number |
| `limit` | integer | No | 20 | Items per page |

### Response (200 OK)

Same format as `List Products`.

## Get Product Variants

Retrieve product variants (size, color, etc.).

### Request

```bash
curl http://localhost:8080/v1/products/prod-abc123/variants
```

### Response (200 OK)

```json
{
  "variants": [
    {
      "id": "var-123",
      "product_id": "prod-abc123",
      "name": "Black",
      "attributes": {"color": "black", "size": "M"},
      "price": {"currency": "JPY", "units": 15000, "nanos": 0},
      "sku": "WH-BT-001-BLK-M",
      "stock_quantity": 50
    }
  ]
}
```

## Caching Strategy

| Operation | Cache TTL | Invalidation |
|------------|------------|--------------|
| List Products | 5 min | Product create/update/delete |
| Get Product | 30 min | Product update/delete |
| Search Products | 5 min | Product update/delete |
| Get Variants | 30 min | Product update/delete |
