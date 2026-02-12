package service

import (
	"context"
	"fmt"
	"time"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/db"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	queries       db.Querier
	productClient productpb.ProductServiceClient
	cache         cache.Cache
	logger        *zap.Logger
}

func NewOrderService(queries db.Querier, productClient productpb.ProductServiceClient, cacheClient cache.Cache, logger *zap.Logger) *OrderService {
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

	var subtotalUnits int64

	for _, item := range req.Items {
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{ProductId: item.ProductId})
		if err != nil {
			s.logger.Error("Failed to get product", zap.String("product_id", item.ProductId), zap.Error(err))
			return nil, fmt.Errorf("failed to get product %s: %w", item.ProductId, err)
		}

		if productResp.Product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("product %s has insufficient stock", item.ProductId)
		}

		itemTotalUnits := productResp.Product.Price.Units * int64(item.Quantity)
		subtotalUnits += itemTotalUnits
	}

	taxUnits := int64(float64(subtotalUnits) * 0.10)
	totalUnits := subtotalUnits + taxUnits

	orderNumber := fmt.Sprintf("ORD-%d", uuid.New().ID())

	orderID, err := s.queries.CreateOrder(ctx, db.CreateOrderParams{
		UserId:           uuid.MustParse(req.UserId),
		Status:           int32(orderpb.OrderStatus_ORDER_STATUS_PENDING),
		SubtotalUnits:    subtotalUnits,
		SubtotalCurrency: "JPY",
		TaxUnits:         taxUnits,
		TaxCurrency:      "JPY",
		DiscountUnits:    0,
		DiscountCurrency: "JPY",
		TotalUnits:       totalUnits,
		TotalCurrency:    "JPY",
		PointsApplied:    0,
		ShippingAddress:  s.addressToMap(req.ShippingAddress),
		PaymentMethod:    int32(req.PaymentMethod),
	})
	if err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range req.Items {
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{ProductId: item.ProductId})
		if err != nil {
			s.logger.Error("Failed to get product for item", zap.String("product_id", item.ProductId), zap.Error(err))
			continue
		}

		itemTotalUnits := productResp.Product.Price.Units * int64(item.Quantity)

		err = s.queries.AddOrderItem(ctx, db.AddOrderItemParams{
			OrderId:            orderID,
			ProductId:          uuid.MustParse(item.ProductId),
			VariantId:          item.VariantId,
			ProductName:        productResp.Product.Name,
			Quantity:           item.Quantity,
			UnitPriceUnits:     productResp.Product.Price.Units,
			UnitPriceCurrency:  productResp.Product.Price.Currency,
			TotalPriceUnits:    itemTotalUnits,
			TotalPriceCurrency: productResp.Product.Price.Currency,
		})
		if err != nil {
			s.logger.Error("Failed to add order item", zap.Error(err))
		}
	}

	return &orderpb.CreateOrderResponse{
		OrderId:     orderID.String(),
		OrderNumber: orderNumber,
		Status:      orderpb.OrderStatus_ORDER_STATUS_PENDING,
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

	orderItems, err := s.queries.GetOrderItems(ctx, orderID)
	if err != nil {
		s.logger.Warn("Failed to get order items", zap.Error(err))
	}

	orderProto := s.orderToProto(order)
	for _, item := range orderItems {
		orderProto.Items = append(orderProto.Items, s.orderItemToProto(item))
	}

	return &orderpb.GetOrderResponse{
		Order: orderProto,
	}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	s.logger.Info("Listing orders", zap.String("user_id", req.UserId))

	orders, err := s.queries.ListUserOrders(ctx, db.ListUserOrdersParams{
		UserId: uuid.MustParse(req.UserId),
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

	return &orderpb.ListOrdersResponse{
		Orders:     orderList,
		Pagination: req.Pagination,
	}, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Updating order status", zap.String("order_id", req.OrderId), zap.String("status", req.Status.String()))

	orderID := uuid.MustParse(req.OrderId)
	if err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Id:     orderID,
		Status: int32(req.Status),
	}); err != nil {
		s.logger.Error("Failed to update order status", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	return &sharedpb.Empty{}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Cancelling order", zap.String("order_id", req.OrderId))

	return s.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{
		OrderId: req.OrderId,
		Status:  orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
	})
}

func (s *OrderService) ApplyPoints(ctx context.Context, req *orderpb.ApplyPointsRequest) (*orderpb.ApplyPointsResponse, error) {
	s.logger.Info("Applying points to order", zap.String("order_id", req.OrderId), zap.Int64("points", req.Points))

	orderID := uuid.MustParse(req.OrderId)

	order, err := s.queries.GetOrder(ctx, orderID)
	if err != nil {
		s.logger.Error("Failed to get order", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != int32(orderpb.OrderStatus_ORDER_STATUS_PENDING) {
		return nil, fmt.Errorf("cannot apply points to non-pending order")
	}

	if req.Points > 10000 {
		return nil, fmt.Errorf("cannot apply more than 10,000 points to a single order")
	}

	pointsValue := req.Points * 10

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	return &orderpb.ApplyPointsResponse{
		Success: true,
		YenValue: &sharedpb.Money{
			Units:    pointsValue,
			Currency: "JPY",
		},
	}, nil
}

func (s *OrderService) ReserveDeliverySlot(ctx context.Context, req *orderpb.ReserveDeliverySlotRequest) (*orderpb.ReserveDeliverySlotResponse, error) {
	s.logger.Info("Reserving delivery slot", zap.String("order_id", req.OrderId), zap.String("slot", req.SlotId))

	orderID := uuid.MustParse(req.OrderId)

	order, err := s.queries.GetOrder(ctx, orderID)
	if err != nil {
		s.logger.Error("Failed to get order", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != int32(orderpb.OrderStatus_ORDER_STATUS_PENDING) {
		return nil, fmt.Errorf("cannot reserve slot for non-pending order")
	}

	if req.SlotId == "" {
		return nil, fmt.Errorf("delivery slot ID is required")
	}

	reservationID := fmt.Sprintf("RES-%s", uuid.New().String())

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	return &orderpb.ReserveDeliverySlotResponse{
		ReservationId: reservationID,
	}, nil
}

func (s *OrderService) orderToProto(o db.Order) *orderpb.Order {
	status := orderpb.OrderStatus(o.Status)

	return &orderpb.Order{
		Id:              o.Id.String(),
		OrderNumber:     o.OrderNumber,
		UserId:          o.UserId.String(),
		Status:          status,
		SubtotalAmount:  s.moneyToProto(o.SubtotalUnits, o.SubtotalCurrency),
		TaxAmount:       s.moneyToProto(o.TaxUnits, o.TaxCurrency),
		DiscountAmount:  s.moneyToProto(o.DiscountUnits, o.DiscountCurrency),
		TotalAmount:     s.moneyToProto(o.TotalUnits, o.TotalCurrency),
		PointsApplied:   o.PointsApplied,
		ShippingAddress: s.addressToProto(o.ShippingAddress),
		PaymentMethod:   orderpb.PaymentMethod(o.PaymentMethod),
		CreatedAt:       protoTime(o.CreatedAt),
		UpdatedAt:       protoTime(o.UpdatedAt),
	}
}

func (s *OrderService) orderItemToProto(i db.OrderItem) *orderpb.OrderItem {
	return &orderpb.OrderItem{
		Id:          i.Id.String(),
		ProductId:   i.ProductId.String(),
		VariantId:   i.VariantId,
		ProductName: i.ProductName,
		Quantity:    i.Quantity,
		UnitPrice:   s.moneyToProto(i.UnitPriceUnits, i.UnitPriceCurrency),
		TotalPrice:  s.moneyToProto(i.TotalPriceUnits, i.TotalPriceCurrency),
	}
}

func (s *OrderService) moneyToProto(units int64, currency string) *sharedpb.Money {
	return &sharedpb.Money{
		Units:    units,
		Currency: currency,
	}
}

func (s *OrderService) addressToProto(addr map[string]interface{}) *orderpb.ShippingAddress {
	return &orderpb.ShippingAddress{
		Name:         getStr(addr, "name"),
		Phone:        getStr(addr, "phone"),
		PostalCode:   getStr(addr, "postal_code"),
		Prefecture:   getStr(addr, "prefecture"),
		City:         getStr(addr, "city"),
		AddressLine1: getStr(addr, "address_line1"),
		AddressLine2: getStr(addr, "address_line2"),
	}
}

func (s *OrderService) addressToMap(addr *orderpb.ShippingAddress) map[string]interface{} {
	if addr == nil {
		return make(map[string]interface{})
	}

	return map[string]interface{}{
		"name":          addr.Name,
		"phone":         addr.Phone,
		"postal_code":   addr.PostalCode,
		"prefecture":    addr.Prefecture,
		"city":          addr.City,
		"address_line1": addr.AddressLine1,
		"address_line2": addr.AddressLine2,
	}
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
