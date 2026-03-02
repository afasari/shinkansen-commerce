package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	paymentpb "github.com/afasari/shinkansen-commerce/gen/proto/go/payment"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
)

// MockPaymentService is a mock implementation of paymentpb.PaymentServiceServer
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, req *paymentpb.CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentpb.CreatePaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPayment(ctx context.Context, req *paymentpb.GetPaymentRequest) (*paymentpb.GetPaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentpb.GetPaymentResponse), args.Error(1)
}

func (m *MockPaymentService) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentpb.ProcessPaymentResponse), args.Error(1)
}

func (m *MockPaymentService) RefundPayment(ctx context.Context, req *paymentpb.RefundPaymentRequest) (*sharedpb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sharedpb.Empty), args.Error(1)
}

func TestHandler_CreatePayment(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPaymentService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &paymentpb.CreatePaymentRequest{
			OrderId: "order-123",
			Method:  paymentpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
		}

		expectedResp := &paymentpb.CreatePaymentResponse{
			PaymentId: "payment-456",
			Status:    paymentpb.PaymentStatus_PAYMENT_STATUS_PENDING,
		}

		mockService.On("CreatePayment", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.CreatePayment(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetPayment(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPaymentService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &paymentpb.GetPaymentRequest{
			PaymentId: "payment-123",
		}

		expectedResp := &paymentpb.GetPaymentResponse{
			Payment: &paymentpb.Payment{
				Id:     "payment-123",
				Status: paymentpb.PaymentStatus_PAYMENT_STATUS_PENDING,
			},
		}

		mockService.On("GetPayment", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.GetPayment(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ProcessPayment(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPaymentService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &paymentpb.ProcessPaymentRequest{
			PaymentId:   "payment-123",
			PaymentData: map[string]string{"card_token": "tok_123"},
		}

		expectedResp := &paymentpb.ProcessPaymentResponse{
			Status:        paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED,
			TransactionId: "txn-456",
		}

		mockService.On("ProcessPayment", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.ProcessPayment(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_RefundPayment(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPaymentService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &paymentpb.RefundPaymentRequest{
			PaymentId: "payment-123",
		}

		expectedResp := &sharedpb.Empty{}

		mockService.On("RefundPayment", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.RefundPayment(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}
