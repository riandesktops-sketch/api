package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents JWT custom claims
type Claims struct {
	UserID     string `json:"user_id"`
	ZodiacSign string `json:"zodiac_sign"`
	TokenType  TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// Manager handles JWT token operations
type Manager struct {
	secretKey        []byte
	accessExpiry     time.Duration
	refreshExpiry    time.Duration
}

// NewManager creates a new JWT manager
func NewManager(secretKey string, accessExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken creates a new access token
// Access tokens are short-lived (15 min) to reduce attack window
// Reference: Pragmatic Programmer - Security Through Simplicity
func (m *Manager) GenerateAccessToken(userID, zodiacSign string) (string, error) {
	claims := Claims{
		UserID:     userID,
		ZodiacSign: zodiacSign,
		TokenType:  AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken creates a new refresh token
// Refresh tokens are long-lived (30 days) for seamless UX
func (m *Manager) GenerateRefreshToken(userID, zodiacSign string) (string, error) {
	claims := Claims{
		UserID:     userID,
		ZodiacSign: zodiacSign,
		TokenType:  RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// VerifyToken verifies and parses a JWT token
func (m *Manager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractUserID extracts user ID from token without full verification
// Useful for logging and non-critical operations
func (m *Manager) ExtractUserID(tokenString string) (string, error) {
	claims, err := m.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ValidateTokenType checks if token is of expected type
func (m *Manager) ValidateTokenType(claims *Claims, expectedType TokenType) error {
	if claims.TokenType != expectedType {
		return ErrInvalidToken
	}
	return nil
}
