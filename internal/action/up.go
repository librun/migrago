package action

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/librun/migrago/internal/storage"
)

// MakeUp do migrate up
func MakeUp(mStorage storage.Storage, cfgPath string, project, database *string) error {
	var projects = make([]string, 0)
	if project != nil {
		projects = append(projects, *project)
	}

	var databases = make([]string, 0)
	if database != nil {
		databases = append(databases, *database)
	}

	config, err := initConfig(cfgPath, projects, databases)
	if err != nil {
		return err
	}

	for _, project := range config.projects {
		log.Println("Project: " + project.name)
		log.Println("----------")
		for _, migration := range project.migrations {
			log.Println("DB: " + migration.database.name)
			// создаем бакет по имени проекта
			if err := mStorage.CreateProjectDB(project.name, migration.database.name); err != nil {
				return err
			}

			// список всех файлов
			filesInDir, err := ioutil.ReadDir(migration.path)
			if err != nil {
				return err
			}

			var keys []string
			for _, f := range filesInDir {
				fileName := f.Name()
				//получаем список файлов на создание миграций, если имя короче 8 символов пропускаем
				if len(fileName) > 7 && fileName[len(fileName)-7:] == migratePostfixUp {
					keys = append(keys, fileName[:len(fileName)-7])
				}
			}
			// сортируем список миграций по дате создания
			sort.Strings(keys)

			_, err = makeMigrationInDB(mStorage, &migration, project.name, keys)
			log.Println("----------")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func makeMigrationInDB(mStorage storage.Storage, migration *projectMigration, projectName string, keys []string) (int, error) {
	var countCompleted int
	var countTotal int
	// откроем соединение с БД
	dbc, errDB := initDB(migration.database)
	if errDB != nil {
		return countCompleted, errDB
	}
	// закроеи подключение к БД
	defer func() {
		log.Println("Completed migrations:", countCompleted, "of", countTotal)
		if err := dbc.close(); err != nil {
			panic(err)
		}
	}()

	// calculate total migrage for run
	var workKeys []string
	for _, version := range keys {
		if haveMigrate, err := mStorage.CheckMigration(projectName, migration.database.name, version); !haveMigrate {
			workKeys = append(workKeys, version)
		} else if err != nil {
			return countCompleted, err
		}
	}
	countTotal = len(workKeys)

	for _, version := range workKeys {
		content, err := ioutil.ReadFile(migration.path + version + migratePostfixUp)
		if err != nil {
			return countCompleted, err
		}

		query := string(content)
		if strings.TrimSpace(query) != "" {
			//выполнение всех запросов из текущего файла
			if errExec := dbc.exec(query); errExec != nil {
				log.Println("migration fail: " + version)
				return countCompleted, errExec
			}
		}

		post := &storage.Migrate{
			Project:   projectName,
			Database:  migration.database.name,
			Version:   version,
			ApplyTime: time.Now().UTC().Unix(),
			RollFlag:  true,
		}

		// если файла с окончанием down.sql не существует, то указываем, что эта миграция не откатываемая
		if _, err := os.Stat(migration.path + version + migratePostfixDown); os.IsNotExist(err) {
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
