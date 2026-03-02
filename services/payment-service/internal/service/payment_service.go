package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentpb "github.com/afasari/shinkansen-commerce/gen/proto/go/payment"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/payment-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/payment-service/internal/db"
	"go.uber.org/zap"
)

type PaymentService struct {
	paymentpb.UnimplementedPaymentServiceServer
	queries db.Querier
	cache   cache.Cache
	logger  *zap.Logger
}

func NewPaymentService(queries db.Querier, cacheClient cache.Cache, logger *zap.Logger) *PaymentService {
	return &PaymentService{
		queries: queries,
		cache:   cacheClient,
		logger:  logger,
	}
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *paymentpb.CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	s.logger.Info("Creating payment", zap.String("order_id", req.OrderId))

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	paymentID, err := s.queries.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:     orderID,
		Method:      req.Method.String(),
		AmountMinor: int(req.Amount.Units),
		Currency:    req.Amount.Currency,
	})

	if err != nil {
		s.logger.Error("Failed to create payment", zap.Error(err))
		if isDuplicateKeyError(err) {
			return nil, status.Error(codes.AlreadyExists, "payment already exists for this order")
		}
		return nil, status.Error(codes.Internal, "failed to create payment")
	}

	return &paymentpb.CreatePaymentResponse{
		PaymentId: paymentID.String(),
		Status:    paymentpb.PaymentStatus_PAYMENT_STATUS_PENDING,
	}, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, req *paymentpb.GetPaymentRequest) (*paymentpb.GetPaymentResponse, error) {
	s.logger.Info("Getting payment", zap.String("payment_id", req.PaymentId))

	paymentID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment_id")
	}

	cacheKey := cache.PaymentCacheKey(req.PaymentId)
	var cached db.Payment
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("Payment cache hit", zap.String("payment_id", req.PaymentId))
		return &paymentpb.GetPaymentResponse{
			Payment: s.paymentToProto(cached),
		}, nil
	}

	payment, err := s.queries.GetPayment(ctx, paymentID)
	if err != nil {
		s.logger.Error("Failed to get payment", zap.Error(err))
		return nil, status.Error(codes.NotFound, "payment not found")
	}

	if err := s.cache.Set(ctx, cacheKey, payment, cache.DefaultTTL); err != nil {
		s.logger.Warn("Failed to cache payment", zap.Error(err))
	}

	return &paymentpb.GetPaymentResponse{
		Payment: s.paymentToProto(payment),
	}, nil
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	s.logger.Info("Processing payment", zap.String("payment_id", req.PaymentId))

	paymentID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment_id")
	}

	payment, err := s.queries.GetPayment(ctx, paymentID)
	if err != nil {
		s.logger.Error("Failed to get payment", zap.Error(err))
		return nil, status.Error(codes.NotFound, "payment not found")
	}

	if payment.Status != "PAYMENT_STATUS_PENDING" {
		return nil, status.Error(codes.FailedPrecondition, "not in pending status")
	}

	paymentStatus, transactionID, err := s.processWithGateway(ctx, payment, req.PaymentData)
	if err != nil {
		return nil, err
	}

	paymentDataBytes, _ := json.Marshal(req.PaymentData)
	if err := s.queries.UpdatePaymentData(ctx, db.UpdatePaymentDataParams{
		ID:          payment.ID,
		PaymentData: paymentDataBytes,
	}); err != nil {
		s.logger.Warn("Failed to update payment data", zap.Error(err))
	}

	if err := s.queries.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:            payment.ID,
		Status:        paymentStatus.String(),
		TransactionID: &transactionID,
	}); err != nil {
		s.logger.Error("Failed to update payment status", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update payment status")
	}

	_ = s.cache.Delete(ctx, cache.PaymentCacheKey(req.PaymentId))
	_ = s.cache.Delete(ctx, cache.PaymentsByOrderCacheKey(payment.OrderID.String()))

	return &paymentpb.ProcessPaymentResponse{
		Status:        paymentStatus,
		TransactionId: transactionID,
	}, nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, req *paymentpb.RefundPaymentRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Refunding payment", zap.String("payment_id", req.PaymentId))

	paymentID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("invalid payment_id: %w", err)
	}

	payment, err := s.queries.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.Status != "PAYMENT_STATUS_COMPLETED" {
		return nil, fmt.Errorf("payment cannot be refunded: %s", payment.Status)
	}

	status, transactionID, err := s.refundWithGateway(ctx, payment, req.Amount)
	if err != nil {
		return nil, err
	}

	if err := s.queries.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:            payment.ID,
		Status:        status.String(),
		TransactionID: &transactionID,
	}); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	_ = s.cache.Delete(ctx, cache.PaymentCacheKey(req.PaymentId))
	_ = s.cache.Delete(ctx, cache.PaymentsByOrderCacheKey(payment.OrderID.String()))

	return &sharedpb.Empty{}, nil
}

func (s *PaymentService) processWithGateway(ctx context.Context, payment db.Payment, paymentData map[string]string) (paymentpb.PaymentStatus, string, error) {
	s.logger.Info("Processing payment with gateway",
		zap.String("payment_id", payment.ID.String()),
		zap.String("method", payment.Method))

	var status paymentpb.PaymentStatus
	var transactionID string

	switch payment.Method {
	case "PAYMENT_METHOD_CREDIT_CARD":
		status, transactionID = s.processCreditCard(payment, paymentData)
	case "PAYMENT_METHOD_PAYPAY":
		status, transactionID = s.processPayPay(payment, paymentData)
	case "PAYMENT_METHOD_RAKUTEN_PAY":
		status, transactionID = s.processRakutenPay(payment, paymentData)
	case "PAYMENT_METHOD_KONBINI_SEVENELEVEN",
		"PAYMENT_METHOD_KONBINI_LAWSON",
		"PAYMENT_METHOD_KONBINI_FAMILYMART":
		status, transactionID = s.processKonbini(payment, paymentData)
	default:
		return paymentpb.PaymentStatus_PAYMENT_STATUS_FAILED, "", fmt.Errorf("unsupported payment method: %s", payment.Method)
	}

	return status, transactionID, nil
}

func (s *PaymentService) processCreditCard(payment db.Payment, paymentData map[string]string) (paymentpb.PaymentStatus, string) {
	transactionID := fmt.Sprintf("CC-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	return paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED, transactionID
}

func (s *PaymentService) processPayPay(payment db.Payment, paymentData map[string]string) (paymentpb.PaymentStatus, string) {
	transactionID := fmt.Sprintf("PAYPAY-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	return paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED, transactionID
}

func (s *PaymentService) processRakutenPay(payment db.Payment, paymentData map[string]string) (paymentpb.PaymentStatus, string) {
	transactionID := fmt.Sprintf("RAKUTEN-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	return paymentpb.PaymentStatus_PAYMENT_STATUS_COMPLETED, transactionID
}

func (s *PaymentService) processKonbini(payment db.Payment, paymentData map[string]string) (paymentpb.PaymentStatus, string) {
	transactionID := fmt.Sprintf("KONBINI-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	return paymentpb.PaymentStatus_PAYMENT_STATUS_PROCESSING, transactionID
}

func (s *PaymentService) refundWithGateway(ctx context.Context, payment db.Payment, amount *sharedpb.Money) (paymentpb.PaymentStatus, string, error) {
	s.logger.Info("Processing refund with gateway",
		zap.String("payment_id", payment.ID.String()),
		zap.String("original_transaction_id", *payment.TransactionID))

	transactionID := fmt.Sprintf("REFUND-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	return paymentpb.PaymentStatus_PAYMENT_STATUS_REFUNDED, transactionID, nil
}

func (s *PaymentService) paymentToProto(p db.Payment) *paymentpb.Payment {
	method := paymentpb.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	if p.Method != "" {
		method = paymentpb.PaymentMethod(paymentpb.PaymentMethod_value[p.Method])
	}

	status := paymentpb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	if p.Status != "" {
		status = paymentpb.PaymentStatus(paymentpb.PaymentStatus_value[p.Status])
	}

	return &paymentpb.Payment{
		Id:            p.ID.String(),
		OrderId:       p.OrderID.String(),
		Method:        method,
		Amount:        &sharedpb.Money{Units: int64(p.AmountMinor), Currency: p.Currency},
		Status:        status,
		TransactionId: toStringPtr(p.TransactionID),
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}

func toStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
