package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const timeout = 3 * time.Second

type PostgresRepo struct {
	DB *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		DB: db,
	}
}

func (p *PostgresRepo) InitTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS refresh_token_blacklist (
			token_id VARCHAR(64) PRIMARY KEY,
			expires_at BIGINT NOT NULL,
			blacklisted_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_refresh_token_blacklist_expires_at 
		ON refresh_token_blacklist(expires_at);
	`

	_, err := p.DB.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create refresh_token_blacklist table: %w", err)
	}

	return nil
}

func (p *PostgresRepo) IsTokenBlacklisted(tokenID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `SELECT EXISTS(SELECT 1 FROM refresh_token_blacklist WHERE token_id = $1)`

	var exists bool
	err := p.DB.QueryRow(ctx, query, tokenID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}

func (p *PostgresRepo) BlacklistToken(tokenID string, expiresAt int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `
		INSERT INTO refresh_token_blacklist (token_id, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token_id) DO NOTHING
	`

	_, err := p.DB.Exec(ctx, query, tokenID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (p *PostgresRepo) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Now().Unix()
	query := `DELETE FROM refresh_token_blacklist WHERE expires_at < $1`

	_, err := p.DB.Exec(ctx, query, now)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}
