package action

import (
	"fmt"
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
	// migratePostfixUp is a postfix for file up migrate.
	migratePostfixUp = "_up.sql"

	// migratePostfixDown is a postfix for file up migrate.
	migratePostfixDown = "_down.sql"
)

// MakeUp applies migrations.
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
		return fmt.Errorf("get config: %w", err)
	}

	for _, project := range cfg.Projects {
		log.Println("Project: " + project.Name)
		log.Println("----------")

		for _, migration := range project.Migrations {
			log.Println("DB: " + migration.Database.Name)
			// Create a bucket by the name of the project.
			if err := mStorage.CreateProjectDB(project.Name, migration.Database.Name); err != nil {
				return fmt.Errorf("create project db: %w", err)
			}

			// All files list.
			filesInDir, err := ioutil.ReadDir(migration.Path)
			if err != nil {
				return fmt.Errorf("get files list: %w", err)
			}

			var keys []string

			for _, f := range filesInDir {
				fileName := f.Name()
				// Get a list of files to create migrations. Skip if the name is shorter
				// than 8 characters.
				if len(fileName) > 7 && fileName[len(fileName)-7:] == migratePostfixUp {
					keys = append(keys, fileName[:len(fileName)-7])
				}
			}

			// Sort the list of migrations by creation date.
			sort.Strings(keys)

			if _, err := makeMigrationInDB(mStorage, migration, project.Name, keys); err != nil {
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
			return countCompleted, fmt.Errorf("check migration: %w", err)
		}
	}

	countTotal = len(workKeys)

	for _, version := range workKeys {
		content, errReadFile := ioutil.ReadFile(migration.Path + version + migratePostfixUp)
		if errReadFile != nil {
			return countCompleted, errReadFile
		}

		dbTx, errBegin := dbc.Begin()
		if errBegin != nil {
			log.Println("migration fail: " + version)

			return countCompleted, errBegin
		}

		query := string(content)
		if strings.TrimSpace(query) != "" {
			// Executing all requests from the current file.
			if errExec := dbTx.Exec(query); errExec != nil {
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

		// If the file with the ending down.sql does not exist, then indicate that
		// this migration is not rolling back.
		if _, err := os.Stat(migration.Path + version + migratePostfixDown); os.IsNotExist(err) {
			post.RollFlag = false
		}

		if err := mStorage.Up(post); err != nil {
			log.Println("migration fail: " + version)

			if errRollback := dbTx.Rollback(); errRollback != nil {
				return countCompleted, errRollback
			}

			return countCompleted, fmt.Errorf("storage up: %w", err)
		}

		if errCommit := dbTx.Commit(); errCommit != nil {
			log.Println("migration fail: " + version)

			return countCompleted, errCommit
		}

		log.Println("migration success: " + version)
		countCompleted++
	}

	return countCompleted, nil
}
