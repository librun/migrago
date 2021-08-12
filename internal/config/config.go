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
		Schema string // TODO: Deprecated
	}

	// YAMLConfig takes configuration values from YAML file.
	YAMLConfig struct {
		Projects  map[string]YAMLConfigProject  `yaml:"projects"`
		Databases map[string]YAMLConfigDatabase `yaml:"databases"`
	}

	// YAMLConfigProject is a block for parse projects in YAML config file.
	YAMLConfigProject struct {
		Migrations []map[string]string `yaml:"migrations"`
	}

	// YAMLConfigDatabase is a block for parse databases in YAML config file.
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
		return Config{}, fmt.Errorf("config open: %w", err)
	}
	defer configFile.Close()

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("config read: %w", err)
	}

	cfg := YAMLConfig{}
	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return Config{}, fmt.Errorf("config format: %w", err)
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

// GetDB gets database object by name.
func (c *Config) GetDB(name string) (Database, error) {
	for _, db := range c.Databases {
		if db.Name == name {
			return db, nil
		}
	}

	return Database{}, fmt.Errorf("database %s not found in Databases", name)
}

// GetProject gets project object by name.
func (c *Config) GetProject(name string) (Project, error) {
	for _, prj := range c.Projects {
		if prj.Name == name {
			return prj, nil
		}
	}

	return Project{}, fmt.Errorf("project %s not found in Projects", name)
}

// GetDB gets database object by name.
func (p *Project) GetDB(name string) (Database, error) {
	for _, migration := range p.Migrations {
		if migration.Database.Name == name {
			return *migration.Database, nil
		}
	}

	return Database{}, fmt.Errorf("database %s not found in Project %s Databases", name, p.Name)
}

// GetProjectMigration gets relation database in project by database name.
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
		// If this database is not needed, skip it.
		if _, ok := dbCurrent[dbName]; dbDelete && !ok {
			continue
		}

		// Check database type.
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
		// If this project is not needed, skip it.
		if _, ok := projectCurrent[prjName]; projectDelete && !ok {
			continue
		}

		project := Project{
			Name: prjName,
		}

		for _, migration := range prjMigration.Migrations {
			for dbName, path := range migration {
				// If this database is not needed, skip it.
				if _, ok := dbCurrent[dbName]; dbDelete && !ok {
					continue
				}

				if db, err := conf.GetDB(dbName); err == nil {
					// Check that the path to migrations exists.
					if fi, err := os.Stat(path); os.IsNotExist(err) || !fi.IsDir() {
						return projects, fmt.Errorf("directory %s not exists", path)
					}

					path, err = filepath.Abs(path)
					if err != nil {
						return projects, fmt.Errorf("get directory %s absolute path: %w", path, err)
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
