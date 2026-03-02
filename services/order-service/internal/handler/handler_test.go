package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
)

// MockOrderService is a mock implementation of orderpb.OrderServiceServer
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orderpb.CreateOrderResponse), args.Error(1)
}

func (m *MockOrderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orderpb.GetOrderResponse), args.Error(1)
}

func (m *MockOrderService) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orderpb.ListOrdersResponse), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*sharedpb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sharedpb.Empty), args.Error(1)
}

func (m *MockOrderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*sharedpb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sharedpb.Empty), args.Error(1)
}

func (m *MockOrderService) ApplyPoints(ctx context.Context, req *orderpb.ApplyPointsRequest) (*orderpb.ApplyPointsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orderpb.ApplyPointsResponse), args.Error(1)
}

func (m *MockOrderService) ReserveDeliverySlot(ctx context.Context, req *orderpb.ReserveDeliverySlotRequest) (*orderpb.ReserveDeliverySlotResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orderpb.ReserveDeliverySlotResponse), args.Error(1)
}

func TestHandler_CreateOrder(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.CreateOrderRequest{
			UserId: "user-123",
			Items: []*orderpb.CreateOrderItem{
				{ProductId: "prod-1", Quantity: 2},
			},
		}

		expectedResp := &orderpb.CreateOrderResponse{
			OrderId: "order-456",
			Status:  orderpb.OrderStatus_ORDER_STATUS_PENDING,
		}

		mockService.On("CreateOrder", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.CreateOrder(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetOrder(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.GetOrderRequest{
			OrderId: "order-123",
		}

		expectedResp := &orderpb.GetOrderResponse{
			Order: &orderpb.Order{
				Id:     "order-123",
				Status: orderpb.OrderStatus_ORDER_STATUS_PENDING,
			},
		}

		mockService.On("GetOrder", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.GetOrder(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("error from service is propagated", func(t *testing.T) {
		req := &orderpb.GetOrderRequest{
			OrderId: "invalid-order",
		}

		mockService.On("GetOrder", mock.Anything, req).Return(nil, status.Error(codes.NotFound, "order not found"))

		resp, err := handler.GetOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ListOrders(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.ListOrdersRequest{
			UserId: "user-123",
			Pagination: &sharedpb.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		expectedResp := &orderpb.ListOrdersResponse{
			Orders: []*orderpb.Order{
				{Id: "order-1"},
				{Id: "order-2"},
			},
			Pagination: req.Pagination,
		}

		mockService.On("ListOrders", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.ListOrders(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateOrderStatus(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.UpdateOrderStatusRequest{
			OrderId: "order-123",
			Status:  orderpb.OrderStatus_ORDER_STATUS_CONFIRMED,
		}

		expectedResp := &sharedpb.Empty{}

		mockService.On("UpdateOrderStatus", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.UpdateOrderStatus(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_CancelOrder(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.CancelOrderRequest{
			OrderId: "order-123",
		}

		expectedResp := &sharedpb.Empty{}

		mockService.On("CancelOrder", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.CancelOrder(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ApplyPoints(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.ApplyPointsRequest{
			OrderId: "order-123",
			Points:  500,
		}

		expectedResp := &orderpb.ApplyPointsResponse{
			Success: true,
		}

		mockService.On("ApplyPoints", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.ApplyPoints(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ReserveDeliverySlot(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockOrderService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &orderpb.ReserveDeliverySlotRequest{
			OrderId: "order-123",
			SlotId:  "slot-456",
		}

		expectedResp := &orderpb.ReserveDeliverySlotResponse{
			ReservationId: "res-789",
		}

		mockService.On("ReserveDeliverySlot", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.ReserveDeliverySlot(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}
