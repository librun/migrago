package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go" // add ClickHouse driver
	_ "github.com/go-sql-driver/mysql"      // add MySQL driver
	_ "github.com/lib/pq"                   // add postgres driver
	"github.com/librun/migrago/internal/config"
)

const (
	dbTypePostgres   = "postgres"
	dbTypeClickhouse = "clickhouse"
	dbTypeMysql      = "mysql"
)

// Errors.
var (
	ErrUnsupportedDB = errors.New("db type not support")
)

// DB is a database handle representing a pool of zero or more
// underlying connections.
type DB struct {
	typeDB  string
	url     *url.URL
	connect *sql.DB
}

// NewDB initializes connection to database.
func NewDB(cfg *config.Database) (*DB, error) {
	db := DB{
		typeDB: cfg.TypeDB,
	}

	var errURLParse error
	if db.url, errURLParse = url.Parse(cfg.DSN); errURLParse != nil {
		return nil, errURLParse
	}

	if !db.checkSupportDatabaseType() {
		return nil, ErrUnsupportedDB
	}

	println(db.getDSN())

	var errConnect error
	if db.connect, errConnect = sql.Open(db.typeDB, db.getDSN()); errConnect != nil {
		return nil, fmt.Errorf("connect: %w", errConnect)
	}

	if err := db.runAfterConnect(cfg); err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}

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

func (db *DB) checkSupportDatabaseType() bool {
	support := false

	for _, dbDriver := range sql.Drivers() {
		if dbDriver == db.typeDB {
			support = true

			break
		}
	}

	return support
}

func (db *DB) runAfterConnect(cfg *config.Database) error {
	if db.typeDB == dbTypePostgres {
		schema := db.url.Query().Get("schema")

		if schema == "" {
			schema = cfg.Schema
		}

		if schema != "" {
			if _, err := db.connect.Exec("SET search_path TO " + cfg.Schema); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *DB) getDSN() string {
	switch db.typeDB {
	case dbTypePostgres:
		q := db.url.Query()
		q.Del("schema")

		db.url.RawQuery = q.Encode()

	case dbTypeClickhouse:
		q := db.url.Query()

		if db.url.User.Username() != "" {
			q.Set("username", db.url.User.Username())
		}

		if pas, set := db.url.User.Password(); set {
			q.Set("password", pas)
		}

		db.url.User = nil

		dbName := db.GetDatabaseName()
		db.url.Path = ""

		q.Set("database", dbName)

		db.url.RawQuery = q.Encode()
	case dbTypeMysql:
		// TODO: реализовать
	}

	return db.url.String()
}

// GetDatabaseName get database name.
func (db *DB) GetDatabaseName() string {
	return strings.TrimLeft(db.url.Path, "/")
}
