package auth

import (
	"errors"
	"os"
	"testing"
	"time"

	domain "rent-app/internal/domain/auth"

	"github.com/golang-jwt/jwt/v5"
)

type MockRefreshTokenRepository struct {
	blacklistedTokens map[string]bool
	blacklistError    error
	isBlacklistedErr  error
	blacklistTokenErr error
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		blacklistedTokens: make(map[string]bool),
	}
}

func (m *MockRefreshTokenRepository) IsTokenBlacklisted(tokenID string) (bool, error) {
	if m.isBlacklistedErr != nil {
		return false, m.isBlacklistedErr
	}
	return m.blacklistedTokens[tokenID], nil
}

func (m *MockRefreshTokenRepository) BlacklistToken(tokenID string, expiresAt int64) error {
	if m.blacklistTokenErr != nil {
		return m.blacklistTokenErr
	}
	m.blacklistedTokens[tokenID] = true
	return nil
}

func (m *MockRefreshTokenRepository) CleanupExpiredTokens() error {
	return nil
}

func TestNewService(t *testing.T) {
	t.Run("creates service with default key when env var not set", func(t *testing.T) {
		originalKey := os.Getenv("JWT_SECRET_KEY")
		os.Unsetenv("JWT_SECRET_KEY")
		defer os.Setenv("JWT_SECRET_KEY", originalKey)

		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		if service == nil {
			t.Fatal("expected service to be created, got nil")
		}
		if service.tokenRepo == nil {
			t.Error("expected tokenRepo to be set")
		}
	})

	t.Run("creates service with env key when set", func(t *testing.T) {
		originalKey := os.Getenv("JWT_SECRET_KEY")
		testKey := "test-secret-key-123"
		os.Setenv("JWT_SECRET_KEY", testKey)
		defer os.Setenv("JWT_SECRET_KEY", originalKey)

		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		if string(service.secretKey) != testKey {
			t.Errorf("expected secretKey to be %s, got %s", testKey, string(service.secretKey))
		}
	})
}

func TestService_GenerateToken(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	service := NewService(mockRepo)

	tests := []struct {
		name       string
		userID     int
		isLandlord bool
		isAdmin    bool
	}{
		{
			name:       "generate token for regular user",
			userID:     1,
			isLandlord: false,
			isAdmin:    false,
		},
		{
			name:       "generate token for landlord",
			userID:     2,
			isLandlord: true,
			isAdmin:    false,
		},
		{
			name:       "generate token for admin",
			userID:     3,
			isLandlord: true,
			isAdmin:    true,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			tokenPair, err := service.GenerateToken(e.userID, e.isLandlord, e.isAdmin)
			if err != nil {
				return
			}

			if tokenPair == nil {
				t.Fatal("expected tokenPair to be not nil")
			}

			if tokenPair.AccessToken == "" {
				t.Error("expected AccessToken to be not empty")
			}

			if tokenPair.RefreshToken == "" {
				t.Error("expected RefreshToken to be not empty")
			}

			if tokenPair.ExpiresIn != int64(domain.AccessTokenTTL.Seconds()) {
				t.Errorf("expected ExpiresIn to be %d, got %d", int64(domain.AccessTokenTTL.Seconds()), tokenPair.ExpiresIn)
			}

			// Проверяем, что access token валиден и содержит правильные claims
			accessClaims, err := service.ValidateAccessToken(tokenPair.AccessToken)
			if err != nil {
				t.Errorf("generated access token should be valid, got error: %v", err)
				return
			}

			if accessClaims.UserID != e.userID {
				t.Errorf("expected UserID %d, got %d", e.userID, accessClaims.UserID)
			}

			if accessClaims.IsLandlord != e.isLandlord {
				t.Errorf("expected IsLandlord %v, got %v", e.isLandlord, accessClaims.IsLandlord)
			}

			if accessClaims.IsAdmin != e.isAdmin {
				t.Errorf("expected IsAdmin %v, got %v", e.isAdmin, accessClaims.IsAdmin)
			}

			if accessClaims.TokenType != domain.TokenTypeAccess {
				t.Errorf("expected TokenType %s, got %s", domain.TokenTypeAccess, accessClaims.TokenType)
			}

			refreshClaims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)
			if err != nil {
				t.Errorf("generated refresh token should be vaild, got error: %v", err)
			}

			if refreshClaims.UserID != e.userID {
				t.Errorf("expected UserID %d, got %d", e.userID, refreshClaims.UserID)
			}

			if refreshClaims.IsLandlord != e.isLandlord {
				t.Errorf("expected IsLandlord %v, got %v", e.isLandlord, refreshClaims.IsLandlord)
			}

			if refreshClaims.IsAdmin != e.isAdmin {
				t.Errorf("expected IsAdmin %v, got %v", e.isAdmin, refreshClaims.IsAdmin)
			}

			if refreshClaims.TokenType != domain.TokenTypeRefresh {
				t.Errorf("expected TokenType %s, got %s", domain.TokenTypeRefresh, refreshClaims.TokenType)
			}
		})
	}
}

func TestService_ValidateAccessToken(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	service := NewService(mockRepo)

	t.Run("valid access token", func(t *testing.T) {
		tokenPair, err := service.GenerateToken(1, false, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("ValidateAccessToken() error = %v, expected nil", err)
			return
		}

		if claims.UserID != 1 {
			t.Errorf("expected UserID 1, got %d", claims.UserID)
		}

		if claims.TokenType != domain.TokenTypeAccess {
			t.Errorf("expected TokenType %s, got %s", domain.TokenTypeAccess, claims.TokenType)
		}
	})

	t.Run("rejects refresh token as access token", func(t *testing.T) {
		tokenPair, err := service.GenerateToken(1, false, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = service.ValidateAccessToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("expected error when validating refresh token as access token, got nil")
			return
		}

		if !errors.Is(err, ErrInvalidTokenType) {
			t.Errorf("expected ErrInvalidTokenType, got %v", err)
		}
	})

	t.Run("rejects invalid token string", func(t *testing.T) {
		_, err := service.ValidateAccessToken("invalid.token.string")
		if err == nil {
			t.Error("expected error for invalid token, got nil")
			return
		}

		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		expiredClaims := &domain.AccessTokenClaims{
			UserID:     1,
			IsLandlord: false,
			IsAdmin:    false,
			TokenType:  domain.TokenTypeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    "rent-app",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		tokenString, err := token.SignedString(service.secretKey)
		if err != nil {
			t.Fatalf("failed to sign expired token: %v", err)
		}

		_, err = service.ValidateAccessToken(tokenString)
		if err == nil {
			t.Error("expected error for expired token, got nil")
			return
		}

		if !errors.Is(err, ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("rejects token with wrong signing method", func(t *testing.T) {
		claims := &domain.AccessTokenClaims{
			UserID:     1,
			IsLandlord: false,
			IsAdmin:    false,
			TokenType:  domain.TokenTypeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "rent-app",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenString, err := token.SignedString([]byte("wrong-secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = service.ValidateAccessToken(tokenString)
		if err == nil {
			t.Error("expected error for token with wrong signing method but got nil")
			return
		}

		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken but got %v", err)
		}
	})
}

func TestService_ValidateRefreshToken(t *testing.T) {
	mockRepo := NewMockRefreshTokenRepository()
	service := NewService(mockRepo)

	t.Run("valid refresh token", func(t *testing.T) {
		tokenPair, err := service.GenerateToken(1, true, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)
		if err != nil {
			t.Errorf("ValidateRefreshToken() error = %v, expected nil", err)
			return
		}

		if claims.UserID != 1 {
			t.Errorf("expected UserID 1, got %d", claims.UserID)
		}

		if !claims.IsLandlord {
			t.Error("expected IsLandlord to be true")
		}

		if claims.TokenType != domain.TokenTypeRefresh {
			t.Errorf("expected TokenType %s, got %s", domain.TokenTypeRefresh, claims.TokenType)
		}
	})

	t.Run("rejects access token as refresh token", func(t *testing.T) {
		tokenPair, err := service.GenerateToken(1, false, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = service.ValidateRefreshToken(tokenPair.AccessToken)
		if err == nil {
			t.Error("expected error when validating access token as refresh token, got nil")
			return
		}

		if !errors.Is(err, ErrInvalidTokenType) {
			t.Errorf("expected ErrInvalidTokenType, got %v", err)
		}
	})

	t.Run("rejects invalid token string", func(t *testing.T) {
		_, err := service.ValidateRefreshToken("invalid.token.string")
		if err == nil {
			t.Error("expected error for invalid token, got nil")
			return
		}

		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		expiredClaims := &domain.RefreshTokenClaims{
			UserID:     1,
			IsLandlord: false,
			IsAdmin:    false,
			TokenType:  domain.TokenTypeRefresh,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    "rent-app",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		tokenString, err := token.SignedString(service.secretKey)
		if err != nil {
			t.Fatalf("failed to sign expired token: %v", err)
		}

		_, err = service.ValidateRefreshToken(tokenString)
		if err == nil {
			t.Error("expected error for expired token, got nil")
			return
		}

		if !errors.Is(err, ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})
}

func TestService_RefreshToken(t *testing.T) {
	t.Run("successful token refresh", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		originalTokenPair, err := service.GenerateToken(1, true, false)
		if err != nil {
			t.Fatalf("failed to generate initial token: %v", err)
		}

		newTokenPair, err := service.RefreshToken(originalTokenPair.RefreshToken)
		if err != nil {
			t.Errorf("RefreshToken() error = %v, expected nil", err)
			return
		}

		if newTokenPair == nil {
			t.Fatal("expected newTokenPair to be not nil")
		}

		if newTokenPair.AccessToken == "" {
			t.Error("expected AccessToken to be not empty")
		}

		if newTokenPair.RefreshToken == "" {
			t.Error("expected RefreshToken to be not empty")
		}

		accessClaims, err := service.ValidateAccessToken(newTokenPair.AccessToken)
		if err != nil {
			t.Errorf("new access token should be valid, got error: %v", err)
			return
		}

		if accessClaims.UserID != 1 {
			t.Errorf("expected UserID 1, got %d", accessClaims.UserID)
		}

		if !accessClaims.IsLandlord {
			t.Error("expected IsLandlord to be true")
		}

		tokenID := tokenIDFromToken(originalTokenPair.RefreshToken)
		isBlacklisted, err := mockRepo.IsTokenBlacklisted(tokenID)
		if err != nil {
			t.Errorf("failed to check blacklist: %v", err)
			return
		}

		if !isBlacklisted {
			t.Error("expected old refresh token to be blacklisted")
		}
	})

	t.Run("rejects invalid refresh token", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		_, err := service.RefreshToken("invalid.token.string")
		if err == nil {
			t.Error("expected error for invalid refresh token, got nil")
			return
		}

		if !errors.Is(err, ErrInvalidToken) && !errors.Is(err, ErrInvalidTokenType) {
			t.Errorf("expected ErrInvalidToken or ErrInvalidTokenType, got %v", err)
		}
	})

	t.Run("rejects expired refresh token", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		expiredClaims := &domain.RefreshTokenClaims{
			UserID:     1,
			IsLandlord: false,
			IsAdmin:    false,
			TokenType:  domain.TokenTypeRefresh,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    "rent-app",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		tokenString, err := token.SignedString(service.secretKey)
		if err != nil {
			t.Fatalf("failed to sign expired token: %v", err)
		}

		_, err = service.RefreshToken(tokenString)
		if err == nil {
			t.Error("expected error for expired refresh token, got nil")
			return
		}

		if !errors.Is(err, ErrTokenExpired) {
			t.Errorf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("rejects blacklisted refresh token", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		tokenPair, err := service.GenerateToken(1, false, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		tokenID := tokenIDFromToken(tokenPair.RefreshToken)
		err = mockRepo.BlacklistToken(tokenID, time.Now().Add(30*24*time.Hour).Unix())
		if err != nil {
			t.Fatalf("failed to blacklist token: %v", err)
		}

		_, err = service.RefreshToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("expected error for blacklisted refresh token, got nil")
			return
		}

		if !errors.Is(err, ErrTokenBlacklisted) {
			t.Errorf("expected ErrTokenBlacklisted, got %v", err)
		}
	})

	t.Run("handles repository error when checking blacklist", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		mockRepo.isBlacklistedErr = errors.New("database error")
		service := NewService(mockRepo)

		tokenPair, err := service.GenerateToken(1, false, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = service.RefreshToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("expected error when repository fails, got nil")
			return
		}

		if err.Error() == "" {
			t.Error("expected error message, got empty string")
		}
	})

	t.Run("preserves user permissions in new tokens", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		originalTokenPair, err := service.GenerateToken(1, true, true)
		if err != nil {
			t.Fatalf("failed to generate initial token: %v", err)
		}

		newTokenPair, err := service.RefreshToken(originalTokenPair.RefreshToken)
		if err != nil {
			t.Fatalf("RefreshToken() error = %v", err)
		}

		accessClaims, err := service.ValidateAccessToken(newTokenPair.AccessToken)
		if err != nil {
			t.Fatalf("failed to validate new access token: %v", err)
		}

		if !accessClaims.IsLandlord {
			t.Error("expected IsLandlord to be true in new token")
		}

		if !accessClaims.IsAdmin {
			t.Error("expected IsAdmin to be true in new token")
		}

		refreshClaims, err := service.ValidateRefreshToken(newTokenPair.RefreshToken)
		if err != nil {
			t.Fatalf("failed to validate new refresh token: %v", err)
		}

		if !refreshClaims.IsLandlord {
			t.Error("expected IsLandlord to be true in new refresh token")
		}

		if !refreshClaims.IsAdmin {
			t.Error("expected IsAdmin to be true in new refresh token")
		}
	})

	t.Run("prevents reuse of refresh token", func(t *testing.T) {
		mockRepo := NewMockRefreshTokenRepository()
		service := NewService(mockRepo)

		tokenPair, err := service.GenerateToken(1, false, false)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = service.RefreshToken(tokenPair.RefreshToken)
		if err != nil {
			t.Fatalf("first RefreshToken() error = %v", err)
		}
		_, err = service.RefreshToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("expected error when reusing refresh token, got nil")
			return
		}

		if !errors.Is(err, ErrTokenBlacklisted) {
			t.Errorf("expected ErrTokenBlacklisted, got %v", err)
		}
	})
}

func TestTokenIDFromToken(t *testing.T) {
	t.Run("generates consistent token IDs", func(t *testing.T) {
		tokenString := "test.token.string"
		id1 := tokenIDFromToken(tokenString)
		id2 := tokenIDFromToken(tokenString)

		if id1 != id2 {
			t.Errorf("expected token IDs to be equal, got %s and %s", id1, id2)
		}

		if len(id1) != 64 {
			t.Errorf("expected token ID length to be 64 (SHA256 hex), got %d", len(id1))
		}
	})

	t.Run("generates different IDs for different tokens", func(t *testing.T) {
		id1 := tokenIDFromToken("token1")
		id2 := tokenIDFromToken("token2")

		if id1 == id2 {
			t.Error("expected different token IDs for different tokens")
		}
	})
}
