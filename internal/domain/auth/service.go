package auth

type Service interface {
	GenerateToken(userID int, isLandLord bool, isAdmin bool) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	RefreshToken(refreshTokenString string) (*TokenPair, error)
}
