package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WebhookService handles payment webhook processing
type WebhookService struct {
	paymentService *PaymentService
	konbiniService *KonbiniService
	secretKey      string
	logger         *zap.Logger
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	paymentService *PaymentService,
	konbiniService *KonbiniService,
	secretKey string,
	logger *zap.Logger,
) *WebhookService {
	return &WebhookService{
		paymentService: paymentService,
		konbiniService: konbiniService,
		secretKey:      secretKey,
		logger:         logger,
	}
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Signature string                 `json:"signature"`
}

// HandleWebhook handles an incoming webhook request
func (s *WebhookService) HandleWebhook(ctx context.Context, event WebhookEvent) error {
	s.logger.Info("Received webhook event",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Type))

	// Verify signature
	if !s.verifySignature(event) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Process event based on type
	switch event.Type {
	case "payment.completed":
		return s.handlePaymentCompleted(ctx, event)
	case "payment.failed":
		return s.handlePaymentFailed(ctx, event)
	case "payment.refunded":
		return s.handlePaymentRefunded(ctx, event)
	case "konbini.paid":
		return s.handleKonbiniPaid(ctx, event)
	case "konbini.expired":
		return s.handleKonbiniExpired(ctx, event)
	default:
		s.logger.Warn("Unknown webhook event type", zap.String("type", event.Type))
		return nil
	}
}

// verifySignature verifies the webhook signature
func (s *WebhookService) verifySignature(event WebhookEvent) bool {
	// Create expected signature
	data := fmt.Sprintf("%s:%s:%d", event.ID, event.Type, event.Timestamp.Unix())

	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(data))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	return event.Signature == expectedSig
}

// handlePaymentCompleted handles a payment completed event
func (s *WebhookService) handlePaymentCompleted(ctx context.Context, event WebhookEvent) error {
	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		return fmt.Errorf("missing payment_id in event data")
	}

	transactionID, _ := event.Data["transaction_id"].(string)
	amount := int64(0)
	if amountFloat, ok := event.Data["amount"].(float64); ok {
		amount = int64(amountFloat)
	}

	s.logger.Info("Payment completed webhook",
		zap.String("payment_id", paymentID),
		zap.String("transaction_id", transactionID),
		zap.Int64("amount", amount))

	// Update payment status in database
	// This would trigger order fulfillment

	return nil
}

// handlePaymentFailed handles a payment failed event
func (s *WebhookService) handlePaymentFailed(ctx context.Context, event WebhookEvent) error {
	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		return fmt.Errorf("missing payment_id in event data")
	}

	reason, _ := event.Data["reason"].(string)

	s.logger.Info("Payment failed webhook",
		zap.String("payment_id", paymentID),
		zap.String("reason", reason))

	// Update payment status and notify customer

	return nil
}

// handlePaymentRefunded handles a payment refunded event
func (s *WebhookService) handlePaymentRefunded(ctx context.Context, event WebhookEvent) error {
	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		return fmt.Errorf("missing payment_id in event data")
	}

	refundAmount := int64(0)
	if amountFloat, ok := event.Data["refund_amount"].(float64); ok {
		refundAmount = int64(amountFloat)
	}

	s.logger.Info("Payment refunded webhook",
		zap.String("payment_id", paymentID),
		zap.Int64("refund_amount", refundAmount))

	// Process refund and update order status

	return nil
}

// handleKonbiniPaid handles a Konbini payment completed event
func (s *WebhookService) handleKonbiniPaid(ctx context.Context, event WebhookEvent) error {
	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		return fmt.Errorf("missing payment_id in event data")
	}

	confirmationNumber, _ := event.Data["confirmation_number"].(string)
	paidAtStr, _ := event.Data["paid_at"].(string)

	paidAt, err := time.Parse(time.RFC3339, paidAtStr)
	if err != nil {
		paidAt = time.Now()
	}

	s.logger.Info("Konbini payment completed webhook",
		zap.String("payment_id", paymentID),
		zap.String("confirmation_number", confirmationNumber))

	if s.konbiniService != nil {
		return s.konbiniService.ProcessKonbiniWebhook(
			ctx,
			paymentID,
			"completed",
			confirmationNumber,
			paidAt,
		)
	}

	return nil
}

// handleKonbiniExpired handles a Konbini payment expired event
func (s *WebhookService) handleKonbiniExpired(ctx context.Context, event WebhookEvent) error {
	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		return fmt.Errorf("missing payment_id in event data")
	}

	s.logger.Info("Konbini payment expired webhook",
		zap.String("payment_id", paymentID))

	// Update payment status to expired
	// Cancel associated order

	return nil
}

// WebhookHandler wraps the webhook service for HTTP handling
type WebhookHandler struct {
	service *WebhookService
}

// NewWebhookHandler creates a new webhook HTTP handler
func NewWebhookHandler(service *WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// ServeHTTP handles HTTP webhook requests
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add event ID if not present
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Add timestamp if not present
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if err := h.service.HandleWebhook(r.Context(), event); err != nil {
		h.service.logger.Error("Failed to handle webhook",
			zap.String("event_id", event.ID),
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GenerateSignature generates a signature for a webhook event
func (s *WebhookService) GenerateSignature(event WebhookEvent) string {
	data := fmt.Sprintf("%s:%s:%d", event.ID, event.Type, event.Timestamp.Unix())

	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// RetryWebhook retries a failed webhook delivery
func (s *WebhookService) RetryWebhook(ctx context.Context, event WebhookEvent, webhookURL string, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		s.logger.Info("Retrying webhook delivery",
			zap.String("event_id", event.ID),
			zap.Int("attempt", attempt))

		// Generate signature for this attempt
		event.Signature = s.GenerateSignature(event)

		// Marshal event
		body, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", event.Signature)

		// Send request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			s.logger.Warn("Webhook delivery attempt failed",
				zap.String("event_id", event.ID),
				zap.Int("attempt", attempt),
				zap.Error(err))
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s.logger.Info("Webhook delivered successfully",
				zap.String("event_id", event.ID),
				zap.Int("attempt", attempt),
				zap.Int("status_code", resp.StatusCode))
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		s.logger.Warn("Webhook delivery attempt failed",
			zap.String("event_id", event.ID),
			zap.Int("attempt", attempt),
			zap.Int("status_code", resp.StatusCode))

		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return fmt.Errorf("failed to deliver webhook after %d attempts: %w", maxRetries, lastErr)
}
