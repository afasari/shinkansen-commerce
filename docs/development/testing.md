# Testing Guide

## Overview

This guide covers testing the Shinkansen Commerce platform.

## Unit Tests

### Run All Unit Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Test Single Service

```bash
cd services/product-service
go test ./... -v -cover
```

## Integration Tests

### Run Integration Tests

```bash
make test-integration
```

Integration tests require infrastructure to be running:

```bash
make up
make test-integration
```

### Test Coverage

Integration tests cover:

- User registration and authentication
- Product listing and search
- Order lifecycle
- Payment processing
- Inventory reservation
- Delivery slot management

## Writing Tests

### Example Unit Test

```go
func TestCreateProduct(t *testing.T) {
    // Setup
    repo := &mockRepository{}
    service := NewProductService(repo)
    
    // Test
    req := &pb.CreateProductRequest{
        Name: "Test Product",
        Price: &pb.Money{Units: 1000},
    }
    resp, err := service.CreateProduct(context.Background(), req)
    
    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if resp.ProductId == "" {
        t.Fatal("expected product ID")
    }
}
```

### Example Integration Test

```go
func TestCompleteOrderFlow(t *testing.T) {
    client := NewTestClient()
    
    // Register user
    userResp, _ := client.PostJSON("/v1/users/register", userReq, &UserResponse{})
    
    // Login
    loginResp, _ := client.PostJSON("/v1/users/login", loginReq, &LoginResponse{})
    client.SetToken(loginResp.AccessToken)
    
    // Create order
    orderResp, _ := client.PostJSON("/v1/orders", orderReq, &OrderResponse{})
    
    // Verify
    if orderResp.OrderId == "" {
        t.Fatal("expected order ID")
    }
}
```

## Test Data

### Test Users

| Email | Password | Role |
|-------|----------|------|
| test@example.com | password123 | customer |
| admin@example.com | admin123 | admin |

### Test Products

| SKU | Name | Price |
|-----|------|-------|
| TEST-001 | Test Product 1 | 1000 JPY |
| TEST-002 | Test Product 2 | 2000 JPY |

## Troubleshooting Tests

### Tests Failing

```bash
# Check service logs
make logs

# Run specific test with verbose output
go test -v ./... -run TestCreateProduct
```

### Database Issues

```bash
# Reset database
make down
docker volume rm shinkansen-commerce_postgres_data
make up
```
