package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	deliverypb "github.com/afasari/shinkansen-commerce/gen/proto/go/delivery"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
)

// MockDeliveryService is a mock implementation of deliverypb.DeliveryServiceServer
type MockDeliveryService struct {
	mock.Mock
}

func (m *MockDeliveryService) GetDeliverySlots(ctx context.Context, req *deliverypb.GetDeliverySlotsRequest) (*deliverypb.GetDeliverySlotsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deliverypb.GetDeliverySlotsResponse), args.Error(1)
}

func (m *MockDeliveryService) ReserveDeliverySlot(ctx context.Context, req *deliverypb.ReserveDeliverySlotRequest) (*deliverypb.ReserveDeliverySlotResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deliverypb.ReserveDeliverySlotResponse), args.Error(1)
}

func (m *MockDeliveryService) GetShipment(ctx context.Context, req *deliverypb.GetShipmentRequest) (*deliverypb.GetShipmentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deliverypb.GetShipmentResponse), args.Error(1)
}

func (m *MockDeliveryService) UpdateShipmentStatus(ctx context.Context, req *deliverypb.UpdateShipmentStatusRequest) (*sharedpb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sharedpb.Empty), args.Error(1)
}

func TestHandler_GetDeliverySlots(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockDeliveryService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &deliverypb.GetDeliverySlotsRequest{
			DeliveryZoneId: "zone-123",
		}

		expectedResp := &deliverypb.GetDeliverySlotsResponse{
			Slots: []*deliverypb.DeliverySlot{
				{Id: "slot-1", Capacity: 10},
			},
		}

		mockService.On("GetDeliverySlots", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.GetDeliverySlots(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ReserveDeliverySlot(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockDeliveryService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &deliverypb.ReserveDeliverySlotRequest{
			SlotId:  "slot-123",
			OrderId: "order-456",
		}

		expectedResp := &deliverypb.ReserveDeliverySlotResponse{
			ReservationId: "res-789",
		}

		mockService.On("ReserveDeliverySlot", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.ReserveDeliverySlot(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetShipment(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockDeliveryService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &deliverypb.GetShipmentRequest{
			ShipmentId: "shipment-123",
		}

		expectedResp := &deliverypb.GetShipmentResponse{
			Shipment: &deliverypb.Shipment{
				Id:     "shipment-123",
				Status: deliverypb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT,
			},
		}

		mockService.On("GetShipment", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.GetShipment(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateShipmentStatus(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockDeliveryService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &deliverypb.UpdateShipmentStatusRequest{
			ShipmentId: "shipment-123",
			Status:     deliverypb.ShipmentStatus_SHIPMENT_STATUS_SHIPPED,
		}

		expectedResp := &sharedpb.Empty{}

		mockService.On("UpdateShipmentStatus", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.UpdateShipmentStatus(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}
