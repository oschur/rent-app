package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	domain "rent-app/internal/domain/auth"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrTokenBlacklisted = errors.New("token has been revoked")
)

type RefreshTokenRepository interface {
	IsTokenBlacklisted(tokenID string) (bool, error)
	BlacklistToken(tokenID string, expiresAt int64) error
	CleanupExpiredTokens() error
}

type Service struct {
	secretKey []byte
	tokenRepo RefreshTokenRepository
}

func NewService(tokenRepo RefreshTokenRepository) *Service {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "DEFAULT_KEY_ONLY_FOR_DEVELOPEMENT" // по окончании разработки это надо убрать
	}

	return &Service{
		secretKey: []byte(secretKey),
		tokenRepo: tokenRepo,
	}
}

func (s *Service) GenerateToken(userID int, isLandlord, isAdmin bool) (*domain.TokenPair, error) {
	accessExpirationTime := time.Now().Add(domain.AccessTokenTTL)
	accessClaims := &domain.AccessTokenClaims{
		UserID:     userID,
		IsLandlord: isLandlord,
		IsAdmin:    isAdmin,
		TokenType:  domain.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rent-app",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshExpirationTime := time.Now().Add(domain.RefreshTokenTTL)
	refreshClaims := &domain.RefreshTokenClaims{
		UserID:     userID,
		IsLandlord: isLandlord,
		IsAdmin:    isAdmin,
		TokenType:  domain.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rent-app",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(domain.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*domain.AccessTokenClaims, error) {
	claims := &domain.AccessTokenClaims{}
	if err := s.parseToken(tokenString, claims); err != nil {
		return nil, err
	}

	if claims.TokenType != domain.TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func (s *Service) ValidateRefreshToken(tokenString string) (*domain.RefreshTokenClaims, error) {
	claims := &domain.RefreshTokenClaims{}
	if err := s.parseToken(tokenString, claims); err != nil {
		return nil, err
	}

	if claims.TokenType != domain.TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func (s *Service) parseToken(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrTokenExpired
		}
		return ErrInvalidToken
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}

// создаем уникалньый ID токена для хранения в БД, что нужно для проверки на черный список
func tokenIDFromToken(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

func (s *Service) RefreshToken(refreshTokenString string) (*domain.TokenPair, error) {
	refreshClaims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	tokenID := tokenIDFromToken(refreshTokenString)
	isBlacklisted, err := s.tokenRepo.IsTokenBlacklisted(tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, ErrTokenBlacklisted
	}

	// если права пользователя изменились, обновим их при следующем логине
	newTokenPair, err := s.GenerateToken(refreshClaims.UserID, refreshClaims.IsLandlord, refreshClaims.IsAdmin)
	if err != nil {
		return nil, err
	}

	// инвалидируем использованный refresh token
	expiresAt := refreshClaims.ExpiresAt.Time.Unix()
	if err := s.tokenRepo.BlacklistToken(tokenID, expiresAt); err != nil {
		_ = err // по-хорошему тут надо прологировать ошибку, но я займусь этим позже
	}

	return newTokenPair, nil
}
