package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	orderpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/order"
	productpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/product"
	sharedpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/shared"
	"github.com/shinkansen-commerce/shinkansen/services/order-service/internal/cache"
	"github.com/shinkansen-commerce/shinkansen/services/order-service/internal/db"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	queries       *db.Queries
	productClient productpb.ProductServiceClient
	cache         cache.Cache
	logger        *zap.Logger
}

func NewOrderService(queries *db.Queries, productClient productpb.ProductServiceClient, cacheClient cache.Cache, logger *zap.Logger) *OrderService {
	return &OrderService{
		queries:       queries,
		productClient: productClient,
		cache:         cacheClient,
		logger:        logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	s.logger.Info("Creating order", zap.String("user_id", req.UserId))

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	var totalUnits int64
	for _, item := range req.Items {
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{ProductId: item.ProductId})
		if err != nil {
			s.logger.Error("Failed to get product", zap.String("product_id", item.ProductId), zap.Error(err))
			return nil, fmt.Errorf("failed to get product %s: %w", item.ProductId, err)
		}

		if productResp.Product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("product %s has insufficient stock", item.ProductId)
		}

		totalUnits += productResp.Product.Price.Units * int64(item.Quantity)
	}

	orderID, err := s.queries.CreateOrder(ctx, db.CreateOrderParams{
		UserID:          uuid.MustParse(req.UserId),
		Status:          "pending",
		TotalUnits:      totalUnits,
		TotalCurrency:   "JPY",
		ShippingAddress: req.ShippingAddress.AsMap(),
	})
	if err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range req.Items {
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{ProductId: item.ProductId})
		if err != nil {
			s.logger.Warn("Failed to get product price", zap.String("product_id", item.ProductId), zap.Error(err))
			continue
		}

		err = s.queries.AddOrderItem(ctx, db.AddOrderItemParams{
			OrderID:       orderID,
			ProductID:     uuid.MustParse(item.ProductId),
			Quantity:      item.Quantity,
			PriceUnits:    productResp.Product.Price.Units,
			PriceCurrency: productResp.Product.Price.Currency,
		})
		if err != nil {
			s.logger.Error("Failed to add order item", zap.Error(err))
		}
	}

	return &orderpb.CreateOrderResponse{
		OrderId: orderID.String(),
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	s.logger.Info("Getting order", zap.String("order_id", req.OrderId))

	cacheKey := cache.OrderCacheKey(req.OrderId)
	var cached db.Order

	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("Order cache hit", zap.String("order_id", req.OrderId))
		return &orderpb.GetOrderResponse{
			Order: s.orderToProto(cached),
		}, nil
	}

	s.logger.Debug("Order cache miss", zap.String("order_id", req.OrderId))

	orderID := uuid.MustParse(req.OrderId)
	order, err := s.queries.GetOrder(ctx, orderID)
	if err != nil {
		s.logger.Error("Failed to get order", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if err := s.cache.Set(ctx, cacheKey, order, cache.DefaultTTL); err != nil {
		s.logger.Warn("Failed to cache order", zap.Error(err))
	}

	return &orderpb.GetOrderResponse{
		Order: s.orderToProto(order),
	}, nil
}

func (s *OrderService) ListUserOrders(ctx context.Context, req *orderpb.ListUserOrdersRequest) (*orderpb.ListUserOrdersResponse, error) {
	s.logger.Info("Listing user orders", zap.String("user_id", req.UserId))

	orders, err := s.queries.ListUserOrders(ctx, db.ListUserOrdersParams{
		UserID: uuid.MustParse(req.UserId),
		Status: req.Status,
		Limit:  req.Pagination.Limit,
		Offset: (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		s.logger.Error("Failed to list orders", zap.Error(err))
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	var orderList []*orderpb.Order
	for _, o := range orders {
		orderList = append(orderList, s.orderToProto(o))
	}

	return &orderpb.ListUserOrdersResponse{
		Orders:     orderList,
		Pagination: req.Pagination,
	}, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	s.logger.Info("Updating order status", zap.String("order_id", req.OrderId), zap.String("status", req.Status))

	orderID := uuid.MustParse(req.OrderId)
	if err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderID,
		Status: req.Status,
	}); err != nil {
		s.logger.Error("Failed to update order status", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	order, err := s.queries.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated order: %w", err)
	}

	return &orderpb.UpdateOrderStatusResponse{
		Order: s.orderToProto(order),
	}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Cancelling order", zap.String("order_id", req.OrderId))

	orderID := uuid.MustParse(req.OrderId)
	if err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderID,
		Status: "cancelled",
	}); err != nil {
		s.logger.Error("Failed to cancel order", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	return &sharedpb.Empty{}, nil
}

func (s *OrderService) orderToProto(o db.Order) *orderpb.Order {
	return &orderpb.Order{
		Id:              o.ID.String(),
		UserId:          o.UserID.String(),
		Status:          o.Status,
		Total:           &sharedpb.Money{Units: o.TotalUnits, Currency: o.TotalCurrency},
		ShippingAddress: s.addressToProto(o.ShippingAddress),
		CreatedAt:       protoTime(o.CreatedAt),
		UpdatedAt:       protoTime(o.UpdatedAt),
	}
}

func (s *OrderService) addressToProto(addr map[string]interface{}) *orderpb.Address {
	if addr == nil {
		return &orderpb.Address{}
	}

	address := &orderpb.Address{
		Street:  getStr(addr, "street"),
		City:    getStr(addr, "city"),
		State:   getStr(addr, "state"),
		Zip:     getStr(addr, "zip"),
		Country: getStr(addr, "country"),
	}
	return address
}

func getStr(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func protoTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
