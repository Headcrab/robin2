package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReloadExpandsEnvRefs(t *testing.T) {
	t.Setenv("ROBIN_DB_TEST_HOST", "127.0.0.1")
	t.Setenv("ROBIN_DB_TEST_USER", "tester")
	t.Setenv("ROBIN_DB_TEST_PASSWORD", "secret")

	cfgBody := `{
		"port": 8008,
		"round": 2,
		"curr_db": "test",
		"curr_cache": "memory",
		"date_formats": ["2006-01-02 15:04:05"],
		"db": [
			{
				"name": "test",
				"type": "mysql",
				"host": "${ROBIN_DB_TEST_HOST}",
				"port": "3306",
				"user": "${ROBIN_DB_TEST_USER}",
				"password": "${ROBIN_DB_TEST_PASSWORD}",
				"database": "runtime",
				"connection_string": "{user}:{password}@tcp({host}:{port})/{database}",
				"query": {}
			}
		],
		"cache": [
			{
				"name": "memory",
				"type": "memory"
			}
		]
	}`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(cfgBody), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := New()
	cfg.FileName = cfgPath
	if err := cfg.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if cfg.CurrDB.Host != "127.0.0.1" {
		t.Fatalf("host was not expanded: %q", cfg.CurrDB.Host)
	}
	if cfg.CurrDB.User != "tester" {
		t.Fatalf("user was not expanded: %q", cfg.CurrDB.User)
	}
	if cfg.CurrDB.Password != "secret" {
		t.Fatalf("password was not expanded: %q", cfg.CurrDB.Password)
	}
}
