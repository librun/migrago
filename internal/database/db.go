package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go" // add ClickHouse driver
	_ "github.com/go-sql-driver/mysql"      // add MySQL driver
	_ "github.com/lib/pq"                   // add postgres driver
	"github.com/librun/migrago/internal/config"
)

const dbTypePostgres = "postgres"

// Errors.
var (
	ErrUnsupportedDB = errors.New("db type not support")
)

// DB is a database handle representing a pool of zero or more
// underlying connections.
type DB struct {
	typeDB  string
	connect *sql.DB
}

// NewDB initializes connection to database.
func NewDB(cfg *config.Database) (*DB, error) {
	db := DB{
		typeDB: cfg.TypeDB,
	}

	if !CheckSupportDatabaseType(cfg.TypeDB) {
		return &db, ErrUnsupportedDB
	}

	connect, err := sql.Open(cfg.TypeDB, cfg.DSN)
	if err != nil {
		return &db, fmt.Errorf("connect: %w", err)
	}

	if cfg.TypeDB == dbTypePostgres && cfg.Schema != "" {
		if _, err := connect.Exec("SET search_path TO " + cfg.Schema); err != nil {
			return nil, fmt.Errorf("exec: %w", err)
		}
	}

	db.connect = connect

	return &db, nil
}

// Exec executes a query.
func (db *DB) Exec(query string) error {
	txn, err := db.connect.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	if _, err := txn.Exec(query); err != nil {
		if err := txn.Rollback(); err != nil {
			return fmt.Errorf("rollback: %w", err)
		}

		return fmt.Errorf("exec: %w", err)
	}

	return txn.Commit()
}

// Close closes connection.
func (db *DB) Close() error {
	return db.connect.Close()
}
