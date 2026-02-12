package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	deliverypb "github.com/afasari/shinkansen-commerce/gen/proto/go/delivery"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeliveryHandler struct {
	client deliverypb.DeliveryServiceClient
}

func NewDeliveryHandler(conn *grpc.ClientConn) *DeliveryHandler {
	return &DeliveryHandler{
		client: deliverypb.NewDeliveryServiceClient(conn),
	}
}

func (h *DeliveryHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/delivery/slots", h.handleDeliverySlots)
	mux.HandleFunc("/v1/delivery/slots/", h.handleDeliverySlotReserve)
	mux.HandleFunc("/v1/shipments/", h.handleShipment)
}

func (h *DeliveryHandler) handleDeliverySlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.getDeliverySlots(w, r, ctx)
	case http.MethodPost:
		h.reserveDeliverySlot(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DeliveryHandler) handleDeliverySlotReserve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slotID := r.URL.Path[len("/v1/delivery/slots/"):]
	parts := splitPath(slotID)
	if len(parts) > 1 && parts[1] == "reserve" {
		if r.Method == http.MethodPost {
			h.reserveDeliverySlotByID(w, r, ctx, parts[0])
			return
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *DeliveryHandler) handleShipment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shipmentID := r.URL.Path[len("/v1/shipments/"):]
	parts := splitPath(shipmentID)
	if len(parts) > 1 && parts[1] == "status" {
		if r.Method == http.MethodPut {
			h.updateShipmentStatus(w, r, ctx, parts[0])
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if shipmentID == "" {
			http.Error(w, "Shipment ID required", http.StatusBadRequest)
			return
		}
		h.getShipment(w, r, ctx, shipmentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DeliveryHandler) getDeliverySlots(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	deliveryZoneID := r.URL.Query().Get("delivery_zone_id")
	if deliveryZoneID == "" {
		http.Error(w, "delivery_zone_id is required", http.StatusBadRequest)
		return
	}

	var date time.Time
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		date = time.Now()
	}

	req := &deliverypb.GetDeliverySlotsRequest{
		DeliveryZoneId: deliveryZoneID,
		Date:           timestamppb.New(date),
	}

	resp, err := h.client.GetDeliverySlots(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DeliveryHandler) reserveDeliverySlot(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req deliverypb.ReserveDeliverySlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.ReserveDeliverySlot(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *DeliveryHandler) reserveDeliverySlotByID(w http.ResponseWriter, r *http.Request, ctx context.Context, slotID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deliverypb.ReserveDeliverySlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.SlotId = slotID

	resp, err := h.client.ReserveDeliverySlot(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *DeliveryHandler) getShipment(w http.ResponseWriter, r *http.Request, ctx context.Context, shipmentID string) {
	req := &deliverypb.GetShipmentRequest{ShipmentId: shipmentID}
	resp, err := h.client.GetShipment(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DeliveryHandler) updateShipmentStatus(w http.ResponseWriter, r *http.Request, ctx context.Context, shipmentID string) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deliverypb.UpdateShipmentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.ShipmentId = shipmentID

	_, err := h.client.UpdateShipmentStatus(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
