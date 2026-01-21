package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	deliverypb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/delivery"
	sharedpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/shared"
	"github.com/shinkansen-commerce/shinkansen/services/delivery-service/internal/db"
	"go.uber.org/zap"
)

type DeliveryService struct {
	deliverypb.UnimplementedDeliveryServiceServer
	queries db.Querier
	logger  *zap.Logger
}

func NewDeliveryService(queries db.Querier, logger *zap.Logger) *DeliveryService {
	return &DeliveryService{
		queries: queries,
		logger:  logger,
	}
}

func (s *DeliveryService) GetDeliverySlots(ctx context.Context, req *deliverypb.GetDeliverySlotsRequest) (*deliverypb.GetDeliverySlotsResponse, error) {
	s.logger.Info("Getting delivery slots",
		zap.String("delivery_zone_id", req.DeliveryZoneId))

	deliveryZoneID := uuid.MustParse(req.DeliveryZoneId)
	date := req.Date.AsTime()
	if date.IsZero() {
		date = time.Now()
	}

	slots, err := s.queries.GetDeliverySlots(ctx, deliveryZoneID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery slots: %w", err)
	}

	protoSlots := make([]*deliverypb.DeliverySlot, 0, len(slots))
	for _, slot := range slots {
		protoSlots = append(protoSlots, s.deliverySlotToProto(slot))
	}

	return &deliverypb.GetDeliverySlotsResponse{
		Slots: protoSlots,
	}, nil
}

func (s *DeliveryService) ReserveDeliverySlot(ctx context.Context, req *deliverypb.ReserveDeliverySlotRequest) (*deliverypb.ReserveDeliverySlotResponse, error) {
	s.logger.Info("Reserving delivery slot",
		zap.String("slot_id", req.SlotId),
		zap.String("order_id", req.OrderId))

	slotID := uuid.MustParse(req.SlotId)
	orderID := uuid.MustParse(req.OrderId)

	reservationID, err := s.queries.ReserveDeliverySlot(ctx, slotID, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve delivery slot: %w", err)
	}

	return &deliverypb.ReserveDeliverySlotResponse{
		ReservationId: reservationID.String(),
		ReservedAt:    timestamppb.Now(),
	}, nil
}

func (s *DeliveryService) GetShipment(ctx context.Context, req *deliverypb.GetShipmentRequest) (*deliverypb.GetShipmentResponse, error) {
	s.logger.Info("Getting shipment", zap.String("shipment_id", req.ShipmentId))

	shipmentID := uuid.MustParse(req.ShipmentId)
	shipment, err := s.queries.GetShipment(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	return &deliverypb.GetShipmentResponse{
		Shipment: s.shipmentToProto(shipment),
	}, nil
}

func (s *DeliveryService) UpdateShipmentStatus(ctx context.Context, req *deliverypb.UpdateShipmentStatusRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Updating shipment status",
		zap.String("shipment_id", req.ShipmentId),
		zap.String("status", req.Status.String()))

	shipmentID := uuid.MustParse(req.ShipmentId)
	err := s.queries.UpdateShipmentStatus(ctx, shipmentID, req.Status.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update shipment status: %w", err)
	}

	return &sharedpb.Empty{}, nil
}

func (s *DeliveryService) deliverySlotToProto(slot db.DeliverySlot) *deliverypb.DeliverySlot {
	return &deliverypb.DeliverySlot{
		Id:             slot.ID.String(),
		DeliveryZoneId: slot.DeliveryZoneID.String(),
		StartTime:      timestamppb.New(slot.StartTime),
		EndTime:        timestamppb.New(slot.EndTime),
		Capacity:       int32(slot.Capacity),
		Reserved:       int32(slot.Reserved),
		Available:      int32(slot.Available),
		Date:           timestamppb.New(slot.Date),
	}
}

func (s *DeliveryService) shipmentToProto(shipment db.Shipment) *deliverypb.Shipment {
	status := deliverypb.ShipmentStatus(deliverypb.ShipmentStatus_value[shipment.Status])

	var estimatedDeliveryAt *timestamppb.Timestamp
	if shipment.EstimatedDeliveryAt != nil {
		estimatedDeliveryAt = timestamppb.New(*shipment.EstimatedDeliveryAt)
	}

	var actualDeliveryAt *timestamppb.Timestamp
	if shipment.ActualDeliveryAt != nil {
		actualDeliveryAt = timestamppb.New(*shipment.ActualDeliveryAt)
	}

	return &deliverypb.Shipment{
		Id:                  shipment.ID.String(),
		OrderId:             shipment.OrderID.String(),
		TrackingNumber:      toStringPtr(shipment.TrackingNumber),
		Status:              status,
		EstimatedDeliveryAt: estimatedDeliveryAt,
		ActualDeliveryAt:    actualDeliveryAt,
		Carrier:             shipment.Carrier,
	}
}

func toStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
