package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	paymentpb "github.com/afasari/shinkansen-commerce/gen/proto/go/payment"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/payment-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/payment-service/internal/db"
)

// MockQuerier is a mock implementation of db.Querier
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreatePayment(ctx context.Context, params db.CreatePaymentParams) (uuid.UUID, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQuerier) GetPayment(ctx context.Context, id uuid.UUID) (db.Payment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Payment), args.Error(1)
}

func (m *MockQuerier) UpdatePaymentData(ctx context.Context, params db.UpdatePaymentDataParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) UpdatePaymentStatus(ctx context.Context, params db.UpdatePaymentStatusParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (db.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return db.Payment{}, args.Error(1)
	}
	return args.Get(0).(db.Payment), args.Error(1)
}

func (m *MockQuerier) ListPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]db.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return []db.Payment{}, args.Error(1)
	}
	return args.Get(0).([]db.Payment), args.Error(1)
}

func TestPaymentService_CreatePayment(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful payment creation", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		orderID := uuid.New()
		paymentID := uuid.New()

		req := &paymentpb.CreatePaymentRequest{
			OrderId: orderID.String(),
			Method:  paymentpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
			Amount:  &sharedpb.Money{Units: 10000, Currency: "JPY"},
		}

		mockQueries.On("CreatePayment", mock.Anything, mock.AnythingOfType("db.CreatePaymentParams")).Return(paymentID, nil)

		resp, err := service.CreatePayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, paymentID.String(), resp.PaymentId)
		assert.Equal(t, paymentpb.PaymentStatus_PAYMENT_STATUS_PENDING, resp.Status)
		mockQueries.AssertExpectations(t)
	})

	t.Run("payment creation with different methods", func(t *testing.T) {
		methods := []paymentpb.PaymentMethod{
			paymentpb.PaymentMethod_PAYMENT_METHOD_PAYPAY,
			paymentpb.PaymentMethod_PAYMENT_METHOD_RAKUTEN_PAY,
			paymentpb.PaymentMethod_PAYMENT_METHOD_KONBINI_SEVENELEVEN,
			paymentpb.PaymentMethod_PAYMENT_METHOD_KONBINI_LAWSON,
		}

		for _, method := range methods {
			t.Run(method.String(), func(t *testing.T) {
				mockQueries := new(MockQuerier)
				mockCache := new(cache.MockCache)
				service := NewPaymentService(mockQueries, mockCache, logger)

				orderID := uuid.New()

				req := &paymentpb.CreatePaymentRequest{
					OrderId: orderID.String(),
					Method:  method,
					Amount:  &sharedpb.Money{Units: 5000, Currency: "JPY"},
				}

				mockQueries.On("CreatePayment", mock.Anything, mock.AnythingOfType("db.CreatePaymentParams")).Return(uuid.New(), nil).Once()

				resp, err := service.CreatePayment(context.Background(), req)

				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, paymentpb.PaymentStatus_PAYMENT_STATUS_PENDING, resp.Status)
			})
		}
	})
}

func TestPaymentService_GetPayment(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful payment retrieval", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_CREDIT_CARD",
			AmountMinor: 10000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*db.Payment")).Return(cache.ErrCacheMiss)
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("db.Payment"), mock.AnythingOfType("time.Duration")).Return(nil)

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)

		req := &paymentpb.GetPaymentRequest{
			PaymentId: paymentID.String(),
		}

		resp, err := service.GetPayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Payment)
		assert.Equal(t, paymentID.String(), resp.Payment.Id)
		mockQueries.AssertExpectations(t)
	})

	t.Run("payment retrieval from cache", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()

		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*db.Payment")).Return(nil)

		req := &paymentpb.GetPaymentRequest{
			PaymentId: paymentID.String(),
		}

		resp, err := service.GetPayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Payment)
	})
}

func TestPaymentService_ProcessPayment(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful credit card payment", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_CREDIT_CARD",
			AmountMinor: 10000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)
		mockQueries.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("db.UpdatePaymentDataParams")).Return(nil)
		mockQueries.On("UpdatePaymentStatus", mock.Anything, mock.AnythingOfType("db.UpdatePaymentStatusParams")).Return(nil)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("[]string")).Return(nil).Twice()

		req := &paymentpb.ProcessPaymentRequest{
			PaymentId:  paymentID.String(),
			PaymentData: map[string]string{
				"card_number": "4111111111111111",
				"expiry":     "12/25",
				"cvv":        "123",
			},
		}

		resp, err := service.ProcessPayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED, resp.Status)
		assert.NotEmpty(t, resp.TransactionId)
		assert.Contains(t, resp.TransactionId, "CC-")
		mockQueries.AssertExpectations(t)
	})

	t.Run("successful PayPay payment", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_PAYPAY",
			AmountMinor: 5000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)
		mockQueries.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("db.UpdatePaymentDataParams")).Return(nil)
		mockQueries.On("UpdatePaymentStatus", mock.Anything, mock.AnythingOfType("db.UpdatePaymentStatusParams")).Return(nil)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("[]string")).Return(nil).Twice()

		req := &paymentpb.ProcessPaymentRequest{
			PaymentId:  paymentID.String(),
			PaymentData: map[string]string{
				"redirect_url": "https://paypay.example.com/redirect",
			},
		}

		resp, err := service.ProcessPayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED, resp.Status)
		assert.Contains(t, resp.TransactionId, "PAYPAY-")
	})

	t.Run("successful Konbini payment", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_KONBINI_SEVENELEVEN",
			AmountMinor: 3000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)
		mockQueries.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("db.UpdatePaymentDataParams")).Return(nil)
		mockQueries.On("UpdatePaymentStatus", mock.Anything, mock.AnythingOfType("db.UpdatePaymentStatusParams")).Return(nil)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("[]string")).Return(nil).Twice()

		req := &paymentpb.ProcessPaymentRequest{
			PaymentId:  paymentID.String(),
			PaymentData: map[string]string{
				"store": "seven_eleven",
			},
		}

		resp, err := service.ProcessPayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, paymentpb.PaymentStatus_PAYMENT_STATUS_PROCESSING, resp.Status)
		assert.Contains(t, resp.TransactionId, "KONBINI-")
	})

	t.Run("cannot process non-pending payment", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_CREDIT_CARD",
			AmountMinor: 10000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_COMPLETED",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)

		req := &paymentpb.ProcessPaymentRequest{
			PaymentId:   paymentID.String(),
			PaymentData: map[string]string{},
		}

		resp, err := service.ProcessPayment(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "not in pending status")
		mockQueries.AssertExpectations(t)
	})

	t.Run("unsupported payment method", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_UNSUPPORTED",
			AmountMinor: 10000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)

		req := &paymentpb.ProcessPaymentRequest{
			PaymentId:   paymentID.String(),
			PaymentData: map[string]string{},
		}

		resp, err := service.ProcessPayment(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "unsupported payment method")
		mockQueries.AssertExpectations(t)
	})
}

func TestPaymentService_RefundPayment(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful refund", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()
		transactionID := "CC-12345-789"

		mockPayment := db.Payment{
			ID:            paymentID,
			OrderID:       orderID,
			Method:        "PAYMENT_METHOD_CREDIT_CARD",
			AmountMinor:   10000,
			Currency:      "JPY",
			Status:        "PAYMENT_STATUS_COMPLETED",
			TransactionID: &transactionID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)
		mockQueries.On("UpdatePaymentStatus", mock.Anything, mock.AnythingOfType("db.UpdatePaymentStatusParams")).Return(nil)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("[]string")).Return(nil).Twice()

		req := &paymentpb.RefundPaymentRequest{
			PaymentId: paymentID.String(),
			Amount:    &sharedpb.Money{Units: 10000, Currency: "JPY"},
		}

		resp, err := service.RefundPayment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockQueries.AssertExpectations(t)
	})

	t.Run("cannot refund non-completed payment", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()

		mockPayment := db.Payment{
			ID:        paymentID,
			OrderID:   orderID,
			Method:    "PAYMENT_METHOD_CREDIT_CARD",
			AmountMinor: 10000,
			Currency:  "JPY",
			Status:    "PAYMENT_STATUS_PENDING",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockQueries.On("GetPayment", mock.Anything, paymentID).Return(mockPayment, nil)

		req := &paymentpb.RefundPaymentRequest{
			PaymentId: paymentID.String(),
			Amount:    &sharedpb.Money{Units: 10000, Currency: "JPY"},
		}

		resp, err := service.RefundPayment(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "cannot be refunded")
		mockQueries.AssertExpectations(t)
	})
}

func TestPaymentService_paymentToProto(t *testing.T) {
	logger := zap.NewNop()

	t.Run("convert payment to proto", func(t *testing.T) {
		mockQueries := new(MockQuerier)
		mockCache := new(cache.MockCache)
		service := NewPaymentService(mockQueries, mockCache, logger)

		paymentID := uuid.New()
		orderID := uuid.New()
		transactionID := "txn-12345"

		payment := db.Payment{
			ID:            paymentID,
			OrderID:       orderID,
			Method:        "PAYMENT_METHOD_CREDIT_CARD",
			AmountMinor:   10000,
			Currency:      "JPY",
			Status:        "PAYMENT_STATUS_COMPLETED",
			TransactionID: &transactionID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Access the private method via testing
		proto := service.paymentToProto(payment)

		assert.Equal(t, paymentID.String(), proto.Id)
		assert.Equal(t, orderID.String(), proto.OrderId)
		assert.Equal(t, paymentpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD, proto.Method)
		assert.Equal(t, int64(10000), proto.Amount.Units)
		assert.Equal(t, "JPY", proto.Amount.Currency)
		assert.Equal(t, paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED, proto.Status)
		assert.Equal(t, transactionID, proto.TransactionId)
	})
}
