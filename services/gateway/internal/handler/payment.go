package handler

import (
	"context"
	"encoding/json"
	"net/http"

	paymentpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/payment"
	"google.golang.org/grpc"
)

type PaymentHandler struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentHandler(conn *grpc.ClientConn) *PaymentHandler {
	return &PaymentHandler{
		client: paymentpb.NewPaymentServiceClient(conn),
	}
}

func (h *PaymentHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/payments", h.handlePayments)
	mux.HandleFunc("/v1/payments/", h.handlePayment)
}

func (h *PaymentHandler) handlePayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		h.createPayment(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PaymentHandler) handlePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentID := r.URL.Path[len("/v1/payments/"):]
	parts := splitPath(paymentID)
	if len(parts) > 1 {
		if parts[1] == "process" {
			h.processPayment(w, r, ctx, parts[0])
			return
		}
		if parts[1] == "refund" {
			h.refundPayment(w, r, ctx, parts[0])
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if paymentID == "" {
			http.Error(w, "Payment ID required", http.StatusBadRequest)
			return
		}
		h.getPayment(w, r, ctx, paymentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PaymentHandler) handlePaymentProcess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentID := r.URL.Path[len("/v1/payments/"):]
	parts := splitPath(paymentID)
	if len(parts) > 1 && parts[1] == "process" {
		if r.Method == http.MethodPost {
			h.processPayment(w, r, ctx, parts[0])
			return
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *PaymentHandler) handlePaymentRefund(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentID := r.URL.Path[len("/v1/payments/"):]
	parts := splitPath(paymentID)
	if len(parts) > 1 && parts[1] == "refund" {
		if r.Method == http.MethodPost {
			h.refundPayment(w, r, ctx, parts[0])
			return
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *PaymentHandler) createPayment(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req paymentpb.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreatePayment(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *PaymentHandler) getPayment(w http.ResponseWriter, r *http.Request, ctx context.Context, paymentID string) {
	req := &paymentpb.GetPaymentRequest{PaymentId: paymentID}
	resp, err := h.client.GetPayment(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) processPayment(w http.ResponseWriter, r *http.Request, ctx context.Context, paymentID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req paymentpb.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.PaymentId = paymentID

	resp, err := h.client.ProcessPayment(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) refundPayment(w http.ResponseWriter, r *http.Request, ctx context.Context, paymentID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req paymentpb.RefundPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.PaymentId = paymentID

	_, err := h.client.RefundPayment(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
