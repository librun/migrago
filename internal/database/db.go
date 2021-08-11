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

type Tx struct {
	tx *sql.Tx
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

// Begin transaction for migration.
func (db *DB) Begin() (*Tx, error) {
	txn, err := db.connect.Begin()
	if err != nil {
		return nil, fmt.Errorf("database begin: %w", err)
	}

	return &Tx{tx: txn}, nil
}

// Close closes connection.
func (db *DB) Close() error {
	return db.connect.Close()
}


// Exec executes a query.
func (t *Tx) Exec(query string) error {
	if _, err := t.tx.Exec(query); err != nil {
		if err := t.tx.Rollback(); err != nil {
			return fmt.Errorf("rollback: %w", err)
		}

		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

// Commit transaction for migration.
func (t *Tx) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("database commit: %w", err)
	}

	return nil
}

// Rollback transaction for migration.
func (t *Tx) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("database rollback: %w", err)
	}

	return nil
}
