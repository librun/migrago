package action

import (
	"log"

	"github.com/librun/migrago/internal/storage"
)

// MakeList show success migrations.
func MakeList(mStorage storage.Storage, cfgPath, projectName, dbName string, rollbackCount *int, skipNoRollback bool) error {
	config, err := initConfig(cfgPath, []string{projectName}, []string{dbName})
	if err != nil {
		return err
	}

	project, err := config.getProject(projectName)
	if err != nil {
		return err
	}

	if _, err := project.getDB(dbName); err != nil {
		return err
	}

	migrations, err := mStorage.GetLast(project.name, dbName, skipNoRollback, rollbackCount)
	if err != nil {
		return err
	}

	log.Println("List migrations:")
	for _, migrate := range migrations {
		t := "migration: " + migrate.Version
		if !migrate.RollFlag {
			t += " (no rollback)"
		}

		log.Println(t)
	}

	return nil
}
