package initialize

import (
	"errors"
	"database/sql"
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
//	_ "github.com/mattn/go-sqlite3"
	_ "github.com/kshvakov/clickhouse"
)

type (
	DB struct {
		Type	string
		Connect *sql.DB
	}

	Database struct {
		Name	string
		Type	string
		Dsn		string
		Schema	string
	}
)

const DB_TYPE_POSTGRES = "postgres"

func CheckSupportDatabaseType(dbType string) bool {
	support := false
	for _, dbDriver := range sql.Drivers() {
		if dbDriver == dbType {
			support = true
			break
		}
	}

	return support
}

func InitDB(dbType, dsnData, dbSchema string) (*DB, error) {
	db := &DB{}

	if !CheckSupportDatabaseType(dbType) {
		return db, errors.New("DB type not support")
	}

	connect, err := sql.Open(dbType, dsnData)

	if err != nil {
		return db, err
	}

	if dbType == DB_TYPE_POSTGRES && dbSchema != "" {
		if _, err := connect.Exec("SET search_path TO " + dbSchema); err != nil {
			return nil, err
		}
	}

	db.Connect = connect

	return db, nil
}


func (db *DB) Exec(query string) error {
	txn, err := db.Connect.Begin()
	if err != nil {
		return err
	}

	if _, err := txn.Exec(query); err != nil {
		txn.Rollback()
		return err
	}

	return txn.Commit()
}

func (db *DB) Close() error {
	if err := db.Connect.Close(); err != nil {
		return err
	}

	return nil
}
