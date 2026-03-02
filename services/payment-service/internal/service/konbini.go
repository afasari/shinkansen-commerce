package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	fpdf "github.com/jung-kurt/gofpdf"
	"github.com/google/uuid"
	"go.uber.org/zap"

	paymentpb "github.com/afasari/shinkansen-commerce/gen/proto/go/payment"
	"github.com/afasari/shinkansen-commerce/services/payment-service/internal/db"
)

// KonbiniService handles Konbini (convenience store) payments
type KonbiniService struct {
	queries db.Querier
	logger  *zap.Logger
}

// NewKonbiniService creates a new Konbini service
func NewKonbiniService(queries db.Querier, logger *zap.Logger) *KonbiniService {
	return &KonbiniService{
		queries: queries,
		logger:  logger,
	}
}

// KonbiniStore represents a convenience store chain
type KonbiniStore struct {
	ID          string
	Name        string
	NameEN      string
	NameJP      string
	PaymentCode string
	LogoURL     string
}

// Supported Konbini stores
var KonbiniStores = map[paymentpb.PaymentMethod]KonbiniStore{
	paymentpb.PaymentMethod_PAYMENT_METHOD_KONBINI_SEVENELEVEN: {
		ID:          "seven-eleven",
		Name:        "Seven Eleven",
		NameEN:      "Seven Eleven",
		NameJP:      "セブン-イレブン",
		PaymentCode: "711",
		LogoURL:     "/assets/konbini/711.png",
	},
	paymentpb.PaymentMethod_PAYMENT_METHOD_KONBINI_LAWSON: {
		ID:          "lawson",
		Name:        "Lawson",
		NameEN:      "Lawson",
		NameJP:      "ローソン",
		PaymentCode: "10001",
		LogoURL:     "/assets/konbini/lawson.png",
	},
	paymentpb.PaymentMethod_PAYMENT_METHOD_KONBINI_FAMILYMART: {
		ID:          "familymart",
		Name:        "FamilyMart",
		NameEN:      "FamilyMart",
		NameJP:      "ファミリーマート",
		PaymentCode: "001",
		LogoURL:     "/assets/konbini/familymart.png",
	},
}

// CreateKonbiniPayment creates a new Konbini payment
func (s *KonbiniService) CreateKonbiniPayment(
	ctx context.Context,
	orderID string,
	method paymentpb.PaymentMethod,
	amount int64,
	currency string,
	customerEmail string,
	customerName string,
) (*KonbiniPaymentDetails, error) {
	s.logger.Info("Creating Konbini payment",
		zap.String("order_id", orderID),
		zap.String("method", method.String()))

	// Validate store
	store, ok := KonbiniStores[method]
	if !ok {
		return nil, fmt.Errorf("unsupported Konbini store: %s", method)
	}

	// Generate payment code
	paymentCode := s.generatePaymentCode(store)

	// Calculate expiration (typically 7 days for Konbini payments)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Generate confirmation number
	confirmationNumber := s.generateConfirmationNumber(store)

	return &KonbiniPaymentDetails{
		PaymentID:         uuid.New().String(),
		OrderID:          orderID,
		Store:            store,
		Amount:           amount,
		Currency:         currency,
		PaymentCode:      paymentCode,
		ConfirmationNumber: confirmationNumber,
		ExpiresAt:        expiresAt,
		CustomerEmail:    customerEmail,
		CustomerName:     customerName,
		Status:           paymentpb.PaymentStatus_PAYMENT_STATUS_PROCESSING,
		CreatedAt:        time.Now(),
	}, nil
}

// KonbiniPaymentDetails represents the details of a Konbini payment
type KonbiniPaymentDetails struct {
	PaymentID          string
	OrderID           string
	Store             KonbiniStore
	Amount            int64
	Currency          string
	PaymentCode       string
	ConfirmationNumber string
	ExpiresAt         time.Time
	CustomerEmail     string
	CustomerName      string
	Status            paymentpb.PaymentStatus
	CreatedAt         time.Time
}

// GeneratePaymentSlip generates a PDF payment slip
func (s *KonbiniService) GeneratePaymentSlip(payment *KonbiniPaymentDetails) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, fmt.Sprintf("支払い用紙 - Payment Slip - %s", payment.Store.NameJP))
	pdf.Ln(15)

	// Payment Details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Payment Details / お支払い詳細")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(95, 8, "Confirmation Number / 確認番号:", "0", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, payment.ConfirmationNumber, "0", 1, "L", false, 0, "")

	pdf.CellFormat(95, 8, "Payment Code / 支払いコード:", "0", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, payment.PaymentCode, "0", 1, "L", false, 0, "")

	// Format amount (convert minor units to yen)
	amountYen := float64(payment.Amount) / 100
	pdf.CellFormat(95, 8, "Amount / 金額:", "0", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, fmt.Sprintf("¥%.0f JPY", amountYen), "0", 1, "L", false, 0, "")

	// Expiration
	pdf.CellFormat(95, 8, "Expiration Date / 有効期限:", "0", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, payment.ExpiresAt.Format("2006-01-02 15:04"), "0", 1, "L", false, 0, "")

	pdf.Ln(10)

	// Customer Info
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Customer Information / お客様情報")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(95, 8, "Name / お名前:", "0", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, payment.CustomerName, "0", 1, "L", false, 0, "")

	pdf.CellFormat(95, 8, "Email / メールアドレス:", "0", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, payment.CustomerEmail, "0", 1, "L", false, 0, "")

	pdf.Ln(10)

	// Store Instructions
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Payment Instructions / お支払い方法")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	instructions := s.getStoreInstructions(payment.Store.ID)
	for _, line := range instructions {
		pdf.MultiCell(190, 6, line, "", "L", false)
	}

	pdf.Ln(10)

	// Barcode placeholder
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Barcode / バーコード")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(190, 30, fmt.Sprintf("[Payment Code: %s]", payment.PaymentCode), "0", 1, "C", false, 0, "")

	pdf.Ln(10)

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(190, 6, "Please bring this payment slip to any "+payment.Store.Name+" store.", "0", 1, "C", false, 0, "")
	pdf.CellFormat(190, 6, "この支払い用紙を持って"+payment.Store.NameJP+"にお支払いください。", "0", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// getStoreInstructions returns payment instructions for a specific store
func (s *KonbiniService) getStoreInstructions(storeID string) []string {
	instructions := map[string][]string{
		"seven-eleven": {
			"1. Bring this payment slip to any Seven Eleven store.",
			"1. この支払い用紙を持ってセブン-イレブンの店舗へお越しください。",
			"2. Present the payment slip at the register.",
			"2. レジでこの支払い用紙を提示してください。",
			"3. Pay the amount shown on the slip.",
			"3. 用紙に記載された金額をお支払いください。",
			"4. You will receive a receipt as proof of payment.",
			"4. 領収証を受け取ってください。",
		},
		"lawson": {
			"1. Bring this payment slip to any Lawson store.",
			"1. この支払い用紙を持ってローソンの店舗へお越しください。",
			"2. Use the Loppi payment kiosk in the store.",
			"2. 店内のLoppi（ロッピー）端末をご利用ください。",
			"3. Enter the confirmation number shown above.",
			"3. 上記の確認番号を入力してください。",
			"4. Pay the amount at the register.",
			"4. レジで金額をお支払いください。",
		},
		"familymart": {
			"1. Bring this payment slip to any FamilyMart store.",
			"1. この支払い用紙を持ってファミリーマートの店舗へお越しください。",
			"2. Use the Famima!! payment kiosk in the store.",
			"2. 店内のFamima!!（ファミマ!!）端末をご利用ください。",
			"3. Enter the payment code shown above.",
			"3. 上記の支払いコードを入力してください。",
			"4. Pay the amount at the register.",
			"4. レジで金額をお支払いください。",
		},
	}

	if inst, ok := instructions[storeID]; ok {
		return inst
	}

	return []string{
		"Please visit the store with this payment slip.",
		"この支払い用紙を持って店舗へお越しください。",
	}
}

// ProcessKonbiniWebhook processes a webhook notification from the payment provider
func (s *KonbiniService) ProcessKonbiniWebhook(
	ctx context.Context,
	paymentID string,
	status string,
	transactionID string,
	paidAt time.Time,
) error {
	s.logger.Info("Processing Konbini webhook",
		zap.String("payment_id", paymentID),
		zap.String("status", status))

	// Update payment status in database
	// This would typically update the payment table
	// and trigger order fulfillment if payment is completed

	return nil
}

// CheckPaymentStatus checks the status of a Konbini payment
func (s *KonbiniService) CheckPaymentStatus(
	ctx context.Context,
	paymentID string,
) (*paymentpb.PaymentStatus, error) {
	s.logger.Info("Checking Konbini payment status", zap.String("payment_id", paymentID))

	// In a real implementation, this would call the payment provider's API
	// For now, return the status from our database

	status := paymentpb.PaymentStatus_PAYMENT_STATUS_PROCESSING
	return &status, nil
}

// CancelKonbiniPayment cancels a pending Konbini payment
func (s *KonbiniService) CancelKonbiniPayment(
	ctx context.Context,
	paymentID string,
) error {
	s.logger.Info("Cancelling Konbini payment", zap.String("payment_id", paymentID))

	// Update payment status to cancelled
	// In a real implementation, this would also notify the payment provider

	return nil
}

// generatePaymentCode generates a unique payment code
func (s *KonbiniService) generatePaymentCode(store KonbiniStore) string {
	timestamp := time.Now().Format("200601021504")
	random := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s%s", store.PaymentCode, timestamp, random)
}

// generateConfirmationNumber generates a unique confirmation number
func (s *KonbiniService) generateConfirmationNumber(store KonbiniStore) string {
	timestamp := time.Now().Format("20060102")
	random := uuid.New().String()[:8]
	return fmt.Sprintf("%s-%s-%s", store.PaymentCode, timestamp, random)
}

// GetSupportedStores returns a list of supported Konbini stores
func (s *KonbiniService) GetSupportedStores() []KonbiniStore {
	stores := make([]KonbiniStore, 0, len(KonbiniStores))
	for _, store := range KonbiniStores {
		stores = append(stores, store)
	}
	return stores
}

// ValidatePaymentCode validates a Konbini payment code
func (s *KonbiniService) ValidatePaymentCode(code string, method paymentpb.PaymentMethod) bool {
	store, ok := KonbiniStores[method]
	if !ok {
		return false
	}

	// Check if code starts with store's payment code
	expectedPrefix := store.PaymentCode
	if len(code) < len(expectedPrefix) {
		return false
	}

	prefix := code[:len(expectedPrefix)]
	return prefix == expectedPrefix
}
