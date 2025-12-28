package auth

type RefreshTokenRepository interface {
	// IsTokenBlacklisted проверяет, находится ли токен в blacklist
	IsTokenBlacklisted(tokenID string) (bool, error)
	// BlacklistToken добавляет токен в blacklist
	BlacklistToken(tokenID string, expiresAt int64) error
	// CleanupExpiredTokens удаляет истекшие токены из blacklist
	CleanupExpiredTokens() error
}
