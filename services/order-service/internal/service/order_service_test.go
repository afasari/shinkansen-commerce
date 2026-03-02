package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/pkg/pgutil"
)

// MockQuerier is a mock implementation of db.Querier
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateOrder(ctx context.Context, params db.CreateOrderParams) (pgtype.UUID, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(pgtype.UUID), args.Error(1)
}

func (m *MockQuerier) AddOrderItem(ctx context.Context, params db.AddOrderItemParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) GetOrder(ctx context.Context, id pgtype.UUID) (db.OrdersOrders, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.OrdersOrders), args.Error(1)
}

func (m *MockQuerier) GetOrderItems(ctx context.Context, orderID pgtype.UUID) ([]db.OrdersOrderItems, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]db.OrdersOrderItems), args.Error(1)
}

func (m *MockQuerier) ListUserOrders(ctx context.Context, params db.ListUserOrdersParams) ([]db.OrdersOrders, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]db.OrdersOrders), args.Error(1)
}

func (m *MockQuerier) GetOrderItem(ctx context.Context, orderID pgtype.UUID) (db.OrdersOrderItems, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(db.OrdersOrderItems), args.Error(1)
}

func (m *MockQuerier) UpdateOrderStatus(ctx context.Context, params db.UpdateOrderStatusParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) UpdateOrderWithPoints(ctx context.Context, params db.UpdateOrderWithPointsParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// MockProductClient is a mock implementation of productpb.ProductServiceClient
type MockProductClient struct {
	mock.Mock
}

func (m *MockProductClient) GetProduct(ctx context.Context, req *productpb.GetProductRequest, opts ...grpc.CallOption) (*productpb.GetProductResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*productpb.GetProductResponse), args.Error(1)
}

func (m *MockProductClient) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest, opts ...grpc.CallOption) (*productpb.CreateProductResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*productpb.CreateProductResponse), args.Error(1)
}

func (m *MockProductClient) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest, opts ...grpc.CallOption) (*sharedpb.Empty, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sharedpb.Empty), args.Error(1)
}

func (m *MockProductClient) GetProductVariants(ctx context.Context, req *productpb.GetProductVariantsRequest, opts ...grpc.CallOption) (*productpb.GetProductVariantsResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*productpb.GetProductVariantsResponse), args.Error(1)
}

func (m *MockProductClient) ListProducts(ctx context.Context, req *productpb.ListProductsRequest, opts ...grpc.CallOption) (*productpb.ListProductsResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*productpb.ListProductsResponse), args.Error(1)
}

func (m *MockProductClient) SearchProducts(ctx context.Context, req *productpb.SearchProductsRequest, opts ...grpc.CallOption) (*productpb.SearchProductsResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*productpb.SearchProductsResponse), args.Error(1)
}

func (m *MockProductClient) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest, opts ...grpc.CallOption) (*productpb.UpdateProductResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*productpb.UpdateProductResponse), args.Error(1)
}

// MockCache is a mock implementation of cache.Cache
type MockCache struct {
	mock.Mock
}

func TestOrderService_CreateOrder(t *testing.T) {
	logger := zap.NewNop()
	mockQueries := new(MockQuerier)
	mockProductClient := new(MockProductClient)
	mockCache := new(cache.MockCache)

	service := NewOrderService(mockQueries, mockProductClient, mockCache, logger)

	userID := uuid.New().String()
	productID := uuid.New().String()
	orderID := uuid.New()

	t.Run("successful order creation", func(t *testing.T) {
		req := &orderpb.CreateOrderRequest{
			UserId: userID,
			Items: []*orderpb.CreateOrderItem{
				{
					ProductId: productID,
					Quantity:  2,
					VariantId: "",
				},
			},
			ShippingAddress: &orderpb.ShippingAddress{
				Name:         "John Doe",
				Phone:        "090-1234-5678",
				PostalCode:   "100-0001",
				Prefecture:   "Tokyo",
				City:         "Chiyoda-ku",
				AddressLine1: "1-1-1 Otemachi",
				AddressLine2: "",
			},
			PaymentMethod: orderpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
		}

		// Mock product service call
		mockProductClient.On("GetProduct", mock.Anything, mock.Anything, mock.Anything).Return(&productpb.GetProductResponse{
			Product: &productpb.Product{
				Id:           productID,
				Name:         "Test Product",
				Price:        &sharedpb.Money{Units: 1000, Currency: "JPY"},
				StockQuantity: 10,
			},
		}, nil).Twice()

		// Mock database calls
		mockQueries.On("CreateOrder", mock.Anything, mock.AnythingOfType("db.CreateOrderParams")).Return(pgutil.ToPG(orderID), nil)
		mockQueries.On("AddOrderItem", mock.Anything, mock.AnythingOfType("db.AddOrderItemParams")).Return(nil)

		resp, err := service.CreateOrder(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.OrderId)
		assert.Equal(t, orderpb.OrderStatus_ORDER_STATUS_PENDING, resp.Status)
		assert.NotEmpty(t, resp.OrderNumber)
	})

	t.Run("order with no items fails", func(t *testing.T) {
		req := &orderpb.CreateOrderRequest{
			UserId:          userID,
			Items:           []*orderpb.CreateOrderItem{},
			ShippingAddress: &orderpb.ShippingAddress{},
			PaymentMethod:   orderpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
		}

		resp, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "at least one item")
	})

	t.Run("order with insufficient stock fails", func(t *testing.T) {
		req := &orderpb.CreateOrderRequest{
			UserId: userID,
			Items: []*orderpb.CreateOrderItem{
				{
					ProductId: productID,
					Quantity:  10,
				},
			},
			ShippingAddress: &orderpb.ShippingAddress{},
			PaymentMethod:   orderpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
		}

		mockProductClient.On("GetProduct", mock.Anything, mock.Anything, mock.Anything).Return(&productpb.GetProductResponse{
			Product: &productpb.Product{
				Id:           productID,
				Name:         "Test Product",
				Price:        &sharedpb.Money{Units: 1000, Currency: "JPY"},
				StockQuantity: 5,
			},
		}, nil).Once()

		resp, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "insufficient stock")
	})
}

func TestOrderService_GetOrder(t *testing.T) {
	logger := zap.NewNop()
	mockQueries := new(MockQuerier)
	mockProductClient := new(MockProductClient)
	mockCache := new(cache.MockCache)

	service := NewOrderService(mockQueries, mockProductClient, mockCache, logger)

	orderID := uuid.New()

	t.Run("successful order retrieval", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*db.OrdersOrders")).Return(cache.ErrCacheMiss)
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockOrder := db.OrdersOrders{
			ID:              pgutil.ToPG(orderID),
			OrderNumber:     "ORD-12345",
			UserID:          pgutil.ToPG(uuid.New()),
			Status:          int32(orderpb.OrderStatus_ORDER_STATUS_PENDING),
			SubtotalUnits:   1000,
			SubtotalCurrency: "JPY",
			TaxUnits:        100,
			TaxCurrency:     "JPY",
			TotalUnits:      1100,
			TotalCurrency:   "JPY",
			CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		mockQueries.On("GetOrder", mock.Anything, pgutil.ToPG(orderID)).Return(mockOrder, nil)
		mockQueries.On("GetOrderItems", mock.Anything, pgutil.ToPG(orderID)).Return([]db.OrdersOrderItems{}, nil)

		req := &orderpb.GetOrderRequest{
			OrderId: orderID.String(),
		}

		resp, err := service.GetOrder(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Order)
		assert.Equal(t, orderID.String(), resp.Order.Id)
		assert.Equal(t, "ORD-12345", resp.Order.OrderNumber)
	})
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	logger := zap.NewNop()
	mockQueries := new(MockQuerier)
	mockProductClient := new(MockProductClient)
	mockCache := new(cache.MockCache)

	service := NewOrderService(mockQueries, mockProductClient, mockCache, logger)

	orderID := uuid.New()

	t.Run("successful status update", func(t *testing.T) {
		mockCache.On("Delete", mock.Anything, mock.Anything).Return(nil)
		mockQueries.On("GetOrder", mock.Anything, mock.Anything).Return(db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			Status: 1, // PENDING
		}, nil)
		mockQueries.On("UpdateOrderStatus", mock.Anything, mock.AnythingOfType("db.UpdateOrderStatusParams")).Return(nil)

		req := &orderpb.UpdateOrderStatusRequest{
			OrderId: orderID.String(),
			Status:  orderpb.OrderStatus_ORDER_STATUS_CONFIRMED,
		}

		resp, err := service.UpdateOrderStatus(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestOrderService_CancelOrder(t *testing.T) {
	logger := zap.NewNop()
	mockQueries := new(MockQuerier)
	mockProductClient := new(MockProductClient)
	mockCache := new(cache.MockCache)

	service := NewOrderService(mockQueries, mockProductClient, mockCache, logger)

	orderID := uuid.New()

	t.Run("successful order cancellation", func(t *testing.T) {
		mockCache.On("Delete", mock.Anything, mock.Anything).Return(nil)
		mockQueries.On("GetOrder", mock.Anything, mock.Anything).Return(db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			Status: 1, // PENDING
		}, nil)
		mockQueries.On("UpdateOrderStatus", mock.Anything, mock.MatchedBy(func(params db.UpdateOrderStatusParams) bool {
			return params.Status == int32(orderpb.OrderStatus_ORDER_STATUS_CANCELLED)
		})).Return(nil)

		req := &orderpb.CancelOrderRequest{
			OrderId: orderID.String(),
		}

		resp, err := service.CancelOrder(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestOrderService_ApplyPoints(t *testing.T) {
	logger := zap.NewNop()
	mockQueries := new(MockQuerier)
	mockProductClient := new(MockProductClient)
	mockCache := new(cache.MockCache)

	service := NewOrderService(mockQueries, mockProductClient, mockCache, logger)

	orderID := uuid.New()
	userID := uuid.New()

	t.Run("successful points application", func(t *testing.T) {
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

		mockOrder := db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			UserID: pgutil.ToPG(userID),
			Status: int32(orderpb.OrderStatus_ORDER_STATUS_PENDING),
		}

		mockQueries.On("GetOrder", mock.Anything, pgutil.ToPG(orderID)).Return(mockOrder, nil).Once()

		req := &orderpb.ApplyPointsRequest{
			OrderId: orderID.String(),
			Points:  500,
		}

		resp, err := service.ApplyPoints(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, int64(5000), resp.YenValue.Units)
	})

	t.Run("cannot apply points to non-pending order", func(t *testing.T) {
		mockOrder := db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			UserID: pgutil.ToPG(userID),
			Status: int32(orderpb.OrderStatus_ORDER_STATUS_CANCELLED),
		}

		mockQueries.On("GetOrder", mock.Anything, pgutil.ToPG(orderID)).Return(mockOrder, nil).Once()

		req := &orderpb.ApplyPointsRequest{
			OrderId: orderID.String(),
			Points:  500,
		}

		resp, err := service.ApplyPoints(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "non-pending")
	})

	t.Run("cannot apply more than 10000 points", func(t *testing.T) {
		mockOrder := db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			UserID: pgutil.ToPG(userID),
			Status: int32(orderpb.OrderStatus_ORDER_STATUS_PENDING),
		}

		mockQueries.On("GetOrder", mock.Anything, pgutil.ToPG(orderID)).Return(mockOrder, nil).Once()

		req := &orderpb.ApplyPointsRequest{
			OrderId: orderID.String(),
			Points:  15000,
		}

		resp, err := service.ApplyPoints(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "more than 10,000")
	})
}

func TestOrderService_ReserveDeliverySlot(t *testing.T) {
	logger := zap.NewNop()
	mockQueries := new(MockQuerier)
	mockProductClient := new(MockProductClient)
	mockCache := new(cache.MockCache)

	service := NewOrderService(mockQueries, mockProductClient, mockCache, logger)

	orderID := uuid.New()
	userID := uuid.New()

	t.Run("successful delivery slot reservation", func(t *testing.T) {
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		mockOrder := db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			UserID: pgutil.ToPG(userID),
			Status: int32(orderpb.OrderStatus_ORDER_STATUS_PENDING),
		}

		mockQueries.On("GetOrder", mock.Anything, pgutil.ToPG(orderID)).Return(mockOrder, nil)

		req := &orderpb.ReserveDeliverySlotRequest{
			OrderId: orderID.String(),
			SlotId:  "slot-123",
		}

		resp, err := service.ReserveDeliverySlot(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.ReservationId)
	})

	t.Run("missing slot id fails", func(t *testing.T) {
		mockOrder := db.OrdersOrders{
			ID:     pgutil.ToPG(orderID),
			UserID: pgutil.ToPG(userID),
			Status: int32(orderpb.OrderStatus_ORDER_STATUS_PENDING),
		}

		mockQueries.On("GetOrder", mock.Anything, pgutil.ToPG(orderID)).Return(mockOrder, nil)

		req := &orderpb.ReserveDeliverySlotRequest{
			OrderId: orderID.String(),
			SlotId:  "",
		}

		resp, err := service.ReserveDeliverySlot(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "required")
	})
}
