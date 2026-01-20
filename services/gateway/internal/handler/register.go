package handler

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

func RegisterHandlers(ctx context.Context, mux *http.ServeMux, conn *grpc.ClientConn) error {
	productHandler := NewProductHandler(conn)
	productHandler.RegisterHandlers(mux)
	return nil
}
