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
	params string
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
	p.params = q.Encode()

	p.dbName = strings.TrimLeft(u.Path, "/")

	u.RawQuery = ""
	u.Path = ""
	p.dsn = u.String()

	return nil
}

func (p *parserDSNPostgres) GetDSN() string {
	params := p.params

	if params != "" {
		params = "?" + params
	}

	dbName := p.dbName
	if dbName != "" {
		dbName = "/" + dbName
	}

	return p.dsn + dbName + params
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
