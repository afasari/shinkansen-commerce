package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"

	userpb "github.com/afasari/shinkansen-commerce/gen/proto/go/user"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
)

// MockUserService is a mock implementation of userpb.UserServiceServer
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.RegisterUserResponse), args.Error(1)
}

func (m *MockUserService) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.LoginUserResponse), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.GetUserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UpdateUserResponse), args.Error(1)
}

func (m *MockUserService) AddAddress(ctx context.Context, req *userpb.AddAddressRequest) (*userpb.AddAddressResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.AddAddressResponse), args.Error(1)
}

func (m *MockUserService) ListAddresses(ctx context.Context, req *userpb.ListAddressesRequest) (*userpb.ListAddressesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.ListAddressesResponse), args.Error(1)
}

func (m *MockUserService) UpdateAddress(ctx context.Context, req *userpb.UpdateAddressRequest) (*userpb.UpdateAddressResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UpdateAddressResponse), args.Error(1)
}

func (m *MockUserService) DeleteAddress(ctx context.Context, req *userpb.DeleteAddressRequest) (*sharedpb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sharedpb.Empty), args.Error(1)
}

func TestHandler_RegisterUser(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.RegisterUserRequest{
			Email:    "test@example.com",
			Password: "SecurePass123!",
			Name:     "Test User",
		}

		expectedResp := &userpb.RegisterUserResponse{
			UserId:       "user-123",
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}

		mockService.On("RegisterUser", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.RegisterUser(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_LoginUser(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.LoginUserRequest{
			Email:    "test@example.com",
			Password: "password",
		}

		expectedResp := &userpb.LoginUserResponse{
			UserId:       "user-123",
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}

		mockService.On("LoginUser", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.LoginUser(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetUser(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.GetUserRequest{
			UserId: "user-123",
		}

		expectedResp := &userpb.GetUserResponse{
			User: &userpb.User{
				Id:    "user-123",
				Email: "test@example.com",
				Name:  "Test User",
			},
		}

		mockService.On("GetUser", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.GetUser(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateUser(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.UpdateUserRequest{
			UserId: "user-123",
			Name:   wrapperspb.String("Updated Name"),
			Phone:  nil,
		}

		expectedResp := &userpb.UpdateUserResponse{
			User: &userpb.User{
				Id:   "user-123",
				Name: "Updated Name",
			},
		}

		mockService.On("UpdateUser", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.UpdateUser(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_AddAddress(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.AddAddressRequest{
			UserId:     "user-123",
			Name:       "John Doe",
			PostalCode: "100-0001",
			Prefecture: "Tokyo",
			City:       "Chiyoda-ku",
		}

		expectedResp := &userpb.AddAddressResponse{
			AddressId: "address-456",
		}

		mockService.On("AddAddress", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.AddAddress(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ListAddresses(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.ListAddressesRequest{
			UserId: "user-123",
		}

		expectedResp := &userpb.ListAddressesResponse{
			Addresses: []*userpb.Address{
				{Id: "address-1", Name: "Home"},
			},
		}

		mockService.On("ListAddresses", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.ListAddresses(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateAddress(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.UpdateAddressRequest{
			AddressId: "address-123",
			Name:      wrapperspb.String("Updated Name"),
			Phone:     nil,
		}

		expectedResp := &userpb.UpdateAddressResponse{
			Address: &userpb.Address{
				Id:   "address-123",
				Name: "Updated Name",
			},
		}

		mockService.On("UpdateAddress", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.UpdateAddress(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_DeleteAddress(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)

	handler := NewHandler(mockService, logger)

	t.Run("valid request is handled", func(t *testing.T) {
		req := &userpb.DeleteAddressRequest{
			AddressId: "address-123",
		}

		expectedResp := &sharedpb.Empty{}

		mockService.On("DeleteAddress", mock.Anything, req).Return(expectedResp, nil)

		resp, err := handler.DeleteAddress(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockService.AssertExpectations(t)
	})
}
