package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	productpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/product"
	sharedpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	client productpb.ProductServiceClient
}

func NewProductHandler(conn *grpc.ClientConn) *ProductHandler {
	return &ProductHandler{
		client: productpb.NewProductServiceClient(conn),
	}
}

func (h *ProductHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/products", h.handleProducts)
	mux.HandleFunc("/v1/products/", h.handleProduct)
	mux.HandleFunc("/v1/products/search", h.handleSearchProducts)
	mux.HandleFunc("/v1/products/variants", h.handleProductVariants)
}

func (h *ProductHandler) handleProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.handleListProducts(w, r, ctx)
	case http.MethodPost:
		h.createProduct(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProductHandler) handleProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Path[len("/v1/products/"):]
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		if productID == "" {
			http.Error(w, "Use /v1/products for listing", http.StatusBadRequest)
			return
		}
		h.getProduct(w, r, ctx, productID)
	case http.MethodPut:
		if productID == "" {
			http.Error(w, "Product ID required", http.StatusBadRequest)
			return
		}
		h.updateProduct(w, r, ctx, productID)
	case http.MethodDelete:
		if productID == "" {
			http.Error(w, "Product ID required", http.StatusBadRequest)
			return
		}
		h.deleteProduct(w, r, ctx, productID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProductHandler) handleListProducts(w http.ResponseWriter, r *http.Request, ctx context.Context) {
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

	req := &productpb.ListProductsRequest{
		CategoryId: r.URL.Query().Get("category_id"),
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
		Pagination: &sharedpb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}

	resp, err := h.client.ListProducts(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) handleSearchProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	req := &productpb.SearchProductsRequest{
		Query:       r.URL.Query().Get("q"),
		CategoryId:  r.URL.Query().Get("category_id"),
		InStockOnly: r.URL.Query().Get("in_stock_only") == "true",
		Pagination: &sharedpb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}

	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		if val, err := strconv.ParseInt(minPrice, 10, 64); err == nil {
			req.MinPrice = &sharedpb.Money{
				Units:    val,
				Currency: "JPY",
			}
		}
	}

	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		if val, err := strconv.ParseInt(maxPrice, 10, 64); err == nil {
			req.MaxPrice = &sharedpb.Money{
				Units:    val,
				Currency: "JPY",
			}
		}
	}

	resp, err := h.client.SearchProducts(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) handleProductVariants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	req := &productpb.GetProductVariantsRequest{ProductId: productID}

	resp, err := h.client.GetProductVariants(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request, ctx context.Context, productID string) {
	req := &productpb.GetProductRequest{ProductId: productID}
	resp, err := h.client.GetProduct(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req productpb.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateProduct(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request, ctx context.Context, productID string) {
	var req productpb.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.ProductId = productID

	resp, err := h.client.UpdateProduct(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request, ctx context.Context, productID string) {
	req := &productpb.DeleteProductRequest{ProductId: productID}
	_, err := h.client.DeleteProduct(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func handleError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	switch st.Code() {
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.AlreadyExists:
		http.Error(w, st.Message(), http.StatusConflict)
	default:
		http.Error(w, st.Message(), http.StatusInternalServerError)
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
