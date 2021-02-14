package action

import (
	"fmt"
	"log"

	"github.com/librun/migrago/internal/config"
	"github.com/librun/migrago/internal/storage"
)

// MakeList shows success applied migrations.
func MakeList(mStorage storage.Storage, cfgPath, projectName, dbName string, rollbackCount *int, skipNoRollback bool) error {
	cfg, err := config.NewConfig(cfgPath, []string{projectName}, []string{dbName})
	if err != nil {
		return fmt.Errorf("config open: %w", err)
	}

	project, err := cfg.GetProject(projectName)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}

	if _, err := project.GetDB(dbName); err != nil {
		return fmt.Errorf("get project db: %w", err)
	}

	migrations, err := mStorage.GetLast(project.Name, dbName, skipNoRollback, rollbackCount)
	if err != nil {
		return fmt.Errorf("get last migration: %w", err)
	}

	log.Println("Migrations list:")

	for _, migrate := range migrations {
		t := "migration: " + migrate.Version
		if !migrate.RollFlag {
			t += " (no rollback)"
		}

		log.Println(t)
	}

	return nil
}
