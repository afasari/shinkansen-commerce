package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	userpb "github.com/afasari/shinkansen-commerce/gen/proto/go/user"

	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/config"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/db"
	"go.uber.org/zap"
)

type UserService struct {
	userpb.UnimplementedUserServiceServer
	queries db.Querier
	cache   cache.Cache
	logger  *zap.Logger
	cfg     *config.Config
}

func NewUserService(queries db.Querier, cacheClient cache.Cache, logger *zap.Logger, cfg *config.Config) *UserService {
	return &UserService{
		queries: queries,
		cache:   cacheClient,
		logger:  logger,
		cfg:     cfg,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	s.logger.Info("Registering user", zap.String("email", req.Email))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Phone:        req.Phone,
	})

	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := s.generateAccessToken(userID, req.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &userpb.RegisterUserResponse{
		UserId:       userID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	s.logger.Info("User login attempt", zap.String("email", req.Email))

	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("User not found", zap.Error(err))
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.Warn("Failed password check", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.Active {
		return nil, fmt.Errorf("user is inactive")
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.cache.Delete(ctx, cache.UserCacheKey(user.ID.String()))

	return &userpb.LoginUserResponse{
		UserId:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	s.logger.Info("Getting user", zap.String("user_id", req.UserId))

	cacheKey := cache.UserCacheKey(req.UserId)
	var cached db.User
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("User cache hit", zap.String("user_id", req.UserId))
		return &userpb.GetUserResponse{
			User: s.userToProto(cached),
		}, nil
	}

	user, err := s.queries.GetUser(ctx, uuid.MustParse(req.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.cache.Set(ctx, cacheKey, user, cache.DefaultTTL); err != nil {
		s.logger.Warn("Failed to cache user", zap.Error(err))
	}

	return &userpb.GetUserResponse{
		User: s.userToProto(user),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	s.logger.Info("Updating user", zap.String("user_id", req.UserId))

	if req.Name == nil && req.Phone == nil {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	userID := uuid.MustParse(req.UserId)
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var name, phone *string
	if req.Name != nil {
		name = &req.Name.Value
	}
	if req.Phone != nil {
		phone = &req.Phone.Value
	}

	params := db.UpdateUserParams{
		ID:    user.ID,
		Name:  name,
		Phone: phone,
	}

	err = s.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	cacheKey := cache.UserCacheKey(userID.String())
	s.cache.Delete(ctx, cacheKey)

	updatedUser, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	return &userpb.UpdateUserResponse{
		User: s.userToProto(updatedUser),
	}, nil
}

func (s *UserService) AddAddress(ctx context.Context, req *userpb.AddAddressRequest) (*userpb.AddAddressResponse, error) {
	s.logger.Info("Adding address", zap.String("user_id", req.UserId))

	userID := uuid.MustParse(req.UserId)
	addressID, err := s.queries.CreateAddress(ctx, db.CreateAddressParams{
		UserID:       userID,
		Name:         req.Name,
		Phone:        req.Phone,
		PostalCode:   req.PostalCode,
		Prefecture:   req.Prefecture,
		City:         req.City,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		IsDefault:    req.IsDefault,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	cacheKey := cache.AddressesCacheKey(userID.String())
	s.cache.Delete(ctx, cacheKey)

	return &userpb.AddAddressResponse{
		AddressId: addressID.String(),
	}, nil
}

func (s *UserService) ListAddresses(ctx context.Context, req *userpb.ListAddressesRequest) (*userpb.ListAddressesResponse, error) {
	s.logger.Info("Listing addresses", zap.String("user_id", req.UserId))

	userID := uuid.MustParse(req.UserId)

	addresses, err := s.queries.ListAddresses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}

	protoAddresses := make([]*userpb.Address, 0, len(addresses))
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, s.addressToProto(addr))
	}

	return &userpb.ListAddressesResponse{
		Addresses: protoAddresses,
	}, nil
}

func (s *UserService) UpdateAddress(ctx context.Context, req *userpb.UpdateAddressRequest) (*userpb.UpdateAddressResponse, error) {
	s.logger.Info("Updating address", zap.String("address_id", req.AddressId))

	addressID := uuid.MustParse(req.AddressId)

	address, err := s.queries.GetAddress(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	var name, phone, postalCode, prefecture, city, addressLine1, addressLine2 *string
	var isDefault *bool

	if req.Name != nil {
		name = &req.Name.Value
	} else {
		name = &address.Name
	}

	if req.Phone != nil {
		phone = &req.Phone.Value
	} else {
		phone = &address.Phone
	}

	if req.PostalCode != nil {
		postalCode = &req.PostalCode.Value
	} else {
		postalCode = &address.PostalCode
	}

	if req.Prefecture != nil {
		prefecture = &req.Prefecture.Value
	} else {
		prefecture = &address.Prefecture
	}

	if req.City != nil {
		city = &req.City.Value
	} else {
		city = &address.City
	}

	if req.AddressLine1 != nil {
		addressLine1 = &req.AddressLine1.Value
	} else {
		addressLine1 = &address.AddressLine1
	}

	if req.AddressLine2 != nil {
		addressLine2 = &req.AddressLine2.Value
	} else {
		addressLine2 = &address.AddressLine2
	}

	if req.IsDefault != nil {
		isDefault = &req.IsDefault.Value
	} else {
		isDefault = &address.IsDefault
	}

	params := db.UpdateAddressParams{
		ID:           address.ID,
		Name:         name,
		Phone:        phone,
		PostalCode:   postalCode,
		Prefecture:   prefecture,
		City:         city,
		AddressLine1: addressLine1,
		AddressLine2: addressLine2,
		IsDefault:    isDefault,
	}

	err = s.queries.UpdateAddress(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	cacheKey := cache.AddressesCacheKey(address.UserID.String())
	s.cache.Delete(ctx, cacheKey)

	updatedAddress, err := s.queries.GetAddress(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated address: %w", err)
	}

	return &userpb.UpdateAddressResponse{
		Address: s.addressToProto(updatedAddress),
	}, nil
}

func (s *UserService) DeleteAddress(ctx context.Context, req *userpb.DeleteAddressRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Deleting address", zap.String("address_id", req.AddressId))

	addressID := uuid.MustParse(req.AddressId)

	address, err := s.queries.GetAddress(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	err = s.queries.DeleteAddress(ctx, db.DeleteAddressParams{ID: addressID})
	if err != nil {
		return nil, fmt.Errorf("failed to delete address: %w", err)
	}

	cacheKey := cache.AddressesCacheKey(address.UserID.String())
	s.cache.Delete(ctx, cacheKey)

	return &sharedpb.Empty{}, nil
}

func (s *UserService) userToProto(u db.User) *userpb.User {
	var updatedAt *timestamppb.Timestamp

	if !u.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(u.UpdatedAt)
	}

	return &userpb.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		Phone:     u.Phone,
		Active:    u.Active,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

func (s *UserService) addressToProto(a db.Address) *userpb.Address {
	return &userpb.Address{
		Id:           a.ID.String(),
		UserId:       a.UserID.String(),
		Name:         a.Name,
		Phone:        a.Phone,
		PostalCode:   a.PostalCode,
		Prefecture:   a.Prefecture,
		City:         a.City,
		AddressLine1: a.AddressLine1,
		AddressLine2: a.AddressLine2,
		IsDefault:    a.IsDefault,
		CreatedAt:    timestamppb.New(a.CreatedAt),
	}
}

func (s *UserService) generateAccessToken(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     time.Now().Add(time.Duration(s.cfg.AccessTokenDuration) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     "shinkansen",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(s.cfg.JWTSecret)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *UserService) generateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Duration(s.cfg.RefreshTokenDuration) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     "shinkansen-refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(s.cfg.JWTSecret)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
