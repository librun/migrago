package database

import (
	"database/sql"
	"errors"
	"regexp"

	"github.com/librun/migrago/internal/config"
)

type parserDSNMysql struct {
	dsn    string
	dbName string
}

// nolint: lll
var parseDSNRegexp = regexp.MustCompile("(?i)" + `^(([a-z0-9\-\_]+?):\/\/)?(([a-z0-9\-\_]+?)\:)?(([a-z0-9\-\_]+?)@)?(([a-z0-9\-\_]+?)\()?(.*?)(\)?)\/(.*?)(\?.*)?$`)

func (p *parserDSNMysql) Parse(cfg *config.Database) error {
	parseResult := parseDSNRegexp.FindAllStringSubmatch(cfg.DSN, -1)

	if len(parseResult) < 1 || len(parseResult[0]) < 12 {
		return errors.New("dsn for mysql not valid")
	}

	p.dbName = parseResult[0][11]

	dsn := ""

	user := parseResult[0][4]
	password := parseResult[0][6]

	if user != "" || password != "" {
		dsn = user

		if password != "" {
			dsn += ":" + password
		}

		dsn += "@"
	}

	host := parseResult[0][9]

	protocol := parseResult[0][2]
	if protocol == "" {
		protocol = parseResult[0][8]
	}

	if protocol != "" {
		host = protocol + "(" + host + ")"
	}

	p.dsn = dsn + host + "/" + p.dbName + parseResult[0][12]

	return nil
}

func (p *parserDSNMysql) GetDSN() string {
	return p.dsn
}

func (p *parserDSNMysql) RunAfterConnect(connect *sql.DB) error {
	return nil
}

func (p *parserDSNMysql) GetDatabaseName() string {
	return p.dbName
}
