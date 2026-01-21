package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	orderpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/order"
	sharedpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/shared"
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
	userID := r.Context().Value("user_id")
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
	var req orderpb.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id")
	if userID != nil {
		req.UserId = userID.(string)
	}

	resp, err := h.client.CreateOrder(ctx, &req)
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
