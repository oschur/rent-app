CREATE TABLE IF NOT EXISTS refresh_token_blacklist (
    token_id VARCHAR(64) PRIMARY KEY,
    expires_at BIGINT NOT NULL,
    blacklisted_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_token_blacklist_expires_at 
ON refresh_token_blacklist(expires_at);