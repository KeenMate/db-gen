package dbGen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
)

// testConnString resolves the connection string for DB-backed tests:
//  1. the DBGEN_TEST_DSN environment variable, if set;
//  2. otherwise the ConnectionString from the committed test/db-gen-copy.json.
func testConnString(t *testing.T) string {
	t.Helper()

	if dsn := os.Getenv("DBGEN_TEST_DSN"); dsn != "" {
		return dsn
	}

	configPath := filepath.Join("..", "..", "test", "db-gen-copy.json")
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Skipf("no DBGEN_TEST_DSN and cannot read %s: %v", configPath, err)
	}

	var cfg struct {
		ConnectionString string
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Skipf("cannot parse %s: %v", configPath, err)
	}
	if cfg.ConnectionString == "" {
		t.Skip("no connection string available for integration tests")
	}
	return cfg.ConnectionString
}

// requireTestDB connects to the test database, (re)applies the fixture schema,
// and returns the connection string. If the database is unreachable the test is
// skipped rather than failed, so the suite stays green without a DB.
func requireTestDB(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping DB-backed test in -short mode")
	}

	connStr := testConnString(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		t.Skipf("test database unreachable, skipping integration test: %v", err)
	}
	defer conn.Close(ctx)

	fixturePath := filepath.Join("..", "..", "test", "fixtures", "schema.sql")
	sqlBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("reading fixtures %s: %v", fixturePath, err)
	}

	// No args -> pgx uses the simple protocol, which runs all statements in the file.
	if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("applying fixtures: %v", err)
	}

	return connStr
}
