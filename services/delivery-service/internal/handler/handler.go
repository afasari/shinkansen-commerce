package handler

import (
	"context"

	deliverypb "github.com/afasari/shinkansen-commerce/gen/proto/go/delivery"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"go.uber.org/zap"
)

// Handler wraps the delivery service and implements the gRPC server interface
type Handler struct {
	deliverypb.UnimplementedDeliveryServiceServer
	service deliverypb.DeliveryServiceServer
	logger  *zap.Logger
}

// NewHandler creates a new delivery service handler
func NewHandler(service deliverypb.DeliveryServiceServer, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) GetDeliverySlots(ctx context.Context, req *deliverypb.GetDeliverySlotsRequest) (*deliverypb.GetDeliverySlotsResponse, error) {
	h.logger.Debug("GetDeliverySlots called", zap.String("delivery_zone_id", req.DeliveryZoneId))
	return h.service.GetDeliverySlots(ctx, req)
}

func (h *Handler) ReserveDeliverySlot(ctx context.Context, req *deliverypb.ReserveDeliverySlotRequest) (*deliverypb.ReserveDeliverySlotResponse, error) {
	h.logger.Debug("ReserveDeliverySlot called", zap.String("slot_id", req.SlotId))
	return h.service.ReserveDeliverySlot(ctx, req)
}

func (h *Handler) GetShipment(ctx context.Context, req *deliverypb.GetShipmentRequest) (*deliverypb.GetShipmentResponse, error) {
	h.logger.Debug("GetShipment called", zap.String("shipment_id", req.ShipmentId))
	return h.service.GetShipment(ctx, req)
}

func (h *Handler) UpdateShipmentStatus(ctx context.Context, req *deliverypb.UpdateShipmentStatusRequest) (*sharedpb.Empty, error) {
	h.logger.Debug("UpdateShipmentStatus called", zap.String("shipment_id", req.ShipmentId))
	return h.service.UpdateShipmentStatus(ctx, req)
}
