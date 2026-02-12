package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	inventorypb "github.com/afasari/shinkansen-commerce/gen/proto/go/inventory"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"google.golang.org/grpc"
)

type InventoryHandler struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryHandler(conn *grpc.ClientConn) *InventoryHandler {
	return &InventoryHandler{
		client: inventorypb.NewInventoryServiceClient(conn),
	}
}

func (h *InventoryHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/inventory/stock", h.handleStock)
	mux.HandleFunc("/v1/inventory/stock/", h.handleStockItems)
	mux.HandleFunc("/v1/inventory/reserve", h.handleReserve)
	mux.HandleFunc("/v1/inventory/release", h.handleRelease)
	mux.HandleFunc("/v1/inventory/movements/", h.handleMovements)
}

func (h *InventoryHandler) handleStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.getStock(w, r, ctx)
	case http.MethodPut:
		h.updateStock(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *InventoryHandler) handleStockItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemID := r.URL.Path[len("/v1/inventory/stock/"):]
	parts := splitPath(itemID)
	if len(parts) > 1 && parts[1] == "movements" {
		if r.Method == http.MethodGet {
			h.getStockMovements(w, r, ctx, parts[0])
			return
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *InventoryHandler) handleMovements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stockItemID := r.URL.Path[len("/v1/inventory/movements/"):]
	if stockItemID == "" {
		http.Error(w, "Stock Item ID required", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.getStockMovements(w, r, ctx, stockItemID)
}

func (h *InventoryHandler) handleReserve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		h.reserveStock(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *InventoryHandler) handleRelease(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		h.releaseStock(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *InventoryHandler) getStock(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	productID := r.URL.Query().Get("product_id")
	variantID := r.URL.Query().Get("variant_id")
	warehouseID := r.URL.Query().Get("warehouse_id")

	if productID == "" || warehouseID == "" {
		http.Error(w, "product_id and warehouse_id are required", http.StatusBadRequest)
		return
	}

	req := &inventorypb.GetStockRequest{
		ProductId:   productID,
		VariantId:   variantID,
		WarehouseId: warehouseID,
	}

	resp, err := h.client.GetStock(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *InventoryHandler) updateStock(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req inventorypb.UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.client.UpdateStock(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func (h *InventoryHandler) reserveStock(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req inventorypb.ReserveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.ReserveStock(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *InventoryHandler) releaseStock(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req inventorypb.ReleaseStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.client.ReleaseStock(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func (h *InventoryHandler) getStockMovements(w http.ResponseWriter, r *http.Request, ctx context.Context, stockItemID string) {
	if stockItemID == "" {
		http.Error(w, "Stock Item ID required", http.StatusBadRequest)
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

	req := &inventorypb.GetStockMovementsRequest{
		StockItemId: stockItemID,
		Pagination: &sharedpb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}

	resp, err := h.client.GetStockMovements(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}
