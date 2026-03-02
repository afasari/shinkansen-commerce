package handler

import (
	"context"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"go.uber.org/zap"
)

// Handler wraps the order service and implements the gRPC server interface
type Handler struct {
	orderpb.UnimplementedOrderServiceServer
	service orderpb.OrderServiceServer
	logger  *zap.Logger
}

// NewHandler creates a new order service handler
func NewHandler(service orderpb.OrderServiceServer, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	h.logger.Debug("CreateOrder called", zap.String("user_id", req.UserId))
	return h.service.CreateOrder(ctx, req)
}

func (h *Handler) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	h.logger.Debug("GetOrder called", zap.String("order_id", req.OrderId))
	return h.service.GetOrder(ctx, req)
}

func (h *Handler) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	h.logger.Debug("ListOrders called", zap.String("user_id", req.UserId))
	return h.service.ListOrders(ctx, req)
}

func (h *Handler) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*sharedpb.Empty, error) {
	h.logger.Debug("UpdateOrderStatus called", zap.String("order_id", req.OrderId))
	return h.service.UpdateOrderStatus(ctx, req)
}

func (h *Handler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*sharedpb.Empty, error) {
	h.logger.Debug("CancelOrder called", zap.String("order_id", req.OrderId))
	return h.service.CancelOrder(ctx, req)
}

func (h *Handler) ApplyPoints(ctx context.Context, req *orderpb.ApplyPointsRequest) (*orderpb.ApplyPointsResponse, error) {
	h.logger.Debug("ApplyPoints called", zap.String("order_id", req.OrderId))
	return h.service.ApplyPoints(ctx, req)
}

func (h *Handler) ReserveDeliverySlot(ctx context.Context, req *orderpb.ReserveDeliverySlotRequest) (*orderpb.ReserveDeliverySlotResponse, error) {
	h.logger.Debug("ReserveDeliverySlot called", zap.String("order_id", req.OrderId))
	return h.service.ReserveDeliverySlot(ctx, req)
}
