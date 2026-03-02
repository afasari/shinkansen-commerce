package handler

import (
	"context"

	paymentpb "github.com/afasari/shinkansen-commerce/gen/proto/go/payment"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"go.uber.org/zap"
)

// Handler wraps the payment service and implements the gRPC server interface
type Handler struct {
	paymentpb.UnimplementedPaymentServiceServer
	service paymentpb.PaymentServiceServer
	logger  *zap.Logger
}

// NewHandler creates a new payment service handler
func NewHandler(service paymentpb.PaymentServiceServer, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreatePayment(ctx context.Context, req *paymentpb.CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	h.logger.Debug("CreatePayment called", zap.String("order_id", req.OrderId))
	return h.service.CreatePayment(ctx, req)
}

func (h *Handler) GetPayment(ctx context.Context, req *paymentpb.GetPaymentRequest) (*paymentpb.GetPaymentResponse, error) {
	h.logger.Debug("GetPayment called", zap.String("payment_id", req.PaymentId))
	return h.service.GetPayment(ctx, req)
}

func (h *Handler) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	h.logger.Debug("ProcessPayment called", zap.String("payment_id", req.PaymentId))
	return h.service.ProcessPayment(ctx, req)
}

func (h *Handler) RefundPayment(ctx context.Context, req *paymentpb.RefundPaymentRequest) (*sharedpb.Empty, error) {
	h.logger.Debug("RefundPayment called", zap.String("payment_id", req.PaymentId))
	return h.service.RefundPayment(ctx, req)
}
