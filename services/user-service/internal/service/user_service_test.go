package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/wrapperspb"

	userpb "github.com/afasari/shinkansen-commerce/gen/proto/go/user"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/config"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/db"
)

// MockQuerier is a mock implementation of db.Querier
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateUser(ctx context.Context, params db.CreateUserParams) (uuid.UUID, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQuerier) GetUser(ctx context.Context, id uuid.UUID) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockQuerier) UpdateUser(ctx context.Context, params db.UpdateUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) CreateAddress(ctx context.Context, params db.CreateAddressParams) (uuid.UUID, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQuerier) ListAddresses(ctx context.Context, userID uuid.UUID) ([]db.Address, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Address), args.Error(1)
}

func (m *MockQuerier) GetAddress(ctx context.Context, id uuid.UUID) (db.Address, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Address), args.Error(1)
}

func (m *MockQuerier) UpdateAddress(ctx context.Context, params db.UpdateAddressParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) DeleteAddress(ctx context.Context, params db.DeleteAddressParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) GetDefaultAddress(ctx context.Context, params db.GetDefaultAddressParams) (db.Address, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return db.Address{}, args.Error(1)
	}
	return args.Get(0).(db.Address), args.Error(1)
}

func (m *MockQuerier) SetDefaultAddress(ctx context.Context, params db.SetDefaultAddressParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func newTestService(t *testing.T) (*UserService, string) {
	logger := zap.NewNop()
	jwtSecret := "test-secret-key-for-jwt-signing"
	cfg := &config.Config{
		JWTSecret:           jwtSecret,
		AccessTokenDuration: 15,
		RefreshTokenDuration: 10080, // 7 days
	}
	mockQueries := new(MockQuerier)
	mockCache := new(cache.MockCache)

	service := NewUserService(mockQueries, mockCache, logger, cfg)
	return service, jwtSecret
}

func TestUserService_RegisterUser(t *testing.T) {
	t.Run("successful user registration", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		req := &userpb.RegisterUserRequest{
			Email:    "test@example.com",
			Password: "SecurePass123!",
			Name:     "Test User",
			Phone:    "090-1234-5678",
		}

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("CreateUser", mock.Anything, mock.AnythingOfType("db.CreateUserParams")).Return(userID, nil)

		resp, err := service.RegisterUser(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID.String(), resp.UserId)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		mockQueries.AssertExpectations(t)
	})

	t.Run("user registration with existing email fails", func(t *testing.T) {
		service, _ := newTestService(t)

		req := &userpb.RegisterUserRequest{
			Email:    "existing@example.com",
			Password: "SecurePass123!",
			Name:     "Test User",
		}

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("CreateUser", mock.Anything, mock.AnythingOfType("db.CreateUserParams")).
			Return(uuid.Nil, assert.AnError)

		resp, err := service.RegisterUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockQueries.AssertExpectations(t)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	// Generate a valid bcrypt hash for "password" for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	t.Run("successful login", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		req := &userpb.LoginUserRequest{
			Email:    "test@example.com",
			Password: "password",
		}

		mockUser := db.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Active:       true,
			CreatedAt:    time.Now(),
		}

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

		resp, err := service.LoginUser(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID.String(), resp.UserId)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		mockQueries.AssertExpectations(t)
	})

	t.Run("login with invalid email fails", func(t *testing.T) {
		service, _ := newTestService(t)

		req := &userpb.LoginUserRequest{
			Email:    "nonexistent@example.com",
			Password: "password",
		}

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").
			Return(db.User{}, assert.AnError)

		resp, err := service.LoginUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "not found")
		mockQueries.AssertExpectations(t)
	})

	t.Run("login with wrong password fails", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		req := &userpb.LoginUserRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockUser := db.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Active:       true,
			CreatedAt:    time.Now(),
		}

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetUserByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

		resp, err := service.LoginUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid credentials")
		mockQueries.AssertExpectations(t)
	})

	t.Run("login with inactive account fails", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		req := &userpb.LoginUserRequest{
			Email:    "inactive@example.com",
			Password: "password",
		}

		mockUser := db.User{
			ID:           userID,
			Email:        "inactive@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Inactive User",
			Active:       false,
			CreatedAt:    time.Now(),
		}

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetUserByEmail", mock.Anything, "inactive@example.com").Return(mockUser, nil)

		resp, err := service.LoginUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "inactive")
		mockQueries.AssertExpectations(t)
	})
}

func TestUserService_GetUser(t *testing.T) {
	t.Run("successful user retrieval", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		mockUser := db.User{
			ID:        userID,
			Email:     "test@example.com",
			Name:      "Test User",
			Phone:     "090-1234-5678",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*db.User")).Return(cache.ErrCacheMiss)
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("db.User"), mock.AnythingOfType("time.Duration")).Return(nil)

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetUser", mock.Anything, userID).Return(mockUser, nil)

		req := &userpb.GetUserRequest{
			UserId: userID.String(),
		}

		resp, err := service.GetUser(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.User)
		assert.Equal(t, userID.String(), resp.User.Id)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.Equal(t, "Test User", resp.User.Name)
		mockQueries.AssertExpectations(t)
	})

	t.Run("user retrieval from cache", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*db.User")).Return(nil)

		req := &userpb.GetUserRequest{
			UserId: userID.String(),
		}

		resp, err := service.GetUser(context.Background(), req)

		// Cache hit scenario - should return nil error (simplified)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Run("successful user update with name", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		mockUser := db.User{
			ID:        userID,
			Email:     "test@example.com",
			Name:      "Old Name",
			Phone:     "090-1234-5678",
			Active:    true,
			CreatedAt: time.Now(),
		}

		updatedUser := mockUser
		updatedUser.Name = "New Name"

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetUser", mock.Anything, userID).Return(mockUser, nil).Once()
		mockQueries.On("UpdateUser", mock.Anything, mock.AnythingOfType("db.UpdateUserParams")).Return(nil).Once()
		mockQueries.On("GetUser", mock.Anything, userID).Return(updatedUser, nil).Once()

		req := &userpb.UpdateUserRequest{
			UserId: userID.String(),
			Name:   wrapperspb.String("New Name"),
		}

		resp, err := service.UpdateUser(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Name", resp.User.Name)
		mockQueries.AssertExpectations(t)
	})

	t.Run("update with no fields fails", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		req := &userpb.UpdateUserRequest{
			UserId: userID.String(),
		}

		resp, err := service.UpdateUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "at least one field")
	})
}

func TestUserService_AddAddress(t *testing.T) {
	t.Run("successful address creation", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()
		addressID := uuid.New()

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("CreateAddress", mock.Anything, mock.AnythingOfType("db.CreateAddressParams")).Return(addressID, nil)

		req := &userpb.AddAddressRequest{
			UserId:       userID.String(),
			Name:         "John Doe",
			Phone:        "090-1234-5678",
			PostalCode:   "100-0001",
			Prefecture:   "Tokyo",
			City:         "Chiyoda-ku",
			AddressLine1: "1-1-1 Otemachi",
			AddressLine2: "",
			IsDefault:    true,
		}

		resp, err := service.AddAddress(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, addressID.String(), resp.AddressId)
		mockQueries.AssertExpectations(t)
	})
}

func TestUserService_ListAddresses(t *testing.T) {
	t.Run("successful address listing", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()

		addresses := []db.Address{
			{
				ID:           uuid.New(),
				UserID:       userID,
				Name:         "Home",
				Phone:        "090-1234-5678",
				PostalCode:   "100-0001",
				Prefecture:   "Tokyo",
				City:         "Chiyoda-ku",
				AddressLine1: "1-1-1 Otemachi",
				AddressLine2: "",
				IsDefault:    true,
				CreatedAt:    time.Now(),
			},
			{
				ID:           uuid.New(),
				UserID:       userID,
				Name:         "Office",
				Phone:        "090-8765-4321",
				PostalCode:   "150-0001",
				Prefecture:   "Tokyo",
				City:         "Shibuya-ku",
				AddressLine1: "1-1-1 Shibuya",
				AddressLine2: "",
				IsDefault:    false,
				CreatedAt:    time.Now(),
			},
		}

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("ListAddresses", mock.Anything, userID).Return(addresses, nil)

		req := &userpb.ListAddressesRequest{
			UserId: userID.String(),
		}

		resp, err := service.ListAddresses(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Addresses, 2)
		assert.Equal(t, "Home", resp.Addresses[0].Name)
		assert.Equal(t, "Office", resp.Addresses[1].Name)
		mockQueries.AssertExpectations(t)
	})
}

func TestUserService_UpdateAddress(t *testing.T) {
	t.Run("successful address update", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()
		addressID := uuid.New()

		existingAddress := db.Address{
			ID:           addressID,
			UserID:       userID,
			Name:         "Old Name",
			Phone:        "090-1234-5678",
			PostalCode:   "100-0001",
			Prefecture:   "Tokyo",
			City:         "Chiyoda-ku",
			AddressLine1: "1-1-1 Otemachi",
			AddressLine2: "",
			IsDefault:    true,
			CreatedAt:    time.Now(),
		}

		updatedAddress := existingAddress
		updatedAddress.Name = "New Name"

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetAddress", mock.Anything, addressID).Return(existingAddress, nil).Once()
		mockQueries.On("UpdateAddress", mock.Anything, mock.AnythingOfType("db.UpdateAddressParams")).Return(nil).Once()
		mockQueries.On("GetAddress", mock.Anything, addressID).Return(updatedAddress, nil).Once()

		req := &userpb.UpdateAddressRequest{
			AddressId: addressID.String(),
			Name:      wrapperspb.String("New Name"),
		}

		resp, err := service.UpdateAddress(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Name", resp.Address.Name)
		mockQueries.AssertExpectations(t)
	})
}

func TestUserService_DeleteAddress(t *testing.T) {
	t.Run("successful address deletion", func(t *testing.T) {
		service, _ := newTestService(t)
		userID := uuid.New()
		addressID := uuid.New()

		existingAddress := db.Address{
			ID:        addressID,
			UserID:    userID,
			Name:      "To Delete",
			CreatedAt: time.Now(),
		}

		mockCache := service.cache.(*cache.MockCache)
		mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		mockQueries := service.queries.(*MockQuerier)
		mockQueries.On("GetAddress", mock.Anything, addressID).Return(existingAddress, nil)
		mockQueries.On("DeleteAddress", mock.Anything, mock.AnythingOfType("db.DeleteAddressParams")).Return(nil)

		req := &userpb.DeleteAddressRequest{
			AddressId: addressID.String(),
		}

		resp, err := service.DeleteAddress(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockQueries.AssertExpectations(t)
	})
}

func TestUserService_generateAccessToken(t *testing.T) {
	service, jwtSecret := newTestService(t)
	userID := uuid.New()

	t.Run("access token generation", func(t *testing.T) {
		token, err := service.generateAccessToken(userID, "test@example.com")

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token structure
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, userID.String(), claims["user_id"])
		assert.Equal(t, "test@example.com", claims["email"])
		assert.Equal(t, "shinkansen", claims["jti"])
	})
}

func TestUserService_generateRefreshToken(t *testing.T) {
	service, jwtSecret := newTestService(t)
	userID := uuid.New()

	t.Run("refresh token generation", func(t *testing.T) {
		token, err := service.generateRefreshToken(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token structure
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, userID.String(), claims["user_id"])
		assert.Equal(t, "shinkansen-refresh", claims["jti"])
	})
}
