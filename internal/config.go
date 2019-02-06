package internal

import (
	"fmt"
	"os"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	i "github.com/ne-ray/migrago.git/internal/initialize"
	"path/filepath"
)

type (
	//Config struct config yaml file
	Config struct {
		Projects	[]Project
		Databases	[]Database
	}

	//Project struct project for config
	Project struct {
		Name		string
		Migrations	[]ProjectMigration
	}

	//ProjectMigration struct relation project with database
	ProjectMigration struct {
		Path		string
		Database	*Database
	}

	//Database struct for database connection
	Database struct {
		Name	string
		Type	string
		Dsn		string
		Schema	string
	}

	config struct {
		Projects	map[string]configProject  `yaml:"projects"`
		Databases	map[string]configDatabase `yaml:"databases"`
	}

	configProject struct {
		Migrations	[]map[string]string  `yaml:"migrations"`
	}

	configDatabase struct {
		Type	string	`yaml:"type"`
		Dsn		string	`yaml:"dsn"`
		Schema	string	`yaml:"schema"`
	}
)

//InitConfig init config from file
func InitConfig(path string, projects, databases []string) (Config, error) {
	configFile, err := os.Open(path)
	defer configFile.Close()
	if err != nil {
		return Config{}, fmt.Errorf("Config open error: %s", err)
	}

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("Config read error: %s", err)
	}

	cfg := config{}

	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return Config{}, fmt.Errorf("Config format error: %s", err)
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

	config := Config{}

	for dbName, db := range cfg.Databases {
		//Если не нужна данная БД пропустим её
		if _, ok := dbCurrent[dbName]; dbDelete && !ok {
			continue
		}

		//Проверим тип БД
		if !i.CheckSupportDatabaseType(db.Type) {
			return Config{}, fmt.Errorf("Type database %s not support", db.Type)
		}
		config.Databases = append(config.Databases, Database{
			Name: dbName,
			Type: db.Type,
			Dsn: db.Dsn,
			Schema: db.Schema,
		})
	}

	for prjName, prjMigration := range cfg.Projects {
		//Если не нужен данный проект пропустим его
		if _, ok := projectCurrent[prjName]; projectDelete && !ok {
			continue
		}

		project := Project{
			Name: prjName,
		}
		for _, migration := range prjMigration.Migrations {
			for dbName, path := range migration {
				//Если не нужна данная БД пропустим её
				if _, ok := dbCurrent[dbName]; dbDelete && !ok {
					continue
				}

				if db, err := config.GetDB(dbName); err == nil {
					//Проверим путь до миграций на существование
					if fi, err := os.Stat(path); os.IsNotExist(err) || !fi.IsDir() {
						return Config{}, fmt.Errorf("Directory %s not exists", path)
					}

					path, err := filepath.Abs(path)

					if err != nil {
						return Config{}, err
					}

					project.Migrations = append(project.Migrations, ProjectMigration{
						Path: path + "/",
						Database: &db,
					})
				} else {
					return Config{}, fmt.Errorf("Database %s not found in databases", dbName)
				}
			}
		}

		config.Projects = append(config.Projects, project)
	}

	return config, nil
}

//GetDB get database object by name
func (c *Config) GetDB(name string) (Database, error) {
	for _, db := range c.Databases {
		if db.Name == name {
			return db, nil
		}
 	}

	return Database{}, fmt.Errorf("Database %s not found in databases", name)
}

//GetProject get project object by name
func (c *Config) GetProject(name string) (Project, error) {
	for _, prj := range c.Projects {
		if prj.Name == name {
			return prj, nil
		}
	}

	return Project{}, fmt.Errorf("Project %s not found in projects", name)
}

//GetDB get database object by name
func (p *Project) GetDB(name string) (Database, error) {
	for _, migration := range p.Migrations {
		if migration.Database.Name == name {
			return *migration.Database, nil
		}
	}

	return Database{}, fmt.Errorf("Database %s not found in project %s databases", name, p.Name)
}

//GetProjectMigration get relation database in project by database name
func (p *Project) GetProjectMigration(dbName string) (ProjectMigration, error) {
	for _, migration := range p.Migrations {
		if migration.Database.Name == dbName {
			return migration, nil
		}
	}

	return ProjectMigration{}, fmt.Errorf("Database %s not found in project %s databases", dbName, p.Name)
}
