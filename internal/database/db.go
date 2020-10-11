package database

import (
	"database/sql"
	"errors"

	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/librun/migrago/internal/config"
)

// dbTypePostgres const for Database postgres.
const dbTypePostgres = "postgres"

// DB struct for Database.
type DB struct {
	typeDB  string
	connect *sql.DB
}

// NewDB initializes connection to Database.
func NewDB(cfg *config.Database) (*DB, error) {
	db := DB{
		typeDB: cfg.TypeDB,
	}

	if !CheckSupportDatabaseType(cfg.TypeDB) {
		return &db, errors.New("DB type not support")
	}

	connect, err := sql.Open(cfg.TypeDB, cfg.DSN)
	if err != nil {
		return &db, err
	}

	if cfg.TypeDB == dbTypePostgres && cfg.Schema != "" {
		if _, err := connect.Exec("SET search_path TO " + cfg.Schema); err != nil {
			return nil, err
		}
	}

	db.connect = connect

	return &db, nil
}

// Exec runs query.
func (db *DB) Exec(query string) error {
	txn, err := db.connect.Begin()
	if err != nil {
		return err
	}

	if _, err := txn.Exec(query); err != nil {
		if err := txn.Rollback(); err != nil {
			return err
		}

		return err
	}

	return txn.Commit()
}

// Close closes connection.
func (db *DB) Close() error {
	if err := db.connect.Close(); err != nil {
		return err
	}

	return nil
}
