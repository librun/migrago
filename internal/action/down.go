package action

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/librun/migrago/internal/config"
	"github.com/librun/migrago/internal/database"
	"github.com/librun/migrago/internal/storage"
)

// MakeDown reverts migrations.
func MakeDown(mStorage storage.Storage, cfgPath, projectName, dbName string, rollbackCount int, skipNoRollback bool) error {
	cfg, err := config.NewConfig(cfgPath, []string{projectName}, []string{dbName})
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	project, err := cfg.GetProject(projectName)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}

	if _, err := project.GetDB(dbName); err != nil {
		return fmt.Errorf("get project db: %w", err)
	}

	projectMigration, err := project.GetProjectMigration(dbName)
	if err != nil {
		return fmt.Errorf("get current migration: %w", err)
	}

	dbc, err := database.NewDB(projectMigration.Database)
	if err != nil {
		return fmt.Errorf("conntect to db: %w", err)
	}
	defer dbc.Close()

	migrations, err := mStorage.GetLast(project.Name, dbName, skipNoRollback, &rollbackCount)
	if err != nil {
		return fmt.Errorf("get last migration: %w", err)
	}

	if len(migrations) < rollbackCount {
		return errors.New("Have " + strconv.Itoa(len(migrations)) + " of " + strconv.Itoa(rollbackCount) + " migration")
	}

	for _, migrate := range migrations {
		if migrate.RollFlag {
			downFile := projectMigration.Path + migrate.Version + migratePostfixDown

			content, err := ioutil.ReadFile(downFile)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			query := string(content)
			if strings.TrimSpace(query) != "" {
				// Executing all requests from the current file.
				if errExec := dbc.Exec(query); errExec != nil {
					return errExec
				}
			}
		}

		migrate := migrate
		if err := mStorage.Delete(&migrate); err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		if migrate.RollFlag {
			log.Println("migration: " + migrate.Version + " roolback completed")
		} else {
			log.Println("migration: " + migrate.Version + " (not roolback) deleted")
		}
	}

	return nil
}
