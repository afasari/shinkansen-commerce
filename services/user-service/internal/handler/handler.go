package handler

import (
	"context"

	userpb "github.com/afasari/shinkansen-commerce/gen/proto/go/user"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"go.uber.org/zap"
)

// Handler wraps the user service and implements the gRPC server interface
type Handler struct {
	userpb.UnimplementedUserServiceServer
	service userpb.UserServiceServer
	logger  *zap.Logger
}

// NewHandler creates a new user service handler
func NewHandler(service userpb.UserServiceServer, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	h.logger.Debug("RegisterUser called", zap.String("email", req.Email))
	return h.service.RegisterUser(ctx, req)
}

func (h *Handler) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	h.logger.Debug("LoginUser called", zap.String("email", req.Email))
	return h.service.LoginUser(ctx, req)
}

func (h *Handler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	h.logger.Debug("GetUser called", zap.String("user_id", req.UserId))
	return h.service.GetUser(ctx, req)
}

func (h *Handler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	h.logger.Debug("UpdateUser called", zap.String("user_id", req.UserId))
	return h.service.UpdateUser(ctx, req)
}

func (h *Handler) AddAddress(ctx context.Context, req *userpb.AddAddressRequest) (*userpb.AddAddressResponse, error) {
	h.logger.Debug("AddAddress called", zap.String("user_id", req.UserId))
	return h.service.AddAddress(ctx, req)
}

func (h *Handler) ListAddresses(ctx context.Context, req *userpb.ListAddressesRequest) (*userpb.ListAddressesResponse, error) {
	h.logger.Debug("ListAddresses called", zap.String("user_id", req.UserId))
	return h.service.ListAddresses(ctx, req)
}

func (h *Handler) UpdateAddress(ctx context.Context, req *userpb.UpdateAddressRequest) (*userpb.UpdateAddressResponse, error) {
	h.logger.Debug("UpdateAddress called", zap.String("address_id", req.AddressId))
	return h.service.UpdateAddress(ctx, req)
}

func (h *Handler) DeleteAddress(ctx context.Context, req *userpb.DeleteAddressRequest) (*sharedpb.Empty, error) {
	h.logger.Debug("DeleteAddress called", zap.String("address_id", req.AddressId))
	return h.service.DeleteAddress(ctx, req)
}
