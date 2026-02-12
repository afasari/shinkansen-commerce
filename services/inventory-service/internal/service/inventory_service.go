package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	inventorypb "github.com/afasari/shinkansen-commerce/gen/proto/go/inventory"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/inventory-service/internal/db"
	"go.uber.org/zap"
)

type InventoryService struct {
	inventorypb.UnimplementedInventoryServiceServer
	queries db.Querier
	logger  *zap.Logger
}

func NewInventoryService(queries db.Querier, logger *zap.Logger) *InventoryService {
	return &InventoryService{
		queries: queries,
		logger:  logger,
	}
}

func (s *InventoryService) GetStock(ctx context.Context, req *inventorypb.GetStockRequest) (*inventorypb.GetStockResponse, error) {
	s.logger.Info("Getting stock",
		zap.String("product_id", req.ProductId),
		zap.String("warehouse_id", req.WarehouseId))

	productID := uuid.MustParse(req.ProductId)
	var variantID uuid.UUID
	if req.VariantId != "" {
		variantID = uuid.MustParse(req.VariantId)
	}
	warehouseID := uuid.MustParse(req.WarehouseId)

	stock, err := s.queries.GetStock(ctx, productID, variantID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	if stock.ID == (uuid.UUID{}) {
		return &inventorypb.GetStockResponse{
			Stock: &inventorypb.StockItem{
				ProductId:         req.ProductId,
				VariantId:         req.VariantId,
				WarehouseId:       req.WarehouseId,
				Quantity:          0,
				ReservedQuantity:  0,
				AvailableQuantity: 0,
				UpdatedAt:         timestamppb.Now(),
			},
		}, nil
	}

	return &inventorypb.GetStockResponse{
		Stock: s.stockToProto(stock),
	}, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	s.logger.Info("Reserving stock", zap.String("order_id", req.OrderId))

	orderID := uuid.MustParse(req.OrderId)
	expiresAt := req.ExpiresAt.AsTime()
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(30 * time.Minute)
	}

	reservationID := uuid.New()
	failedItems := []string{}
	success := true

	for _, item := range req.Items {
		productID := uuid.MustParse(item.ProductId)
		var variantID uuid.UUID
		if item.VariantId != "" {
			variantID = uuid.MustParse(item.VariantId)
		}
		warehouseID := uuid.MustParse(item.WarehouseId)

		stock, err := s.queries.GetStock(ctx, productID, variantID, warehouseID)
		if err != nil {
			s.logger.Error("Failed to get stock for reservation", zap.Error(err))
			failedItems = append(failedItems, item.ProductId)
			continue
		}

		if stock.ID == (uuid.UUID{}) || stock.AvailableQuantity < int(item.Quantity) {
			failedItems = append(failedItems, item.ProductId)
			success = false
			continue
		}

		err = s.queries.ReserveStock(ctx, db.ReserveStockParams{
			OrderID:     orderID,
			StockItemID: stock.ID,
			Quantity:    int(item.Quantity),
			ExpiresAt:   expiresAt,
		})
		if err != nil {
			s.logger.Error("Failed to reserve stock", zap.Error(err))
			failedItems = append(failedItems, item.ProductId)
			success = false
			continue
		}

		reason := fmt.Sprintf("Order: %s", req.OrderId)
		s.queries.CreateStockMovement(ctx, db.CreateStockMovementParams{
			StockItemID:  stock.ID,
			MovementType: "MOVEMENT_TYPE_RESERVATION",
			Quantity:     int(item.Quantity),
			Reference:    &reason,
		})
	}

	return &inventorypb.ReserveStockResponse{
		ReservationId: reservationID.String(),
		Success:       success,
		FailedItems:   failedItems,
	}, nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, req *inventorypb.ReleaseStockRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Releasing stock", zap.String("reservation_id", req.ReservationId))

	err := s.queries.ReleaseStock(ctx, db.ReleaseStockParams{
		OrderID: uuid.MustParse(req.ReservationId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to release stock: %w", err)
	}

	return &sharedpb.Empty{}, nil
}

func (s *InventoryService) UpdateStock(ctx context.Context, req *inventorypb.UpdateStockRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Updating stock",
		zap.String("product_id", req.ProductId),
		zap.Int32("delta", req.QuantityDelta))

	productID := uuid.MustParse(req.ProductId)
	var variantID *uuid.UUID
	var variantIDValue uuid.UUID
	if req.VariantId != "" {
		vid := uuid.MustParse(req.VariantId)
		variantID = &vid
		variantIDValue = vid
	}
	warehouseID := uuid.MustParse(req.WarehouseId)

	err := s.queries.UpdateStockQuantity(ctx, db.UpdateStockParams{
		ProductID:   productID,
		VariantID:   variantID,
		WarehouseID: warehouseID,
		Delta:       int(req.QuantityDelta),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	stock, _ := s.queries.GetStock(ctx, productID, variantIDValue, warehouseID)
	if stock.ID != (uuid.UUID{}) {
		movementType := "MOVEMENT_TYPE_INBOUND"
		if req.QuantityDelta < 0 {
			movementType = "MOVEMENT_TYPE_OUTBOUND"
		}

		s.queries.CreateStockMovement(ctx, db.CreateStockMovementParams{
			StockItemID:  stock.ID,
			MovementType: movementType,
			Quantity:     int(req.QuantityDelta),
			Reference:    &req.Reason,
		})
	}

	return &sharedpb.Empty{}, nil
}

func (s *InventoryService) GetStockMovements(ctx context.Context, req *inventorypb.GetStockMovementsRequest) (*inventorypb.GetStockMovementsResponse, error) {
	s.logger.Info("Getting stock movements", zap.String("stock_item_id", req.StockItemId))

	stockItemID := uuid.MustParse(req.StockItemId)

	limit := 50
	offset := 0
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
		if req.Pagination.Page > 0 {
			offset = int(req.Pagination.Page) * limit
		}
	}

	movements, err := s.queries.ListStockMovements(ctx, stockItemID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock movements: %w", err)
	}

	protoMovements := make([]*inventorypb.StockMovement, 0, len(movements))
	for _, m := range movements {
		protoMovements = append(protoMovements, s.movementToProto(m))
	}

	return &inventorypb.GetStockMovementsResponse{
		Movements: protoMovements,
	}, nil
}

func (s *InventoryService) stockToProto(stock db.StockItem) *inventorypb.StockItem {
	return &inventorypb.StockItem{
		Id:                stock.ID.String(),
		ProductId:         stock.ProductID.String(),
		VariantId:         uuidToString(stock.VariantID),
		WarehouseId:       stock.WarehouseID.String(),
		Quantity:          int32(stock.Quantity),
		ReservedQuantity:  int32(stock.ReservedQuantity),
		AvailableQuantity: int32(stock.AvailableQuantity),
		UpdatedAt:         timestamppb.New(stock.UpdatedAt),
	}
}

func (s *InventoryService) movementToProto(m db.StockMovement) *inventorypb.StockMovement {
	movementType := inventorypb.MovementType(inventorypb.MovementType_value[m.MovementType])
	return &inventorypb.StockMovement{
		Id:          m.ID.String(),
		StockItemId: m.StockItemID.String(),
		Type:        movementType,
		Quantity:    int32(m.Quantity),
		Reference:   toStringPtr(m.Reference),
		CreatedAt:   timestamppb.New(m.CreatedAt),
	}
}

func uuidToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func toStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
