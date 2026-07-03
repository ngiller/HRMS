package database

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(databaseURL, encryptionKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("unable to parse database URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	// Set session-level config parameters on each new connection
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if encryptionKey != "" {
			if _, err := conn.Exec(ctx, "SELECT set_config('app.encryption_key', $1, false)", encryptionKey); err != nil {
				return fmt.Errorf("set encryption_key: %w", err)
			}
		}
		// Initialize app.current_user_id so SET LOCAL works in WithUserContext
		// and PL/pgSQL triggers/functions can safely cast it to UUID
		if _, err := conn.Exec(ctx, "SELECT set_config('app.current_user_id', '00000000-0000-0000-0000-000000000000', false)"); err != nil {
			return fmt.Errorf("set current_user_id: %w", err)
		}
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool
	log.Println("Database connected successfully")
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("Database connection closed")
	}
}

// WithUserContext executes fn inside a transaction with app.current_user_id set.
// This is required for audit log triggers that use current_setting('app.current_user_id').
// The setting is scoped to the transaction (SET LOCAL), so it's automatically cleaned up.
func WithUserContext(ctx context.Context, userID string, fn func(tx pgx.Tx) error) error {
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if userID != "" {
		_, err = tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID)
		if err != nil {
			return fmt.Errorf("set user context: %w", err)
		}
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UserIDFromContext extracts a user ID string from fiber context locals.
// The auth middleware stores user_id as uuid.UUID (which is [16]byte).
func UserIDFromContext(userID interface{}) string {
	if userID == nil {
		return ""
	}
	switch v := userID.(type) {
	case [16]byte:
		// Format [16]byte as UUID string: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		return hex.EncodeToString(v[:4]) + "-" +
			hex.EncodeToString(v[4:6]) + "-" +
			hex.EncodeToString(v[6:8]) + "-" +
			hex.EncodeToString(v[8:10]) + "-" +
			hex.EncodeToString(v[10:])
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
