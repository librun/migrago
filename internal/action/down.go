package action

import (
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/librun/migrago/internal/config"
	"github.com/librun/migrago/internal/database"
	"github.com/librun/migrago/internal/storage"
)

// MakeDown do revert migrate.
func MakeDown(mStorage storage.Storage, cfgPath, projectName, dbName string, rollbackCount int, skipNoRollback bool) error {
	cfg, err := config.NewConfig(cfgPath, []string{projectName}, []string{dbName})
	if err != nil {
		return err
	}

	project, err := cfg.GetProject(projectName)
	if err != nil {
		return err
	}

	if _, err := project.GetDB(dbName); err != nil {
		return err
	}

	projectMigration, err := project.GetProjectMigration(dbName)
	if err != nil {
		return err
	}

	dbc, errDB := database.NewDB(projectMigration.Database)
	if errDB != nil {
		return errDB
	}
	defer dbc.Close()

	migrations, err := mStorage.GetLast(project.Name, dbName, skipNoRollback, &rollbackCount)
	if err != nil {
		return err
	}

	if len(migrations) < rollbackCount {
		return errors.New("Have " + strconv.Itoa(len(migrations)) + " of " + strconv.Itoa(rollbackCount) + " migration")
	}

	for _, migrate := range migrations {
		if migrate.RollFlag {
			downFile := projectMigration.Path + migrate.Version + migratePostfixDown

			content, err := ioutil.ReadFile(downFile)
			if err != nil {
				return err
			}

			query := string(content)
			if strings.TrimSpace(query) != "" {
				// выполнение всех запросов из текущего файла.
				if errExec := dbc.Exec(query); errExec != nil {
					return errExec
				}
			}
		}

		migrate := migrate
		if err := mStorage.Delete(&migrate); err != nil {
			return err
		}

		if migrate.RollFlag {
			log.Println("migration: " + migrate.Version + " roolback completed")
		} else {
			log.Println("migration: " + migrate.Version + " (not roolback) deleted")
		}
	}

	return nil
}
