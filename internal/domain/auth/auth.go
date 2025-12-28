package auth

import (
	"context"
	"time"

	contextUserInfo "rent-app/internal/context"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type AccessTokenClaims struct {
	UserID     int    `json:"user_id"`
	IsLandlord bool   `json:"is_landlord"`
	IsAdmin    bool   `json:"is_admin"`
	TokenType  string `json:"token_type"` // "access" или "refresh"
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID     int    `json:"user_id"`
	IsLandlord bool   `json:"is_landlord"`
	IsAdmin    bool   `json:"is_admin"`
	TokenType  string `json:"token_type"` // "access" или "refresh"
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	TokenPair
	User struct {
		ID         int    `json:"id"`
		Email      string `json:"email"`
		FirstName  string `json:"firstname"`
		LastName   string `json:"lastname"`
		IsLandlord bool   `json:"islandlord"`
		IsAdmin    bool   `json:"isadmin"`
	} `json:"user"`
}

const (
	AccessTokenTTL  = 2 * time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour
	TokenType       = "Bearer"
)

type ContextKey string

const UserContextKey ContextKey = "user"

func SetUserContext(ctx context.Context, claims *AccessTokenClaims) context.Context {
	userInfo := &contextUserInfo.UserInfo{
		UserID:     claims.UserID,
		IsLandlord: claims.IsLandlord,
		IsAdmin:    claims.IsAdmin,
	}
	return contextUserInfo.SetUserInfo(ctx, userInfo)
}
