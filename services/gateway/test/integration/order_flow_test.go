package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL      = "http://localhost:8080"
	testEmail    = "test@example.com"
	testPassword = "testPassword123"
	testName     = "Test User"
	testPhone    = "090-1234-5678"
)

type TestClient struct {
	httpClient *http.Client
	authToken  string
}

func NewTestClient() *TestClient {
	return &TestClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TestClient) doRequest(method, url string, body interface{}, authToken string) (*http.Response, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, baseURL+url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	return c.httpClient.Do(req)
}

func (c *TestClient) getJSON(url string, response interface{}) error {
	return c.requestJSON("GET", url, nil, response, c.authToken)
}

func (c *TestClient) postJSON(url string, body, response interface{}) error {
	return c.requestJSON("POST", url, body, response, c.authToken)
}

func (c *TestClient) requestJSON(method, url string, body, response interface{}, authToken string) error {
	resp, err := c.doRequest(method, url, body, authToken)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if response != nil {
		return json.NewDecoder(resp.Body).Decode(response)
	}
	return nil
}

func TestCompleteOrderFlow(t *testing.T) {
	ctx := context.Background()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewTestClient()

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.doRequest("GET", "/health", nil, "")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	var userID string

	t.Run("Register User", func(t *testing.T) {
		registerReq := map[string]interface{}{
			"email":    testEmail,
			"password": testPassword,
			"name":     testName,
			"phone":    testPhone,
		}

		var resp map[string]interface{}
		err := client.postJSON("/v1/users/register", registerReq, &resp)
		if err != nil {
			t.Fatalf("Register user failed: %v", err)
		}

		if resp["user_id"] == nil {
			t.Error("user_id not in response")
		} else {
			userID = resp["user_id"].(string)
		}

		if resp["access_token"] == nil {
			t.Error("access_token not in response")
		} else {
			client.authToken = resp["access_token"].(string)
		}
	})

	t.Run("Get Current User", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.getJSON("/v1/users/me", &resp)
		if err != nil {
			t.Fatalf("Get current user failed: %v", err)
		}

		user := resp["user"].(map[string]interface{})
		if user["email"] != testEmail {
			t.Errorf("Expected email %s, got %s", testEmail, user["email"])
		}
	})

	t.Run("Add Address", func(t *testing.T) {
		addressReq := map[string]interface{}{
			"name":          "Test Address",
			"phone":         testPhone,
			"postal_code":   "100-0001",
			"prefecture":    "Tokyo",
			"city":          "Chiyoda-ku",
			"address_line1": "1-1 Chiyoda",
			"is_default":    true,
		}

		var resp map[string]interface{}
		err := client.postJSON("/v1/users/me/addresses", addressReq, &resp)
		if err != nil {
			t.Fatalf("Add address failed: %v", err)
		}

		if resp["address_id"] == nil {
			t.Error("address_id not in response")
		}
	})

	var addressID string
	t.Run("List Addresses", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.getJSON("/v1/users/me/addresses", &resp)
		if err != nil {
			t.Fatalf("List addresses failed: %v", err)
		}

		addresses := resp["addresses"].([]interface{})
		if len(addresses) == 0 {
			t.Error("No addresses found")
		} else if len(addresses) > 0 {
			addr := addresses[0].(map[string]interface{})
			addressID = addr["id"].(string)
		}
	})

	var productID string
	t.Run("Create Product (Admin)", func(t *testing.T) {
		productReq := map[string]interface{}{
			"name":        "Test Product",
			"description": "A test product for integration testing",
			"price":       map[string]interface{}{"units": int64(1000), "currency": "JPY"},
			"active":      true,
		}

		var resp map[string]interface{}
		err := client.postJSON("/v1/products", productReq, &resp)
		if err != nil {
			t.Logf("Create product failed (may need admin permissions): %v", err)
		} else {
			productID = resp["product_id"].(string)
		}
	})

	t.Run("List Products", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.getJSON("/v1/products?page=1&limit=10", &resp)
		if err != nil {
			t.Fatalf("List products failed: %v", err)
		}

		products := resp["products"].([]interface{})
		t.Logf("Found %d products", len(products))
	})

	t.Run("Search Products", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.getJSON("/v1/products/search?q=test&limit=10", &resp)
		if err != nil {
			t.Fatalf("Search products failed: %v", err)
		}

		products := resp["products"].([]interface{})
		t.Logf("Search returned %d products", len(products))
	})

	var orderID string
	t.Run("Create Order", func(t *testing.T) {
		testProductID := productID
		if testProductID == "" {
			testProductID = uuid.New().String()
		}

		orderReq := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id": testProductID,
					"quantity":   2,
				},
			},
			"shipping_address_id": addressID,
		}

		var resp map[string]interface{}
		err := client.postJSON("/v1/orders", orderReq, &resp)
		if err != nil {
			t.Fatalf("Create order failed: %v", err)
		}

		if resp["order_id"] == nil {
			t.Error("order_id not in response")
		} else {
			orderID = resp["order_id"].(string)
		}

		if resp["order_number"] == nil {
			t.Error("order_number not in response")
		}
	})

	t.Run("Get Order", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.getJSON("/v1/orders/"+orderID, &resp)
		if err != nil {
			t.Fatalf("Get order failed: %v", err)
		}

		order := resp["order"].(map[string]interface{})
		if order["user_id"] != userID {
			t.Errorf("Expected user_id %s, got %s", userID, order["user_id"])
		}
	})

	t.Run("List Orders", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.getJSON("/v1/orders?page=1&limit=10", &resp)
		if err != nil {
			t.Fatalf("List orders failed: %v", err)
		}

		orders := resp["orders"].([]interface{})
		t.Logf("Found %d orders", len(orders))
	})

	var paymentID string
	t.Run("Create Payment", func(t *testing.T) {
		paymentReq := map[string]interface{}{
			"order_id": orderID,
			"method":   "PAYMENT_METHOD_CREDIT_CARD",
			"amount":   map[string]interface{}{"units": int64(2000), "currency": "JPY"},
		}

		var resp map[string]interface{}
		err := client.postJSON("/v1/payments", paymentReq, &resp)
		if err != nil {
			t.Fatalf("Create payment failed: %v", err)
		}

		if resp["payment_id"] == nil {
			t.Error("payment_id not in response")
		} else {
			paymentID = resp["payment_id"].(string)
		}
	})

	t.Run("Process Payment", func(t *testing.T) {
		paymentReq := map[string]interface{}{
			"payment_data": map[string]string{
				"card_number": "4111111111111111",
				"expiry":      "12/25",
				"cvv":         "123",
			},
		}

		var resp map[string]interface{}
		err := client.postJSON("/v1/payments/"+paymentID+"/process", paymentReq, &resp)
		if err != nil {
			t.Fatalf("Process payment failed: %v", err)
		}

		status := resp["status"]
		if status == nil || status.(string) != "PAYMENT_STATUS_COMPLETED" {
			t.Logf("Payment status: %v", status)
		}
	})

	t.Run("Update Order Status", func(t *testing.T) {
		statusReq := map[string]interface{}{
			"status": "ORDER_STATUS_CONFIRMED",
		}

		resp, err := client.doRequest("POST", "/v1/orders/"+orderID+"/status", statusReq, client.authToken)
		if err != nil {
			t.Fatalf("Update order status failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNoContent {
			body, _ := getResponseBytes(resp)
			t.Errorf("Expected status 204, got %d: %s", resp.StatusCode, string(body))
		}
	})

	_ = ctx
}

func getResponseBytes(resp *http.Response) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	return buf.Bytes(), err
}

func TestMain(m *testing.M) {
	// This is a placeholder for test setup/teardown
	// In real usage, you would:
	// 1. Start docker-compose if not running
	// 2. Wait for services to be healthy
	// 3. Run tests
	// 4. Cleanup

	fmt.Println("Make sure Docker Compose is running with: make up")
	fmt.Println("Wait for all services to be healthy with: docker-compose ps")
	m.Run()
}
