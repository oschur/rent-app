package config

import (
	"testing"
)

func TestConfig_Load(t *testing.T) {

	t.Setenv("PORT", "3000")
	t.Setenv("ENV", "test")
	t.Setenv("DB_DSN", "postgres://test:test@localhost/testdb")

	cfg, err := ConfigLoad()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Port != "3000" {
		t.Error("PORT incorrect environment value: expected 3000 but got", cfg.Port)
	}
	if cfg.Env != "test" {
		t.Error("ENV incorrect environment value: expected test but got", cfg.Env)
	}
	if cfg.DSN != "postgres://test:test@localhost/testdb" {
		t.Error("DB_DSN incorrect environment value: expected postgres://test:test@localhost/testdb but got", cfg.DSN)
	}
}

func TestConfigLoad_Defaults(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://test:test@localhost/testdb")

	cfg, err := ConfigLoad()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Port != "8080" {
		t.Error("expected default PORT 8080, got", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Error("expected default ENV development, got", cfg.Env)
	}
}

func TestConfigLoad_MissingDSN(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("ENV", "test")
	t.Setenv("DB_DSN", "")

	_, err := ConfigLoad()
	if err == nil {
		t.Fatal("expected error when DB_DSN is missing, got nil")
	}
}
