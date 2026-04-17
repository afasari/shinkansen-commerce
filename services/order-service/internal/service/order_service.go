package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/pkg/pgutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	queries        db.Querier
	productClient  productpb.ProductServiceClient
	cache          cache.Cache
	cartService    *CartService
	stateMachine   *OrderStateMachine
	eventPublisher *OrderEventPublisher
	logger         *zap.Logger
}

func NewOrderService(
	queries db.Querier,
	productClient productpb.ProductServiceClient,
	cacheClient cache.Cache,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		queries:        queries,
		productClient:  productClient,
		cache:          cacheClient,
		cartService:    nil,                          // Optional - can be set later if needed
		stateMachine:   NewOrderStateMachine(logger), // Create default state machine
		eventPublisher: nil,                          // Optional - event publishing can be added later
		logger:         logger,
	}
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}

// SetEventPublisher sets the event publisher (optional)
func (s *OrderService) SetEventPublisher(publisher *OrderEventPublisher) {
	s.eventPublisher = publisher
}

// SetCartService sets the cart service (optional)
func (s *OrderService) SetCartService(cartService *CartService) {
	s.cartService = cartService
}

// SetStateMachine sets the state machine (optional)
func (s *OrderService) SetStateMachine(stateMachine *OrderStateMachine) {
	s.stateMachine = stateMachine
}

func (s *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	ctx, span := otel.Tracer("order-service").Start(ctx, "OrderService.CreateOrder",
		trace.WithAttributes(attribute.String("order.user_id", req.UserId)),
	)
	defer span.End()

	s.logger.Info("Creating order", zap.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "order must have at least one item")
	}

	var subtotalUnits int64

	for _, item := range req.Items {
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{ProductId: item.ProductId})
		if err != nil {
			s.logger.Error("Failed to get product", zap.String("product_id", item.ProductId), zap.Error(err))
			return nil, status.Error(codes.NotFound, fmt.Sprintf("product %s not found", item.ProductId))
		}

		if productResp.Product.StockQuantity < item.Quantity {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("product %s has insufficient stock", item.ProductId))
		}

		itemTotalUnits := productResp.Product.Price.Units * int64(item.Quantity)
		subtotalUnits += itemTotalUnits
	}

	taxUnits := int64(float64(subtotalUnits) * 0.10)
	totalUnits := subtotalUnits + taxUnits

	orderNumber := fmt.Sprintf("ORD-%s", uuid.New().String())

	orderID, err := s.queries.CreateOrder(ctx, db.CreateOrderParams{
		OrderNumber:      orderNumber,
		UserID:           pgutil.ToPG(userID),
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
		ShippingAddress:  s.addressToBytes(req.ShippingAddress),
		PaymentMethod:    int32(req.PaymentMethod),
	})
	if err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		if isDuplicateKeyError(err) {
			return nil, status.Error(codes.AlreadyExists, "order number already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	// Publish order created event
	if s.eventPublisher != nil {
		orderProto := &orderpb.Order{
			Id:              pgutil.FromPG(orderID),
			OrderNumber:     orderNumber,
			UserId:          req.UserId,
			Status:          orderpb.OrderStatus_ORDER_STATUS_PENDING,
			SubtotalAmount:  s.moneyToProto(subtotalUnits, "JPY"),
			TaxAmount:       s.moneyToProto(taxUnits, "JPY"),
			DiscountAmount:  s.moneyToProto(0, "JPY"),
			TotalAmount:     s.moneyToProto(totalUnits, "JPY"),
			PointsApplied:   0,
			ShippingAddress: req.ShippingAddress,
			PaymentMethod:   req.PaymentMethod,
			CreatedAt:       timestamppb.Now(),
		}
		if err := s.eventPublisher.PublishOrderCreated(ctx, orderProto); err != nil {
			s.logger.Warn("Failed to publish order created event", zap.Error(err))
		}
	}

	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductId)
		if err != nil {
			s.logger.Error("Failed to parse product ID", zap.String("product_id", item.ProductId), zap.Error(err))
			continue
		}

		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{ProductId: item.ProductId})
		if err != nil {
			s.logger.Error("Failed to get product for item", zap.String("product_id", item.ProductId), zap.Error(err))
			continue
		}

		itemTotalUnits := productResp.Product.Price.Units * int64(item.Quantity)

		err = s.queries.AddOrderItem(ctx, db.AddOrderItemParams{
			OrderID:            orderID,
			ProductID:          pgutil.ToPG(productID),
			VariantID:          pgutil.ToPGFromString(item.VariantId),
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
		OrderId:     pgutil.FromPG(orderID),
		OrderNumber: orderNumber,
		Status:      orderpb.OrderStatus_ORDER_STATUS_PENDING,
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	ctx, span := otel.Tracer("order-service").Start(ctx, "OrderService.GetOrder",
		trace.WithAttributes(attribute.String("order.id", req.OrderId)),
	)
	defer span.End()

	s.logger.Info("Getting order", zap.String("order_id", req.OrderId))

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	cacheKey := cache.OrderCacheKey(req.OrderId)
	var cached db.OrdersOrders

	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("Order cache hit", zap.String("order_id", req.OrderId))
		return &orderpb.GetOrderResponse{
			Order: s.orderToProto(cached),
		}, nil
	}

	s.logger.Debug("Order cache miss", zap.String("order_id", req.OrderId))

	orderIDpg := pgutil.ToPG(orderID)
	order, err := s.queries.GetOrder(ctx, orderIDpg)
	if err != nil {
		s.logger.Error("Failed to get order", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, status.Error(codes.NotFound, "order not found")
	}

	if err := s.cache.Set(ctx, cacheKey, order, cache.DefaultTTL); err != nil {
		s.logger.Warn("Failed to cache order", zap.Error(err))
	}

	orderItems, err := s.queries.GetOrderItems(ctx, orderIDpg)
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

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	orders, err := s.queries.ListUserOrders(ctx, db.ListUserOrdersParams{
		UserID: pgutil.ToPG(userID),
		Limit:  req.Pagination.Limit,
		Offset: (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		s.logger.Error("Failed to list orders", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list orders")
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
	ctx, span := otel.Tracer("order-service").Start(ctx, "OrderService.UpdateOrderStatus",
		trace.WithAttributes(
			attribute.String("order.id", req.OrderId),
			attribute.String("order.new_status", req.Status.String()),
		),
	)
	defer span.End()

	s.logger.Info("Updating order status", zap.String("order_id", req.OrderId), zap.String("status", req.Status.String()))

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	orderIDpg := pgutil.ToPG(orderID)

	// Get current order for state transition validation
	currentOrder, err := s.queries.GetOrder(ctx, orderIDpg)
	if err != nil {
		s.logger.Error("Failed to get current order", zap.Error(err))
		return nil, status.Error(codes.NotFound, "order not found")
	}

	currentStatus := orderpb.OrderStatus(currentOrder.Status)

	// Validate state transition
	if s.stateMachine != nil {
		if err := s.stateMachine.Transition(req.OrderId, currentStatus, req.Status); err != nil {
			return nil, err
		}
	}

	if err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderIDpg,
		Status: int32(req.Status),
	}); err != nil {
		s.logger.Error("Failed to update order status", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update order status")
	}

	// Publish status change event
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishOrderStatusChanged(
			ctx,
			req.OrderId,
			pgutil.FromPG(currentOrder.UserID),
			currentStatus,
			req.Status,
			"",
		); err != nil {
			s.logger.Warn("Failed to publish status change event", zap.Error(err))
		}
	}

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	return &sharedpb.Empty{}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*sharedpb.Empty, error) {
	ctx, span := otel.Tracer("order-service").Start(ctx, "OrderService.CancelOrder",
		trace.WithAttributes(attribute.String("order.id", req.OrderId)),
	)
	defer span.End()

	s.logger.Info("Cancelling order", zap.String("order_id", req.OrderId))

	orderID := pgutil.ToPG(uuid.MustParse(req.OrderId))

	// Get current order for validation
	currentOrder, err := s.queries.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current order: %w", err)
	}

	currentStatus := orderpb.OrderStatus(currentOrder.Status)

	// Check if order can be cancelled
	if s.stateMachine != nil && !s.stateMachine.IsCancellable(currentStatus) {
		return nil, fmt.Errorf("order cannot be cancelled in status: %s", currentStatus)
	}

	if err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderID,
		Status: int32(orderpb.OrderStatus_ORDER_STATUS_CANCELLED),
	}); err != nil {
		s.logger.Error("Failed to cancel order", zap.String("order_id", req.OrderId), zap.Error(err))
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	cacheKey := cache.OrderCacheKey(req.OrderId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate order cache", zap.Error(err))
	}

	// Publish cancel event
	if s.eventPublisher != nil {
		orderProto := s.orderToProto(currentOrder)
		orderProto.Status = orderpb.OrderStatus_ORDER_STATUS_CANCELLED
		if err := s.eventPublisher.PublishOrderCancelled(ctx, orderProto, req.Reason); err != nil {
			s.logger.Warn("Failed to publish cancel event", zap.Error(err))
		}
	}

	return &sharedpb.Empty{}, nil
}

func (s *OrderService) ApplyPoints(ctx context.Context, req *orderpb.ApplyPointsRequest) (*orderpb.ApplyPointsResponse, error) {
	s.logger.Info("Applying points to order", zap.String("order_id", req.OrderId), zap.Int64("points", req.Points))

	orderID := pgutil.ToPG(uuid.MustParse(req.OrderId))

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

	orderID := pgutil.ToPG(uuid.MustParse(req.OrderId))

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

func (s *OrderService) orderToProto(o db.OrdersOrders) *orderpb.Order {
	status := orderpb.OrderStatus(o.Status)

	return &orderpb.Order{
		Id:              pgutil.FromPG(o.ID),
		OrderNumber:     o.OrderNumber,
		UserId:          pgutil.FromPG(o.UserID),
		Status:          status,
		SubtotalAmount:  s.moneyToProto(o.SubtotalUnits, o.SubtotalCurrency),
		TaxAmount:       s.moneyToProto(o.TaxUnits, o.TaxCurrency),
		DiscountAmount:  s.moneyToProto(o.DiscountUnits, o.DiscountCurrency),
		TotalAmount:     s.moneyToProto(o.TotalUnits, o.TotalCurrency),
		PointsApplied:   int64(o.PointsApplied),
		ShippingAddress: s.bytesToAddress(o.ShippingAddress),
		PaymentMethod:   orderpb.PaymentMethod(o.PaymentMethod),
		CreatedAt:       protoTimeFromTimestamptz(o.CreatedAt),
		UpdatedAt:       protoTimeFromTimestamptz(o.UpdatedAt),
	}
}

func (s *OrderService) orderItemToProto(i db.OrdersOrderItems) *orderpb.OrderItem {
	variantID := i.VariantID.String()

	return &orderpb.OrderItem{
		Id:          pgutil.FromPG(i.ID),
		ProductId:   pgutil.FromPG(i.ProductID),
		VariantId:   variantID,
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

func (s *OrderService) addressToBytes(addr *orderpb.ShippingAddress) []byte {
	if addr == nil {
		return []byte("{}")
	}

	data, _ := json.Marshal(map[string]interface{}{
		"name":          addr.Name,
		"phone":         addr.Phone,
		"postal_code":   addr.PostalCode,
		"prefecture":    addr.Prefecture,
		"city":          addr.City,
		"address_line1": addr.AddressLine1,
		"address_line2": addr.AddressLine2,
	})
	return data
}

func getStr(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func protoTimeFromTimestamptz(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}

func (s *OrderService) bytesToAddress(data []byte) *orderpb.ShippingAddress {
	if len(data) == 0 {
		return &orderpb.ShippingAddress{}
	}
	var addrMap map[string]interface{}
	if err := json.Unmarshal(data, &addrMap); err != nil {
		return &orderpb.ShippingAddress{}
	}
	return s.addressToProto(addrMap)
}
