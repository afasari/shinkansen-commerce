package handler

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

func RegisterHandlers(ctx context.Context, mux *http.ServeMux, productConn, orderConn, userConn, paymentConn, inventoryConn, deliveryConn *grpc.ClientConn) error {
	productHandler := NewProductHandler(productConn)
	productHandler.RegisterHandlers(mux)

	orderHandler := NewOrderHandler(orderConn)
	orderHandler.RegisterHandlers(mux)

	userHandler := NewUserHandler(userConn)
	userHandler.RegisterHandlers(mux)

	paymentHandler := NewPaymentHandler(paymentConn)
	paymentHandler.RegisterHandlers(mux)

	inventoryHandler := NewInventoryHandler(inventoryConn)
	inventoryHandler.RegisterHandlers(mux)

	deliveryHandler := NewDeliveryHandler(deliveryConn)
	deliveryHandler.RegisterHandlers(mux)

	return nil
}
