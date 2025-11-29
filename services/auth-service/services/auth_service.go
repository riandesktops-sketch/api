package services

import (
	"context"
	"errors"

	"zodiac-ai-backend/pkg/jwt"
	"zodiac-ai-backend/pkg/utils"
	"zodiac-ai-backend/pkg/validator"
	"zodiac-ai-backend/services/auth-service/models"
	"zodiac-ai-backend/services/auth-service/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// AuthService handles authentication business logic
// Reference: Pragmatic Programmer - Dependency Injection for testability
type AuthService struct {
	userRepo         *repositories.UserRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
	jwtManager       *jwt.Manager
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repositories.UserRepository,
	refreshTokenRepo *repositories.RefreshTokenRepository,
	jwtManager *jwt.Manager,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate input
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	_, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}
	if err != repositories.ErrUserNotFound {
		return nil, err
	}

	// Calculate zodiac sign from date of birth
	zodiacSign := utils.CalculateZodiac(req.DateOfBirth)

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:       req.Email,
		Password:    hashedPassword,
		FullName:    req.FullName,
		DisplayName: req.FullName, // Default to full name
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		ZodiacSign:  string(zodiacSign),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if err == repositories.ErrUserAlreadyExists {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.Hex(), user.ZodiacSign)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.Hex(), user.ZodiacSign)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, user.ID, refreshToken); err != nil {
		return nil, err
	}

	// Clear password before returning
	user.Password = ""

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Validate input
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.Hex(), user.ZodiacSign)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.Hex(), user.ZodiacSign)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, user.ID, refreshToken); err != nil {
		return nil, err
	}

	// Clear password before returning
	user.Password = ""

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshAccessToken refreshes access token using refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenString string) (string, error) {
	// Verify refresh token
	claims, err := s.jwtManager.VerifyToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Validate token type
	if err := s.jwtManager.ValidateTokenType(claims, jwt.RefreshToken); err != nil {
		return "", err
	}

	// Check if refresh token exists in database
	_, err = s.refreshTokenRepo.FindByToken(ctx, refreshTokenString)
	if err != nil {
		return "", err
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(claims.UserID, claims.ZodiacSign)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GetProfile gets user profile
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Clear password
	user.Password = ""
	return user, nil
}

// UpdateProfile updates user profile
func (s *AuthService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.User, error) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Build update document
	update := bson.M{}
	if req.DisplayName != "" {
		update["display_name"] = req.DisplayName
	}
	if req.Bio != "" {
		update["bio"] = req.Bio
	}
	if req.AvatarURL != "" {
		update["avatar_url"] = req.AvatarURL
	}

	// Update user
	if err := s.userRepo.Update(ctx, id, update); err != nil {
		return nil, err
	}

	// Get updated user
	return s.GetProfile(ctx, userID)
}
