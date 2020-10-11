package action

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/librun/migrago/internal/config"
	"github.com/librun/migrago/internal/database"
	"github.com/librun/migrago/internal/storage"
)

const (
	// migratePostfixUp postfix for file up migrate.
	migratePostfixUp = "_up.sql"

	// migratePostfixDown postfix for file up migrate.
	migratePostfixDown = "_down.sql"
)

// MakeUp do migrate up.
func MakeUp(mStorage storage.Storage, cfgPath string, project, dbName *string) error {
	projects := make([]string, 0)
	if project != nil {
		projects = append(projects, *project)
	}

	databases := make([]string, 0)
	if dbName != nil {
		databases = append(databases, *dbName)
	}

	cfg, err := config.NewConfig(cfgPath, projects, databases)
	if err != nil {
		return err
	}

	for _, project := range cfg.Projects {
		log.Println("Project: " + project.Name)
		log.Println("----------")

		for _, migration := range project.Migrations {
			log.Println("DB: " + migration.Database.Name)
			// создаем бакет по имени проекта
			if err := mStorage.CreateProjectDB(project.Name, migration.Database.Name); err != nil {
				return err
			}

			// список всех файлов
			filesInDir, err := ioutil.ReadDir(migration.Path)
			if err != nil {
				return err
			}

			var keys []string

			for _, f := range filesInDir {
				fileName := f.Name()
				// получаем список файлов на создание миграций, если имя короче 8 символов пропускаем
				if len(fileName) > 7 && fileName[len(fileName)-7:] == migratePostfixUp {
					keys = append(keys, fileName[:len(fileName)-7])
				}
			}

			// сортируем список миграций по дате создания
			sort.Strings(keys)

			_, err = makeMigrationInDB(mStorage, migration, project.Name, keys)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func makeMigrationInDB(mStorage storage.Storage, migration config.ProjectMigration, projectName string, keys []string) (int, error) {
	defer log.Println("----------")

	var countCompleted int
	var countTotal int

	dbc, errDB := database.NewDB(migration.Database)
	if errDB != nil {
		return countCompleted, errDB
	}

	defer func() {
		log.Println("Completed migrations:", countCompleted, "of", countTotal)

		if err := dbc.Close(); err != nil {
			panic(err)
		}
	}()

	// Calculate total migrations to run.
	var workKeys []string

	for _, version := range keys {
		if haveMigrate, err := mStorage.CheckMigration(projectName, migration.Database.Name, version); !haveMigrate {
			workKeys = append(workKeys, version)
		} else if err != nil {
			return countCompleted, err
		}
	}

	countTotal = len(workKeys)

	for _, version := range workKeys {
		content, err := ioutil.ReadFile(migration.Path + version + migratePostfixUp)
		if err != nil {
			return countCompleted, err
		}

		query := string(content)
		if strings.TrimSpace(query) != "" {
			// выполнение всех запросов из текущего файла
			if errExec := dbc.Exec(query); errExec != nil {
				log.Println("migration fail: " + version)
				return countCompleted, errExec
			}
		}

		post := &storage.Migrate{
			Project:   projectName,
			Database:  migration.Database.Name,
			Version:   version,
			ApplyTime: time.Now().UTC().Unix(),
			RollFlag:  true,
		}

		// если файла с окончанием down.sql не существует, то указываем, что эта миграция не откатываемая
		if _, err := os.Stat(migration.Path + version + migratePostfixDown); os.IsNotExist(err) {
			post.RollFlag = false
		}

		if err := mStorage.Up(post); err != nil {
			log.Println("migration fail: " + version)
			return countCompleted, err
		}

		log.Println("migration success: " + version)
		countCompleted++
	}

	return countCompleted, nil
}
