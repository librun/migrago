package database

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/librun/migrago/internal/config"
)

type parserDSNPostgres struct {
	dsn    string
	schema string
	dbName string
}

func (p *parserDSNPostgres) Parse(cfg *config.Database) error {
	u, errURLParse := url.Parse(cfg.DSN)
	if errURLParse != nil {
		return errURLParse
	}

	schema := u.Query().Get("schema")
	if schema == "" {
		p.schema = cfg.Schema
	}

	q := u.Query()
	q.Del("schema")
	u.RawQuery = q.Encode()

	p.dsn = u.String()

	p.dbName = strings.TrimLeft(u.Path, "/")

	return nil
}

func (p *parserDSNPostgres) GetDSN() string {
	return p.dsn
}

func (p *parserDSNPostgres) RunAfterConnect(connect *sql.DB) error {
	if p.schema != "" {
		if _, err := connect.Exec("SET search_path TO " + p.schema); err != nil {
			return err
		}
	}

	return nil
}

func (p *parserDSNPostgres) GetDatabaseName() string {
	return p.dbName
}
