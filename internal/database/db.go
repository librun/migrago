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
type (
	DB struct {
		typeDB  string
		connect *sql.DB
		pDSN    parserDSN
	}

	parserDSN interface {
		Parse(*config.Database) error
		GetDSN() string
		GetDatabaseName() string
		RunAfterConnect(*sql.DB) error
	}
)

// NewDBConnect initializes connection to database.
func NewDB(cfg *config.Database) (*DB, error) {
	db := DB{
		typeDB: cfg.TypeDB,
	}

	if pDSN, errGetPDSN := db.getParserDSN(cfg); errGetPDSN == nil {
		db.pDSN = pDSN
	} else {
		return nil, errGetPDSN
	}

	var errConnect error
	if db.connect, errConnect = sql.Open(db.typeDB, db.pDSN.GetDSN()); errConnect != nil {
		return nil, fmt.Errorf("connect: %w", errConnect)
	}

	if err := db.pDSN.RunAfterConnect(db.connect); err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}

	return &db, nil
}

// NewDBWithoutConnect initializes database without connect.
func NewDBWithoutConnect(cfg *config.Database) (*DB, error) {
	db := DB{
		typeDB: cfg.TypeDB,
	}

	if pDSN, errGetPDSN := db.getParserDSN(cfg); errGetPDSN == nil {
		db.pDSN = pDSN
	} else {
		return nil, errGetPDSN
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

// GetDatabaseName get database name.
func (db *DB) GetDatabaseName() string {
	return db.pDSN.GetDatabaseName()
}

func (db *DB) getParserDSN(cfg *config.Database) (parserDSN, error) {
	support := false

	for _, dbDriver := range sql.Drivers() {
		if dbDriver == db.typeDB {
			support = true

			break
		}
	}

	if support {
		var dbConnect parserDSN

		switch db.typeDB {
		case dbTypePostgres:
			dbConnect = &parserDSNPostgres{}

		case dbTypeClickhouse:
			dbConnect = &parserDSNClickhouse{}

		case dbTypeMysql:
			dbConnect = &parserDSNMysql{}

		default:
			return nil, ErrUnsupportedDB
		}

		if err := dbConnect.Parse(cfg); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return dbConnect, nil
	}

	return nil, ErrUnsupportedDB
}
