package action

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/librun/migrago/internal/config"
	"gopkg.in/yaml.v2"
)

// Migrations create modes.
const (
	CreateModeUp   = "up"
	CreateModeDown = "down"
	CreateModeBoth = "both"
)

// DefaultMigrationNameFormat contains template for create migration name.
const DefaultMigrationNameFormat = "%d%02d%02d_%02d%02d%02d_%s_%s.sql"

// MakeCreate creates new migration file.
func MakeCreate(cfgPath, name, mode, project, db string) error {
	cfg := config.YAMLConfig{}

	configFile, err := os.Open(cfgPath)
	if err != nil {
		return fmt.Errorf("config open: %w", err)
	}

	defer configFile.Close()

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		return fmt.Errorf("config read: %w", err)
	}

	if err := yaml.Unmarshal(configByte, &cfg); err != nil {
		return fmt.Errorf("config format: %w", err)
	}

	// Get migration directory.
	directory := ""

	for projectName, value := range cfg.Projects {
		if projectName != project {
			continue
		}

		for _, migrations := range value.Migrations {
			var ok bool

			if directory, ok = migrations[db]; !ok {
				continue
			} else {
				break
			}
		}
	}

	if directory == "" {
		return errors.New("invalid project or db")
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err = os.MkdirAll(directory, 0777); err != nil {
			return fmt.Errorf("create directory %s error: %w", directory, err)
		}
	}

	// Create file.
	if mode == "both" {
		for _, modeItem := range []string{"up", "down"} {
			if err := createFile(name, modeItem, directory); err != nil {
				return err
			}
		}

		return nil
	}

	return createFile(name, mode, directory)
}

func createFile(name, mode, directory string) error {
	now := time.Now()
	filename := fmt.Sprintf(
		DefaultMigrationNameFormat,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		name,
		mode,
	)

	_, err := os.Create(fmt.Sprintf("%s/%s", directory, filename))
	if err == nil {
		log.Printf("migration: %s created", filename)
	}

	return err
}
