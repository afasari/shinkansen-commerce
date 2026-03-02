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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	deliverypb "github.com/afasari/shinkansen-commerce/gen/proto/go/delivery"
	"github.com/afasari/shinkansen-commerce/services/delivery-service/internal/db"
)

// MockQuerier is a mock implementation of db.Querier
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) GetDeliverySlots(ctx context.Context, deliveryZoneID uuid.UUID, date time.Time) ([]db.DeliverySlot, error) {
	args := m.Called(ctx, deliveryZoneID, date)
	return args.Get(0).([]db.DeliverySlot), args.Error(1)
}

func (m *MockQuerier) ReserveDeliverySlot(ctx context.Context, slotID uuid.UUID, orderID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, slotID, orderID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQuerier) GetShipment(ctx context.Context, id uuid.UUID) (db.Shipment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Shipment), args.Error(1)
}

func (m *MockQuerier) UpdateShipmentStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockQuerier) CreateShipment(ctx context.Context, orderID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQuerier) GetDeliverySlot(ctx context.Context, id uuid.UUID) (db.DeliverySlot, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.DeliverySlot), args.Error(1)
}

func (m *MockQuerier) GetShipmentByOrderID(ctx context.Context, orderID uuid.UUID) (db.Shipment, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(db.Shipment), args.Error(1)
}

func (m *MockQuerier) ReleaseDeliverySlot(ctx context.Context, orderID uuid.UUID) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func TestDeliveryService_GetDeliverySlots(t *testing.T) {
	t.Run("successful delivery slots retrieval", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		deliveryZoneID := uuid.New()
		slotDate := time.Now().UTC().Truncate(24 * time.Hour)

		slot1 := db.DeliverySlot{
			ID:             uuid.New(),
			DeliveryZoneID: deliveryZoneID,
			StartTime:      time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			Capacity:       10,
			Reserved:       3,
			Available:      7,
			Date:           slotDate,
		}

		slot2 := db.DeliverySlot{
			ID:             uuid.New(),
			DeliveryZoneID: deliveryZoneID,
			StartTime:      time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2025, 1, 1, 18, 0, 0, 0, time.UTC),
			Capacity:       15,
			Reserved:       5,
			Available:      10,
			Date:           slotDate,
		}

		mockQueries.On("GetDeliverySlots", mock.Anything, deliveryZoneID, mock.Anything).Return([]db.DeliverySlot{slot1, slot2}, nil)

		req := &deliverypb.GetDeliverySlotsRequest{
			DeliveryZoneId: deliveryZoneID.String(),
			Date:           timestamppb.New(slotDate),
		}

		resp, err := service.GetDeliverySlots(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Slots, 2)
		assert.Equal(t, int32(10), resp.Slots[0].Capacity)
		assert.Equal(t, int32(3), resp.Slots[0].Reserved)
		assert.Equal(t, int32(7), resp.Slots[0].Available)
		mockQueries.AssertExpectations(t)
	})

	t.Run("invalid delivery zone ID", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		slotDate := time.Now().UTC().Truncate(24 * time.Hour)

		req := &deliverypb.GetDeliverySlotsRequest{
			DeliveryZoneId: "invalid-uuid",
			Date:           timestamppb.New(slotDate),
		}

		resp, err := service.GetDeliverySlots(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, status.Convert(err).Message(), "invalid delivery_zone_id")
	})

	t.Run("no slots available", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		deliveryZoneID := uuid.New()
		slotDate := time.Now().UTC().Truncate(24 * time.Hour)

		mockQueries.On("GetDeliverySlots", mock.Anything, deliveryZoneID, slotDate).Return([]db.DeliverySlot{}, nil)

		req := &deliverypb.GetDeliverySlotsRequest{
			DeliveryZoneId: deliveryZoneID.String(),
			Date:           timestamppb.New(slotDate),
		}

		resp, err := service.GetDeliverySlots(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Slots, 0)
		mockQueries.AssertExpectations(t)
	})
}

func TestDeliveryService_ReserveDeliverySlot(t *testing.T) {
	t.Run("successful slot reservation", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		slotID := uuid.New()
		orderID := uuid.New()
		reservationID := uuid.New()

		mockQueries.On("ReserveDeliverySlot", mock.Anything, slotID, orderID).Return(reservationID, nil)

		req := &deliverypb.ReserveDeliverySlotRequest{
			SlotId:  slotID.String(),
			OrderId: orderID.String(),
		}

		resp, err := service.ReserveDeliverySlot(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, reservationID.String(), resp.ReservationId)
		assert.NotNil(t, resp.ReservedAt)
		mockQueries.AssertExpectations(t)
	})

	t.Run("invalid slot ID", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		orderID := uuid.New()

		req := &deliverypb.ReserveDeliverySlotRequest{
			SlotId:  "invalid-uuid",
			OrderId: orderID.String(),
		}

		resp, err := service.ReserveDeliverySlot(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, status.Convert(err).Message(), "invalid slot_id")
	})

	t.Run("invalid order ID", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		slotID := uuid.New()

		req := &deliverypb.ReserveDeliverySlotRequest{
			SlotId:  slotID.String(),
			OrderId: "invalid-uuid",
		}

		resp, err := service.ReserveDeliverySlot(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, status.Convert(err).Message(), "invalid order_id")
	})

	t.Run("slot no longer available", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		slotID := uuid.New()
		orderID := uuid.New()

		mockQueries.On("ReserveDeliverySlot", mock.Anything, slotID, orderID).Return(uuid.Nil, assert.AnError)

		req := &deliverypb.ReserveDeliverySlotRequest{
			SlotId:  slotID.String(),
			OrderId: orderID.String(),
		}

		resp, err := service.ReserveDeliverySlot(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockQueries.AssertExpectations(t)
	})
}

func TestDeliveryService_GetShipment(t *testing.T) {
	t.Run("successful shipment retrieval", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		shipmentID := uuid.New()
		orderID := uuid.New()
		estDelivery := time.Now().Add(24 * time.Hour)
		trackingNumber := "TRACK-12345"

		shipment := db.Shipment{
			ID:                  shipmentID,
			OrderID:             orderID,
			TrackingNumber:      &trackingNumber,
			Status:              "SHIPMENT_STATUS_IN_TRANSIT",
			EstimatedDeliveryAt: &estDelivery,
			ActualDeliveryAt:    nil,
			Carrier:             "Yamato Transport",
		}

		mockQueries.On("GetShipment", mock.Anything, shipmentID).Return(shipment, nil)

		req := &deliverypb.GetShipmentRequest{
			ShipmentId: shipmentID.String(),
		}

		resp, err := service.GetShipment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Shipment)
		assert.Equal(t, shipmentID.String(), resp.Shipment.Id)
		assert.Equal(t, orderID.String(), resp.Shipment.OrderId)
		assert.Equal(t, trackingNumber, resp.Shipment.TrackingNumber)
		assert.Equal(t, deliverypb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT, resp.Shipment.Status)
		assert.Equal(t, "Yamato Transport", resp.Shipment.Carrier)
		assert.NotNil(t, resp.Shipment.EstimatedDeliveryAt)
		mockQueries.AssertExpectations(t)
	})

	t.Run("shipment not found", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		shipmentID := uuid.New()

		mockQueries.On("GetShipment", mock.Anything, shipmentID).Return(db.Shipment{}, assert.AnError)

		req := &deliverypb.GetShipmentRequest{
			ShipmentId: shipmentID.String(),
		}

		resp, err := service.GetShipment(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockQueries.AssertExpectations(t)
	})

	t.Run("invalid shipment ID", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		req := &deliverypb.GetShipmentRequest{
			ShipmentId: "invalid-uuid",
		}

		resp, err := service.GetShipment(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, status.Convert(err).Message(), "invalid shipment_id")
	})

	t.Run("delivered shipment", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		shipmentID := uuid.New()
		orderID := uuid.New()
		estDelivery := time.Now().Add(24 * time.Hour)
		actualDelivery := time.Now()
		trackingNumber := "TRACK-12345"

		shipment := db.Shipment{
			ID:                  shipmentID,
			OrderID:             orderID,
			TrackingNumber:      &trackingNumber,
			Status:              "SHIPMENT_STATUS_DELIVERED",
			EstimatedDeliveryAt: &estDelivery,
			ActualDeliveryAt:    &actualDelivery,
			Carrier:             "Sagawa Express",
		}

		mockQueries.On("GetShipment", mock.Anything, shipmentID).Return(shipment, nil)

		req := &deliverypb.GetShipmentRequest{
			ShipmentId: shipmentID.String(),
		}

		resp, err := service.GetShipment(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, deliverypb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED, resp.Shipment.Status)
		assert.NotNil(t, resp.Shipment.ActualDeliveryAt)
		mockQueries.AssertExpectations(t)
	})
}

func TestDeliveryService_UpdateShipmentStatus(t *testing.T) {
	t.Run("successful status update", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		shipmentID := uuid.New()

		mockQueries.On("UpdateShipmentStatus", mock.Anything, shipmentID, "SHIPMENT_STATUS_SHIPPED").Return(nil)

		req := &deliverypb.UpdateShipmentStatusRequest{
			ShipmentId: shipmentID.String(),
			Status:     deliverypb.ShipmentStatus_SHIPMENT_STATUS_SHIPPED,
		}

		resp, err := service.UpdateShipmentStatus(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockQueries.AssertExpectations(t)
	})

	t.Run("invalid shipment ID", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		req := &deliverypb.UpdateShipmentStatusRequest{
			ShipmentId: "invalid-uuid",
			Status:     deliverypb.ShipmentStatus_SHIPMENT_STATUS_SHIPPED,
		}

		resp, err := service.UpdateShipmentStatus(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, status.Convert(err).Message(), "invalid shipment_id")
	})

	t.Run("update to delivered", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		shipmentID := uuid.New()

		mockQueries.On("UpdateShipmentStatus", mock.Anything, shipmentID, "SHIPMENT_STATUS_DELIVERED").Return(nil)

		req := &deliverypb.UpdateShipmentStatusRequest{
			ShipmentId: shipmentID.String(),
			Status:     deliverypb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED,
		}

		resp, err := service.UpdateShipmentStatus(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockQueries.AssertExpectations(t)
	})

	t.Run("update to failed delivery", func(t *testing.T) {
		logger := zap.NewNop()
		mockQueries := new(MockQuerier)

		service := NewDeliveryService(mockQueries, logger)

		shipmentID := uuid.New()

		mockQueries.On("UpdateShipmentStatus", mock.Anything, shipmentID, "SHIPMENT_STATUS_FAILED_DELIVERY").Return(nil)

		req := &deliverypb.UpdateShipmentStatusRequest{
			ShipmentId: shipmentID.String(),
			Status:     deliverypb.ShipmentStatus_SHIPMENT_STATUS_FAILED_DELIVERY,
		}

		resp, err := service.UpdateShipmentStatus(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockQueries.AssertExpectations(t)
	})
}
