package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"go.uber.org/zap"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
)

// OrderEventPublisher publishes order-related events to Kafka
type OrderEventPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   *zap.Logger
}

// NewOrderEventPublisher creates a new order event publisher
func NewOrderEventProducer(
	brokers []string,
	topic string,
	logger *zap.Logger,
) (*OrderEventPublisher, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &OrderEventPublisher{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

// Close closes the Kafka producer
func (p *OrderEventPublisher) Close() error {
	return p.producer.Close()
}

// OrderEvent represents an order event
type OrderEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	OrderID   string                 `json:"order_id"`
	UserID    string                 `json:"user_id"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PublishOrderCreated publishes an order created event
func (p *OrderEventPublisher) PublishOrderCreated(ctx context.Context, order *orderpb.Order) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.created",
		OrderID:   order.Id,
		UserID:    order.UserId,
		Status:    order.Status.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"order_number":     order.OrderNumber,
			"total_amount":     order.TotalAmount.Units,
			"currency":         order.TotalAmount.Currency,
			"payment_method":   order.PaymentMethod.String(),
			"item_count":       len(order.Items),
			"shipping_address": order.ShippingAddress,
			"points_applied":   order.PointsApplied,
		},
	}

	return p.publish(ctx, event)
}

// PublishOrderStatusChanged publishes an order status change event
func (p *OrderEventPublisher) PublishOrderStatusChanged(
	ctx context.Context,
	orderID, userID string,
	oldStatus, newStatus orderpb.OrderStatus,
	reason string,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.status_changed",
		OrderID:   orderID,
		UserID:    userID,
		Status:    newStatus.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"old_status": oldStatus.String(),
			"new_status": newStatus.String(),
			"reason":     reason,
		},
	}

	return p.publish(ctx, event)
}

// PublishOrderCancelled publishes an order cancelled event
func (p *OrderEventPublisher) PublishOrderCancelled(
	ctx context.Context,
	order *orderpb.Order,
	reason string,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.cancelled",
		OrderID:   order.Id,
		UserID:    order.UserId,
		Status:    order.Status.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"order_number":  order.OrderNumber,
			"cancel_reason": reason,
			"total_amount":  order.TotalAmount.Units,
			"currency":      order.TotalAmount.Currency,
		},
	}

	return p.publish(ctx, event)
}

// PublishOrderPaid publishes an order payment completed event
func (p *OrderEventPublisher) PublishOrderPaid(
	ctx context.Context,
	orderID, userID string,
	paymentID string,
	amount int64,
	currency string,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.paid",
		OrderID:   orderID,
		UserID:    userID,
		Status:    orderpb.OrderStatus_ORDER_STATUS_CONFIRMED.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"payment_id": paymentID,
			"amount":     amount,
			"currency":   currency,
		},
	}

	return p.publish(ctx, event)
}

// PublishOrderShipped publishes an order shipped event
func (p *OrderEventPublisher) PublishOrderShipped(
	ctx context.Context,
	orderID, userID string,
	trackingNumber string,
	carrier string,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.shipped",
		OrderID:   orderID,
		UserID:    userID,
		Status:    orderpb.OrderStatus_ORDER_STATUS_SHIPPED.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"tracking_number": trackingNumber,
			"carrier":         carrier,
		},
	}

	return p.publish(ctx, event)
}

// PublishOrderDelivered publishes an order delivered event
func (p *OrderEventPublisher) PublishOrderDelivered(
	ctx context.Context,
	orderID, userID string,
	deliveryTime time.Time,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.delivered",
		OrderID:   orderID,
		UserID:    userID,
		Status:    orderpb.OrderStatus_ORDER_STATUS_DELIVERED.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"delivered_at": deliveryTime.Format(time.RFC3339),
		},
	}

	return p.publish(ctx, event)
}

// PublishPointsApplied publishes a points applied event
func (p *OrderEventPublisher) PublishPointsApplied(
	ctx context.Context,
	orderID, userID string,
	pointsApplied int64,
	yenValue int64,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.points_applied",
		OrderID:   orderID,
		UserID:    userID,
		Status:    "",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"points_applied": pointsApplied,
			"yen_value":      yenValue,
		},
	}

	return p.publish(ctx, event)
}

// PublishDeliverySlotReserved publishes a delivery slot reserved event
func (p *OrderEventPublisher) PublishDeliverySlotReserved(
	ctx context.Context,
	orderID, userID string,
	slotID, reservationID string,
	deliveryDate time.Time,
) error {
	event := OrderEvent{
		EventID:   uuid.New().String(),
		EventType: "order.delivery_slot_reserved",
		OrderID:   orderID,
		UserID:    userID,
		Status:    "",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"slot_id":        slotID,
			"reservation_id": reservationID,
			"delivery_date":  deliveryDate.Format(time.RFC3339),
		},
	}

	return p.publish(ctx, event)
}

// publish sends an event to Kafka
func (p *OrderEventPublisher) publish(ctx context.Context, event OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := sarama.StringEncoder(event.OrderID)
	message := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       key,
		Value:     sarama.ByteEncoder(data),
		Timestamp: time.Now(),
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		p.logger.Error("Failed to publish order event",
			zap.String("event_type", event.EventType),
			zap.String("order_id", event.OrderID),
			zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.logger.Info("Published order event",
		zap.String("event_type", event.EventType),
		zap.String("order_id", event.OrderID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// OrderEventConsumer consumes order events
type OrderEventConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	handler  OrderEventHandler
	logger   *zap.Logger
}

// OrderEventHandler handles order events
type OrderEventHandler interface {
	HandleOrderCreated(ctx context.Context, event OrderEvent) error
	HandleOrderStatusChanged(ctx context.Context, event OrderEvent) error
	HandleOrderCancelled(ctx context.Context, event OrderEvent) error
}

// NewOrderEventConsumer creates a new order event consumer
func NewOrderEventConsumer(
	brokers []string,
	groupID, topic string,
	handler OrderEventHandler,
	logger *zap.Logger,
) (*OrderEventConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	balanceStrategy := sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Group.Rebalance.Strategy = balanceStrategy
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &OrderEventConsumer{
		consumer: consumer,
		topic:    topic,
		handler:  handler,
		logger:   logger,
	}, nil
}

// Consume starts consuming events
func (c *OrderEventConsumer) Consume(ctx context.Context) error {
	for {
		if err := c.consumer.Consume(ctx, []string{c.topic}, c); err != nil {
			return fmt.Errorf("error from consumer: %w", err)
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

// Setup is called at the beginning of a new session
func (c *OrderEventConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called at the end of a session
func (c *OrderEventConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages
func (c *OrderEventConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event OrderEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Error("Failed to unmarshal event", zap.Error(err))
			continue
		}

		var err error
		switch event.EventType {
		case "order.created":
			err = c.handler.HandleOrderCreated(context.Background(), event)
		case "order.status_changed":
			err = c.handler.HandleOrderStatusChanged(context.Background(), event)
		case "order.cancelled":
			err = c.handler.HandleOrderCancelled(context.Background(), event)
		}

		if err != nil {
			c.logger.Error("Failed to handle event",
				zap.String("event_type", event.EventType),
				zap.String("order_id", event.OrderID),
				zap.Error(err))
		}

		session.MarkMessage(msg, "")
	}

	return nil
}

// Close closes the consumer
func (c *OrderEventConsumer) Close() error {
	return c.consumer.Close()
}
