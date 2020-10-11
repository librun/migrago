package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type (
	// Config struct Config yaml file.
	Config struct {
		Projects  []Project
		Databases []Database
	}

	// Project struct Project for Config.
	Project struct {
		Name       string
		Migrations []ProjectMigration
	}

	// ProjectMigration struct relation Project with Database.
	ProjectMigration struct {
		Path     string
		Database *Database
	}

	// Database struct for Database connection.
	Database struct {
		Name   string
		TypeDB string
		DSN    string
		Schema string
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

// NewConfig init Config from file.
func NewConfig(path string, projects, databases []string) (Config, error) {
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

	conf := Config{
		Databases: cfg.parseDatabases(dbCurrent, dbDelete),
	}

	conf.Projects, err = cfg.parseProjects(conf, projectCurrent, projectDelete, dbCurrent, dbDelete)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

// GetDB get Database object by Name.
func (c *Config) GetDB(name string) (Database, error) {
	for _, db := range c.Databases {
		if db.Name == name {
			return db, nil
		}
	}

	return Database{}, fmt.Errorf("database %s not found in Databases", name)
}

// GetProject get Project object by Name.
func (c *Config) GetProject(name string) (Project, error) {
	for _, prj := range c.Projects {
		if prj.Name == name {
			return prj, nil
		}
	}

	return Project{}, fmt.Errorf("project %s not found in Projects", name)
}

// GetDB get Database object by Name.
func (p *Project) GetDB(name string) (Database, error) {
	for _, migration := range p.Migrations {
		if migration.Database.Name == name {
			return *migration.Database, nil
		}
	}

	return Database{}, fmt.Errorf("database %s not found in Project %s Databases", name, p.Name)
}

// GetProjectMigration get relation Database in Project by Database Name.
func (p *Project) GetProjectMigration(dbName string) (ProjectMigration, error) {
	for _, migration := range p.Migrations {
		if migration.Database.Name == dbName {
			return migration, nil
		}
	}

	return ProjectMigration{}, fmt.Errorf("database %s not found in Project %s Databases", dbName, p.Name)
}

func (cfg *YAMLConfig) parseDatabases(dbCurrent map[string]bool, dbDelete bool) []Database {
	databases := make([]Database, 0, len(cfg.Databases))

	for dbName, db := range cfg.Databases {
		// Если не нужна данная БД пропустим её.
		if _, ok := dbCurrent[dbName]; dbDelete && !ok {
			continue
		}

		// Проверим тип БД.
		// if !database.CheckSupportDatabaseType(db.Type) {
		//	 return Config{}, fmt.Errorf("type Database %s not support", db.Type)
		// }

		databases = append(databases, Database{
			Name:   dbName,
			TypeDB: db.Type,
			DSN:    db.DSN,
			Schema: db.Schema,
		})
	}

	return databases
}

func (cfg *YAMLConfig) parseProjects(conf Config, projectCurrent map[string]bool, projectDelete bool, dbCurrent map[string]bool, dbDelete bool) ([]Project, error) {
	projects := make([]Project, 0, len(cfg.Projects))

	for prjName, prjMigration := range cfg.Projects {
		// Если не нужен данный проект пропустим его.
		if _, ok := projectCurrent[prjName]; projectDelete && !ok {
			continue
		}

		project := Project{
			Name: prjName,
		}

		for _, migration := range prjMigration.Migrations {
			for dbName, path := range migration {
				// Если не нужна данная БД пропустим её.
				if _, ok := dbCurrent[dbName]; dbDelete && !ok {
					continue
				}

				if db, err := conf.GetDB(dbName); err == nil {
					// Проверим путь до миграций на существование.
					if fi, err := os.Stat(path); os.IsNotExist(err) || !fi.IsDir() {
						return projects, fmt.Errorf("directory %s not exists", path)
					}

					path, err = filepath.Abs(path)
					if err != nil {
						return projects, err
					}

					project.Migrations = append(project.Migrations, ProjectMigration{
						Path:     path + "/",
						Database: &db,
					})
				} else {
					return projects, fmt.Errorf("database %s not found in Databases", dbName)
				}
			}
		}

		projects = append(projects, project)
	}

	return projects, nil
}
