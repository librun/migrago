package action

import (
	"log"
	"strings"
	"io/ioutil"
	"github.com/ne-ray/migrago/internal"
	i "github.com/ne-ray/migrago/internal/initialize"
)

//MakeDown do revert migrate
func MakeDown(project *internal.Project, dbName string, rollbackCount int) error {
	projectMigration, err := project.GetProjectMigration(dbName)
	if err != nil {
		return err
	}
	db, errDB := i.InitDB(projectMigration.Database.Type, projectMigration.Database.Dsn, projectMigration.Database.Schema)
	if errDB != nil {
		return errDB
	}
	defer db.Close()

	migrations, err := internal.DbBolt.GetLast(project.Name, dbName, rollbackCount)
	if err != nil {
		return err
	}


	for _, migrate := range migrations {
		if migrate.RollFlag {
			downFile := projectMigration.Path + migrate.FileName + internal.MigratePostfixDown
			content, err := ioutil.ReadFile(downFile)
			if err != nil {
				return err
			}

			query := string(content)
			if strings.TrimSpace(query) != "" {
				//выполнение всех запросов из текущего файла
				if errExec := db.Exec(query); errExec != nil {
					return errExec
				}
			}
		}

		if err := internal.DbBolt.Delete(project.Name, dbName, migrate.MigrationID); err != nil {
			return err
		}

		log.Println("migration: " + migrate.FileName + " complete")
	}

	return nil
}
