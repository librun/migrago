package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	// migratePostfixUp postfix for file up migrate
	migratePostfixUp = "_up.sql"

	// migratePostfixDown postfix for file up migrate
	migratePostfixDown = "_down.sql"
)

type (
	// Config struct Config yaml file
	Config struct {
		projects  []Project
		databases []Database
	}

	// Project struct Project for Config
	Project struct {
		name       string
		migrations []ProjectMigration
	}

	// ProjectMigration struct relation Project with Database
	ProjectMigration struct {
		path     string
		database *Database
	}

	// Database struct for Database connection
	Database struct {
		name   string
		typeDB string
		dsn    string
		schema string
	}

	YAMLConfig struct {
		Projects  map[string]YAMLConfigProject  `yaml:"projects"`
		Databases map[string]YAMLConfigDatabase `yaml:"databases"`
	}

	YAMLConfigProject struct {
		Migrations []map[string]string `yaml:"migrations"`
	}

	YAMLConfigDatabase struct {
		Type   string `yaml:"type"`
		DSN    string `yaml:"dsn"`
		Schema string `yaml:"schema"`
	}
)

// initConfig init Config from file.
func initConfig(path string, projects, databases []string) (Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("config open error: %s", err)
	}
	defer configFile.Close()

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("config read error: %s", err)
	}

	cfg := YAMLConfig{}

	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return Config{}, fmt.Errorf("config format error: %s", err)
	}

	projectDelete := false
	projectCurrent := map[string]bool{}
	if len(projects) > 0 {
		projectDelete = true
		for _, projectName := range projects {
			projectCurrent[projectName] = true
		}
	}

	dbDelete := false
	dbCurrent := map[string]bool{}
	if len(databases) > 0 {
		dbDelete = true
		for _, projectName := range databases {
			dbCurrent[projectName] = true
		}
	}

	conf := Config{}

	for dbName, db := range cfg.Databases {
		// Если не нужна данная БД пропустим её
		if _, ok := dbCurrent[dbName]; dbDelete && !ok {
			continue
		}

		// Проверим тип БД
		if !checkSupportDatabaseType(db.Type) {
			return Config{}, fmt.Errorf("type database %s not support", db.Type)
		}

		conf.databases = append(conf.databases, Database{
			name:   dbName,
			typeDB: db.Type,
			dsn:    db.DSN,
			schema: db.Schema,
		})
	}

	for prjName, prjMigration := range cfg.Projects {
		// Если не нужен данный проект пропустим его
		if _, ok := projectCurrent[prjName]; projectDelete && !ok {
			continue
		}

		project := Project{
			name: prjName,
		}

		for _, migration := range prjMigration.Migrations {
			for dbName, path := range migration {
				// Если не нужна данная БД пропустим её
				if _, ok := dbCurrent[dbName]; dbDelete && !ok {
					continue
				}

				if db, err := conf.getDB(dbName); err == nil {
					// Проверим путь до миграций на существование
					if fi, err := os.Stat(path); os.IsNotExist(err) || !fi.IsDir() {
						return Config{}, fmt.Errorf("directory %s not exists", path)
					}

					path, err = filepath.Abs(path)
					if err != nil {
						return Config{}, err
					}

					project.migrations = append(project.migrations, ProjectMigration{
						path:     path + "/",
						database: &db,
					})
				} else {
					return Config{}, fmt.Errorf("database %s not found in databases", dbName)
				}
			}
		}

		conf.projects = append(conf.projects, project)
	}

	return conf, nil
}

// getDB get Database object by name.
func (c *Config) getDB(name string) (Database, error) {
	for _, db := range c.databases {
		if db.name == name {
			return db, nil
		}
	}

	return Database{}, fmt.Errorf("database %s not found in databases", name)
}

// getProject get Project object by name.
func (c *Config) getProject(name string) (Project, error) {
	for _, prj := range c.projects {
		if prj.name == name {
			return prj, nil
		}
	}

	return Project{}, fmt.Errorf("project %s not found in projects", name)
}

// getDB get Database object by name.
func (p *Project) getDB(name string) (Database, error) {
	for _, migration := range p.migrations {
		if migration.database.name == name {
			return *migration.database, nil
		}
	}

	return Database{}, fmt.Errorf("database %s not found in Project %s databases", name, p.name)
}

// getProjectMigration get relation Database in Project by Database name.
func (p *Project) getProjectMigration(dbName string) (ProjectMigration, error) {
	for _, migration := range p.migrations {
		if migration.database.name == dbName {
			return migration, nil
		}
	}

	return ProjectMigration{}, fmt.Errorf("database %s not found in Project %s databases", dbName, p.name)
}
