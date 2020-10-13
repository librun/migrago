package storage

import (
	"database/sql"
	"strconv"

	_ "github.com/lib/pq" // init postgresql driver.
)

// PostgreSQL is a database handle representing a pool of zero or more
// underlying connections.
type PostgreSQL struct {
	connect *sql.DB
}

// Init opens a database specified by its database driver name and a
// driver-specific data source name.
func (p *PostgreSQL) Init(cfg *Config) error {
	var err error
	p.connect, err = sql.Open(TypePostgres, cfg.DSN)

	if err != nil {
		return err
	}

	if cfg.Schema != "" {
		if _, err := p.connect.Exec("SET search_path TO " + cfg.Schema); err != nil {
			return err
		}
	}

	return nil
}

// PreInit creates migrago table.
func (p *PostgreSQL) PreInit(cfg *Config) error {
	var err error

	p.connect, err = sql.Open(TypePostgres, cfg.DSN)
	if err != nil {
		return err
	}

	if cfg.Schema != "" {
		if _, err := p.connect.Exec("SET search_path TO " + cfg.Schema); err != nil {
			return err
		}
	}

	if _, err := p.connect.Exec("CREATE TABLE migration (" +
		"\"project\" varchar NOT NULL, \"database\" varchar NOT NULL,\"version\" varchar NOT NULL, " +
		"\"apply_time\" bigint NOT NULL DEFAULT 0, \"rollback\" bool NOT NULL DEFAULT true, " +
		"CONSTRAINT migration_pk PRIMARY KEY (\"project\",\"database\",\"version\"));"); err != nil {
		return err
	}

	return nil
}

// Close closes the database and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server
// to finish.
func (p *PostgreSQL) Close() error {
	if p.connect != nil {
		return p.connect.Close()
	}

	return nil
}

// CreateProjectDB does nothing. Function is not needed for postgres.
// Introduced to implement the Storage interface.
func (*PostgreSQL) CreateProjectDB(string, string) error {
	return nil
}

// CheckMigration checks the migration was done successfully.
func (p *PostgreSQL) CheckMigration(projectName, dbName, version string) (bool, error) {
	row := p.connect.QueryRow(
		"SELECT count(*) FROM migration WHERE project = $1 AND database = $2 AND version = $3 LIMIT 1",
		projectName, dbName, version,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// Up runs migration up.
func (p *PostgreSQL) Up(post *Migrate) error {
	if _, err := p.connect.Exec(
		"INSERT INTO migration (project, database, version, apply_time, rollback) VALUES ($1, $2, $3, $4, $5)",
		post.Project, post.Database, post.Version, post.ApplyTime, post.RollFlag,
	); err != nil {
		return err
	}

	return nil
}

// GetLast gets a list of recent migrations.
func (p *PostgreSQL) GetLast(projectName, dbName string, skipNoRollback bool, limit *int) ([]Migrate, error) {
	result := make([]Migrate, 0)

	query := "SELECT project, database, version, apply_time, rollback FROM migration WHERE project = $1 AND database = $2"

	// Flag for skip non-rolling migrations.
	if skipNoRollback {
		query += " AND rollback is true"
	}

	query += " ORDER BY version DESC"

	if limit != nil {
		query += " LIMIT " + strconv.Itoa(*limit)
	}

	rows, err := p.connect.Query(query, projectName, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mi Migrate
		if err := rows.Scan(&mi.Project, &mi.Database, &mi.Version, &mi.ApplyTime, &mi.RollFlag); err != nil {
			continue
		}

		result = append(result, mi)
	}

	return result, rows.Err()
}

// Delete calls migration down.
func (p *PostgreSQL) Delete(post *Migrate) error {
	_, err := p.connect.Exec(
		"DELETE FROM migration WHERE project = $1 AND database = $2 AND version = $3",
		post.Project, post.Database, post.Version,
	)

	// TODO: check if there are migrations after the current one.

	return err
}
