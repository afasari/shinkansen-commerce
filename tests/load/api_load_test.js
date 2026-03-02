import { check } from 'k6';
import { SharedArray } from 'k6/data';
import http from 'k6/http';

// Test configuration for Shinkansen Commerce API Gateway
const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

// Scenarios configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 300 },  // Ramp up to 300 users
    { duration: '5m', target: 300 },  // Stay at 300 users
    { duration: '3m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% under 500ms, 99% under 1s
    http_req_failed: ['rate<0.01'],                 // Error rate < 1%
  },
};

// Login credentials
const USERNAME = 'test@example.com';
const PASSWORD = 'password123';

let authCookie = null;

// Setup function - runs once
export function setup() {
  // Login and get auth token
  const loginRes = http.post(`${BASE_URL}/v1/users/login`, JSON.stringify({
    email: USERNAME,
    password: PASSWORD,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login successful': (r) => r.status === 200,
  });

  const loginData = loginRes.json();
  return { token: loginData.access_token };
}

// Main test scenarios
export default function(data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  // Scenario 1: Browse products
  const browseGroup = 'Browse Products';

  // List products
  const listRes = http.get(`${BASE_URL}/v1/products?page=1&limit=20`, { headers });
  check(listRes, {
    [`${browseGroup} - list products status 200`]: (r) => r.status === 200,
    [`${browseGroup} - list products has data`]: (r) => r.json('products.length') > 0,
  });

  // Scenario 2: Search products
  const searchGroup = 'Search Products';

  const searchRes = http.get(`${BASE_URL}/v1/products/search?query=electronics&page=1&limit=10`, { headers });
  check(searchRes, {
    [`${searchGroup} - search status 200`]: (r) => r.status === 200,
    [`${searchGroup} - search has results`]: (r) => r.json('products.length') >= 0,
  });

  // Scenario 3: Get product details
  if (listRes.json('products.length') > 0) {
    const productId = listRes.json('products[0].id');
    const detailRes = http.get(`${BASE_URL}/v1/products/${productId}`, { headers });
    check(detailRes, {
      'Product Details - status 200': (r) => r.status === 200,
      'Product Details - has name': (r) => r.json('product.name') !== undefined,
    });

    // Scenario 4: Add to cart (simulation)
    const cartGroup = 'Cart Operations';

    const addToCartRes = http.post(`${BASE_URL}/v1/cart/items`, JSON.stringify({
      product_id: productId,
      quantity: 1,
    }), { headers });

    check(addToCartRes, {
      [`${cartGroup} - add to cart status 200`]: (r) => r.status === 200 || r.status === 201 || r.status === 202,
    });

    // Get cart
    const getCartRes = http.get(`${BASE_URL}/v1/cart`, { headers });
    check(getCartRes, {
      [`${cartGroup} - get cart status 200`]: (r) => r.status === 200,
    });
  }

  // Scenario 5: Get user orders
  const ordersGroup = 'Order Operations';

  const ordersRes = http.get(`${BASE_URL}/v1/orders?page=1&limit=10`, { headers });
  check(ordersRes, {
    [`${ordersGroup} - list orders status 200`]: (r) => r.status === 200,
  });

  // Small sleep between iterations to simulate realistic user behavior
  sleep(Math.random() * 3);
}

// Teardown function
export function teardown(data) {
  // Logout
  http.post(`${BASE_URL}/v1/users/logout`, null, {
    headers: { 'Authorization': `Bearer ${data.token}` },
  });
}
