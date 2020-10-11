package storage

import (
	"database/sql"
	"strconv"

	_ "github.com/lib/pq" // init postgresql driver.
)

// StorageTypePostgreSQL константа для типа хранилища postgres.
const StorageTypePostgreSQL = "postgres"

// PostgreSQL тип хранилища postgres.
type PostgreSQL struct {
	connect *sql.DB
}

// Init инициализация соединения с хранилищем.
func (p *PostgreSQL) Init(cfg *Config) error {
	var err error
	p.connect, err = sql.Open(StorageTypePostgreSQL, cfg.DSN)

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

// PreInit подготовка БД к работе.
func (p *PostgreSQL) PreInit(cfg *Config) error {
	var err error

	p.connect, err = sql.Open(StorageTypePostgreSQL, cfg.DSN)
	if err != nil {
		return err
	}

	if cfg.Schema != "" {
		if _, err := p.connect.Exec("SET search_path TO " + cfg.Schema); err != nil {
			return err
		}
	}

	// выполним запрос на создание таблицы для миграций.
	if _, err := p.connect.Exec("CREATE TABLE migration (" +
		"\"project\" varchar NOT NULL, \"database\" varchar NOT NULL,\"version\" varchar NOT NULL, " +
		"\"apply_time\" bigint NOT NULL DEFAULT 0, \"rollback\" bool NOT NULL DEFAULT true, " +
		"CONSTRAINT migration_pk PRIMARY KEY (\"project\",\"database\",\"version\"));"); err != nil {
		return err
	}

	return nil
}

// Close закрытие соединения с хранилищем.
func (p *PostgreSQL) Close() error {
	if p.connect != nil {
		return p.connect.Close()
	}

	return nil
}

// CreateProjectDB для postgres не требуется.
func (*PostgreSQL) CreateProjectDB(string, string) error {
	return nil
}

// CheckMigration проверить выполнялась ли миграция.
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

// Up выполнить миграцию.
func (p *PostgreSQL) Up(post *Migrate) error {
	if _, err := p.connect.Exec(
		"INSERT INTO migration (project, database, version, apply_time, rollback) VALUES ($1, $2, $3, $4, $5)",
		post.Project, post.Database, post.Version, post.ApplyTime, post.RollFlag,
	); err != nil {
		return err
	}

	return nil
}

// GetLast получить список последних миграций.
func (p *PostgreSQL) GetLast(projectName, dbName string, skipNoRollback bool, limit *int) ([]Migrate, error) {
	var result = []Migrate{}
	query := "SELECT project, database, version, apply_time, rollback FROM migration WHERE project = $1 AND database = $2"

	// флаг пропускать неоткатываемые миграции
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

// Delete откатить миграцию.
func (p *PostgreSQL) Delete(post *Migrate) error {
	_, err := p.connect.Exec(
		"DELETE FROM migration WHERE project = $1 AND database = $2 AND version = $3",
		post.Project, post.Database, post.Version,
	)

	// TODO: сделать проверку вдруг есть миграции после текущей

	return err
}
