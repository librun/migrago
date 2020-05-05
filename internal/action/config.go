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
	// config struct config yaml file
	config struct {
		projects  []project
		databases []database
	}

	// project struct project for config
	project struct {
		name       string
		migrations []projectMigration
	}

	// projectMigration struct relation project with database
	projectMigration struct {
		path     string
		database *database
	}

	// Database struct for database connection
	database struct {
		name   string
		typeDB string
		dsn    string
		schema string
	}

	yamlConfig struct {
		Projects  map[string]yamlConfigProject  `yaml:"projects"`
		Databases map[string]yamlConfigDatabase `yaml:"databases"`
	}

	yamlConfigProject struct {
		Migrations []map[string]string `yaml:"migrations"`
	}

	yamlConfigDatabase struct {
		Type   string `yaml:"type"`
		DSN    string `yaml:"dsn"`
		Schema string `yaml:"schema"`
	}
)

// initConfig init config from file
func initConfig(path string, projects, databases []string) (config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return config{}, fmt.Errorf("Config open error: %s", err)
	}
	defer configFile.Close()

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config{}, fmt.Errorf("Config read error: %s", err)
	}

	cfg := yamlConfig{}

	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return config{}, fmt.Errorf("Config format error: %s", err)
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

	conf := config{}

	for dbName, db := range cfg.Databases {
		//Если не нужна данная БД пропустим её
		if _, ok := dbCurrent[dbName]; dbDelete && !ok {
			continue
		}

		//Проверим тип БД
		if !checkSupportDatabaseType(db.Type) {
			return config{}, fmt.Errorf("Type database %s not support", db.Type)
		}
		conf.databases = append(conf.databases, database{
			name:   dbName,
			typeDB: db.Type,
			dsn:    db.DSN,
			schema: db.Schema,
		})
	}

	for prjName, prjMigration := range cfg.Projects {
		//Если не нужен данный проект пропустим его
		if _, ok := projectCurrent[prjName]; projectDelete && !ok {
			continue
		}

		project := project{
			name: prjName,
		}
		for _, migration := range prjMigration.Migrations {
			for dbName, path := range migration {
				//Если не нужна данная БД пропустим её
				if _, ok := dbCurrent[dbName]; dbDelete && !ok {
					continue
				}

				if db, err := conf.getDB(dbName); err == nil {
					//Проверим путь до миграций на существование
					if fi, err := os.Stat(path); os.IsNotExist(err) || !fi.IsDir() {
						return config{}, fmt.Errorf("Directory %s not exists", path)
					}

					path, err := filepath.Abs(path)

					if err != nil {
						return config{}, err
					}

					project.migrations = append(project.migrations, projectMigration{
						path:     path + "/",
						database: &db,
					})
				} else {
					return config{}, fmt.Errorf("Database %s not found in databases", dbName)
				}
			}
		}

		conf.projects = append(conf.projects, project)
	}

	return conf, nil
}

// getDB get database object by name
func (c *config) getDB(name string) (database, error) {
	for _, db := range c.databases {
		if db.name == name {
			return db, nil
		}
	}

	return database{}, fmt.Errorf("Database %s not found in databases", name)
}

// getProject get project object by name
func (c *config) getProject(name string) (project, error) {
	for _, prj := range c.projects {
		if prj.name == name {
			return prj, nil
		}
	}

	return project{}, fmt.Errorf("Project %s not found in projects", name)
}

// getDB get database object by name
func (p *project) getDB(name string) (database, error) {
	for _, migration := range p.migrations {
		if migration.database.name == name {
			return *migration.database, nil
		}
	}

	return database{}, fmt.Errorf("Database %s not found in project %s databases", name, p.name)
}

// getProjectMigration get relation database in project by database name
func (p *project) getProjectMigration(dbName string) (projectMigration, error) {
	for _, migration := range p.migrations {
		if migration.database.name == dbName {
			return migration, nil
		}
	}

	return projectMigration{}, fmt.Errorf("Database %s not found in project %s databases", dbName, p.name)
}
