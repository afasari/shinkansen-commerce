package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/gateway/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type OrderHandler struct {
	client orderpb.OrderServiceClient
}

func NewOrderHandler(conn *grpc.ClientConn) *OrderHandler {
	return &OrderHandler{
		client: orderpb.NewOrderServiceClient(conn),
	}
}

func (h *OrderHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/orders", h.handleOrders)
	mux.HandleFunc("/v1/orders/", h.handleOrder)
	mux.HandleFunc("/v1/orders/apply-points", h.applyPoints)
}

func (h *OrderHandler) handleOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.listOrders(w, r, ctx)
	case http.MethodPost:
		h.createOrder(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OrderHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderID := r.URL.Path[len("/v1/orders/"):]
	parts := splitPath(orderID)
	if len(parts) > 1 {
		if parts[1] == "status" {
			h.updateOrderStatus(w, r, ctx, parts[0])
			return
		}
		if parts[1] == "cancel" {
			h.cancelOrder(w, r, ctx, parts[0])
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if orderID == "" {
			http.Error(w, "Order ID required", http.StatusBadRequest)
			return
		}
		h.getOrder(w, r, ctx, orderID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	page := int32(1)
	limit := int32(20)
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.ParseInt(p, 10, 32); err == nil {
			page = int32(val)
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.ParseInt(l, 10, 32); err == nil {
			limit = int32(val)
		}
	}

	var status string
	if s := r.URL.Query().Get("status"); s != "" {
		status = s
	}

	req := &orderpb.ListOrdersRequest{
		UserId: userID.(string),
		Status: wrapperspb.String(status),
		Pagination: &sharedpb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}

	resp, err := h.client.ListOrders(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Decode JSON into a map to handle proto wrapper types
	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build the proto request manually
	req := &orderpb.CreateOrderRequest{}

	// Set user_id from JWT if authenticated, otherwise use the one from request
	userID := r.Context().Value(middleware.UserIDKey)
	if userID != nil {
		req.UserId = userID.(string)
	} else if uid := getString(raw, "user_id"); uid != "" {
		req.UserId = uid
	}

	// Handle points_to_apply (Int64Value wrapper)
	if pts := getString(raw, "points_to_apply"); pts != "" {
		if v, err := strconv.ParseInt(pts, 10, 64); err == nil {
			req.PointsToApply = &wrapperspb.Int64Value{Value: v}
		}
	}

	// Handle delivery_slot_id (StringValue wrapper)
	if slotID := getString(raw, "delivery_slot_id"); slotID != "" {
		req.DeliverySlotId = &wrapperspb.StringValue{Value: slotID}
	}

	// Parse payment_method (enum — may be string name or numeric)
	if pm := getString(raw, "payment_method"); pm != "" {
		if v, ok := orderpb.PaymentMethod_value[pm]; ok {
			req.PaymentMethod = orderpb.PaymentMethod(v)
		} else if n, err := strconv.Atoi(pm); err == nil {
			req.PaymentMethod = orderpb.PaymentMethod(n)
		}
	} else if v := getFloat(raw, "payment_method"); v > 0 {
		req.PaymentMethod = orderpb.PaymentMethod(int32(v))
	}

	// Parse shipping_address
	if addr, ok := raw["shipping_address"].(map[string]interface{}); ok {
		req.ShippingAddress = &orderpb.ShippingAddress{
			Name:         getString(addr, "name"),
			Phone:        getString(addr, "phone"),
			PostalCode:   getString(addr, "postal_code"),
			Prefecture:   getString(addr, "prefecture"),
			City:         getString(addr, "city"),
			AddressLine1: getString(addr, "address_line1"),
			AddressLine2: getString(addr, "address_line2"),
		}
	}

	// Parse items
	if items, ok := raw["items"].([]interface{}); ok {
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				req.Items = append(req.Items, &orderpb.CreateOrderItem{
					ProductId: getString(m, "product_id"),
					VariantId: getString(m, "variant_id"),
					Quantity:  int32(getFloat(m, "quantity")),
				})
			}
		}
	}

	resp, err := h.client.CreateOrder(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request, ctx context.Context, orderID string) {
	req := &orderpb.GetOrderRequest{OrderId: orderID}
	resp, err := h.client.GetOrder(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) updateOrderStatus(w http.ResponseWriter, r *http.Request, ctx context.Context, orderID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req orderpb.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.OrderId = orderID

	_, err := h.client.UpdateOrderStatus(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func (h *OrderHandler) cancelOrder(w http.ResponseWriter, r *http.Request, ctx context.Context, orderID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &orderpb.CancelOrderRequest{OrderId: orderID}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		req.Reason = body.Reason
	}
	_, err := h.client.CancelOrder(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func (h *OrderHandler) applyPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req orderpb.ApplyPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.ApplyPoints(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i, ch := range path {
		if ch == '/' {
			if start < i {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
