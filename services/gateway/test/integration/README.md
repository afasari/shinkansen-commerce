# Integration Tests

This directory contains end-to-end integration tests for the Shinkansen Commerce platform.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ installed
- Make

## Test Flow

The integration test `order_flow_test.go` tests the complete order lifecycle:

1. **User Registration & Authentication**
   - Register a new user
   - Login and get JWT token
   - Get current user profile

2. **Address Management**
   - Add a shipping address
   - List user addresses
   - Set default address

3. **Product Management**
   - Create a product (requires admin permissions)
   - List all products
   - Search products

4. **Order Lifecycle**
   - Create an order with items
   - Get order details
   - List user orders

5. **Payment Processing**
   - Create a payment
   - Process payment with mock gateway
   - Verify payment status

6. **Order Status Updates**
   - Update order status to confirmed
   - Verify status change

## Running Tests

### Quick Start

```bash
# Start all services
make up

# Run integration tests
make test-integration
```

### Detailed Steps

1. **Start Infrastructure**
   ```bash
   make up
   ```
   This starts:
   - PostgreSQL database
   - Redis cache
   - All 7 microservices
   - API Gateway

2. **Wait for Services**
   Services will be started automatically with health checks.
   Wait 30-60 seconds for all services to become healthy.

   Check status:
   ```bash
   docker-compose ps
   ```

3. **Run Integration Tests**
   ```bash
   make test-integration
   ```
   
   Or manually:
   ```bash
   cd services/gateway
   go test -v ./test/integration/... -timeout 10m
   ```

4. **Stop Infrastructure** (optional)
   ```bash
   make down
   ```

## Test Configuration

The tests use these default values:

- Gateway URL: `http://localhost:8080`
- Test User: `test@example.com` / `testPassword123`
- Test Address: Tokyo, Chiyoda-ku

## Troubleshooting

### Gateway Not Responding

```bash
# Check gateway logs
docker-compose logs gateway

# Check if service is running
docker-compose ps
```

### Database Connection Issues

```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Verify database schema
docker exec shinkansen-postgres psql -U shinkansen -d shinkansen -c "\dn"
```

### Tests Failing

1. Ensure all services are healthy
2. Check service logs: `make logs`
3. Restart services: `make down && make up`

## Test Output

Successful test run:
```
=== RUN   TestCompleteOrderFlow
=== RUN   TestCompleteOrderFlow/Health_Check
=== RUN   TestCompleteOrderFlow/Register_User
=== RUN   TestCompleteOrderFlow/Get_Current_User
...
--- PASS: TestCompleteOrderFlow (45.23s)
PASS
ok      github.com/shinkansen-commerce/shinkansen/services/gateway/test/integration    45.234s
```

## Coverage

The integration tests cover:

- ✅ User registration and login
- ✅ JWT authentication
- ✅ Profile management
- ✅ Address CRUD operations
- ✅ Product listing and search
- ✅ Order creation and retrieval
- ✅ Payment processing (mock gateway)
- ✅ Order status updates
- ✅ API error handling
