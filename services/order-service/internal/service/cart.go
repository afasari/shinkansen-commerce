package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
)

// CartService handles shopping cart operations
type CartService struct {
	productClient productpb.ProductServiceClient
	redisClient   *redis.Client
	logger        *zap.Logger
}

// NewCartService creates a new cart service
func NewCartService(
	productClient productpb.ProductServiceClient,
	redisClient *redis.Client,
	logger *zap.Logger,
) *CartService {
	return &CartService{
		productClient: productClient,
		redisClient:   redisClient,
		logger:        logger,
	}
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID     string    `json:"product_id"`
	VariantID     string    `json:"variant_id"`
	Quantity      int32     `json:"quantity"`
	UnitPrice     int64     `json:"unit_price"`
	PriceCurrency string    `json:"price_currency"`
	AddedAt       time.Time `json:"added_at"`
}

// Cart represents a shopping cart
type Cart struct {
	UserID    string     `json:"user_id"`
	SessionID string     `json:"session_id,omitempty"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// CartKey generates a Redis key for a cart
func (s *CartService) CartKey(userID string, sessionID string) string {
	if sessionID != "" {
		return fmt.Sprintf("cart:session:%s", sessionID)
	}
	return fmt.Sprintf("cart:user:%s", userID)
}

// GetCart retrieves a cart from Redis or creates a new one
func (s *CartService) GetCart(ctx context.Context, userID, sessionID string) (*Cart, error) {
	key := s.CartKey(userID, sessionID)

	val, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cart doesn't exist, create new one
		return &Cart{
			UserID:    userID,
			SessionID: sessionID,
			Items:     []CartItem{},
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	var cart Cart
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	return &cart, nil
}

// SaveCart saves a cart to Redis
func (s *CartService) SaveCart(ctx context.Context, cart *Cart) error {
	key := s.CartKey(cart.UserID, cart.SessionID)
	cart.UpdatedAt = time.Now()

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	// Set cart to expire in 7 days
	ttl := time.Until(cart.ExpiresAt.Add(7 * 24 * time.Hour))
	if ttl < 0 {
		ttl = 7 * 24 * time.Hour
	}

	return s.redisClient.Set(ctx, key, data, ttl).Err()
}

// AddItem adds an item to the cart
func (s *CartService) AddItem(ctx context.Context, userID, sessionID, productID, variantID string, quantity int32) (*Cart, error) {
	s.logger.Info("Adding item to cart",
		zap.String("user_id", userID),
		zap.String("product_id", productID),
		zap.Int32("quantity", quantity))

	// Validate product exists and get price
	product, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product.Product.StockQuantity < quantity {
		return nil, fmt.Errorf("insufficient stock: available %d, requested %d",
			product.Product.StockQuantity, quantity)
	}

	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists in cart
	for i, item := range cart.Items {
		if item.ProductID == productID && item.VariantID == variantID {
			cart.Items[i].Quantity += quantity
			cart.Items[i].UnitPrice = product.Product.Price.Units
			if err := s.SaveCart(ctx, cart); err != nil {
				return nil, err
			}
			return cart, nil
		}
	}

	// Add new item
	cart.Items = append(cart.Items, CartItem{
		ProductID:     productID,
		VariantID:     variantID,
		Quantity:      quantity,
		UnitPrice:     product.Product.Price.Units,
		PriceCurrency: product.Product.Price.Currency,
		AddedAt:       time.Now(),
	})

	if err := s.SaveCart(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// UpdateItem updates the quantity of an item in the cart
func (s *CartService) UpdateItem(ctx context.Context, userID, sessionID, productID, variantID string, quantity int32) (*Cart, error) {
	s.logger.Info("Updating cart item",
		zap.String("user_id", userID),
		zap.String("product_id", productID),
		zap.Int32("quantity", quantity))

	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	for i, item := range cart.Items {
		if item.ProductID == productID && item.VariantID == variantID {
			if quantity <= 0 {
				// Remove item
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			} else {
				// Validate stock
				product, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{
					ProductId: productID,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get product: %w", err)
				}

				if product.Product.StockQuantity < quantity {
					return nil, fmt.Errorf("insufficient stock: available %d, requested %d",
						product.Product.StockQuantity, quantity)
				}

				cart.Items[i].Quantity = quantity
				cart.Items[i].UnitPrice = product.Product.Price.Units
			}

			if err := s.SaveCart(ctx, cart); err != nil {
				return nil, err
			}
			return cart, nil
		}
	}

	return nil, fmt.Errorf("item not found in cart")
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(ctx context.Context, userID, sessionID, productID, variantID string) (*Cart, error) {
	s.logger.Info("Removing item from cart",
		zap.String("user_id", userID),
		zap.String("product_id", productID))

	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	for i, item := range cart.Items {
		if item.ProductID == productID && item.VariantID == variantID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)

			if err := s.SaveCart(ctx, cart); err != nil {
				return nil, err
			}
			return cart, nil
		}
	}

	return nil, fmt.Errorf("item not found in cart")
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(ctx context.Context, userID, sessionID string) error {
	s.logger.Info("Clearing cart", zap.String("user_id", userID))

	key := s.CartKey(userID, sessionID)
	return s.redisClient.Del(ctx, key).Err()
}

// MergeCart merges a session cart into a user cart
func (s *CartService) MergeCart(ctx context.Context, userID, sessionID string) (*Cart, error) {
	s.logger.Info("Merging session cart into user cart",
		zap.String("user_id", userID),
		zap.String("session_id", sessionID))

	sessionCart, err := s.GetCart(ctx, "", sessionID)
	if err != nil {
		return nil, err
	}

	// If session cart is empty, just return user cart
	if len(sessionCart.Items) == 0 {
		return s.GetCart(ctx, userID, "")
	}

	userCart, err := s.GetCart(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	// Merge items
	for _, sessionItem := range sessionCart.Items {
		found := false
		for i, userItem := range userCart.Items {
			if userItem.ProductID == sessionItem.ProductID && userItem.VariantID == sessionItem.VariantID {
				userCart.Items[i].Quantity += sessionItem.Quantity
				userCart.Items[i].UnitPrice = sessionItem.UnitPrice
				found = true
				break
			}
		}
		if !found {
			userCart.Items = append(userCart.Items, sessionItem)
		}
	}

	// Save merged cart
	if err := s.SaveCart(ctx, userCart); err != nil {
		return nil, err
	}

	// Delete session cart
	if err := s.ClearCart(ctx, "", sessionID); err != nil {
		s.logger.Warn("Failed to delete session cart after merge", zap.Error(err))
	}

	return userCart, nil
}

// CartToOrderItems converts cart items to order items
func (s *CartService) CartToOrderItems(cart *Cart) []*orderpb.OrderItem {
	items := make([]*orderpb.OrderItem, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = &orderpb.OrderItem{
			ProductId: item.ProductID,
			VariantId: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: &sharedpb.Money{
				Units:    item.UnitPrice,
				Currency: item.PriceCurrency,
			},
		}
	}
	return items
}

// GetCartSummary returns a summary of the cart
func (s *CartService) GetCartSummary(ctx context.Context, userID, sessionID string) (*orderpb.CartSummary, error) {
	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	var subtotalUnits int64
	currency := "JPY"

	for _, item := range cart.Items {
		itemTotal := item.UnitPrice * int64(item.Quantity)
		subtotalUnits += itemTotal
		currency = item.PriceCurrency
	}

	itemCount := int32(0)
	for _, item := range cart.Items {
		itemCount += item.Quantity
	}

	return &orderpb.CartSummary{
		ItemCount: itemCount,
		Subtotal: &sharedpb.Money{
			Units:    subtotalUnits,
			Currency: currency,
		},
	}, nil
}
