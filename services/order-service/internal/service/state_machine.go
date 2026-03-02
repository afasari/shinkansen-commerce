package service

import (
	"fmt"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	"go.uber.org/zap"
)

// OrderStateMachine manages order status transitions
type OrderStateMachine struct {
	logger *zap.Logger
}

// NewOrderStateMachine creates a new order state machine
func NewOrderStateMachine(logger *zap.Logger) *OrderStateMachine {
	return &OrderStateMachine{
		logger: logger,
	}
}

// Valid transitions map
var validTransitions = map[orderpb.OrderStatus][]orderpb.OrderStatus{
	orderpb.OrderStatus_ORDER_STATUS_PENDING: {
		orderpb.OrderStatus_ORDER_STATUS_CONFIRMED,
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
		orderpb.OrderStatus_ORDER_STATUS_EXPIRED,
	},
	orderpb.OrderStatus_ORDER_STATUS_CONFIRMED: {
		orderpb.OrderStatus_ORDER_STATUS_PROCESSING,
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
	},
	orderpb.OrderStatus_ORDER_STATUS_PROCESSING: {
		orderpb.OrderStatus_ORDER_STATUS_SHIPPED,
		orderpb.OrderStatus_ORDER_STATUS_READY_FOR_PICKUP,
	},
	orderpb.OrderStatus_ORDER_STATUS_SHIPPED: {
		orderpb.OrderStatus_ORDER_STATUS_IN_TRANSIT,
	},
	orderpb.OrderStatus_ORDER_STATUS_IN_TRANSIT: {
		orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
		orderpb.OrderStatus_ORDER_STATUS_FAILED_DELIVERY,
	},
	orderpb.OrderStatus_ORDER_STATUS_READY_FOR_PICKUP: {
		orderpb.OrderStatus_ORDER_STATUS_PICKED_UP,
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
	},
	orderpb.OrderStatus_ORDER_STATUS_PICKED_UP: {
		orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
	},
	orderpb.OrderStatus_ORDER_STATUS_FAILED_DELIVERY: {
		orderpb.OrderStatus_ORDER_STATUS_RETURNED,
		orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
	},
	orderpb.OrderStatus_ORDER_STATUS_DELIVERED: {
		orderpb.OrderStatus_ORDER_STATUS_RETURNED,
	},
	orderpb.OrderStatus_ORDER_STATUS_CANCELLED:    {},
	orderpb.OrderStatus_ORDER_STATUS_EXPIRED:      {},
	orderpb.OrderStatus_ORDER_STATUS_RETURNED:     {},
}

// CanTransition checks if a status transition is valid
func (sm *OrderStateMachine) CanTransition(from, to orderpb.OrderStatus) bool {
	allowedStates, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}

	return false
}

// Transition performs a status transition
func (sm *OrderStateMachine) Transition(
	orderID string,
	currentStatus orderpb.OrderStatus,
	newStatus orderpb.OrderStatus,
) error {
	if !sm.CanTransition(currentStatus, newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s for order %s",
			currentStatus, newStatus, orderID)
	}

	sm.logger.Info("Order status transition",
		zap.String("order_id", orderID),
		zap.String("from", currentStatus.String()),
		zap.String("to", newStatus.String()))

	return nil
}

// IsFinal checks if a status is a final state (no further transitions)
func (sm *OrderStateMachine) IsFinal(status orderpb.OrderStatus) bool {
	finalStates := []orderpb.OrderStatus{
		orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
		orderpb.OrderStatus_ORDER_STATUS_EXPIRED,
		orderpb.OrderStatus_ORDER_STATUS_RETURNED,
	}

	for _, s := range finalStates {
		if s == status {
			return true
		}
	}

	return false
}

// IsCancellable checks if an order can be cancelled
func (sm *OrderStateMachine) IsCancellable(status orderpb.OrderStatus) bool {
	cancellableStates := []orderpb.OrderStatus{
		orderpb.OrderStatus_ORDER_STATUS_PENDING,
		orderpb.OrderStatus_ORDER_STATUS_CONFIRMED,
		orderpb.OrderStatus_ORDER_STATUS_READY_FOR_PICKUP,
	}

	for _, s := range cancellableStates {
		if s == status {
			return true
		}
	}

	return false
}

// IsRefundable checks if an order is eligible for refund
func (sm *OrderStateMachine) IsRefundable(status orderpb.OrderStatus) bool {
	refundableStates := []orderpb.OrderStatus{
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
		orderpb.OrderStatus_ORDER_STATUS_RETURNED,
		orderpb.OrderStatus_ORDER_STATUS_FAILED_DELIVERY,
	}

	for _, s := range refundableStates {
		if s == status {
			return true
		}
	}

	return false
}

// GetStatusDescription returns a human-readable description of the status
func (sm *OrderStateMachine) GetStatusDescription(status orderpb.OrderStatus) string {
	descriptions := map[orderpb.OrderStatus]string{
		orderpb.OrderStatus_ORDER_STATUS_PENDING:          "Order received, awaiting confirmation",
		orderpb.OrderStatus_ORDER_STATUS_CONFIRMED:        "Order confirmed, preparing for shipment",
		orderpb.OrderStatus_ORDER_STATUS_PROCESSING:       "Processing order items",
		orderpb.OrderStatus_ORDER_STATUS_SHIPPED:          "Order has been shipped",
		orderpb.OrderStatus_ORDER_STATUS_IN_TRANSIT:       "Order is in transit to delivery address",
		orderpb.OrderStatus_ORDER_STATUS_DELIVERED:        "Order has been delivered",
		orderpb.OrderStatus_ORDER_STATUS_READY_FOR_PICKUP: "Order ready for pickup at store",
		orderpb.OrderStatus_ORDER_STATUS_PICKED_UP:        "Order has been picked up",
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED:        "Order has been cancelled",
		orderpb.OrderStatus_ORDER_STATUS_EXPIRED:          "Order payment expired",
		orderpb.OrderStatus_ORDER_STATUS_FAILED_DELIVERY:  "Delivery attempt failed",
		orderpb.OrderStatus_ORDER_STATUS_RETURNED:         "Order has been returned",
	}

	if desc, ok := descriptions[status]; ok {
		return desc
	}
	return "Unknown status"
}

// GetCustomerAction returns recommended action for customer based on status
func (sm *OrderStateMachine) GetCustomerAction(status orderpb.OrderStatus) string {
	actions := map[orderpb.OrderStatus]string{
		orderpb.OrderStatus_ORDER_STATUS_PENDING:          "Please complete payment to confirm your order",
		orderpb.OrderStatus_ORDER_STATUS_CONFIRMED:        "Your order is being prepared",
		orderpb.OrderStatus_ORDER_STATUS_PROCESSING:       "Your order is being processed",
		orderpb.OrderStatus_ORDER_STATUS_SHIPPED:          "Track your package using the tracking number",
		orderpb.OrderStatus_ORDER_STATUS_IN_TRANSIT:       "Your package is on its way",
		orderpb.OrderStatus_ORDER_STATUS_DELIVERED:        "Enjoy your purchase! Thank you for shopping with us",
		orderpb.OrderStatus_ORDER_STATUS_READY_FOR_PICKUP: "Your order is ready for pickup at the store",
		orderpb.OrderStatus_ORDER_STATUS_PICKED_UP:        "You have picked up your order",
		orderpb.OrderStatus_ORDER_STATUS_CANCELLED:        "Your order has been cancelled",
		orderpb.OrderStatus_ORDER_STATUS_EXPIRED:          "Your order payment has expired. Please place a new order",
		orderpb.OrderStatus_ORDER_STATUS_FAILED_DELIVERY:  "Delivery failed. Please contact customer support",
		orderpb.OrderStatus_ORDER_STATUS_RETURNED:         "Your order has been returned. Refund will be processed",
	}

	if action, ok := actions[status]; ok {
		return action
	}
	return "Please contact customer support for assistance"
}

// AutoTransition handles automatic state transitions based on conditions
func (sm *OrderStateMachine) AutoTransition(
	orderID string,
	status orderpb.OrderStatus,
	condition string,
) (orderpb.OrderStatus, bool) {
	transitions := map[string]map[orderpb.OrderStatus]orderpb.OrderStatus{
		"payment_timeout": {
			orderpb.OrderStatus_ORDER_STATUS_PENDING: orderpb.OrderStatus_ORDER_STATUS_EXPIRED,
		},
		"payment_completed": {
			orderpb.OrderStatus_ORDER_STATUS_PENDING: orderpb.OrderStatus_ORDER_STATUS_CONFIRMED,
		},
		"pickup_timeout": {
			orderpb.OrderStatus_ORDER_STATUS_READY_FOR_PICKUP: orderpb.OrderStatus_ORDER_STATUS_CANCELLED,
		},
		"delivery_confirmed": {
			orderpb.OrderStatus_ORDER_STATUS_IN_TRANSIT: orderpb.OrderStatus_ORDER_STATUS_DELIVERED,
		},
	}

	if conditionMap, ok := transitions[condition]; ok {
		if newStatus, ok := conditionMap[status]; ok {
			return newStatus, true
		}
	}

	return status, false
}
