package database

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/librun/migrago/internal/config"
)

type parserDSNClickhouse struct {
	dsn    string
	dbName string
}

func (p *parserDSNClickhouse) Parse(cfg *config.Database) error {
	u, errURLParse := url.Parse(cfg.DSN)
	if errURLParse != nil {
		return errURLParse
	}

	q := u.Query()

	if u.User.Username() != "" {
		q.Set("username", u.User.Username())
	}

	if pas, set := u.User.Password(); set {
		q.Set("password", pas)
	}

	u.User = nil

	p.dbName = strings.TrimLeft(u.Path, "/")
	u.Path = ""

	if p.dbName == "" {
		p.dbName = q.Get("database")
	}

	q.Del("database")

	if p.dbName != "" {
		q.Set("database", p.dbName)
	}

	u.RawQuery = q.Encode()
	p.dsn = u.String()

	return nil
}

func (p *parserDSNClickhouse) GetDSN() string {
	return p.dsn
}

func (p *parserDSNClickhouse) RunAfterConnect(connect *sql.DB) error {
	return nil
}

func (p *parserDSNClickhouse) GetDatabaseName() string {
	return p.dbName
}
