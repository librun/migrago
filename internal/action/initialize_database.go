package action

import (
	"errors"

	"database/sql"
	// init sql drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/lib/pq"
	//	_ "github.com/mattn/go-sqlite3"
)

type (
	// db struct for database
	db struct {
		typeDB  string
		connect *sql.DB
	}
)

// dbTypePostgres const for database postgres
const dbTypePostgres = "postgres"

// checkSupportDatabaseType check support database driver
func checkSupportDatabaseType(dbType string) bool {
	support := false
	for _, dbDriver := range sql.Drivers() {
		if dbDriver == dbType {
			support = true
			break
		}
	}

	return support
}

// initDB initialize connection with database
func initDB(cfg *database) (*db, error) {
	dbc := &db{
		typeDB: cfg.typeDB,
	}

	if !checkSupportDatabaseType(cfg.typeDB) {
		return dbc, errors.New("DB type not support")
	}

	connect, err := sql.Open(cfg.typeDB, cfg.dsn)

	if err != nil {
		return dbc, err
	}

	if cfg.typeDB == dbTypePostgres && cfg.schema != "" {
		if _, err := connect.Exec("SET search_path TO " + cfg.schema); err != nil {
			return nil, err
		}
	}

	dbc.connect = connect

	return dbc, nil
}

// Exec run query
func (dbc *db) exec(query string) error {
	txn, err := dbc.connect.Begin()
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

// Close close connection
func (dbc *db) close() error {
	if err := dbc.connect.Close(); err != nil {
		return err
	}

	return nil
}
